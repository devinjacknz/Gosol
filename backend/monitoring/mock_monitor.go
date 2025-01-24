package monitoring

import (
	"context"
	"sync"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockMonitor is a mock implementation of monitoring
type MockMonitor struct {
	mock.Mock
	metrics     map[string]float64
	healthState map[string]bool
	mu          sync.RWMutex
}

// NewMockMonitor creates a new mock monitor
func NewMockMonitor() *MockMonitor {
	return &MockMonitor{
		metrics:     make(map[string]float64),
		healthState: make(map[string]bool),
	}
}

// RecordEvent records a monitoring event
func (m *MockMonitor) RecordEvent(ctx context.Context, event Event) {
	m.Called(ctx, event)
	m.mu.Lock()
	defer m.mu.Unlock()

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

// RecordMetric records a metric value
func (m *MockMonitor) RecordMetric(ctx context.Context, name string, value float64) {
	m.Called(ctx, name, value)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics[name] = value
}

// RecordMarketData records market data processing
func (m *MockMonitor) RecordMarketData(ctx context.Context, data interface{}, duration time.Duration) {
	m.Called(ctx, data, duration)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics["market_data_processing_time"] += duration.Seconds()
	m.metrics["market_data_count"]++
}

// GetMetrics gets metrics
func (m *MockMonitor) GetMetrics(ctx context.Context) map[string]float64 {
	args := m.Called(ctx)
	if metrics, ok := args.Get(0).(map[string]float64); ok {
		return metrics
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	metrics := make(map[string]float64, len(m.metrics))
	for k, v := range m.metrics {
		metrics[k] = v
	}
	return metrics
}

// CheckHealth checks system health
func (m *MockMonitor) CheckHealth(ctx context.Context) (bool, map[string]bool) {
	args := m.Called(ctx)
	if health, ok := args.Get(1).(map[string]bool); ok {
		return args.Bool(0), health
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	health := make(map[string]bool, len(m.healthState))
	for k, v := range m.healthState {
		health[k] = v
	}
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
func (m *MockMonitor) SetHealth(component string, healthy bool) {
	m.Called(component, healthy)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.healthState[component] = healthy
}

// Reset resets all metrics
func (m *MockMonitor) Reset() {
	m.Called()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics = make(map[string]float64)
}

// ExpectRecordEvent sets up expectation for RecordEvent
func (m *MockMonitor) ExpectRecordEvent(ctx context.Context, event Event) *mock.Call {
	return m.On("RecordEvent", ctx, event)
}

// ExpectRecordMetric sets up expectation for RecordMetric
func (m *MockMonitor) ExpectRecordMetric(ctx context.Context, name string, value float64) *mock.Call {
	return m.On("RecordMetric", ctx, name, value)
}

// ExpectRecordMarketData sets up expectation for RecordMarketData
func (m *MockMonitor) ExpectRecordMarketData(ctx context.Context, data interface{}, duration time.Duration) *mock.Call {
	return m.On("RecordMarketData", ctx, data, duration)
}

// ExpectGetMetrics sets up expectation for GetMetrics
func (m *MockMonitor) ExpectGetMetrics(ctx context.Context, metrics map[string]float64) *mock.Call {
	return m.On("GetMetrics", ctx).Return(metrics)
}

// ExpectCheckHealth sets up expectation for CheckHealth
func (m *MockMonitor) ExpectCheckHealth(ctx context.Context, healthy bool, states map[string]bool) *mock.Call {
	return m.On("CheckHealth", ctx).Return(healthy, states)
}

// ExpectSetHealth sets up expectation for SetHealth
func (m *MockMonitor) ExpectSetHealth(component string, healthy bool) *mock.Call {
	return m.On("SetHealth", component, healthy)
}

// ExpectReset sets up expectation for Reset
func (m *MockMonitor) ExpectReset() *mock.Call {
	return m.On("Reset")
}
