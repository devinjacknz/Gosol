package analysis

import (
	"math"
)

// RSI calculates the Relative Strength Index
func (a *MarketAnalyzer) RSI(period int) float64 {
	if len(a.historicalData) < period+1 {
		return 0
	}

	var gains, losses float64
	for i := 1; i <= period; i++ {
		change := a.historicalData[len(a.historicalData)-i].Price - a.historicalData[len(a.historicalData)-i-1].Price
		if change >= 0 {
			gains += change
		} else {
			losses -= change
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	if avgLoss == 0 {
		return 100
	}
	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs))
}

// MACD calculates Moving Average Convergence Divergence
// Returns: MACD line, Signal line, Histogram
func (a *MarketAnalyzer) MACD() ([]float64, []float64, []float64) {
	shortPeriod := 12
	longPeriod := 26
	signalPeriod := 9

	if len(a.historicalData) < longPeriod {
		return nil, nil, nil
	}

	// Calculate EMAs
	shortEMA := a.calculateEMA(shortPeriod)
	longEMA := a.calculateEMA(longPeriod)

	// Calculate MACD line
	macdLine := make([]float64, len(shortEMA))
	for i := 0; i < len(shortEMA); i++ {
		if i < len(longEMA) {
			macdLine[i] = shortEMA[i] - longEMA[i]
		}
	}

	// Calculate Signal line (9-day EMA of MACD line)
	signalLine := a.calculateEMAFromValues(macdLine, signalPeriod)

	// Calculate Histogram
	histogram := make([]float64, len(signalLine))
	for i := 0; i < len(signalLine); i++ {
		histogram[i] = macdLine[i+len(macdLine)-len(signalLine)] - signalLine[i]
	}

	return macdLine, signalLine, histogram
}

// BollingerBands calculates Bollinger Bands
// Returns: Middle band, Upper band, Lower band
func (a *MarketAnalyzer) BollingerBands(period int, stdDevMultiplier float64) ([]float64, []float64, []float64) {
	if len(a.historicalData) < period {
		return nil, nil, nil
	}

	// Calculate middle band (SMA)
	middleBand := a.calculateMA(period)

	// Calculate standard deviation
	upperBand := make([]float64, len(middleBand))
	lowerBand := make([]float64, len(middleBand))

	for i := 0; i < len(middleBand); i++ {
		var sumSquaredDiff float64
		for j := 0; j < period; j++ {
			diff := a.historicalData[i+j].Price - middleBand[i]
			sumSquaredDiff += diff * diff
		}
		stdDev := math.Sqrt(sumSquaredDiff / float64(period))

		upperBand[i] = middleBand[i] + stdDevMultiplier*stdDev
		lowerBand[i] = middleBand[i] - stdDevMultiplier*stdDev
	}

	return middleBand, upperBand, lowerBand
}

// VolumeAnalysis performs volume analysis
// Returns: Volume MA, Volume trend ("increasing", "decreasing", "stable")
func (a *MarketAnalyzer) VolumeAnalysis(period int) ([]float64, string) {
	if len(a.historicalData) < period {
		return nil, "insufficient data"
	}

	// Calculate volume moving average
	volumeMA := make([]float64, len(a.historicalData)-period+1)
	for i := 0; i <= len(a.historicalData)-period; i++ {
		sum := 0.0
		for j := 0; j < period; j++ {
			sum += a.historicalData[i+j].Volume
		}
		volumeMA[i] = sum / float64(period)
	}

	// Analyze volume trend
	recentVol := volumeMA[len(volumeMA)-1]
	prevVol := volumeMA[len(volumeMA)-2]
	volChange := (recentVol - prevVol) / prevVol

	if volChange > 0.1 {
		return volumeMA, "increasing"
	} else if volChange < -0.1 {
		return volumeMA, "decreasing"
	}
	return volumeMA, "stable"
}

// Helper function to calculate EMA
func (a *MarketAnalyzer) calculateEMA(period int) []float64 {
	if len(a.historicalData) < period {
		return nil
	}

	multiplier := 2.0 / float64(period+1)
	ema := make([]float64, len(a.historicalData)-period+1)

	// Initialize EMA with SMA
	var sum float64
	for i := 0; i < period; i++ {
		sum += a.historicalData[i].Price
	}
	ema[0] = sum / float64(period)

	// Calculate EMA
	for i := 1; i < len(ema); i++ {
		price := a.historicalData[i+period-1].Price
		ema[i] = (price-ema[i-1])*multiplier + ema[i-1]
	}

	return ema
}

// Helper function to calculate EMA from a series of values
func (a *MarketAnalyzer) calculateEMAFromValues(values []float64, period int) []float64 {
	if len(values) < period {
		return nil
	}

	multiplier := 2.0 / float64(period+1)
	ema := make([]float64, len(values)-period+1)

	// Initialize EMA with SMA
	var sum float64
	for i := 0; i < period; i++ {
		sum += values[i]
	}
	ema[0] = sum / float64(period)

	// Calculate EMA
	for i := 1; i < len(ema); i++ {
		value := values[i+period-1]
		ema[i] = (value-ema[i-1])*multiplier + ema[i-1]
	}

	return ema
}
