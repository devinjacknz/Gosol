package analysis

import (
	"fmt"
	"math"
	"github.com/leonzhao/trading-system/backend/models"
)

// SMA calculates Simple Moving Average
func SMA(prices []float64, period int) (float64, error) {
	if len(prices) < period {
		return 0, fmt.Errorf("insufficient data points (%d) for SMA period %d", len(prices), period)
	}
	
	var sum float64
	for _, price := range prices[len(prices)-period:] {
		sum += price
	}
	return sum / float64(period), nil
}

// EMA calculates Exponential Moving Average
func EMA(prices []float64, period int) (float64, error) {
	if len(prices) < period {
		return 0, fmt.Errorf("insufficient data points (%d) for EMA period %d", len(prices), period)
	}
	
	// Calculate SMA as initial EMA value
	sma, err := SMA(prices[:period], period)
	if err != nil {
		return 0, err
	}

	multiplier := 2.0 / (float64(period) + 1)
	ema := sma

	// Calculate EMA for remaining values
	for _, price := range prices[period:] {
		ema = (price-ema)*multiplier + ema
	}
	return ema, nil
}

// RSI calculates Relative Strength Index
func RSI(prices []float64, period int) (float64, error) {
	if len(prices) <= period {
		return 0, fmt.Errorf("insufficient data points (%d) for RSI period %d", len(prices), period)
	}
	if period <= 0 {
		return 0, fmt.Errorf("invalid period: %d", period)
	}

	var gains, losses float64
	for i := 1; i <= period; i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains += change
		} else {
			losses -= change
		}
	}

	// Handle edge cases
	if gains == 0 {
		return 0, nil
	}
	if losses == 0 {
		return 100, nil
	}

	// Smoothing factor
	alpha := 1.0 / float64(period)
	
	// Calculate initial averages
	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	// Calculate remaining averages
	for i := period + 1; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			avgGain = (change * alpha) + (avgGain * (1 - alpha))
			avgLoss = avgLoss * (1 - alpha)
		} else {
			avgLoss = (-change * alpha) + (avgLoss * (1 - alpha))
			avgGain = avgGain * (1 - alpha)
		}
	}

	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs)), nil
}

// BollingerBands calculates Bollinger Bands
func BollingerBands(prices []float64, period int, stdDev float64) (models.BollingerBands, error) {
	if len(prices) < period {
		return models.BollingerBands{}, fmt.Errorf("insufficient data points (%d) for period %d", len(prices), period)
	}
	if period <= 0 {
		return models.BollingerBands{}, fmt.Errorf("invalid period: %d", period)
	}
	if stdDev <= 0 {
		return models.BollingerBands{}, fmt.Errorf("invalid standard deviation: %f", stdDev)
	}

	// Calculate SMA as middle band
	middle, err := SMA(prices[len(prices)-period:], period)
	if err != nil {
		return models.BollingerBands{}, err
	}

	// Calculate standard deviation
	var sum float64
	for _, price := range prices[len(prices)-period:] {
		sum += math.Pow(price - middle, 2)
	}
	std := math.Sqrt(sum / float64(period))

	return models.BollingerBands{
		Middle: middle,
		Upper:  middle + stdDev*std,
		Lower:  middle - stdDev*std,
	}, nil
}

// MACD calculates Moving Average Convergence Divergence
func MACD(prices []float64, fast, slow, signal int) (float64, float64, error) {
	if len(prices) < slow {
		return 0, 0, fmt.Errorf("insufficient data points (%d) for MACD slow period %d", len(prices), slow)
	}

	fastEMA, err := EMA(prices, fast)
	if err != nil {
		return 0, 0, err
	}

	slowEMA, err := EMA(prices, slow)
	if err != nil {
		return 0, 0, err
	}

	macdLine := fastEMA - slowEMA
	
	// Get prices for signal line calculation
	if len(prices) < signal {
		return 0, 0, fmt.Errorf("insufficient data points (%d) for MACD signal period %d", len(prices), signal)
	}
	signalPrices := prices[len(prices)-signal:]
	
	signalEMA, err := EMA(signalPrices, signal)
	if err != nil {
		return 0, 0, err
	}

	return macdLine, macdLine - signalEMA, nil
}
