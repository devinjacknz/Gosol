package analysis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/leonzhao/trading-system/backend/models"
)

func TestMarketAnalyzerIntegration(t *testing.T) {
	ctx := context.Background()
	analyzer := NewMarketAnalyzer()

	// Setup test data
	now := time.Now()
		testData := []models.MarketData{
			{ClosePrice: 110, Volume: 2200, Timestamp: now.Add(-55 * time.Minute)},
			{ClosePrice: 112, Volume: 2800, Timestamp: now.Add(-50 * time.Minute)},
			{ClosePrice: 115, Volume: 3000, Timestamp: now.Add(-45 * time.Minute)},
			{ClosePrice: 114, Volume: 3200, Timestamp: now.Add(-40 * time.Minute)},
			{ClosePrice: 116, Volume: 3500, Timestamp: now.Add(-35 * time.Minute)},
			{ClosePrice: 117, Volume: 3800, Timestamp: now.Add(-30 * time.Minute)},
			{ClosePrice: 115, Volume: 4000, Timestamp: now.Add(-25 * time.Minute)},
			{ClosePrice: 118, Volume: 4200, Timestamp: now.Add(-20 * time.Minute)},
			{ClosePrice: 119, Volume: 4500, Timestamp: now.Add(-15 * time.Minute)},
			{ClosePrice: 120, Volume: 4800, Timestamp: now.Add(-10 * time.Minute)},
			{ClosePrice: 118, Volume: 5000, Timestamp: now.Add(-5 * time.Minute)},
			{ClosePrice: 119, Volume: 5200, Timestamp: now},
	}

	// Add test data to analyzer
	for _, data := range testData {
		analyzer.AddMarketData(data)
	}

	// Add test trades for metrics calculation
	analyzer.tradeHistory = []models.Trade{
		{Price: 100, Amount: 1, Value: 100, Status: models.TradeExecuted},
		{Price: 110, Amount: 1, Value: 110, Status: models.TradeExecuted},
		{Price: 105, Amount: 1, Value: 105, Status: models.TradeExecuted},
	}

	t.Run("Market Analysis Flow", func(t *testing.T) {
		// Test trend analysis
		trend, strength := analyzer.AnalyzeTrend("1h")
		assert.Contains(t, []string{"bullish", "consolidation", "neutral"}, trend)
		assert.GreaterOrEqual(t, strength, 0.0)

		// Test volatility analysis
		volatility, err := analyzer.AnalyzeVolatility("1h")
		assert.NoError(t, err)
		assert.Greater(t, volatility, 0.0)

		// Test support and resistance levels
		support, resistance := analyzer.calculateSupportResistance(14)
		assert.Greater(t, support, 110.0)
		assert.Greater(t, resistance, support)
		assert.Less(t, resistance, 200.0)

		// Test price prediction
		predictedPrice, confidence := analyzer.predictNextPrice()
		assert.Greater(t, predictedPrice, 0.0)
		assert.Greater(t, confidence, 0.0)
	})

	t.Run("Trading Performance Analysis", func(t *testing.T) {
		trades := []models.Trade{
			{Price: 100, Amount: 1.0, Value: 100.0, Timestamp: now.Add(-5 * time.Hour)},
			{Price: 105, Amount: 1.0, Value: 105.0, Timestamp: now.Add(-4 * time.Hour)},
			{Price: 103, Amount: 1.0, Value: 103.0, Timestamp: now.Add(-3 * time.Hour)},
			{Price: 108, Amount: 1.0, Value: 108.0, Timestamp: now.Add(-2 * time.Hour)},
		}

		metrics := analyzer.CalculateMetrics(trades)
		assert.InDelta(t, 0.33, metrics["win_rate"], 0.1)  // 1 win out of 3 trades
		assert.InDelta(t, 5.0, metrics["max_drawdown"], 2.0)
		assert.InDelta(t, 0.5, metrics["sharpe"], 0.3)
	})

	t.Run("Report Generation", func(t *testing.T) {
		report, err := analyzer.GenerateReport(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, report)

		assert.NotEmpty(t, report.Trend)
		assert.Greater(t, report.TrendStrength, 0.0)
		assert.Greater(t, report.Volatility, 0.0)
		assert.Greater(t, report.Support, 0.0)
		assert.Greater(t, report.Resistance, report.Support)
	})

	t.Run("Edge Cases", func(t *testing.T) {
		emptyAnalyzer := NewMarketAnalyzer()

		// Test with no data
		trend, strength := emptyAnalyzer.AnalyzeTrend("1h")
		assert.Equal(t, "neutral", trend)
		assert.Equal(t, 0.0, strength)

		_, err := emptyAnalyzer.AnalyzeVolatility("1h")
		assert.Error(t, err)

		support, resistance := emptyAnalyzer.calculateSupportResistance(14)
		assert.Equal(t, 0.0, support)
		assert.Equal(t, 0.0, resistance)
	})
}
