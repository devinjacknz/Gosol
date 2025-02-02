package streaming

import (
	"context"
	"time"
)

// Price represents a price point in time
type Price struct {
	Timestamp time.Time
	Value     float64
	Volume    float64
}

// IndicatorValue represents the calculated value of an indicator
type IndicatorValue struct {
	Timestamp time.Time
	Name      string
	Value     float64
}

// Indicator defines the interface for all technical indicators
type Indicator interface {
	// Update processes a new price point and returns the new indicator value
	Update(ctx context.Context, price Price) (*IndicatorValue, error)
	// Reset clears the indicator's state
	Reset()
	// Name returns the indicator's name
	Name() string
}

// WindowedIndicator represents indicators that operate on a time window
type WindowedIndicator interface {
	Indicator
	// SetWindow updates the calculation window
	SetWindow(period int) error
}

// IndicatorFactory creates new indicator instances
type IndicatorFactory interface {
	// CreateRSI creates a new RSI indicator
	CreateRSI(period int) (WindowedIndicator, error)
	// CreateEMA creates a new EMA indicator
	CreateEMA(period int) (WindowedIndicator, error)
	// CreateMACD creates a new MACD indicator
	CreateMACD(fastPeriod, slowPeriod, signalPeriod int) (Indicator, error)
}
