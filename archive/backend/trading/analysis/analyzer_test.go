package analysis

import (
	"context"
	"testing"
	"time"

	"github.com/leonzhao/trading-system/backend/models"

	"github.com/stretchr/testify/assert"
)

func TestNewMarketAnalyzer(t *testing.T) {
	analyzer := NewMarketAnalyzer()
	assert.NotNil(t, analyzer)
	assert.Empty(t, analyzer.historicalData)
}

func TestAddMarketData(t *testing.T) {
	analyzer := NewMarketAnalyzer()
	data := models.MarketData{
		ClosePrice:  100,
		OpenPrice:   95,
		HighPrice:   105,
		LowPrice:    90,
		Volume:      1000,
		Timestamp:   time.Now(),
	}

	analyzer.AddMarketData(data)
	
	analyzer.mu.Lock()
	defer analyzer.mu.Unlock()
	assert.Len(t, analyzer.marketData, 1, "Should have 1 market data entry")
	assert.Equal(t, data.ClosePrice, analyzer.priceHistory[0], "Price history should match")
}

func TestCalculateMetrics(t *testing.T) {
	analyzer := NewMarketAnalyzer()
	trades := []models.Trade{
		{
			Price:     100,
			Amount:    1,
			Value:     110,
			Fee:       0.1,
			Status:    models.TradeExecuted,
			Timestamp: time.Now(),
		},
		{
			Price:     110,
			Amount:    1,
			Value:     105,
			Fee:       0.1,
			Status:    models.TradeExecuted,
			Timestamp: time.Now(),
		},
		{
			Price:     105,
			Amount:    1,
			Value:     115,
			Fee:       0.1,
			Status:    models.TradeExecuted,
			Timestamp: time.Now(),
		},
	}

	metrics := analyzer.CalculateMetrics(trades)
	// Expected P&L: [9.9, -5.1, 9.8] 
	// Drawdowns: [0, 5.1, 0] -> Max 5.1
	assert.InDelta(t, 0.667, metrics["win_rate"], 0.001)
	assert.InDelta(t, 5.1, metrics["max_drawdown"], 0.1)
	assert.InDelta(t, 1.8, metrics["sharpe"], 0.001)
}

func TestAnalyzeTrend(t *testing.T) {
	analyzer := NewMarketAnalyzer()
	now := time.Now()

	// Add test data for uptrend
	for i := 0; i < 30; i++ {
		analyzer.AddMarketData(models.MarketData{
			ClosePrice: 100 + float64(i),
			Timestamp:  now.Add(time.Duration(i) * time.Hour),
		})
	}

	trend, strength := analyzer.AnalyzeTrend("1d")
	assert.Equal(t, "bullish", trend)
	assert.Greater(t, strength, 0.0)

	// Test downtrend
	analyzer = NewMarketAnalyzer()
	for i := 0; i < 30; i++ {
		analyzer.AddMarketData(models.MarketData{
			ClosePrice: 100 - float64(i),
			Timestamp:  now.Add(time.Duration(i) * time.Hour),
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
			prices:        []float64{100.0, 100.1, 99.9, 100.05, 100.0},
			expectHighVol: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer = NewMarketAnalyzer()
			for i, price := range tt.prices {
				analyzer.AddMarketData(models.MarketData{
					ClosePrice: price,
					Timestamp:  now.Add(time.Duration(i) * time.Hour),
				})
			}

			volatility, _ := analyzer.AnalyzeVolatility("1d")
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

	// Add market data
	for i := 0; i < 30; i++ {
		analyzer.AddMarketData(models.MarketData{
			ClosePrice: 100 + float64(i%5),
			Timestamp:  now.Add(time.Duration(i) * time.Hour),
		})
	}
	
	// Add trades to generate meaningful metrics
	analyzer.tradeHistory = []models.Trade{
		{Price: 100, Amount: 1, Value: 110, Status: models.TradeExecuted},
		{Price: 110, Amount: 1, Value: 105, Status: models.TradeExecuted},
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
		analyzer.AddMarketData(models.MarketData{
			ClosePrice: price,
			Timestamp:  now.Add(time.Duration(i) * time.Hour),
		})
	}

	// Test 3-period MA
	ma := analyzer.calculateMA(3)
	// Last 3 prices of [10,20,30,40,50] are 30,40,50 = average 40
	assert.InDelta(t, 40.0, ma, 0.001)
}

func TestCalculateMaxDrawdown(t *testing.T) {
	analyzer := NewMarketAnalyzer()
	trades := []models.Trade{
		{Value: 10},
		{Value: -5},
		{Value: -3},
		{Value: 7},
		{Value: -6},
	}

	maxDrawdown := analyzer.calculateMaxDrawdown(trades)
	// Trades: [10, -5, -3, 7, -6] => P&L: [10,5,2,9,3]
	// Drawdowns: [0,5,8,1,7] => Max 8
	assert.InDelta(t, 8.0, maxDrawdown, 0.001)
}

func TestCalculateSharpeRatio(t *testing.T) {
	analyzer := NewMarketAnalyzer()
	
	// Test with empty analyzer state
	sharpeRatio := analyzer.calculateSharpeRatio()
	assert.Greater(t, sharpeRatio, 0.0)
}

func TestCalculateSupportResistance(t *testing.T) {
	analyzer := NewMarketAnalyzer()
	now := time.Now()

	prices := []float64{100, 105, 95, 110, 98, 108, 96, 112, 97, 109}
	for i, price := range prices {
		analyzer.AddMarketData(models.MarketData{
			ClosePrice: price,
			Timestamp:  now.Add(time.Duration(i) * time.Hour),
		})
	}

	support, resistance := analyzer.calculateSupportResistance(14)
	// Test data has prices from 95 to 112
	assert.InDelta(t, 95.0, support, 1.0)
	assert.InDelta(t, 112.0, resistance, 1.0)
}
