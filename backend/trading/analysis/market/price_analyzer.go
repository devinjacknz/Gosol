package market

import (
	"context"
	"math"
)

// PriceAnalyzer analyzes price data
type PriceAnalyzer struct{}

// NewPriceAnalyzer creates a new price analyzer
func NewPriceAnalyzer() *PriceAnalyzer {
	return &PriceAnalyzer{}
}

// PriceAnalysis contains price analysis results
type PriceAnalysis struct {
	Volatility     float64
	PriceRange     PriceRange
	MovingAverages MovingAverages
	Support        float64
	Resistance     float64
}

// PriceRange represents price range information
type PriceRange struct {
	High   float64
	Low    float64
	Open   float64
	Close  float64
	Change float64
}

// MovingAverages contains various moving averages
type MovingAverages struct {
	SMA20  float64
	EMA20  float64
	SMA50  float64
	EMA50  float64
	SMA200 float64
}

// Analyze performs price analysis
func (pa *PriceAnalyzer) Analyze(ctx context.Context, data MarketData) (*PriceAnalysis, error) {
	if len(data.Prices) < 200 {
		return nil, ErrInsufficientData
	}

	// Calculate price range
	priceRange := calculatePriceRange(data.Prices)

	// Calculate volatility (standard deviation)
	volatility := calculateVolatility(data.Prices)

	// Calculate moving averages
	ma := calculateMovingAverages(data.Prices)

	// Calculate support and resistance
	support, resistance := calculateSupportResistance(data.Prices, data.OrderBook)

	return &PriceAnalysis{
		Volatility:     volatility,
		PriceRange:     priceRange,
		MovingAverages: ma,
		Support:        support,
		Resistance:     resistance,
	}, nil
}

func calculatePriceRange(prices []float64) PriceRange {
	if len(prices) == 0 {
		return PriceRange{}
	}

	high := prices[0]
	low := prices[0]
	for _, price := range prices {
		if price > high {
			high = price
		}
		if price < low {
			low = price
		}
	}

	open := prices[0]
	close := prices[len(prices)-1]
	change := (close - open) / open * 100

	return PriceRange{
		High:   high,
		Low:    low,
		Open:   open,
		Close:  close,
		Change: change,
	}
}

func calculateVolatility(prices []float64) float64 {
	if len(prices) < 2 {
		return 0
	}

	// Calculate mean
	mean := 0.0
	for _, price := range prices {
		mean += price
	}
	mean /= float64(len(prices))

	// Calculate variance
	variance := 0.0
	for _, price := range prices {
		diff := price - mean
		variance += diff * diff
	}
	variance /= float64(len(prices) - 1)

	return math.Sqrt(variance)
}

func calculateMovingAverages(prices []float64) MovingAverages {
	return MovingAverages{
		SMA20:  calculateSMA(prices, 20),
		EMA20:  calculateEMA(prices, 20),
		SMA50:  calculateSMA(prices, 50),
		EMA50:  calculateEMA(prices, 50),
		SMA200: calculateSMA(prices, 200),
	}
}

func calculateSMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	sum := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		sum += prices[i]
	}
	return sum / float64(period)
}

func calculateEMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	multiplier := 2.0 / float64(period+1)
	ema := prices[0]

	for i := 1; i < len(prices); i++ {
		ema = (prices[i]-ema)*multiplier + ema
	}

	return ema
}

func calculateSupportResistance(prices []float64, orderBook OrderBook) (support, resistance float64) {
	// Simple implementation using order book
	if len(orderBook.Bids) > 0 && len(orderBook.Asks) > 0 {
		// Find strongest bid level
		maxBidVolume := 0.0
		for _, bid := range orderBook.Bids {
			if bid.Amount > maxBidVolume {
				maxBidVolume = bid.Amount
				support = bid.Price
			}
		}

		// Find strongest ask level
		maxAskVolume := 0.0
		for _, ask := range orderBook.Asks {
			if ask.Amount > maxAskVolume {
				maxAskVolume = ask.Amount
				resistance = ask.Price
			}
		}
	}

	return support, resistance
}
