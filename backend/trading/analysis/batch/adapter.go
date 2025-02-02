package batch

import (
	"context"
	"fmt"
	"time"

	"github.com/kwanRoshi/Gosol/backend/trading/analysis/streaming"
)

// BatchPrice represents a price point with OHLCV data
type BatchPrice struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

// BatchAdapter converts streaming indicators to batch processing mode
type BatchAdapter struct {
	indicator streaming.Indicator
	prices    []BatchPrice
}

// NewBatchAdapter creates a new batch adapter for a streaming indicator
func NewBatchAdapter(indicator streaming.Indicator) *BatchAdapter {
	return &BatchAdapter{
		indicator: indicator,
		prices:    make([]BatchPrice, 0),
	}
}

// ProcessBatch processes a batch of price data and returns the indicator values
func (b *BatchAdapter) ProcessBatch(ctx context.Context, prices []BatchPrice) ([]streaming.IndicatorValue, error) {
	results := make([]streaming.IndicatorValue, 0, len(prices))

	// Reset indicator state
	b.indicator.Reset()

	// Process each price point
	for _, price := range prices {
		streamPrice := streaming.Price{
			Timestamp: price.Timestamp,
			Value:     price.Close, // Use closing price by default
			Volume:    price.Volume,
		}

		value, err := b.indicator.Update(ctx, streamPrice)
		if err != nil {
			return nil, fmt.Errorf("failed to process price: %w", err)
		}

		results = append(results, *value)
	}

	return results, nil
}

// Name returns the name of the underlying indicator
func (b *BatchAdapter) Name() string {
	return fmt.Sprintf("Batch(%s)", b.indicator.Name())
}

// GetLastValue returns the last calculated indicator value
func (b *BatchAdapter) GetLastValue() (*streaming.IndicatorValue, error) {
	if len(b.prices) == 0 {
		return nil, fmt.Errorf("no data available")
	}

	lastPrice := b.prices[len(b.prices)-1]
	return &streaming.IndicatorValue{
		Timestamp: lastPrice.Timestamp,
		Name:      b.Name(),
		Value:     lastPrice.Close,
	}, nil
}
