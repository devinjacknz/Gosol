package market

import (
	"context"
	"math"
)

// TrendAnalyzer analyzes market trends
type TrendAnalyzer struct{}

// NewTrendAnalyzer creates a new trend analyzer
func NewTrendAnalyzer() *TrendAnalyzer {
	return &TrendAnalyzer{}
}

// TrendAnalysis contains trend analysis results
type TrendAnalysis struct {
	TrendDirection TrendDirection
	TrendStrength  float64
	Momentum       float64
	RSI            float64
	MACD           MACDData
	Patterns       []Pattern
}

// TrendDirection represents the direction of a trend
type TrendDirection int

const (
	TrendUp TrendDirection = iota + 1
	TrendDown
	TrendSideways
)

// MACDData contains MACD indicator data
type MACDData struct {
	MACD      float64
	Signal    float64
	Histogram float64
}

// Pattern represents a chart pattern
type Pattern struct {
	Type     PatternType
	StartIdx int
	EndIdx   int
	Strength float64
}

// PatternType represents different chart patterns
type PatternType int

const (
	DoubleTop PatternType = iota + 1
	DoubleBottom
	HeadAndShoulders
	InverseHeadAndShoulders
	Triangle
	Channel
)

// Analyze performs trend analysis
func (ta *TrendAnalyzer) Analyze(ctx context.Context, data MarketData) (*TrendAnalysis, error) {
	if len(data.Prices) < 50 {
		return nil, ErrInsufficientData
	}

	// Determine trend direction and strength
	direction, strength := analyzeTrendDirection(data.Prices)

	// Calculate momentum
	momentum := calculateMomentum(data.Prices)

	// Calculate RSI
	rsi := calculateRSI(data.Prices)

	// Calculate MACD
	macd := calculateMACD(data.Prices)

	// Identify patterns
	patterns := identifyPatterns(data.Prices)

	return &TrendAnalysis{
		TrendDirection: direction,
		TrendStrength:  strength,
		Momentum:       momentum,
		RSI:            rsi,
		MACD:           macd,
		Patterns:       patterns,
	}, nil
}

func analyzeTrendDirection(prices []float64) (TrendDirection, float64) {
	if len(prices) < 2 {
		return TrendSideways, 0
	}

	// Calculate linear regression
	n := float64(len(prices))
	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumXX := 0.0

	for i, price := range prices {
		x := float64(i)
		sumX += x
		sumY += price
		sumXY += x * price
		sumXX += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)
	strength := math.Abs(slope)

	if slope > 0.01 {
		return TrendUp, strength
	} else if slope < -0.01 {
		return TrendDown, strength
	}
	return TrendSideways, strength
}

func calculateMomentum(prices []float64) float64 {
	if len(prices) < 14 {
		return 0
	}

	return (prices[len(prices)-1] / prices[len(prices)-14]) * 100
}

func calculateRSI(prices []float64) float64 {
	if len(prices) < 14 {
		return 50
	}

	period := 14
	gains := 0.0
	losses := 0.0

	for i := len(prices) - period; i < len(prices)-1; i++ {
		change := prices[i+1] - prices[i]
		if change >= 0 {
			gains += change
		} else {
			losses -= change
		}
	}

	if losses == 0 {
		return 100
	}

	rs := gains / losses
	return 100 - (100 / (1 + rs))
}

func calculateMACD(prices []float64) MACDData {
	if len(prices) < 26 {
		return MACDData{}
	}

	// Calculate EMAs
	ema12 := calculateEMA(prices, 12)
	ema26 := calculateEMA(prices, 26)
	macd := ema12 - ema26

	// Calculate signal line (9-day EMA of MACD)
	signal := macd // Simplified for this example
	histogram := macd - signal

	return MACDData{
		MACD:      macd,
		Signal:    signal,
		Histogram: histogram,
	}
}

func identifyPatterns(prices []float64) []Pattern {
	patterns := make([]Pattern, 0)

	// This is a simplified pattern recognition implementation
	// In a real system, you would implement more sophisticated pattern recognition algorithms

	// Example: Look for potential double tops
	if len(prices) > 20 {
		peak1 := 0.0
		peak1Idx := 0
		peak2 := 0.0
		peak2Idx := 0

		for i := 1; i < len(prices)-1; i++ {
			if prices[i] > prices[i-1] && prices[i] > prices[i+1] {
				if peak1 == 0 {
					peak1 = prices[i]
					peak1Idx = i
				} else {
					peak2 = prices[i]
					peak2Idx = i
				}
			}
		}

		if math.Abs(peak1-peak2)/peak1 < 0.02 && peak2Idx-peak1Idx > 5 {
			patterns = append(patterns, Pattern{
				Type:     DoubleTop,
				StartIdx: peak1Idx,
				EndIdx:   peak2Idx,
				Strength: 0.8,
			})
		}
	}

	return patterns
}
