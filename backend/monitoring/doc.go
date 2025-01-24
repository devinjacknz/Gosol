/*
Package monitoring provides a flexible and thread-safe monitoring system for tracking metrics,
events, and component health in a Go application.

Basic usage:

	monitor := monitoring.NewMonitor()
	api := monitoring.NewAPI(monitor)

	// Register API endpoints
	mux := http.NewServeMux()
	api.RegisterRoutes(mux)

	// Record metrics
	ctx := context.Background()
	monitor.RecordMetric(ctx, "requests_total", 42)

	// Record events
	monitor.RecordEvent(ctx, monitoring.Event{
		Type:      monitoring.MetricTrading,
		Severity:  monitoring.SeverityInfo,
		Message:   "Trade executed",
		Details:   map[string]interface{}{"volume": 100.0},
		Timestamp: time.Now(),
	})

	// Monitor component health
	monitor.SetHealth("database", true)
	monitor.SetHealth("api", true)

The package provides several key types:

	type Monitor struct { ... }      // Main monitoring implementation
	type Event struct { ... }        // Represents a monitoring event
	type API struct { ... }          // HTTP API for metrics and health
	type MockMonitor struct { ... }  // Mock implementation for testing

Metric Types:

	MetricTrading        // Trading-related metrics
	MetricMarketData     // Market data processing metrics
	MetricProcessing     // General processing metrics
	MetricTaskCompletion // Task completion metrics
	MetricSystem         // System-level metrics

Severity Levels:

	SeverityInfo     // Informational events
	SeverityWarning  // Warning events
	SeverityError    // Error events
	SeverityCritical // Critical events

HTTP Endpoints:

	GET /api/metrics - Returns all recorded metrics as JSON
	GET /api/health  - Returns system health status as JSON

Thread Safety:

All monitor operations are thread-safe through the use of a read-write mutex.
The monitor can be safely used from multiple goroutines.

Testing:

The package includes a mock implementation for testing:

	func TestYourFunction(t *testing.T) {
		mockMonitor := monitoring.NewMockMonitor()
		mockMonitor.ExpectRecordMetric(ctx, "test_metric", 42.0)
		YourFunction(mockMonitor)
		mockMonitor.AssertExpectations(t)
	}

Best Practices:

1. Use meaningful metric names that follow a consistent naming convention
2. Include relevant details in events to aid debugging
3. Regularly check component health and set up alerts for unhealthy states
4. Use the mock monitor in tests to verify monitoring behavior
5. Consider implementing metric aggregation for high-volume metrics
*/
package monitoring
