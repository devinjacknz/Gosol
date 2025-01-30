package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonitor_RecordEvent(t *testing.T) {
	monitor := NewMonitor()
	defer monitor.Close()

	ctx := context.Background()
	event := NewTestEvent(MetricTrading, SeverityInfo, "Test event")
	monitor.RecordEvent(ctx, event)

	events := monitor.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, "Test event", events[0].Message)
}

func TestMonitor_RecordMetric(t *testing.T) {
	monitor := NewMonitor()
	defer monitor.Close()

	ctx := context.Background()
	monitor.RecordMetric(ctx, "test_metric", 42.0, map[string]string{"tag": "value"})

	events := monitor.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, MetricSystem, events[0].Type)

	details, ok := events[0].Details.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test_metric", details["name"])
	assert.Equal(t, 42.0, details["value"])
}

func TestMonitor_RecordMarketData(t *testing.T) {
	monitor := NewMonitor()
	defer monitor.Close()

	ctx := context.Background()
	monitor.RecordMarketData(ctx, "SOL", 100.0, 1000.0)

	events := monitor.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, MetricMarketData, events[0].Type)

	details, ok := events[0].Details.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "SOL", details["tokenAddress"])
	assert.Equal(t, 100.0, details["price"])
	assert.Equal(t, 1000.0, details["volume"])
}

func TestMonitor_Health(t *testing.T) {
	monitor := NewMonitor()
	defer monitor.Close()

	ctx := context.Background()

	// Initial state should be healthy
	healthy, subsystems := monitor.CheckHealth(ctx)
	assert.True(t, healthy)
	assert.True(t, subsystems["database"])
	assert.True(t, subsystems["api"])
	assert.True(t, subsystems["processing"])

	// Set unhealthy state
	monitor.SetHealth("unhealthy")
	healthy, subsystems = monitor.CheckHealth(ctx)
	assert.False(t, healthy)
	assert.False(t, subsystems["database"])
	assert.False(t, subsystems["api"])
	assert.False(t, subsystems["processing"])

	// Reset to healthy state
	monitor.SetHealth("healthy")
	healthy, subsystems = monitor.CheckHealth(ctx)
	assert.True(t, healthy)
	assert.True(t, subsystems["database"])
}

func TestMonitor_Metrics(t *testing.T) {
	monitor := NewMonitor()
	defer monitor.Close()

	ctx := context.Background()

	// Record some trades
	monitor.RecordEvent(ctx, NewTestEvent(MetricTrading, SeverityInfo, "Trade execution attempt"))
	monitor.RecordEvent(ctx, NewTestEvent(MetricTrading, SeverityInfo, "Trade executed successfully"))
	monitor.RecordEvent(ctx, NewTestEvent(MetricTrading, SeverityInfo, "Trade execution failed"))

	metrics := monitor.GetMetrics()
	assert.Equal(t, int64(1), metrics.TotalTrades)
	assert.Equal(t, int64(1), metrics.SuccessfulTrades)
	assert.Equal(t, int64(1), metrics.FailedTrades)

	// Test success rate
	rate := monitor.GetSuccessRate()
	assert.Equal(t, 1.0, rate)
}

func TestMonitor_Reset(t *testing.T) {
	monitor := NewMonitor()
	defer monitor.Close()

	ctx := context.Background()

	// Record some events
	monitor.RecordEvent(ctx, NewTestEvent(MetricTrading, SeverityInfo, "Test event"))
	assert.Len(t, monitor.GetEvents(), 1)

	// Reset monitor
	monitor.Reset()
	assert.Len(t, monitor.GetEvents(), 0)

	metrics := monitor.GetMetrics()
	assert.Equal(t, int64(0), metrics.TotalTrades)
	assert.Equal(t, int64(0), metrics.SuccessfulTrades)
}

func TestMonitor_EventHandlers(t *testing.T) {
	monitor := NewMonitor()
	defer monitor.Close()

	ctx := context.Background()
	handlerCalled := false

	// Add event handler
	monitor.AddEventHandler(MetricTrading, func(event Event) {
		handlerCalled = true
		assert.Equal(t, "Test event", event.Message)
	})

	// Record event
	monitor.RecordEvent(ctx, NewTestEvent(MetricTrading, SeverityInfo, "Test event"))
	time.Sleep(100 * time.Millisecond) // Wait for event processing

	assert.True(t, handlerCalled)
}

func TestMonitor_GetEventsByType(t *testing.T) {
	monitor := NewMonitor()
	defer monitor.Close()

	ctx := context.Background()

	// Record events of different types
	monitor.RecordEvent(ctx, NewTestEvent(MetricTrading, SeverityInfo, "Trading event"))
	monitor.RecordEvent(ctx, NewTestEvent(MetricSystem, SeverityInfo, "System event"))
	monitor.RecordEvent(ctx, NewTestEvent(MetricTrading, SeverityInfo, "Another trading event"))

	// Get trading events
	tradingEvents := monitor.GetEventsByType(MetricTrading)
	assert.Len(t, tradingEvents, 2)
	for _, event := range tradingEvents {
		assert.Equal(t, MetricTrading, event.Type)
	}

	// Get system events
	systemEvents := monitor.GetEventsByType(MetricSystem)
	assert.Len(t, systemEvents, 1)
	assert.Equal(t, MetricSystem, systemEvents[0].Type)
}
