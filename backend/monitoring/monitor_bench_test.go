package monitoring

import (
	"context"
	"testing"
)

func BenchmarkMonitor_RecordEvent(b *testing.B) {
	monitor := NewMonitor()
	defer monitor.Close()

	ctx := context.Background()
	event := NewTestEvent(MetricTrading, SeverityInfo, "Test event")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.RecordEvent(ctx, event)
	}
}

func BenchmarkMonitor_RecordMetric(b *testing.B) {
	monitor := NewMonitor()
	defer monitor.Close()

	ctx := context.Background()
	tags := map[string]string{"tag": "value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.RecordMetric(ctx, "test_metric", float64(i), tags)
	}
}

func BenchmarkMonitor_RecordMarketData(b *testing.B) {
	monitor := NewMonitor()
	defer monitor.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.RecordMarketData(ctx, "SOL", float64(i), float64(i*100))
	}
}

func BenchmarkMonitor_GetMetrics(b *testing.B) {
	monitor := NewMonitor()
	defer monitor.Close()

	ctx := context.Background()
	for i := 0; i < 1000; i++ {
		monitor.RecordEvent(ctx, NewTestEvent(MetricTrading, SeverityInfo, "Trade executed successfully"))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.GetMetrics()
	}
}

func BenchmarkMonitor_CheckHealth(b *testing.B) {
	monitor := NewMonitor()
	defer monitor.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.CheckHealth(ctx)
	}
}

func BenchmarkMonitor_GetEvents(b *testing.B) {
	monitor := NewMonitor()
	defer monitor.Close()

	ctx := context.Background()
	for i := 0; i < 1000; i++ {
		monitor.RecordEvent(ctx, NewTestEvent(MetricTrading, SeverityInfo, "Test event"))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.GetEvents()
	}
}

func BenchmarkMonitor_GetEventsByType(b *testing.B) {
	monitor := NewMonitor()
	defer monitor.Close()

	ctx := context.Background()
	for i := 0; i < 1000; i++ {
		monitor.RecordEvent(ctx, NewTestEvent(MetricTrading, SeverityInfo, "Test event"))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.GetEventsByType(MetricTrading)
	}
}

func BenchmarkMonitor_Concurrent(b *testing.B) {
	monitor := NewMonitor()
	defer monitor.Close()

	ctx := context.Background()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			monitor.RecordEvent(ctx, NewTestEvent(MetricTrading, SeverityInfo, "Test event"))
			monitor.GetMetrics()
			monitor.CheckHealth(ctx)
		}
	})
}
