package analysis

import (
	"math"
)

// CalculateRSI calculates the Relative Strength Index
func CalculateRSI(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 50 // Default to neutral if not enough data
	}

	var gains, losses float64
	for i := 1; i <= period; i++ {
		change := prices[i] - prices[i-1]
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

// CalculateEMA calculates the Exponential Moving Average
func CalculateEMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return prices[len(prices)-1] // Return last price if not enough data
	}

	multiplier := 2.0 / float64(period+1)
	ema := prices[0]

	for i := 1; i < len(prices); i++ {
		ema = (prices[i]-ema)*multiplier + ema
	}

	return ema
}

// CalculateVolatility calculates price volatility
func CalculateVolatility(prices []float64) float64 {
	if len(prices) < 2 {
		return 0
	}

	var sum, sumSquared float64
	for _, price := range prices {
		sum += price
		sumSquared += price * price
	}

	mean := sum / float64(len(prices))
	variance := (sumSquared/float64(len(prices)) - mean*mean)
	return math.Sqrt(variance) // Return standard deviation
}

// CalculateTrend determines price trend
func CalculateTrend(prices []float64, period int) string {
	if len(prices) < period {
		return "neutral"
	}

	shortEMA := CalculateEMA(prices, period/2)
	longEMA := CalculateEMA(prices, period)

	if shortEMA > longEMA {
		return "bullish"
	} else if shortEMA < longEMA {
		return "bearish"
	}
	return "neutral"
}

// CalculateMACD calculates Moving Average Convergence Divergence
func CalculateMACD(prices []float64) (macd, signal, histogram float64) {
	if len(prices) < 26 {
		return 0, 0, 0
	}

	ema12 := CalculateEMA(prices, 12)
	ema26 := CalculateEMA(prices, 26)
	macd = ema12 - ema26
	signal = CalculateEMA([]float64{macd}, 9)
	histogram = macd - signal

	return macd, signal, histogram
}

// CalculateBollingerBands calculates Bollinger Bands
func CalculateBollingerBands(prices []float64, period int, stdDev float64) (middle, upper, lower float64) {
	if len(prices) < period {
		return prices[len(prices)-1], prices[len(prices)-1], prices[len(prices)-1]
	}

	// Calculate SMA for middle band
	var sum float64
	for _, price := range prices[len(prices)-period:] {
		sum += price
	}
	middle = sum / float64(period)

	// Calculate standard deviation
	var variance float64
	for _, price := range prices[len(prices)-period:] {
		diff := price - middle
		variance += diff * diff
	}
	variance /= float64(period)
	sd := math.Sqrt(variance)

	upper = middle + (sd * stdDev)
	lower = middle - (sd * stdDev)

	return middle, upper, lower
}

// CalculateATR calculates Average True Range
func CalculateATR(highs, lows, closes []float64, period int) float64 {
	if len(highs) < period+1 || len(lows) < period+1 || len(closes) < period+1 {
		return 0
	}

	var tr []float64
	for i := 1; i < len(closes); i++ {
		// True Range is the greatest of:
		// 1. Current High - Current Low
		// 2. |Current High - Previous Close|
		// 3. |Current Low - Previous Close|
		high := highs[i]
		low := lows[i]
		prevClose := closes[i-1]

		tr = append(tr, math.Max(
			high-low,
			math.Max(
				math.Abs(high-prevClose),
				math.Abs(low-prevClose),
			),
		))
	}

	// Calculate ATR as EMA of TR
	return CalculateEMA(tr, period)
}

// CalculateStochRSI calculates Stochastic RSI
func CalculateStochRSI(prices []float64, period int) float64 {
	if len(prices) < period*2 {
		return 50
	}

	var rsiValues []float64
	for i := period; i < len(prices); i++ {
		rsi := CalculateRSI(prices[i-period:i+1], period)
		rsiValues = append(rsiValues, rsi)
	}

	if len(rsiValues) < period {
		return 50
	}

	// Get the highest and lowest RSI values over the period
	var highestRSI, lowestRSI float64
	highestRSI = rsiValues[0]
	lowestRSI = rsiValues[0]

	for _, rsi := range rsiValues[len(rsiValues)-period:] {
		if rsi > highestRSI {
			highestRSI = rsi
		}
		if rsi < lowestRSI {
			lowestRSI = rsi
		}
	}

	// Calculate StochRSI
	if highestRSI == lowestRSI {
		return 50
	}

	currentRSI := rsiValues[len(rsiValues)-1]
	stochRSI := (currentRSI - lowestRSI) / (highestRSI - lowestRSI) * 100

	return stochRSI
}
