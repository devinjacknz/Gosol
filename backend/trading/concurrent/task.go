package concurrent

import (
	"context"
	"fmt"
	"time"

	"solmeme-trader/models"
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
}

// BaseTask provides common task functionality
type BaseTask struct {
	name      string
	taskType  TaskType
	startTime time.Time
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
