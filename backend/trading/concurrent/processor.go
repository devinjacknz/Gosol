package concurrent

import (
	"context"
	"time"

	"solmeme-trader/dex"
	"solmeme-trader/models"
	"solmeme-trader/monitoring"
	"solmeme-trader/repository"
	"solmeme-trader/trading/risk"
)

// ProcessorConfig defines processor configuration
type ProcessorConfig struct {
	NumWorkers     int           `json:"num_workers"`
	BatchSize      int           `json:"batch_size"`
	ProcessTimeout time.Duration `json:"process_timeout"`
}

// Processor handles concurrent task processing
type Processor struct {
	config        ProcessorConfig
	pool          *Pool
	repo          repository.Repository
	dexClient     *dex.DexClient
	marketService *dex.MarketDataService
	riskManager   *risk.RiskManager
	monitor       *monitoring.Monitor
	marketProc    *MarketDataProcessor
	tradeProc     *TradeProcessor
}

// NewProcessor creates a new processor
func NewProcessor(config ProcessorConfig, repo repository.Repository, dexClient *dex.DexClient, marketService *dex.MarketDataService, riskManager *risk.RiskManager, monitor *monitoring.Monitor) *Processor {
	p := &Processor{
		config:        config,
		repo:          repo,
		dexClient:     dexClient,
		marketService: marketService,
		riskManager:   riskManager,
		monitor:       monitor,
	}

	// Initialize sub-processors
	p.marketProc = NewMarketDataProcessor(repo, dexClient, marketService, monitor)
	p.tradeProc = NewTradeProcessor(repo, dexClient, riskManager, monitor)

	// Initialize worker pool
	p.pool = NewPool(config.NumWorkers, monitor)

	return p
}

// Start starts the processor
func (p *Processor) Start(ctx context.Context) {
	p.pool.Start(ctx)
}

// ProcessMarketData processes market data
func (p *Processor) ProcessMarketData(ctx context.Context, data *models.MarketData) error {
	task := NewMarketDataTask(p.marketProc, data.TokenAddress, data)
	p.pool.Submit(task)
	return nil
}

// ProcessTrade processes a trade
func (p *Processor) ProcessTrade(ctx context.Context, trade *models.Trade) error {
	task := NewTradeTask(p.tradeProc, trade)
	p.pool.Submit(task)
	return nil
}

// ProcessBatch processes a batch of tasks
func (p *Processor) ProcessBatch(ctx context.Context, tasks []Task) error {
	for _, task := range tasks {
		p.pool.Submit(task)
	}
	return nil
}

// Wait waits for all tasks to complete
func (p *Processor) Wait() {
	p.pool.Wait()
}

// GetStats gets processor statistics
func (p *Processor) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"num_workers": p.config.NumWorkers,
		"batch_size":  p.config.BatchSize,
		"timeout":     p.config.ProcessTimeout.String(),
	}
}
