package concurrent

import (
	"context"
	"fmt"
	"time"

	"github.com/leonzhao/trading-system/backend/models"
)

// TaskType represents the type of task
type TaskType string

const (
	TaskTypeMarketData TaskType = "market_data"
	TaskTypeTrade      TaskType = "trade"
)

// Task represents a processable task
type Task interface {
	Execute(ctx context.Context) error
	GetName() string
	GetType() TaskType
	GetPriority() int
	SetPriority(priority int)
	GetTimeout() time.Duration
	SetTimeout(timeout time.Duration)
	GetRetryCount() int
	SetRetryConfig(maxRetries int, delay time.Duration)
	ShouldRetry() bool
}

// BaseTask provides common task functionality
type BaseTask struct {
	name       string
	taskType   TaskType
	startTime  time.Time
	priority   int
	timeout    time.Duration
	maxRetries int
	retryCount int
	retryDelay time.Duration
}

// GetPriority returns the task priority
func (t *BaseTask) GetPriority() int {
	return t.priority
}

// SetPriority sets the task priority
func (t *BaseTask) SetPriority(priority int) {
	t.priority = priority
}

// GetTimeout returns the task timeout
func (t *BaseTask) GetTimeout() time.Duration {
	return t.timeout
}

// SetTimeout sets the task timeout
func (t *BaseTask) SetTimeout(timeout time.Duration) {
	t.timeout = timeout
}

// GetRetryCount returns the current retry count
func (t *BaseTask) GetRetryCount() int {
	return t.retryCount
}

// SetRetryConfig sets retry configuration
func (t *BaseTask) SetRetryConfig(maxRetries int, delay time.Duration) {
	t.maxRetries = maxRetries
	t.retryDelay = delay
}

// ShouldRetry checks if the task should be retried
func (t *BaseTask) ShouldRetry() bool {
	return t.retryCount < t.maxRetries
}

// GetName returns the task name
func (t *BaseTask) GetName() string {
	return t.name
}

// GetType returns the task type
func (t *BaseTask) GetType() TaskType {
	return t.taskType
}

// MarketDataTask represents a market data processing task
type MarketDataTask struct {
	BaseTask
	processor    *MarketDataProcessor
	tokenAddress string
	data         *models.MarketData
}

// NewMarketDataTask creates a new market data task
func NewMarketDataTask(processor *MarketDataProcessor, tokenAddress string, data *models.MarketData) *MarketDataTask {
	return &MarketDataTask{
		BaseTask: BaseTask{
			name:      fmt.Sprintf("market_data_%s", tokenAddress),
			taskType:  TaskTypeMarketData,
			startTime: time.Now(),
		},
		processor:    processor,
		tokenAddress: tokenAddress,
		data:         data,
	}
}

// Execute processes the market data task
func (t *MarketDataTask) Execute(ctx context.Context) error {
	return t.processor.ProcessMarketData(ctx, t.data)
}

// TradeTask represents a trade processing task
type TradeTask struct {
	BaseTask
	processor *TradeProcessor
	trade     *models.Trade
}

// NewTradeTask creates a new trade task
func NewTradeTask(processor *TradeProcessor, trade *models.Trade) *TradeTask {
	return &TradeTask{
		BaseTask: BaseTask{
			name:      fmt.Sprintf("trade_%s", trade.ID),
			taskType:  TaskTypeTrade,
			startTime: time.Now(),
		},
		processor: processor,
		trade:     trade,
	}
}

// Execute processes the trade task
func (t *TradeTask) Execute(ctx context.Context) error {
	return t.processor.ProcessTrade(ctx, t.trade)
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	Task      Task
	Error     error
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// NewTaskResult creates a new task result
func NewTaskResult(task Task, err error, startTime time.Time) *TaskResult {
	endTime := time.Now()
	return &TaskResult{
		Task:      task,
		Error:     err,
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
	}
}
