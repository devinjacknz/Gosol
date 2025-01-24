package analysis

import (
	"math"
	"solmeme-trader/models"
	"time"
)

// CalculateSMA calculates Simple Moving Average
func CalculateSMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return math.NaN()
	}

	var sum float64
	for _, price := range prices[len(prices)-period:] {
		sum += price
	}
	return sum / float64(period)
}

// CalculateEMA calculates Exponential Moving Average with improved validation
func CalculateEMA(prices []float64, period int) float64 {
	if len(prices) < period || period < 2 {
		return math.NaN()
	}

	// Calculate initial SMA using first 'period' prices
	sma := CalculateSMA(prices[:period], period)
	if math.IsNaN(sma) {
		return math.NaN()
	}

	multiplier := 2.0 / (float64(period) + 1.0)
	ema := sma

	// Calculate EMA using remaining prices
	for _, price := range prices[period:] {
		ema = (price-ema)*multiplier + ema
	}
	return ema
}

// CalculateRSI calculates Relative Strength Index using rolling window
// Implements Wilder's Smoothing method with precision improvements
func CalculateRSI(prices []float64, period int) float64 {
	if len(prices) < period*2 {
		return math.NaN()
	}

	// Use the last 2*period prices for calculation
	window := prices[len(prices)-period*2:]
	
	var avgGain, avgLoss float64
	for i := 1; i <= period; i++ {
		change := window[i] - window[i-1]
		if change > 0 {
			avgGain += change
		} else {
			avgLoss -= change
		}
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// Calculate remaining periods with smoothing
	for i := period + 1; i < len(window); i++ {
		change := window[i] - window[i-1]
		currentGain := 0.0
		currentLoss := 0.0
		
		if change > 0 {
			currentGain = change
		} else {
			currentLoss = -change
		}
		
		// Wilder's smoothing
		avgGain = (avgGain*(float64(period)-1) + currentGain) / float64(period)
		avgLoss = (avgLoss*(float64(period)-1) + currentLoss) / float64(period)
	}

	if math.Abs(avgLoss) < 1e-8 {
		return 100.0
	}
	rs := avgGain / avgLoss
	rsi := 100.0 - (100.0 / (1.0 + rs))
	
	// Clamp value between 0-100 to handle edge cases
	return math.Max(0.0, math.Min(100.0, rsi))
}

// CalculateBollingerBands calculates Bollinger Bands with improved precision and sample deviation
func CalculateBollingerBands(prices []float64, period int) models.BollingerBands {
	if len(prices) < period || period < 2 {
		return models.BollingerBands{}
	}

	// Calculate SMA using Kahan summation algorithm for improved accuracy
	var sma, c float64
	window := prices[len(prices)-period:]
	for _, price := range window {
		y := price - c
		t := sma + y
		c = (t - sma) - y
		sma = t
	}
	sma /= float64(period)

	// Calculate variance with Bessel's correction (sample deviation)
	var variance, c2 float64
	for _, price := range window {
		dev := price - sma
		y := dev*dev - c2
		t := variance + y
		c2 = (t - variance) - y
		variance = t
	}
	variance /= float64(period - 1)
	stdDev := math.Sqrt(variance)

	return models.BollingerBands{
		Upper:  sma + 2*stdDev,
		Middle: sma,
		Lower:  sma - 2*stdDev,
	}
}

// CalculateMACD calculates Moving Average Convergence Divergence with improved EMA handling
func CalculateMACD(prices []float64) models.MACD {
	const (
		fastPeriod   = 12
		slowPeriod   = 26
		signalPeriod = 9
	)
	
	// Validate input length for minimum calculation requirements
	if len(prices) < slowPeriod || len(prices) < fastPeriod {
		return models.MACD{}
	}

	// Calculate EMAs with error handling
	emaFast := CalculateEMA(prices, fastPeriod)
	emaSlow := CalculateEMA(prices, slowPeriod)
	if math.IsNaN(emaFast) || math.IsNaN(emaSlow) {
		return models.MACD{}
	}
	macdLine := emaFast - emaSlow

	// Calculate MACD values for signal line
	macdValues := make([]float64, 0, len(prices))
	for i := 0; i < len(prices); i++ {
		if i < slowPeriod-1 { // Wait until we have enough data for slow EMA
			continue
		}
		fast := CalculateEMA(prices[:i+1], fastPeriod)
		slow := CalculateEMA(prices[:i+1], slowPeriod)
		if math.IsNaN(fast) || math.IsNaN(slow) {
			continue
		}
		macdValues = append(macdValues, fast - slow)
	}
	
	// Calculate signal line with validated data
	if len(macdValues) < signalPeriod {
		return models.MACD{}
	}
	signalLine := CalculateEMA(macdValues, signalPeriod)
	if math.IsNaN(signalLine) {
		return models.MACD{}
	}

	// Normalize precision and calculate histogram
	roundedMACD := math.Round(macdLine*1e8)/1e8
	roundedSignal := math.Round(signalLine*1e8)/1e8
	return models.MACD{
		MACDLine:   roundedMACD,
		SignalLine: roundedSignal,
		Histogram:  math.Round((roundedMACD - roundedSignal)*1e8)/1e8,
	}
}

// GenerateSignals creates trading signals based on technical indicators
func GenerateSignals(indicators *models.TechnicalIndicators) []models.TradeSignal {
	var signals []models.TradeSignal
	now := time.Now()

	// Bullish signal when price crosses above upper Bollinger Band
	if indicators.Price > indicators.BollingerBands.Upper {
		signals = append(signals, models.TradeSignal{
			SignalType:  models.SignalTypeOverbought,
			Strength:    models.SignalStrengthStrong,
			Timestamp:   now,
			Description: "Price above upper Bollinger Band - Overbought condition",
		})
	}

	// Bearish signal when price crosses below lower Bollinger Band
	if indicators.Price < indicators.BollingerBands.Lower {
		signals = append(signals, models.TradeSignal{
			SignalType:  models.SignalTypeOversold,
			Strength:    models.SignalStrengthStrong,
			Timestamp:   now,
			Description: "Price below lower Bollinger Band - Oversold condition",
		})
	}

	// MACD crossover signals
	if indicators.MACD.Histogram > 0 && indicators.MACD.MACDLine > indicators.MACD.SignalLine {
		signals = append(signals, models.TradeSignal{
			SignalType:  models.SignalTypeBullish,
			Strength:    models.SignalStrengthMedium,
			Timestamp:   now,
			Description: "MACD bullish crossover detected",
		})
	} else if indicators.MACD.Histogram < 0 && indicators.MACD.MACDLine < indicators.MACD.SignalLine {
		signals = append(signals, models.TradeSignal{
			SignalType:  models.SignalTypeBearish,
			Strength:    models.SignalStrengthMedium,
			Timestamp:   now,
			Description: "MACD bearish crossover detected",
		})
	}

	// RSI based signals
	if indicators.RSI > 70 {
		signals = append(signals, models.TradeSignal{
			SignalType:  models.SignalTypeOverbought,
			Strength:    models.SignalStrengthWeak,
			Timestamp:   now,
			Description: "RSI above 70 - Overbought condition",
		})
	} else if indicators.RSI < 30 {
		signals = append(signals, models.TradeSignal{
			SignalType:  models.SignalTypeOversold,
			Strength:    models.SignalStrengthWeak,
			Timestamp:   now,
			Description: "RSI below 30 - Oversold condition",
		})
	}

	return signals
}
