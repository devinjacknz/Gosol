package market

import (
	"context"
	"time"
)

// PricePoint represents a price data point
type PricePoint struct {
	Price     float64   `json:"price"`
	Volume    float64   `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

// MarketSignal represents a market signal
type MarketSignal struct {
	Type      string    `json:"type"` // "buy", "sell", "hold"
	Strength  float64   `json:"strength"` // 0-1
	Price     float64   `json:"price"`
	Volume    float64   `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

// Analyzer defines market analysis operations
type Analyzer interface {
	// Analyze analyzes market data and returns signals
	Analyze(ctx context.Context, data []*PricePoint) (*MarketSignal, error)

	// CalculateVolatility calculates market volatility
	CalculateVolatility(data []*PricePoint) float64

	// CalculateTrend calculates market trend
	CalculateTrend(data []*PricePoint) float64
}

// MarketAnalyzer implements the Analyzer interface
type MarketAnalyzer struct {
	lookbackPeriod time.Duration
	minDataPoints  int
}

// NewMarketAnalyzer creates a new market analyzer
func NewMarketAnalyzer(lookbackPeriod time.Duration, minDataPoints int) *MarketAnalyzer {
	return &MarketAnalyzer{
		lookbackPeriod: lookbackPeriod,
		minDataPoints:  minDataPoints,
	}
}

// Analyze analyzes market data and returns signals
func (a *MarketAnalyzer) Analyze(ctx context.Context, data []*PricePoint) (*MarketSignal, error) {
	if len(data) < a.minDataPoints {
		return &MarketSignal{
			Type:      "hold",
			Strength:  0,
			Price:     data[len(data)-1].Price,
			Volume:    data[len(data)-1].Volume,
			Timestamp: time.Now(),
		}, nil
	}

	volatility := a.CalculateVolatility(data)
	trend := a.CalculateTrend(data)

	signal := &MarketSignal{
		Price:     data[len(data)-1].Price,
		Volume:    data[len(data)-1].Volume,
		Timestamp: time.Now(),
	}

	// Simple trend following strategy
	if trend > 0.7 && volatility < 0.3 {
		signal.Type = "buy"
		signal.Strength = trend * (1 - volatility)
	} else if trend < -0.7 && volatility < 0.3 {
		signal.Type = "sell"
		signal.Strength = -trend * (1 - volatility)
	} else {
		signal.Type = "hold"
		signal.Strength = 0
	}

	return signal, nil
}

// CalculateVolatility calculates market volatility
func (a *MarketAnalyzer) CalculateVolatility(data []*PricePoint) float64 {
	if len(data) < 2 {
		return 0
	}

	var sumReturns, sumSquaredReturns float64
	n := float64(len(data) - 1)

	for i := 1; i < len(data); i++ {
		ret := (data[i].Price - data[i-1].Price) / data[i-1].Price
		sumReturns += ret
		sumSquaredReturns += ret * ret
	}

	meanReturn := sumReturns / n
	variance := (sumSquaredReturns/n - meanReturn*meanReturn)
	
	return variance
}

// CalculateTrend calculates market trend
func (a *MarketAnalyzer) CalculateTrend(data []*PricePoint) float64 {
	if len(data) < 2 {
		return 0
	}

	// Simple linear regression
	var sumX, sumY, sumXY, sumX2 float64
	n := float64(len(data))

	for i := 0; i < len(data); i++ {
		x := float64(i)
		y := data[i].Price

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	maxPrice := data[0].Price
	for _, p := range data {
		if p.Price > maxPrice {
			maxPrice = p.Price
		}
	}

	// Normalize slope to [-1, 1]
	normalizedSlope := slope / maxPrice
	if normalizedSlope > 1 {
		normalizedSlope = 1
	} else if normalizedSlope < -1 {
		normalizedSlope = -1
	}

	return normalizedSlope
}
