package market

import (
	"context"
	"math"
)

// VolumeAnalyzer analyzes volume data
type VolumeAnalyzer struct{}

// NewVolumeAnalyzer creates a new volume analyzer
func NewVolumeAnalyzer() *VolumeAnalyzer {
	return &VolumeAnalyzer{}
}

// VolumeAnalysis contains volume analysis results
type VolumeAnalysis struct {
	VolumeRatio    float64
	AverageVolume  float64
	VolumeProfile  VolumeProfile
	Liquidity      float64
	VolumeChange   float64
	RelativeVolume float64
}

// VolumeProfile represents volume distribution at different price levels
type VolumeProfile struct {
	PriceLevels []float64
	Volumes     []float64
}

// Analyze performs volume analysis
func (va *VolumeAnalyzer) Analyze(ctx context.Context, data MarketData) (*VolumeAnalysis, error) {
	if len(data.Volumes) < 20 {
		return nil, ErrInsufficientData
	}

	// Calculate volume ratio (current volume / average volume)
	avgVolume := calculateAverageVolume(data.Volumes)
	volumeRatio := data.Volumes[len(data.Volumes)-1] / avgVolume

	// Calculate volume profile
	volumeProfile := calculateVolumeProfile(data.Prices, data.Volumes)

	// Calculate liquidity (sum of bid and ask volumes)
	liquidity := calculateLiquidity(data.OrderBook)

	// Calculate volume change (percentage change from previous period)
	volumeChange := calculateVolumeChange(data.Volumes)

	// Calculate relative volume (compared to n-day average)
	relativeVolume := calculateRelativeVolume(data.Volumes)

	return &VolumeAnalysis{
		VolumeRatio:    volumeRatio,
		AverageVolume:  avgVolume,
		VolumeProfile:  volumeProfile,
		Liquidity:      liquidity,
		VolumeChange:   volumeChange,
		RelativeVolume: relativeVolume,
	}, nil
}

func calculateAverageVolume(volumes []float64) float64 {
	if len(volumes) == 0 {
		return 0
	}

	sum := 0.0
	for _, volume := range volumes {
		sum += volume
	}
	return sum / float64(len(volumes))
}

func calculateVolumeProfile(prices, volumes []float64) VolumeProfile {
	if len(prices) != len(volumes) || len(prices) == 0 {
		return VolumeProfile{}
	}

	// Find price range
	minPrice := prices[0]
	maxPrice := prices[0]
	for _, price := range prices {
		if price < minPrice {
			minPrice = price
		}
		if price > maxPrice {
			maxPrice = price
		}
	}

	// Create price levels (10 levels)
	numLevels := 10
	priceStep := (maxPrice - minPrice) / float64(numLevels)
	priceLevels := make([]float64, numLevels)
	volumeLevels := make([]float64, numLevels)

	for i := 0; i < numLevels; i++ {
		priceLevels[i] = minPrice + float64(i)*priceStep
	}

	// Aggregate volumes at each price level
	for i, price := range prices {
		level := int(math.Floor((price - minPrice) / priceStep))
		if level >= numLevels {
			level = numLevels - 1
		}
		volumeLevels[level] += volumes[i]
	}

	return VolumeProfile{
		PriceLevels: priceLevels,
		Volumes:     volumeLevels,
	}
}

func calculateLiquidity(orderBook OrderBook) float64 {
	bidVolume := 0.0
	askVolume := 0.0

	for _, bid := range orderBook.Bids {
		bidVolume += bid.Amount
	}

	for _, ask := range orderBook.Asks {
		askVolume += ask.Amount
	}

	return bidVolume + askVolume
}

func calculateVolumeChange(volumes []float64) float64 {
	if len(volumes) < 2 {
		return 0
	}

	prev := volumes[len(volumes)-2]
	curr := volumes[len(volumes)-1]

	if prev == 0 {
		return 0
	}

	return (curr - prev) / prev * 100
}

func calculateRelativeVolume(volumes []float64) float64 {
	if len(volumes) < 20 {
		return 0
	}

	// Calculate 20-day average volume
	sum := 0.0
	for i := len(volumes) - 20; i < len(volumes)-1; i++ {
		sum += volumes[i]
	}
	avgVolume := sum / 19 // Exclude current volume

	currentVolume := volumes[len(volumes)-1]
	if avgVolume == 0 {
		return 0
	}

	return currentVolume / avgVolume
}
