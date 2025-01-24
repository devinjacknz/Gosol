package concurrent

import (
	"context"
	"fmt"
	"time"

	"solmeme-trader/dex"
	"solmeme-trader/models"
	"solmeme-trader/monitoring"
	"solmeme-trader/repository"
)

// MarketDataProcessor handles market data processing
type MarketDataProcessor struct {
	repo          repository.Repository
	dexClient     *dex.DexClient
	marketService *dex.MarketDataService
	monitor       *monitoring.Monitor
}

// NewMarketDataProcessor creates a new market data processor
func NewMarketDataProcessor(repo repository.Repository, dexClient *dex.DexClient, marketService *dex.MarketDataService, monitor *monitoring.Monitor) *MarketDataProcessor {
	return &MarketDataProcessor{
		repo:          repo,
		dexClient:     dexClient,
		marketService: marketService,
		monitor:       monitor,
	}
}

// ProcessMarketData processes market data for a token
func (p *MarketDataProcessor) ProcessMarketData(ctx context.Context, data *models.MarketData) error {
	start := time.Now()

	// Get market info from DEX
	marketInfo, err := p.dexClient.GetMarketInfo(ctx, data.TokenAddress)
	if err != nil {
		p.monitor.RecordEvent(ctx, monitoring.Event{
			Type:      monitoring.MetricMarketData,
			Severity:  monitoring.SeverityError,
			Message:   "Failed to get market info",
			Details:   map[string]interface{}{"error": err.Error()},
			Token:     &data.TokenAddress,
			Timestamp: time.Now(),
		})
		return fmt.Errorf("failed to get market info: %w", err)
	}

	// Get liquidity info
	liquidity, err := p.dexClient.GetLiquidity(ctx, data.TokenAddress)
	if err != nil {
		p.monitor.RecordEvent(ctx, monitoring.Event{
			Type:      monitoring.MetricMarketData,
			Severity:  monitoring.SeverityError,
			Message:   "Failed to get liquidity info",
			Details:   map[string]interface{}{"error": err.Error()},
			Token:     &data.TokenAddress,
			Timestamp: time.Now(),
		})
		return fmt.Errorf("failed to get liquidity info: %w", err)
	}

	// Get recent trades
	trades, err := p.dexClient.GetRecentTrades(ctx, marketInfo.Address, 1000)
	if err != nil {
		p.monitor.RecordEvent(ctx, monitoring.Event{
			Type:      monitoring.MetricMarketData,
			Severity:  monitoring.SeverityError,
			Message:   "Failed to get recent trades",
			Details:   map[string]interface{}{"error": err.Error()},
			Token:     &data.TokenAddress,
			Timestamp: time.Now(),
		})
		return fmt.Errorf("failed to get recent trades: %w", err)
	}

	// Calculate 24h volume
	var volume24h float64
	oneDayAgo := time.Now().Add(-24 * time.Hour)
	for _, trade := range trades {
		if trade.Timestamp.After(oneDayAgo) {
			volume24h += trade.Price * trade.Size
		}
	}

	// Update market data
	data.Price = marketInfo.LastPrice
	data.Volume24h = volume24h
	data.MarketCap = marketInfo.LastPrice * liquidity.TotalSupply
	data.Liquidity = liquidity.TVL
	data.PriceImpact = 0 // Calculate from orderbook if needed
	data.Timestamp = time.Now()

	// Save to database
	if err := p.repo.SaveMarketData(ctx, data); err != nil {
		p.monitor.RecordEvent(ctx, monitoring.Event{
			Type:      monitoring.MetricMarketData,
			Severity:  monitoring.SeverityError,
			Message:   "Failed to save market data",
			Details:   map[string]interface{}{"error": err.Error()},
			Token:     &data.TokenAddress,
			Timestamp: time.Now(),
		})
		return fmt.Errorf("failed to save market data: %w", err)
	}

	// Record successful processing
	p.monitor.RecordMarketData(ctx, data, time.Since(start))

	return nil
}

// EnhancedMarketDataTask represents an enhanced market data processing task
type EnhancedMarketDataTask struct {
	processor    *MarketDataProcessor
	TokenAddress string
	MarketData   *models.MarketData
}

// NewEnhancedMarketDataTask creates a new enhanced market data task
func NewEnhancedMarketDataTask(processor *MarketDataProcessor, tokenAddress string, data *models.MarketData) *EnhancedMarketDataTask {
	return &EnhancedMarketDataTask{
		processor:    processor,
		TokenAddress: tokenAddress,
		MarketData:   data,
	}
}

// Execute executes the market data task
func (t *EnhancedMarketDataTask) Execute(ctx context.Context) error {
	return t.processor.ProcessMarketData(ctx, t.MarketData)
}

// Name returns the task name
func (t *EnhancedMarketDataTask) Name() string {
	return fmt.Sprintf("MarketData-%s", t.TokenAddress)
}
