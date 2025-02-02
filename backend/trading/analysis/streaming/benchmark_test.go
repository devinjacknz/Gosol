package streaming

import (
	"context"
	"testing"
	"time"
)

func generateTestPrices(n int) []Price {
	prices := make([]Price, n)
	now := time.Now()
	basePrice := 100.0

	for i := 0; i < n; i++ {
		prices[i] = Price{
			Timestamp: now.Add(time.Duration(i) * time.Minute),
			Value:     basePrice + float64(i%10),
			Volume:    1000.0 + float64(i%100),
		}
	}
	return prices
}

func BenchmarkRSI(b *testing.B) {
	ctx := context.Background()
	rsi, _ := NewRSI(14)
	prices := generateTestPrices(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, price := range prices {
			_, err := rsi.Update(ctx, price)
			if err != nil {
				b.Fatalf("Failed to update RSI: %v", err)
			}
		}
		rsi.Reset()
	}
}

func BenchmarkEMA(b *testing.B) {
	ctx := context.Background()
	ema, _ := NewEMA(10)
	prices := generateTestPrices(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, price := range prices {
			_, err := ema.Update(ctx, price)
			if err != nil {
				b.Fatalf("Failed to update EMA: %v", err)
			}
		}
		ema.Reset()
	}
}

func BenchmarkMACD(b *testing.B) {
	ctx := context.Background()
	macd, _ := NewMACD(12, 26, 9)
	prices := generateTestPrices(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, price := range prices {
			_, err := macd.Update(ctx, price)
			if err != nil {
				b.Fatalf("Failed to update MACD: %v", err)
			}
		}
		macd.Reset()
	}
}

func BenchmarkIndicatorPipeline(b *testing.B) {
	ctx := context.Background()
	pipeline := NewIndicatorPipeline()
	factory := NewIndicatorFactory()

	// Add indicators
	rsi, _ := factory.CreateRSI(14)
	ema, _ := factory.CreateEMA(10)
	macd, _ := factory.CreateMACD(12, 26, 9)

	pipeline.AddIndicator(rsi)
	pipeline.AddIndicator(ema)
	pipeline.AddIndicator(macd)

	prices := generateTestPrices(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, price := range prices {
			err := pipeline.ProcessPrice(ctx, price)
			if err != nil {
				b.Fatalf("Failed to process price: %v", err)
			}
		}
		pipeline.Reset()
	}
}
