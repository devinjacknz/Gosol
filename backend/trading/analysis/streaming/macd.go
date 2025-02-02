package streaming

import (
	"context"
	"fmt"
)

// MACD implements the Moving Average Convergence Divergence indicator
type MACD struct {
	fastEMA     *EMA
	slowEMA     *EMA
	signalEMA   *EMA
	initialized bool
}

// NewMACD creates a new MACD indicator instance
func NewMACD(fastPeriod, slowPeriod, signalPeriod int) (*MACD, error) {
	if fastPeriod >= slowPeriod {
		return nil, fmt.Errorf("fast period must be less than slow period")
	}

	fastEMA, err := NewEMA(fastPeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to create fast EMA: %w", err)
	}

	slowEMA, err := NewEMA(slowPeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to create slow EMA: %w", err)
	}

	signalEMA, err := NewEMA(signalPeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to create signal EMA: %w", err)
	}

	return &MACD{
		fastEMA:   fastEMA,
		slowEMA:   slowEMA,
		signalEMA: signalEMA,
	}, nil
}

// Name returns the indicator name
func (m *MACD) Name() string {
	return fmt.Sprintf("MACD(%d,%d,%d)",
		m.fastEMA.period,
		m.slowEMA.period,
		m.signalEMA.period)
}

// Reset clears the indicator's state
func (m *MACD) Reset() {
	m.fastEMA.Reset()
	m.slowEMA.Reset()
	m.signalEMA.Reset()
	m.initialized = false
}

// Update processes a new price point and returns the new MACD values
func (m *MACD) Update(ctx context.Context, price Price) (*IndicatorValue, error) {
	// Update EMAs
	fastValue, err := m.fastEMA.Update(ctx, price)
	if err != nil {
		return nil, fmt.Errorf("failed to update fast EMA: %w", err)
	}

	slowValue, err := m.slowEMA.Update(ctx, price)
	if err != nil {
		return nil, fmt.Errorf("failed to update slow EMA: %w", err)
	}

	// Calculate MACD line
	macdLine := fastValue.Value - slowValue.Value

	// Update signal line
	signalPrice := Price{
		Timestamp: price.Timestamp,
		Value:     macdLine,
	}
	signalValue, err := m.signalEMA.Update(ctx, signalPrice)
	if err != nil {
		return nil, fmt.Errorf("failed to update signal EMA: %w", err)
	}

	// Calculate histogram
	histogram := macdLine - signalValue.Value

	return &IndicatorValue{
		Timestamp: price.Timestamp,
		Name:      m.Name(),
		Value:     histogram,
	}, nil
}
