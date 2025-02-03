package market

import (
	"context"
	"time"

"github.com/devinjacknz/godydxhyber/backend/trading/analysis/monitoring"
)

// MarketAnalyzer provides market analysis functionality
type MarketAnalyzer struct {
	priceAnalyzer     *PriceAnalyzer
	volumeAnalyzer    *VolumeAnalyzer
	trendAnalyzer     *TrendAnalyzer
	liquidityAnalyzer *LiquidityAnalyzer
}

// NewMarketAnalyzer creates a new market analyzer
func NewMarketAnalyzer() *MarketAnalyzer {
	return &MarketAnalyzer{
		priceAnalyzer:     NewPriceAnalyzer(),
		volumeAnalyzer:    NewVolumeAnalyzer(),
		trendAnalyzer:     NewTrendAnalyzer(),
		liquidityAnalyzer: NewLiquidityAnalyzer(),
	}
}

// Analysis contains market analysis results
type Analysis struct {
	PriceAnalysis     PriceAnalysis
	VolumeAnalysis    VolumeAnalysis
	TrendAnalysis     TrendAnalysis
	LiquidityAnalysis LiquidityAnalysis
	Timestamp         time.Time
}

// Analyze performs comprehensive market analysis
func (ma *MarketAnalyzer) Analyze(ctx context.Context, symbol string, data MarketData) (*Analysis, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		monitoring.RecordIndicatorCalculation("market_analysis", duration)
	}()

	priceAnalysis, err := ma.priceAnalyzer.Analyze(ctx, data)
	if err != nil {
		monitoring.RecordIndicatorError("price_analysis", err.Error())
		return nil, err
	}

	volumeAnalysis, err := ma.volumeAnalyzer.Analyze(ctx, data)
	if err != nil {
		monitoring.RecordIndicatorError("volume_analysis", err.Error())
		return nil, err
	}

	trendAnalysis, err := ma.trendAnalyzer.Analyze(ctx, data)
	if err != nil {
		monitoring.RecordIndicatorError("trend_analysis", err.Error())
		return nil, err
	}

	liquidityAnalysis, err := ma.liquidityAnalyzer.Analyze(ctx, data)
	if err != nil {
		monitoring.RecordIndicatorError("liquidity_analysis", err.Error())
		return nil, err
	}

	analysis := &Analysis{
		PriceAnalysis:     *priceAnalysis,
		VolumeAnalysis:    *volumeAnalysis,
		TrendAnalysis:     *trendAnalysis,
		LiquidityAnalysis: *liquidityAnalysis,
		Timestamp:         time.Now(),
	}

	// Record key metrics
	monitoring.RecordIndicatorValue("price_volatility", priceAnalysis.Volatility)
	monitoring.RecordIndicatorValue("volume_ratio", volumeAnalysis.VolumeRatio)
	monitoring.RecordIndicatorValue("trend_strength", trendAnalysis.TrendStrength)
	monitoring.RecordIndicatorValue("liquidity_depth", liquidityAnalysis.MarketDepth)

	return analysis, nil
}

// MarketData represents market data for analysis
type MarketData struct {
	Prices    []float64
	Volumes   []float64
	OrderBook OrderBook
	Timestamp time.Time
}

// OrderBook represents market order book data
type OrderBook struct {
	Bids []OrderBookLevel
	Asks []OrderBookLevel
}

// OrderBookLevel represents a level in the order book
type OrderBookLevel struct {
	Price  float64
	Amount float64
}
