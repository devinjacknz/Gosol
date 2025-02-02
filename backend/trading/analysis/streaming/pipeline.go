package streaming

import (
	"context"
	"fmt"
	"sync"
)

// PriceHandler defines the interface for components that process price updates
type PriceHandler interface {
	HandlePrice(ctx context.Context, price Price) error
}

// IndicatorPipeline manages a collection of technical indicators and processes price updates
type IndicatorPipeline struct {
	indicators []Indicator
	handlers   []PriceHandler
	mu         sync.RWMutex
}

// NewIndicatorPipeline creates a new IndicatorPipeline instance
func NewIndicatorPipeline() *IndicatorPipeline {
	return &IndicatorPipeline{
		indicators: make([]Indicator, 0),
		handlers:   make([]PriceHandler, 0),
	}
}

// AddIndicator adds a new indicator to the pipeline
func (p *IndicatorPipeline) AddIndicator(indicator Indicator) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.indicators = append(p.indicators, indicator)
}

// AddHandler adds a new price handler to the pipeline
func (p *IndicatorPipeline) AddHandler(handler PriceHandler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers = append(p.handlers, handler)
}

// ProcessPrice processes a new price update through all indicators and handlers
func (p *IndicatorPipeline) ProcessPrice(ctx context.Context, price Price) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Process through indicators
	for _, indicator := range p.indicators {
		value, err := indicator.Update(ctx, price)
		if err != nil {
			return fmt.Errorf("failed to update indicator %s: %w", indicator.Name(), err)
		}

		// Notify handlers
		for _, handler := range p.handlers {
			if err := handler.HandlePrice(ctx, price); err != nil {
				return fmt.Errorf("handler failed to process price: %w", err)
			}
		}

		// Log indicator value (in production, you might want to send this to a proper handler)
		fmt.Printf("Indicator %s: %f\n", value.Name, value.Value)
	}

	return nil
}

// Reset resets all indicators in the pipeline
func (p *IndicatorPipeline) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, indicator := range p.indicators {
		indicator.Reset()
	}
}
