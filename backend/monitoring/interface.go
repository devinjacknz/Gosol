package monitoring

import (
	"context"
	"time"
)

// Monitor defines the monitoring interface
type IMonitor interface {
	// RecordEvent records a monitoring event
	RecordEvent(ctx context.Context, event Event)

	// RecordMetric records a metric value
	RecordMetric(ctx context.Context, name string, value float64)

	// RecordMarketData records market data processing
	RecordMarketData(ctx context.Context, data interface{}, duration time.Duration)

	// GetMetrics gets metrics with context
	GetMetrics(ctx context.Context) map[string]float64

	// CheckHealth checks system health
	CheckHealth(ctx context.Context) (bool, map[string]bool)

	// SetHealth sets component health status
	SetHealth(component string, healthy bool)

	// Reset resets all metrics
	Reset()
}

// Ensure Monitor implements IMonitor
var _ IMonitor = (*Monitor)(nil)

// Ensure MockMonitor implements IMonitor
var _ IMonitor = (*MockMonitor)(nil)
