package streaming

import (
	"testing"
	"time"

"github.com/devinjacknz/godydxhyber/backend/trading/analysis/monitoring"
)

func BenchmarkIndicatorWithMonitoring(b *testing.B) {
	indicators := []struct {
		name     string
		function func([]float64) float64
	}{
		{"EMA10", func(data []float64) float64 {
			start := time.Now()
			result := EMA(data, 10)
			duration := time.Since(start)
			monitoring.RecordIndicatorCalculation("EMA10", duration)
			monitoring.RecordIndicatorValue("EMA10", result)
			return result
		}},
		{"MACD", func(data []float64) float64 {
			start := time.Now()
			result := MACD(data, 12, 26, 9)
			duration := time.Since(start)
			monitoring.RecordIndicatorCalculation("MACD", duration)
			monitoring.RecordIndicatorValue("MACD", result)
			return result
		}},
		{"RSI14", func(data []float64) float64 {
			start := time.Now()
			result := RSI(data, 14)
			duration := time.Since(start)
			monitoring.RecordIndicatorCalculation("RSI14", duration)
			monitoring.RecordIndicatorValue("RSI14", result)
			return result
		}},
	}

	data := generateTestData(1000)
	batchSize := 100

	for _, ind := range indicators {
		b.Run(ind.name, func(b *testing.B) {
			start := time.Now()
			for i := 0; i < b.N; i++ {
				for j := 0; j < len(data)-batchSize; j += batchSize {
					batch := data[j : j+batchSize]
					ind.function(batch)
				}
			}
			duration := time.Since(start)
			monitoring.RecordBatchProcessing(ind.name, duration, batchSize)
		})
	}
}

func generateTestData(n int) []float64 {
	data := make([]float64, n)
	for i := range data {
		data[i] = float64(i)
	}
	return data
}
