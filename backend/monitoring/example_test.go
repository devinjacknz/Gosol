package monitoring_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"solmeme-trader/monitoring"
)

func Example() {
	// Create a new monitor
	monitor := monitoring.NewMonitor()

	// Create monitoring API
	api := monitoring.NewAPI(monitor)

	// Register API routes
	mux := http.NewServeMux()
	api.RegisterRoutes(mux)

	// Record some metrics
	ctx := context.Background()
	monitor.RecordMetric(ctx, "requests_total", 42)
	monitor.RecordMetric(ctx, "response_time_ms", 123.45)

	// Record an event
	monitor.RecordEvent(ctx, monitoring.Event{
		Type:      monitoring.MetricTrading,
		Severity:  monitoring.SeverityInfo,
		Message:   "Trade executed",
		Details:   map[string]interface{}{"volume": 100.0},
		Timestamp: time.Now(),
	})

	// Set component health
	monitor.SetHealth("database", true)
	monitor.SetHealth("api", true)
	monitor.SetHealth("cache", false)

	// Start HTTP server
	go func() {
		if err := http.ListenAndServe(":8080", mux); err != nil {
			log.Fatal(err)
		}
	}()

	// Example of checking metrics
	metrics := monitor.GetMetrics(ctx)
	fmt.Printf("Total requests: %v\n", metrics["requests_total"])
	fmt.Printf("Average response time: %v ms\n", metrics["response_time_ms"])

	// Example of checking health
	healthy, states := monitor.CheckHealth(ctx)
	fmt.Printf("System healthy: %v\n", healthy)
	for component, state := range states {
		fmt.Printf("Component %s health: %v\n", component, state)
	}

	// Output:
	// Total requests: 42
	// Average response time: 123.45 ms
	// System healthy: false
	// Component database health: true
	// Component api health: true
	// Component cache health: false
}

func ExampleMonitor_RecordEvent() {
	monitor := monitoring.NewMonitor()
	ctx := context.Background()

	// Record a trading event
	monitor.RecordEvent(ctx, monitoring.Event{
		Type:      monitoring.MetricTrading,
		Severity:  monitoring.SeverityInfo,
		Message:   "Trade executed",
		Details:   map[string]interface{}{"volume": 100.0},
		Timestamp: time.Now(),
	})

	// Get metrics
	metrics := monitor.GetMetrics(ctx)
	fmt.Printf("Trade count: %v\n", metrics["trade_count"])
	fmt.Printf("Trade volume: %v\n", metrics["trade_volume"])

	// Output:
	// Trade count: 1
	// Trade volume: 100
}

func ExampleMonitor_RecordMetric() {
	monitor := monitoring.NewMonitor()
	ctx := context.Background()

	// Record some metrics
	monitor.RecordMetric(ctx, "requests", 42)
	monitor.RecordMetric(ctx, "latency_ms", 123.45)

	// Get metrics
	metrics := monitor.GetMetrics(ctx)
	for name, value := range metrics {
		fmt.Printf("%s: %v\n", name, value)
	}

	// Output:
	// latency_ms: 123.45
	// requests: 42
}

func ExampleMonitor_CheckHealth() {
	monitor := monitoring.NewMonitor()
	ctx := context.Background()

	// Set component health
	monitor.SetHealth("api", true)
	monitor.SetHealth("database", false)

	// Check health
	healthy, states := monitor.CheckHealth(ctx)
	fmt.Printf("System healthy: %v\n", healthy)
	for component, state := range states {
		fmt.Printf("%s healthy: %v\n", component, state)
	}

	// Output:
	// System healthy: false
	// api healthy: true
	// database healthy: false
}

func ExampleAPI() {
	monitor := monitoring.NewMonitor()
	api := monitoring.NewAPI(monitor)

	// Register routes
	mux := http.NewServeMux()
	api.RegisterRoutes(mux)

	// Start server
	go func() {
		if err := http.ListenAndServe(":8080", mux); err != nil {
			log.Fatal(err)
		}
	}()

	fmt.Println("Monitoring API endpoints:")
	fmt.Println("- GET /api/metrics")
	fmt.Println("- GET /api/health")

	// Output:
	// Monitoring API endpoints:
	// - GET /api/metrics
	// - GET /api/health
}
