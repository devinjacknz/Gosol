package monitoring

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMonitor_RecordEvent(t *testing.T) {
	monitor := NewMonitor()
	ctx := TestContext(t)

	// Test trading event
	monitor.RecordEvent(ctx, TestEvent(MetricTrading, SeverityInfo).
		WithDetails(map[string]interface{}{"volume": 100.0}))

	metrics := monitor.GetMetrics(ctx)
	assert.Equal(t, float64(1), metrics["trade_count"])
	assert.Equal(t, float64(100), metrics["trade_volume"])

	// Test market data event
	monitor.RecordEvent(ctx, TestEvent(MetricMarketData, SeverityInfo))

	metrics = monitor.GetMetrics(ctx)
	assert.Equal(t, float64(1), metrics["market_data_count"])

	// Test processing event
	monitor.RecordEvent(ctx, TestEvent(MetricProcessing, SeverityInfo).
		WithDetails(map[string]interface{}{"duration": "1s"}))

	metrics = monitor.GetMetrics(ctx)
	assert.Equal(t, float64(1), metrics["processing_time"])
}

func TestMonitor_RecordMetric(t *testing.T) {
	monitor := NewMonitor()
	ctx := TestContext(t)

	// Record metrics
	monitor.RecordMetric(ctx, "test_metric", 42.0)
	monitor.RecordMetric(ctx, "another_metric", 123.0)

	// Get metrics
	metrics := monitor.GetMetrics(ctx)

	// Verify metrics
	assert.Equal(t, float64(42), metrics["test_metric"])
	assert.Equal(t, float64(123), metrics["another_metric"])

	// Update metric
	monitor.RecordMetric(ctx, "test_metric", 84.0)
	metrics = monitor.GetMetrics(ctx)
	assert.Equal(t, float64(84), metrics["test_metric"])
}

func TestMonitor_RecordMarketData(t *testing.T) {
	monitor := NewMonitor()
	ctx := TestContext(t)

	// Record market data
	monitor.RecordMarketData(ctx, "test_data", 2*time.Second)

	// Get metrics
	metrics := monitor.GetMetrics(ctx)

	// Verify metrics
	assert.Equal(t, float64(1), metrics["market_data_count"])
	assert.Equal(t, float64(2), metrics["market_data_processing_time"])

	// Record another market data
	monitor.RecordMarketData(ctx, "more_data", 3*time.Second)
	metrics = monitor.GetMetrics(ctx)
	assert.Equal(t, float64(2), metrics["market_data_count"])
	assert.Equal(t, float64(5), metrics["market_data_processing_time"])
}

func TestMonitor_Health(t *testing.T) {
	monitor := NewMonitor()
	ctx := TestContext(t)

	// Initially no health states
	healthy, states := monitor.CheckHealth(ctx)
	assert.True(t, healthy)
	assert.Empty(t, states)

	// Set some health states
	monitor.SetHealth("component1", true)
	monitor.SetHealth("component2", true)

	healthy, states = monitor.CheckHealth(ctx)
	assert.True(t, healthy)
	assert.Equal(t, map[string]bool{
		"component1": true,
		"component2": true,
	}, states)

	// Make one component unhealthy
	monitor.SetHealth("component2", false)

	healthy, states = monitor.CheckHealth(ctx)
	assert.False(t, healthy)
	assert.Equal(t, map[string]bool{
		"component1": true,
		"component2": false,
	}, states)
}

func TestMonitor_Reset(t *testing.T) {
	monitor := NewMonitor()
	ctx := TestContext(t)

	// Record some metrics
	monitor.RecordMetric(ctx, "test_metric", 42.0)
	monitor.RecordMetric(ctx, "another_metric", 123.0)

	// Set some health states
	monitor.SetHealth("component1", true)
	monitor.SetHealth("component2", false)

	// Verify initial state
	metrics := monitor.GetMetrics(ctx)
	assert.Equal(t, float64(42), metrics["test_metric"])
	assert.Equal(t, float64(123), metrics["another_metric"])

	healthy, states := monitor.CheckHealth(ctx)
	assert.False(t, healthy)
	assert.NotEmpty(t, states)

	// Reset monitor
	monitor.Reset()

	// Verify metrics are cleared
	metrics = monitor.GetMetrics(ctx)
	assert.Empty(t, metrics)

	// Verify health states are preserved
	healthy, states = monitor.CheckHealth(ctx)
	assert.False(t, healthy)
	assert.NotEmpty(t, states)
}

// Helper method to add details to test event
func (e Event) WithDetails(details map[string]interface{}) Event {
	e.Details = details
	return e
}
