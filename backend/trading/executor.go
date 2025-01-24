package trading

import (
	"context"
	"fmt"
	"time"

	"solmeme-trader/dex"
	"solmeme-trader/models"
	"solmeme-trader/monitoring"
	"solmeme-trader/repository"
	"solmeme-trader/trading/concurrent"
	"solmeme-trader/trading/risk"
)

// Executor handles trade execution
type Executor struct {
	repo      repository.Repository
	dexClient *dex.DexClient
	processor *concurrent.Processor
	monitor   *monitoring.Monitor
}

// NewExecutor creates a new trade executor
func NewExecutor(repo repository.Repository, dexClient *dex.DexClient, riskManager *risk.RiskManager, monitor *monitoring.Monitor) *Executor {
	config := concurrent.ProcessorConfig{
		NumWorkers:     10,
		BatchSize:      100,
		ProcessTimeout: 30 * time.Second,
	}

	marketService := dex.NewMarketDataService(dexClient, 5*time.Minute)
	processor := concurrent.NewProcessor(config, repo, dexClient, marketService, riskManager, monitor)

	return &Executor{
		repo:      repo,
		dexClient: dexClient,
		processor: processor,
		monitor:   monitor,
	}
}

// Start starts the executor
func (e *Executor) Start(ctx context.Context) {
	e.processor.Start(ctx)
}

// ExecuteSignal executes a trade signal
func (e *Executor) ExecuteSignal(ctx context.Context, signal *models.TradeSignalMessage) error {
	start := time.Now()

	// Create trade from signal
	trade := &models.Trade{
		TokenAddress:  signal.TokenAddress,
		Type:         models.OrderTypeMarket,
		Side:         signal.Action,
		Amount:       signal.Amount,
		Price:        signal.TargetPrice,
		Status:       models.TradeStatusPending,
		Timestamp:    time.Now(),
		UpdateTime:   time.Now(),
	}

	// Process trade
	if err := e.processor.ProcessTrade(ctx, trade); err != nil {
		e.monitor.RecordEvent(ctx, monitoring.Event{
			Type:      monitoring.MetricTrading,
			Severity:  monitoring.SeverityError,
			Message:   "Failed to process trade",
			Details:   map[string]interface{}{"error": err.Error()},
			Token:     &signal.TokenAddress,
			Timestamp: time.Now(),
		})
		return fmt.Errorf("failed to process trade: %w", err)
	}

	// Record successful execution
	e.monitor.RecordEvent(ctx, monitoring.Event{
		Type:      monitoring.MetricTrading,
		Severity:  monitoring.SeverityInfo,
		Message:   "Trade signal executed",
		Details: map[string]interface{}{
			"duration": time.Since(start).String(),
			"trade":    trade,
		},
		Token:     &signal.TokenAddress,
		Timestamp: time.Now(),
	})

	return nil
}

// UpdatePositions updates all open positions with current market data
func (e *Executor) UpdatePositions(ctx context.Context) error {
	// Get all open positions
	positions, err := e.repo.GetOpenPositions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get open positions: %w", err)
	}

	if len(positions) == 0 {
		return nil
	}

	// Get market data for all positions
	for _, pos := range positions {
		marketData := &models.MarketData{
			TokenAddress: pos.TokenAddress,
			Timestamp:   time.Now(),
		}

		// Process market data
		if err := e.processor.ProcessMarketData(ctx, marketData); err != nil {
			e.monitor.RecordEvent(ctx, monitoring.Event{
				Type:      monitoring.MetricTrading,
				Severity:  monitoring.SeverityError,
				Message:   "Failed to process market data",
				Details:   map[string]interface{}{"error": err.Error()},
				Token:     &pos.TokenAddress,
				Timestamp: time.Now(),
			})
		}
	}

	return nil
}

// GetStats gets executor statistics
func (e *Executor) GetStats() map[string]interface{} {
	return e.processor.GetStats()
}
