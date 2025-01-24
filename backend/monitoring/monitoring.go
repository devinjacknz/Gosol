package monitoring

import (
	"context"
	"sync"
	"time"
)

// Metric types
const (
	MetricTrading        = "trading"
	MetricMarketData     = "market_data"
	MetricProcessing     = "processing"
	MetricTaskCompletion = "task_completion"
	MetricSystem         = "system"
)

// Severity levels
const (
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityError    = "error"
	SeverityCritical = "critical"
)

// Event represents a monitoring event
type Event struct {
	Type      string                 `json:"type"`
	Severity  string                 `json:"severity"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Token     *string                `json:"token,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// Monitor handles system monitoring
type Monitor struct {
	metrics     map[string]float64
	healthState map[string]bool
	mu          sync.RWMutex
}

// NewMonitor creates a new monitor
func NewMonitor() *Monitor {
	return &Monitor{
		metrics:     make(map[string]float64),
		healthState: make(map[string]bool),
	}
}

// RecordEvent records a monitoring event
func (m *Monitor) RecordEvent(ctx context.Context, event Event) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update metrics based on event type
	switch event.Type {
	case MetricTrading:
		m.metrics["trade_count"]++
		if volume, ok := event.Details["volume"].(float64); ok {
			m.metrics["trade_volume"] += volume
		}
	case MetricMarketData:
		m.metrics["market_data_count"]++
	case MetricProcessing:
		if duration, ok := event.Details["duration"].(string); ok {
			if d, err := time.ParseDuration(duration); err == nil {
				m.metrics["processing_time"] += d.Seconds()
			}
		}
	}
}

// RecordMetric records a metric value with name and value
func (m *Monitor) RecordMetric(ctx context.Context, name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics[name] = value
}

// RecordMarketData records market data processing
func (m *Monitor) RecordMarketData(ctx context.Context, data interface{}, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics["market_data_processing_time"] += duration.Seconds()
	m.metrics["market_data_count"]++
}

// GetMetrics gets metrics with context
func (m *Monitor) GetMetrics(ctx context.Context) map[string]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Copy metrics to prevent map mutation while reading
	metrics := make(map[string]float64, len(m.metrics))
	for k, v := range m.metrics {
		metrics[k] = v
	}

	return metrics
}

// CheckHealth checks system health
func (m *Monitor) CheckHealth(ctx context.Context) (bool, map[string]bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Copy health state to prevent map mutation while reading
	health := make(map[string]bool, len(m.healthState))
	for k, v := range m.healthState {
		health[k] = v
	}

	// Overall health is true only if all components are healthy
	allHealthy := true
	for _, healthy := range health {
		if !healthy {
			allHealthy = false
			break
		}
	}

	return allHealthy, health
}

// SetHealth sets component health status
func (m *Monitor) SetHealth(component string, healthy bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.healthState[component] = healthy
}

// Reset resets all metrics
func (m *Monitor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics = make(map[string]float64)
}
