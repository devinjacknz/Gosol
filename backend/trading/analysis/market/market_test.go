package market

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMarketAnalyzer(t *testing.T) {
	analyzer := NewMarketAnalyzer()
	ctx := context.Background()

	t.Run("Analyze with sufficient data", func(t *testing.T) {
		data := MarketData{
			Prices:    generateTestPrices(),
			Volumes:   generateTestVolumes(),
			OrderBook: generateTestOrderBook(),
			Timestamp: time.Now(),
		}

		analysis, err := analyzer.Analyze(ctx, "BTC/USD", data)
		assert.NoError(t, err)
		assert.NotNil(t, analysis)

		// Verify price analysis
		assert.NotZero(t, analysis.PriceAnalysis.Volatility)
		assert.NotZero(t, analysis.PriceAnalysis.PriceRange.High)
		assert.NotZero(t, analysis.PriceAnalysis.MovingAverages.SMA20)

		// Verify volume analysis
		assert.NotZero(t, analysis.VolumeAnalysis.VolumeRatio)
		assert.NotZero(t, analysis.VolumeAnalysis.AverageVolume)

		// Verify trend analysis
		assert.NotZero(t, analysis.TrendAnalysis.TrendStrength)
		assert.NotZero(t, analysis.TrendAnalysis.RSI)

		// Verify liquidity analysis
		assert.NotZero(t, analysis.LiquidityAnalysis.MarketDepth)
		assert.NotZero(t, analysis.LiquidityAnalysis.BidAskSpread)
	})

	t.Run("Analyze with insufficient data", func(t *testing.T) {
		data := MarketData{
			Prices:    []float64{100.0},
			Volumes:   []float64{1000.0},
			OrderBook: OrderBook{},
			Timestamp: time.Now(),
		}

		analysis, err := analyzer.Analyze(ctx, "BTC/USD", data)
		assert.Error(t, err)
		assert.Nil(t, analysis)
		assert.Equal(t, ErrInsufficientData, err)
	})

	t.Run("Analyze with invalid order book", func(t *testing.T) {
		data := MarketData{
			Prices:    generateTestPrices(),
			Volumes:   generateTestVolumes(),
			OrderBook: OrderBook{},
			Timestamp: time.Now(),
		}

		analysis, err := analyzer.Analyze(ctx, "BTC/USD", data)
		assert.Error(t, err)
		assert.Nil(t, analysis)
		assert.Equal(t, ErrInvalidOrderBook, err)
	})
}

func generateTestPrices() []float64 {
	prices := make([]float64, 200)
	basePrice := 100.0
	for i := range prices {
		// Generate a simple trend with some noise
		trend := float64(i) * 0.1
		noise := float64(i%5) * 0.2
		prices[i] = basePrice + trend + noise
	}
	return prices
}

func generateTestVolumes() []float64 {
	volumes := make([]float64, 200)
	baseVolume := 1000.0
	for i := range volumes {
		// Generate volumes with some variation
		variation := float64(i%10) * 100
		volumes[i] = baseVolume + variation
	}
	return volumes
}

func generateTestOrderBook() OrderBook {
	bids := make([]OrderBookLevel, 10)
	asks := make([]OrderBookLevel, 10)

	basePrice := 100.0
	baseAmount := 1.0

	for i := range bids {
		bids[i] = OrderBookLevel{
			Price:  basePrice - float64(i)*0.1,
			Amount: baseAmount + float64(i)*0.1,
		}
		asks[i] = OrderBookLevel{
			Price:  basePrice + float64(i)*0.1,
			Amount: baseAmount + float64(i)*0.1,
		}
	}

	return OrderBook{
		Bids: bids,
		Asks: asks,
	}
}

func BenchmarkMarketAnalyzer(b *testing.B) {
	analyzer := NewMarketAnalyzer()
	ctx := context.Background()
	data := MarketData{
		Prices:    generateTestPrices(),
		Volumes:   generateTestVolumes(),
		OrderBook: generateTestOrderBook(),
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analysis, err := analyzer.Analyze(ctx, "BTC/USD", data)
		if err != nil {
			b.Fatal(err)
		}
		if analysis == nil {
			b.Fatal("analysis should not be nil")
		}
	}
}
