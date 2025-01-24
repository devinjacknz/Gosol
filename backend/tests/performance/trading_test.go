package performance

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func BenchmarkTradeExecution(b *testing.B) {
	ctx := context.Background()
	market := setupTestMarket()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trade := &Trade{
			TokenAddress: "TestToken",
			Action:       "buy",
			Amount:       1.0,
			Price:        100.0,
			Timestamp:    time.Now(),
		}
		_ = ExecuteTrade(ctx, trade, market)
	}
}

func BenchmarkMarketDataProcessing(b *testing.B) {
	ctx := context.Background()
	analyzer := setupTestAnalyzer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = analyzer.AnalyzeMarket(ctx, "TestToken")
	}
}

func BenchmarkConcurrentTrading(b *testing.B) {
	ctx := context.Background()
	market := setupTestMarket()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			trade := &Trade{
				TokenAddress: "TestToken",
				Action:       "buy",
				Amount:       1.0,
				Price:        100.0,
				Timestamp:    time.Now(),
			}
			_ = ExecuteTrade(ctx, trade, market)
		}
	})
}

func BenchmarkRiskManagement(b *testing.B) {
	rm := NewRiskManager(0.05, 0.10)
	position := Position{
		EntryPrice: 100.0,
		Size:       1.0,
		Side:       "long",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rm.CheckPosition(position, 105.0)
	}
}

func TestTradingLatency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping latency test in short mode")
	}

	ctx := context.Background()
	market := setupTestMarket()

	// Measure trade execution latency
	trades := 1000
	latencies := make([]time.Duration, trades)

	for i := 0; i < trades; i++ {
		trade := &Trade{
			TokenAddress: "TestToken",
			Action:       "buy",
			Amount:       1.0,
			Price:        100.0,
			Timestamp:    time.Now(),
		}

		start := time.Now()
		err := ExecuteTrade(ctx, trade, market)
		latencies[i] = time.Since(start)

		require.NoError(t, err)
	}

	// Calculate statistics
	var total time.Duration
	var max time.Duration
	for _, lat := range latencies {
		total += lat
		if lat > max {
			max = lat
		}
	}

	avg := total / time.Duration(trades)
	t.Logf("Average latency: %v", avg)
	t.Logf("Maximum latency: %v", max)

	// Assert performance requirements
	require.Less(t, avg, 10*time.Millisecond, "Average latency too high")
	require.Less(t, max, 50*time.Millisecond, "Maximum latency too high")
}

func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory usage test in short mode")
	}

	ctx := context.Background()
	market := setupTestMarket()
	analyzer := setupTestAnalyzer()

	// Monitor memory usage during intensive operations
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	initialAlloc := m.Alloc

	// Perform intensive operations
	for i := 0; i < 10000; i++ {
		trade := &Trade{
			TokenAddress: "TestToken",
			Action:       "buy",
			Amount:       1.0,
			Price:        100.0,
			Timestamp:    time.Now(),
		}
		_ = ExecuteTrade(ctx, trade, market)
		_, _ = analyzer.AnalyzeMarket(ctx, "TestToken")
	}

	runtime.ReadMemStats(&m)
	finalAlloc := m.Alloc

	memoryIncrease := finalAlloc - initialAlloc
	t.Logf("Memory increase: %d bytes", memoryIncrease)

	// Assert reasonable memory usage
	require.Less(t, memoryIncrease, uint64(50*1024*1024), "Memory usage too high")
}

func TestConcurrencyHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrency test in short mode")
	}

	ctx := context.Background()
	market := setupTestMarket()

	// Test concurrent trade execution
	trades := 100
	concurrency := 10
	errors := make(chan error, trades)
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < trades/concurrency; j++ {
				trade := &Trade{
					TokenAddress: "TestToken",
					Action:       "buy",
					Amount:       1.0,
					Price:        100.0,
					Timestamp:    time.Now(),
				}
				if err := ExecuteTrade(ctx, trade, market); err != nil {
					errors <- err
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for any errors during concurrent execution
	for err := range errors {
		require.NoError(t, err)
	}
}

// Helper functions to set up test environment
func setupTestMarket() MarketService {
	// Implementation depends on your actual market service
	return NewTestMarketService()
}

func setupTestAnalyzer() MarketAnalyzer {
	// Implementation depends on your actual analyzer
	return NewTestMarketAnalyzer()
}
