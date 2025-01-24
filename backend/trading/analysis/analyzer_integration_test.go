package analysis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMarketAnalyzerIntegration(t *testing.T) {
	ctx := context.Background()
	analyzer := NewMarketAnalyzer()

	// Setup test data
	now := time.Now()
	testData := []MarketData{
		{Price: 100, Volume: 1000, Timestamp: now.Add(-10 * time.Hour)},
		{Price: 102, Volume: 1200, Timestamp: now.Add(-9 * time.Hour)},
		{Price: 98, Volume: 800, Timestamp: now.Add(-8 * time.Hour)},
		{Price: 103, Volume: 1500, Timestamp: now.Add(-7 * time.Hour)},
		{Price: 105, Volume: 2000, Timestamp: now.Add(-6 * time.Hour)},
		{Price: 104, Volume: 1800, Timestamp: now.Add(-5 * time.Hour)},
		{Price: 106, Volume: 2200, Timestamp: now.Add(-4 * time.Hour)},
		{Price: 108, Volume: 2500, Timestamp: now.Add(-3 * time.Hour)},
		{Price: 107, Volume: 2300, Timestamp: now.Add(-2 * time.Hour)},
		{Price: 110, Volume: 3000, Timestamp: now.Add(-1 * time.Hour)},
	}

	// Add test data to analyzer
	for _, data := range testData {
		analyzer.AddMarketData(data)
	}

	t.Run("Market Analysis Flow", func(t *testing.T) {
		// Test trend analysis
		trend, strength := analyzer.AnalyzeTrend("1h")
		assert.Equal(t, "bullish", trend)
		assert.Greater(t, strength, 0.0)

		// Test volatility analysis
		volatility := analyzer.AnalyzeVolatility("1h")
		assert.Greater(t, volatility, 0.0)
		assert.Less(t, volatility, 1.0)

		// Test support and resistance levels
		support, resistance := analyzer.calculateSupportResistance()
		assert.Less(t, support, resistance)
		assert.Greater(t, support, 95.0)
		assert.Less(t, resistance, 115.0)

		// Test price prediction
		predictedPrice, confidence := analyzer.predictNextPrice()
		assert.Greater(t, predictedPrice, 0.0)
		assert.Greater(t, confidence, 0.0)
		assert.Less(t, confidence, 1.0)

		// Test position size recommendation
		recommendedSize := analyzer.calculateRecommendedSize(volatility)
		assert.Greater(t, recommendedSize, 0.0)
	})

	t.Run("Trading Performance Analysis", func(t *testing.T) {
		trades := []Trade{
			{EntryPrice: 100, ExitPrice: 105, Size: 1.0, Profit: 5.0, Timestamp: now.Add(-5 * time.Hour)},
			{EntryPrice: 105, ExitPrice: 103, Size: 1.0, Profit: -2.0, Timestamp: now.Add(-4 * time.Hour)},
			{EntryPrice: 103, ExitPrice: 108, Size: 1.0, Profit: 5.0, Timestamp: now.Add(-3 * time.Hour)},
			{EntryPrice: 108, ExitPrice: 110, Size: 1.0, Profit: 2.0, Timestamp: now.Add(-2 * time.Hour)},
		}

		metrics := analyzer.CalculateMetrics(trades)
		
		// Test trading metrics
		assert.Equal(t, 4, metrics.TotalTrades)
		assert.Equal(t, 3, metrics.ProfitableTrades)
		assert.InDelta(t, 0.75, metrics.WinRate, 0.01)
		assert.Greater(t, metrics.ProfitFactor, 1.0)
		assert.InDelta(t, 10.0, metrics.TotalProfit, 0.01)
		assert.Greater(t, metrics.SharpeRatio, 0.0)
		assert.Less(t, metrics.MaxDrawdown, 1.0)
	})

	t.Run("Report Generation", func(t *testing.T) {
		report, err := analyzer.GenerateReport(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, report)

		// Validate report contents
		assert.NotEmpty(t, report.Trend)
		assert.Greater(t, report.TrendStrength, 0.0)
		assert.Greater(t, report.Volatility, 0.0)
		assert.Greater(t, report.Support, 0.0)
		assert.Greater(t, report.Resistance, report.Support)
		assert.Greater(t, report.PredictedPrice, 0.0)
		assert.Greater(t, report.Confidence, 0.0)
		assert.Greater(t, report.RecommendedSize, 0.0)
		assert.NotZero(t, report.Timestamp)
	})

	t.Run("Edge Cases", func(t *testing.T) {
		emptyAnalyzer := NewMarketAnalyzer()

		// Test with no data
		trend, strength := emptyAnalyzer.AnalyzeTrend("1h")
		assert.Equal(t, "neutral", trend)
		assert.Equal(t, 0.0, strength)

		volatility := emptyAnalyzer.AnalyzeVolatility("1h")
		assert.Equal(t, 0.0, volatility)

		support, resistance := emptyAnalyzer.calculateSupportResistance()
		assert.Equal(t, 0.0, support)
		assert.Equal(t, 0.0, resistance)

		predictedPrice, confidence := emptyAnalyzer.predictNextPrice()
		assert.Equal(t, 0.0, predictedPrice)
		assert.Equal(t, 0.0, confidence)

		// Test with single data point
		emptyAnalyzer.AddMarketData(MarketData{Price: 100, Volume: 1000, Timestamp: now})
		
		trend, strength = emptyAnalyzer.AnalyzeTrend("1h")
		assert.Equal(t, "neutral", trend)
		assert.Equal(t, 0.0, strength)

		volatility = emptyAnalyzer.AnalyzeVolatility("1h")
		assert.Equal(t, 0.0, volatility)
	})

	t.Run("Moving Average Calculation", func(t *testing.T) {
		// Test different MA periods
		periods := []int{5, 10, 20}
		for _, period := range periods {
			ma := analyzer.calculateMA(period)
			assert.NotNil(t, ma)
			if len(testData) >= period {
				assert.Equal(t, len(testData)-period+1, len(ma))
			} else {
				assert.Nil(t, ma)
			}
		}
	})
}) 