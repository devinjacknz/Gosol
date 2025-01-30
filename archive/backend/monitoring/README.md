# Monitoring Package

The monitoring package provides a flexible and thread-safe monitoring system for tracking metrics, events, and component health in a Go application.

## Features

- Metric tracking with support for counters and gauges
- Event recording with severity levels and custom details
- Component health monitoring
- Thread-safe operations
- HTTP API endpoints for metrics and health checks
- Mock implementation for testing

## Installation

```bash
go get solmeme-trader/monitoring
```

## Quick Start

```go
import (
    "context"
    "net/http"
    "solmeme-trader/monitoring"
)

func main() {
    // Create monitor
    monitor := monitoring.NewMonitor()

    // Create and register API endpoints
    api := monitoring.NewAPI(monitor)
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

    // Start HTTP server
    http.ListenAndServe(":8080", mux)
}
```

## API Endpoints

### GET /api/metrics

Returns all recorded metrics as JSON.

Example response:
```json
{
    "requests_total": 42,
    "response_time_ms": 123.45,
    "trade_count": 10,
    "trade_volume": 1000
}
```

### GET /api/health

Returns system health status as JSON.

Example response:
```json
{
    "healthy": true,
    "state": {
        "database": true,
        "api": true,
        "cache": true
    }
}
```

## Metric Types

- `MetricTrading`: Trading-related metrics
- `MetricMarketData`: Market data processing metrics
- `MetricProcessing`: General processing metrics
- `MetricTaskCompletion`: Task completion metrics
- `MetricSystem`: System-level metrics

## Severity Levels

- `SeverityInfo`: Informational events
- `SeverityWarning`: Warning events
- `SeverityError`: Error events
- `SeverityCritical`: Critical events

## Testing

The package includes a mock implementation for testing:

```go
func TestYourFunction(t *testing.T) {
    mockMonitor := monitoring.NewMockMonitor()
    
    // Setup expectations
    mockMonitor.ExpectRecordMetric(ctx, "test_metric", 42.0)
    
    // Run your test
    YourFunction(mockMonitor)
    
    // Verify expectations
    mockMonitor.AssertExpectations(t)
}
```

## Benchmarks

Run benchmarks to measure performance:

```bash
go test -bench=. -benchmem
```

Example benchmark results:
```
BenchmarkMonitor_RecordEvent-8         1000000    1234 ns/op    789 B/op    5 allocs/op
BenchmarkMonitor_RecordMetric-8        2000000     567 ns/op    234 B/op    3 allocs/op
BenchmarkMonitor_GetMetrics-8          3000000     345 ns/op    123 B/op    2 allocs/op
```

## Thread Safety

All monitor operations are thread-safe through the use of a read-write mutex. The monitor can be safely used from multiple goroutines.

## Best Practices

1. Use meaningful metric names that follow a consistent naming convention
2. Include relevant details in events to aid debugging
3. Regularly check component health and set up alerts for unhealthy states
4. Use the mock monitor in tests to verify monitoring behavior
5. Consider implementing metric aggregation for high-volume metrics

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
