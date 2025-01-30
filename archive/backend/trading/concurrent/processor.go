package concurrent

import (
	"context"
	"sync"
	"time"

	"github.com/leonzhao/trading-system/backend/dex"
	"github.com/leonzhao/trading-system/backend/models"
	"github.com/leonzhao/trading-system/backend/monitoring"
	"github.com/leonzhao/trading-system/backend/repository"
	"github.com/leonzhao/trading-system/backend/trading/risk"
)

// Priority levels for tasks
const (
	PriorityHigh   = 3
	PriorityMedium = 2
	PriorityLow    = 1
)

// ProcessorConfig defines processor configuration
type ProcessorConfig struct {
	NumWorkers     int           `json:"num_workers"`
	BatchSize      int           `json:"batch_size"`
	ProcessTimeout time.Duration `json:"process_timeout"`
	MaxRetries     int           `json:"max_retries"`
	RetryDelay     time.Duration `json:"retry_delay"`
}

// ProcessorStats represents processor statistics
type ProcessorStats struct {
	TasksProcessed   int64
	TasksFailed      int64
	TasksRetried     int64
	AverageLatency   time.Duration
	QueueSize        int
	ActiveWorkers    int
	LastUpdated      time.Time
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
	stats         ProcessorStats
	statsMutex    sync.RWMutex
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

// ProcessMarketData processes market data with priority
func (p *Processor) ProcessMarketData(ctx context.Context, data *models.MarketData, priority int) error {
	task := NewMarketDataTask(p.marketProc, data.TokenAddress, data)
	task.SetPriority(priority)
	task.SetTimeout(p.config.ProcessTimeout)
	task.SetRetryConfig(p.config.MaxRetries, p.config.RetryDelay)
	
	start := time.Now()
	err := p.pool.Submit(task)
	if err == nil {
		p.recordTaskLatency(time.Since(start))
		p.incrementTasksProcessed()
	} else {
		p.incrementTasksFailed()
	}
	return err
}

// ProcessTrade processes a trade with priority
func (p *Processor) ProcessTrade(ctx context.Context, trade *models.Trade, priority int) error {
	task := NewTradeTask(p.tradeProc, trade)
	task.SetPriority(priority)
	task.SetTimeout(p.config.ProcessTimeout)
	task.SetRetryConfig(p.config.MaxRetries, p.config.RetryDelay)
	
	start := time.Now()
	err := p.pool.Submit(task)
	if err == nil {
		p.recordTaskLatency(time.Since(start))
		p.incrementTasksProcessed()
	} else {
		p.incrementTasksFailed()
	}
	return err
}

func (p *Processor) recordTaskLatency(duration time.Duration) {
	p.statsMutex.Lock()
	defer p.statsMutex.Unlock()
	
	// Update average latency using exponential moving average
	alpha := 0.1 // Smoothing factor
	p.stats.AverageLatency = time.Duration(float64(p.stats.AverageLatency)*(1-alpha) + float64(duration)*alpha)
	p.stats.LastUpdated = time.Now()
}

func (p *Processor) incrementTasksProcessed() {
	p.statsMutex.Lock()
	defer p.statsMutex.Unlock()
	p.stats.TasksProcessed++
	p.stats.LastUpdated = time.Now()
}

func (p *Processor) incrementTasksFailed() {
	p.statsMutex.Lock()
	defer p.statsMutex.Unlock()
	p.stats.TasksFailed++
	p.stats.LastUpdated = time.Now()
}

func (p *Processor) incrementTasksRetried() {
	p.statsMutex.Lock()
	defer p.statsMutex.Unlock()
	p.stats.TasksRetried++
	p.stats.LastUpdated = time.Now()
}

// ProcessBatch processes a batch of tasks with priority
func (p *Processor) ProcessBatch(ctx context.Context, tasks []Task, priority int) error {
	for _, task := range tasks {
		task.SetPriority(priority)
		task.SetTimeout(p.config.ProcessTimeout)
		task.SetRetryConfig(p.config.MaxRetries, p.config.RetryDelay)
		
		if err := p.pool.Submit(task); err != nil {
			return err
		}
	}
	return nil
}

// Wait waits for all tasks to complete
func (p *Processor) Wait() {
	p.pool.Wait()
}

// GetStats gets detailed processor statistics
func (p *Processor) GetStats() ProcessorStats {
	p.statsMutex.RLock()
	defer p.statsMutex.RUnlock()
	
	stats := p.stats
	stats.QueueSize = p.pool.QueueSize()
	stats.ActiveWorkers = p.pool.ActiveWorkers()
	return stats
}

// GetHealthStatus returns the health status of the processor
func (p *Processor) GetHealthStatus() string {
	stats := p.GetStats()
	if stats.TasksFailed > stats.TasksProcessed/10 { // More than 10% failure rate
		return "unhealthy"
	}
	if stats.QueueSize > p.config.BatchSize*2 { // Queue backing up
		return "degraded"
	}
	return "healthy"
}
