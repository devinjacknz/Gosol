package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func BenchmarkMonitor_RecordEvent(b *testing.B) {
	monitor := NewMonitor()
	ctx := context.Background()
	event := TestEvent(MetricTrading, SeverityInfo).
		WithDetails(map[string]interface{}{"volume": 100.0})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.RecordEvent(ctx, event)
	}
}

func BenchmarkMonitor_RecordMetric(b *testing.B) {
	monitor := NewMonitor()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.RecordMetric(ctx, fmt.Sprintf("metric_%d", i), float64(i))
	}
}

func BenchmarkMonitor_RecordMarketData(b *testing.B) {
	monitor := NewMonitor()
	ctx := context.Background()
	data := "test_data"
	duration := time.Second

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.RecordMarketData(ctx, data, duration)
	}
}

func BenchmarkMonitor_GetMetrics(b *testing.B) {
	monitor := NewMonitor()
	ctx := context.Background()

	// Setup some metrics
	for i := 0; i < 1000; i++ {
		monitor.RecordMetric(ctx, fmt.Sprintf("metric_%d", i), float64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.GetMetrics(ctx)
	}
}

func BenchmarkMonitor_CheckHealth(b *testing.B) {
	monitor := NewMonitor()
	ctx := context.Background()

	// Setup some health states
	for i := 0; i < 100; i++ {
		monitor.SetHealth(fmt.Sprintf("component_%d", i), i%2 == 0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.CheckHealth(ctx)
	}
}

func BenchmarkMonitor_SetHealth(b *testing.B) {
	monitor := NewMonitor()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.SetHealth(fmt.Sprintf("component_%d", i), i%2 == 0)
	}
}

func BenchmarkMonitor_Concurrent(b *testing.B) {
	monitor := NewMonitor()
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			switch i % 5 {
			case 0:
				monitor.RecordEvent(ctx, TestEvent(MetricTrading, SeverityInfo))
			case 1:
				monitor.RecordMetric(ctx, fmt.Sprintf("metric_%d", i), float64(i))
			case 2:
				monitor.RecordMarketData(ctx, "test_data", time.Second)
			case 3:
				monitor.GetMetrics(ctx)
			case 4:
				monitor.CheckHealth(ctx)
			}
			i++
		}
	})
}

func BenchmarkAPI_HandleMetrics(b *testing.B) {
	monitor := NewMonitor()
	api := NewAPI(monitor)
	ctx := context.Background()

	// Setup some metrics
	for i := 0; i < 1000; i++ {
		monitor.RecordMetric(ctx, fmt.Sprintf("metric_%d", i), float64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := TestRequest(nil, "GET", "/api/metrics")
		w := &benchmarkResponseRecorder{
			headers: make(http.Header),
		}
		api.handleMetrics(w, req)
	}
}

// benchmarkResponseRecorder implements http.ResponseWriter for benchmarking
type benchmarkResponseRecorder struct {
	headers http.Header
	code    int
	body    []byte
}

func (r *benchmarkResponseRecorder) Header() http.Header {
	return r.headers
}

func (r *benchmarkResponseRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return len(b), nil
}

func (r *benchmarkResponseRecorder) WriteHeader(code int) {
	r.code = code
}
