package analysis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMarketAnalyzer(t *testing.T) {
	analyzer := NewMarketAnalyzer()
	assert.NotNil(t, analyzer)
	assert.Empty(t, analyzer.historicalData)
}

func TestAddMarketData(t *testing.T) {
	analyzer := NewMarketAnalyzer()
	data := MarketData{
		Price:     100,
		Volume:    1000,
		Timestamp: time.Now(),
	}

	analyzer.AddMarketData(data)
	assert.Len(t, analyzer.historicalData, 1)
	assert.Equal(t, data, analyzer.historicalData[0])
}

func TestCalculateMetrics(t *testing.T) {
	analyzer := NewMarketAnalyzer()
	trades := []Trade{
		{
			EntryPrice: 100,
			ExitPrice:  110,
			Size:       1,
			Profit:     10,
			Timestamp:  time.Now(),
		},
		{
			EntryPrice: 110,
			ExitPrice:  105,
			Size:       1,
			Profit:     -5,
			Timestamp:  time.Now(),
		},
		{
			EntryPrice: 105,
			ExitPrice:  115,
			Size:       1,
			Profit:     10,
			Timestamp:  time.Now(),
		},
	}

	metrics := analyzer.CalculateMetrics(trades)
	assert.Equal(t, 3, metrics.TotalTrades)
	assert.Equal(t, 2, metrics.ProfitableTrades)
	assert.InDelta(t, 0.667, metrics.WinRate, 0.001)
	assert.InDelta(t, 4.0, metrics.ProfitFactor, 0.001) // (20)/(5)
	assert.InDelta(t, 15.0, metrics.TotalProfit, 0.001)
	assert.InDelta(t, 10.0, metrics.AverageWin, 0.001)
	assert.InDelta(t, 5.0, metrics.AverageLoss, 0.001)
}

func TestAnalyzeTrend(t *testing.T) {
	analyzer := NewMarketAnalyzer()
	now := time.Now()

	// Add test data for uptrend
	for i := 0; i < 30; i++ {
		analyzer.AddMarketData(MarketData{
			Price:     100 + float64(i),
			Volume:    1000,
			Timestamp: now.Add(time.Duration(i) * time.Hour),
		})
	}

	trend, strength := analyzer.AnalyzeTrend("1d")
	assert.Equal(t, "bullish", trend)
	assert.Greater(t, strength, 0.0)

	// Test downtrend
	analyzer = NewMarketAnalyzer()
	for i := 0; i < 30; i++ {
		analyzer.AddMarketData(MarketData{
			Price:     100 - float64(i),
			Volume:    1000,
			Timestamp: now.Add(time.Duration(i) * time.Hour),
		})
	}

	trend, strength = analyzer.AnalyzeTrend("1d")
	assert.Equal(t, "bearish", trend)
	assert.Greater(t, strength, 0.0)
}

func TestAnalyzeVolatility(t *testing.T) {
	analyzer := NewMarketAnalyzer()
	now := time.Now()

	tests := []struct {
		name          string
		prices        []float64
		expectHighVol bool
	}{
		{
			name:          "High volatility",
			prices:        []float64{100, 120, 90, 110, 95},
			expectHighVol: true,
		},
		{
			name:          "Low volatility",
			prices:        []float64{100, 101, 99, 100, 101},
			expectHighVol: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer = NewMarketAnalyzer()
			for i, price := range tt.prices {
				analyzer.AddMarketData(MarketData{
					Price:     price,
					Volume:    1000,
					Timestamp: now.Add(time.Duration(i) * time.Hour),
				})
			}

			volatility := analyzer.AnalyzeVolatility("1d")
			if tt.expectHighVol {
				assert.Greater(t, volatility, 0.1)
			} else {
				assert.Less(t, volatility, 0.1)
			}
		})
	}
}

func TestGenerateReport(t *testing.T) {
	analyzer := NewMarketAnalyzer()
	now := time.Now()

	// Add test data
	for i := 0; i < 30; i++ {
		analyzer.AddMarketData(MarketData{
			Price:     100 + float64(i%5),
			Volume:    1000 + float64(i*100),
			Timestamp: now.Add(time.Duration(i) * time.Hour),
		})
	}

	ctx := context.Background()
	report, err := analyzer.GenerateReport(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.NotEmpty(t, report.Trend)
	assert.Greater(t, report.Support, 0.0)
	assert.Greater(t, report.Resistance, report.Support)
}

func TestCalculateMA(t *testing.T) {
	analyzer := NewMarketAnalyzer()
	now := time.Now()

	// Add test data with known average
	prices := []float64{10, 20, 30, 40, 50}
	for i, price := range prices {
		analyzer.AddMarketData(MarketData{
			Price:     price,
			Volume:    1000,
			Timestamp: now.Add(time.Duration(i) * time.Hour),
		})
	}

	// Test 3-period MA
	ma := analyzer.calculateMA(3)
	assert.Len(t, ma, 3)
	assert.InDelta(t, 20, ma[0], 0.001) // (10+20+30)/3
	assert.InDelta(t, 30, ma[1], 0.001) // (20+30+40)/3
	assert.InDelta(t, 40, ma[2], 0.001) // (30+40+50)/3
}

func TestCalculateMaxDrawdown(t *testing.T) {
	analyzer := NewMarketAnalyzer()
	trades := []Trade{
		{Profit: 10},
		{Profit: -5},
		{Profit: -3},
		{Profit: 7},
		{Profit: -6},
	}

	maxDrawdown := analyzer.calculateMaxDrawdown(trades)
	assert.InDelta(t, 0.4, maxDrawdown, 0.001) // Maximum drawdown from peak 10 to 2
}

func TestCalculateSharpeRatio(t *testing.T) {
	analyzer := NewMarketAnalyzer()
	trades := []Trade{
		{Profit: 10},
		{Profit: 8},
		{Profit: -5},
		{Profit: 12},
		{Profit: -3},
	}

	sharpeRatio := analyzer.calculateSharpeRatio(trades)
	assert.Greater(t, sharpeRatio, 0.0)
}

func TestCalculateSupportResistance(t *testing.T) {
	analyzer := NewMarketAnalyzer()
	now := time.Now()

	prices := []float64{100, 105, 95, 110, 98, 108, 96, 112, 97, 109}
	for i, price := range prices {
		analyzer.AddMarketData(MarketData{
			Price:     price,
			Volume:    1000,
			Timestamp: now.Add(time.Duration(i) * time.Hour),
		})
	}

	support, resistance := analyzer.calculateSupportResistance()
	assert.InDelta(t, 96.0, support, 1.0)     // Around 10th percentile
	assert.InDelta(t, 110.0, resistance, 1.0) // Around 90th percentile
}
