package streaming

import (
	"context"
	"fmt"
)

// EMA implements the Exponential Moving Average indicator
type EMA struct {
	period      int
	multiplier  float64
	currentEMA  float64
	initialized bool
}

// NewEMA creates a new EMA indicator instance
func NewEMA(period int) (*EMA, error) {
	if period < 1 {
		return nil, fmt.Errorf("period must be >= 1, got %d", period)
	}
	return &EMA{
		period:     period,
		multiplier: 2.0 / float64(period+1),
	}, nil
}

// Name returns the indicator name
func (e *EMA) Name() string {
	return fmt.Sprintf("EMA(%d)", e.period)
}

// SetWindow updates the calculation window
func (e *EMA) SetWindow(period int) error {
	if period < 1 {
		return fmt.Errorf("period must be >= 1, got %d", period)
	}
	e.period = period
	e.multiplier = 2.0 / float64(period+1)
	e.Reset()
	return nil
}

// Reset clears the indicator's state
func (e *EMA) Reset() {
	e.currentEMA = 0
	e.initialized = false
}

// Update processes a new price point and returns the new EMA value
func (e *EMA) Update(ctx context.Context, price Price) (*IndicatorValue, error) {
	if !e.initialized {
		e.currentEMA = price.Value
		e.initialized = true
		return &IndicatorValue{
			Timestamp: price.Timestamp,
			Name:      e.Name(),
			Value:     e.currentEMA,
		}, nil
	}

	// EMA = (Close - EMA(previous)) * multiplier + EMA(previous)
	e.currentEMA = (price.Value-e.currentEMA)*e.multiplier + e.currentEMA

	return &IndicatorValue{
		Timestamp: price.Timestamp,
		Name:      e.Name(),
		Value:     e.currentEMA,
	}, nil
}
