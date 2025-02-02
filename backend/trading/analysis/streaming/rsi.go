package streaming

import (
	"context"
	"fmt"
	"math"
)

// RSI implements the Relative Strength Index indicator
type RSI struct {
	period      int
	lastPrice   float64
	gains       []float64
	losses      []float64
	avgGain     float64
	avgLoss     float64
	initialized bool
}

// NewRSI creates a new RSI indicator instance
func NewRSI(period int) (*RSI, error) {
	if period < 2 {
		return nil, fmt.Errorf("period must be >= 2, got %d", period)
	}
	return &RSI{
		period: period,
		gains:  make([]float64, 0, period),
		losses: make([]float64, 0, period),
	}, nil
}

// Name returns the indicator name
func (r *RSI) Name() string {
	return fmt.Sprintf("RSI(%d)", r.period)
}

// SetWindow updates the calculation window
func (r *RSI) SetWindow(period int) error {
	if period < 2 {
		return fmt.Errorf("period must be >= 2, got %d", period)
	}
	r.period = period
	r.Reset()
	return nil
}

// Reset clears the indicator's state
func (r *RSI) Reset() {
	r.gains = make([]float64, 0, r.period)
	r.losses = make([]float64, 0, r.period)
	r.avgGain = 0
	r.avgLoss = 0
	r.initialized = false
}

// Update processes a new price point and returns the new RSI value
func (r *RSI) Update(ctx context.Context, price Price) (*IndicatorValue, error) {
	if !r.initialized {
		r.lastPrice = price.Value
		r.initialized = true
		return &IndicatorValue{
			Timestamp: price.Timestamp,
			Name:      r.Name(),
			Value:     50, // Default value when not enough data
		}, nil
	}

	change := price.Value - r.lastPrice
	r.lastPrice = price.Value

	gain := math.Max(change, 0)
	loss := math.Max(-change, 0)

	r.gains = append(r.gains, gain)
	r.losses = append(r.losses, loss)

	if len(r.gains) > r.period {
		r.gains = r.gains[1:]
		r.losses = r.losses[1:]
	}

	if len(r.gains) < r.period {
		return &IndicatorValue{
			Timestamp: price.Timestamp,
			Name:      r.Name(),
			Value:     50, // Default value when not enough data
		}, nil
	}

	r.avgGain = average(r.gains)
	r.avgLoss = average(r.losses)

	rs := r.avgGain / (r.avgLoss + 0.000001) // Avoid division by zero
	rsi := 100 - (100 / (1 + rs))

	return &IndicatorValue{
		Timestamp: price.Timestamp,
		Name:      r.Name(),
		Value:     rsi,
	}, nil
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
