package trading

import (
	"context"
	"fmt"
	"time"

	"github.com/leonzhao/trading-system/backend/dex"
	"github.com/leonzhao/trading-system/backend/models"
	"github.com/leonzhao/trading-system/backend/monitoring"
	"github.com/leonzhao/trading-system/backend/repository"
	"github.com/leonzhao/trading-system/backend/trading/concurrent"
	"github.com/leonzhao/trading-system/backend/trading/risk"
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

	// Create market data service
	marketService := dex.NewMarketDataService(
		dex.NewRaydiumClient("https://api.raydium.io"),
		dex.NewJupiterClient("https://api.jup.ag"),
	)
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
		TokenAddress:  signal.Signal.Symbol, // Using Symbol as TokenAddress
		Type:         models.OrderTypeMarket,
		Side:         mapSignalTypeToTradeSide(signal.Signal.SignalType),
		Amount:       signal.Signal.Size,
		Price:        signal.Signal.Price,
		Status:       models.TradeStatusPending,
		Timestamp:    time.Now(),
		UpdateTime:   time.Now(),
	}

	// Process trade with high priority
	if err := e.processor.ProcessTrade(ctx, trade, concurrent.PriorityHigh); err != nil {
		e.monitor.RecordEvent(ctx, monitoring.Event{
			Type:      monitoring.MetricTrading,
			Severity:  monitoring.SeverityError,
			Message:   "Failed to process trade",
			Details: map[string]interface{}{
				"error":        err.Error(),
				"tokenAddress": signal.Signal.Symbol,
			},
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
			"duration":     time.Since(start).String(),
			"trade":       trade,
			"tokenAddress": signal.Signal.Symbol,
		},
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
		if err := e.processor.ProcessMarketData(ctx, marketData, concurrent.PriorityMedium); err != nil {
			e.monitor.RecordEvent(ctx, monitoring.Event{
				Type:      monitoring.MetricTrading,
				Severity:  monitoring.SeverityError,
				Message:   "Failed to process market data",
				Details: map[string]interface{}{
					"error":        err.Error(),
					"tokenAddress": pos.TokenAddress,
				},
				Timestamp: time.Now(),
			})
		}
	}

	return nil
}

// GetStats gets executor statistics
func (e *Executor) GetStats() concurrent.ProcessorStats {
	return e.processor.GetStats()
}

// GetHealthStatus returns the executor health status
func (e *Executor) GetHealthStatus() string {
	return e.processor.GetHealthStatus()
}

// mapSignalTypeToTradeSide maps signal type to trade side
func mapSignalTypeToTradeSide(signalType models.TradeSignalType) models.TradeSide {
	switch signalType {
	case models.SignalTypeBuy, models.SignalTypeOversold:
		return models.TradeSideBuy
	case models.SignalTypeSell, models.SignalTypeOverbought:
		return models.TradeSideSell
	default:
		return models.TradeSideBuy // Default to buy for unknown signal types
	}
}
