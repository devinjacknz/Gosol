package concurrent

import (
	"context"
	"fmt"
	"time"

	"github.com/leonzhao/trading-system/backend/dex"
	"github.com/leonzhao/trading-system/backend/models"
	"github.com/leonzhao/trading-system/backend/monitoring"
	"github.com/leonzhao/trading-system/backend/repository"
	"github.com/leonzhao/trading-system/backend/trading/risk"
)

// TradeProcessor handles trade processing
type TradeProcessor struct {
	repo        repository.Repository
	dexClient   *dex.DexClient
	riskManager *risk.RiskManager
	monitor     *monitoring.Monitor
}

// NewTradeProcessor creates a new trade processor
func NewTradeProcessor(repo repository.Repository, dexClient *dex.DexClient, riskManager *risk.RiskManager, monitor *monitoring.Monitor) *TradeProcessor {
	return &TradeProcessor{
		repo:        repo,
		dexClient:   dexClient,
		riskManager: riskManager,
		monitor:     monitor,
	}
}

// ProcessTrade processes a trade
func (p *TradeProcessor) ProcessTrade(ctx context.Context, trade *models.Trade) error {
	start := time.Now()

	// Get market data
	marketData, err := p.repo.GetLatestMarketData(ctx, trade.TokenAddress)
	if err != nil {
		p.monitor.RecordEvent(ctx, monitoring.Event{
			Type:      monitoring.MetricTrading,
			Severity:  monitoring.SeverityError,
			Message:   "Failed to get market data",
			Details:   map[string]interface{}{
				"error": err.Error(),
				"tokenAddress": trade.TokenAddress,
			},
			Timestamp: time.Now(),
		})
		return fmt.Errorf("failed to get market data: %w", err)
	}

	// Check if we can open position
	if err := p.riskManager.CanOpenPosition(ctx, trade.TokenAddress, trade.Amount, marketData); err != nil {
		p.monitor.RecordEvent(ctx, monitoring.Event{
			Type:      monitoring.MetricTrading,
			Severity:  monitoring.SeverityError,
			Message:   "Cannot open position",
			Details:   map[string]interface{}{
				"error": err.Error(),
				"tokenAddress": trade.TokenAddress,
			},
			Timestamp: time.Now(),
		})
		return fmt.Errorf("cannot open position: %w", err)
	}

	// Create position
	position := &models.Position{
		ID:           trade.ID,
		TokenAddress: trade.TokenAddress,
		Side:         string(trade.Side),
		EntryPrice:   trade.Price,
		CurrentPrice: marketData.ClosePrice,
		Size:         trade.Amount,
		Leverage:     1.0, // Default to no leverage
		OpenTime:     time.Now(),
		LastUpdated:  time.Now(),
	}

	// Open position
	if err := p.riskManager.OpenPosition(ctx, risk.FromModelPosition(position), marketData); err != nil {
		p.monitor.RecordEvent(ctx, monitoring.Event{
			Type:      monitoring.MetricTrading,
			Severity:  monitoring.SeverityError,
			Message:   "Failed to open position",
			Details:   map[string]interface{}{
				"error": err.Error(),
				"tokenAddress": trade.TokenAddress,
			},
			Timestamp: time.Now(),
		})
		return fmt.Errorf("failed to open position: %w", err)
	}

	// Save position
	if err := p.repo.SavePosition(ctx, position); err != nil {
		p.monitor.RecordEvent(ctx, monitoring.Event{
			Type:      monitoring.MetricTrading,
			Severity:  monitoring.SeverityError,
			Message:   "Failed to save position",
			Details:   map[string]interface{}{
				"error": err.Error(),
				"tokenAddress": trade.TokenAddress,
			},
			Timestamp: time.Now(),
		})
		return fmt.Errorf("failed to save position: %w", err)
	}

	// Update trade status
	trade.Status = models.TradeStatusCompleted
	trade.UpdateTime = time.Now()
	if err := p.repo.UpdateTrade(ctx, trade); err != nil {
		p.monitor.RecordEvent(ctx, monitoring.Event{
			Type:      monitoring.MetricTrading,
			Severity:  monitoring.SeverityError,
			Message:   "Failed to update trade",
			Details:   map[string]interface{}{
				"error": err.Error(),
				"tokenAddress": trade.TokenAddress,
			},
			Timestamp: time.Now(),
		})
		return fmt.Errorf("failed to update trade: %w", err)
	}

	// Record successful processing
	p.monitor.RecordEvent(ctx, monitoring.Event{
		Type:      monitoring.MetricTrading,
		Severity:  monitoring.SeverityInfo,
		Message:   "Trade processed",
		Details: map[string]interface{}{
			"duration": time.Since(start).String(),
			"trade":    trade,
			"position": position,
			"tokenAddress": trade.TokenAddress,
		},
		Timestamp: time.Now(),
	})

	return nil
}
