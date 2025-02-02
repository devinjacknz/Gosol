package streaming

import (
	"context"
	"sync"
	"testing"
	"time"
)

// MockHandler implements PriceHandler for testing
type MockHandler struct {
	prices []Price
	mu     sync.Mutex
}

func NewMockHandler() *MockHandler {
	return &MockHandler{
		prices: make([]Price, 0),
	}
}

func (h *MockHandler) HandlePrice(ctx context.Context, price Price) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.prices = append(h.prices, price)
	return nil
}

func (h *MockHandler) GetPrices() []Price {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.prices
}

func TestIndicatorPipeline(t *testing.T) {
	// Create pipeline
	pipeline := NewIndicatorPipeline()

	// Create indicators
	factory := NewIndicatorFactory()

	rsi, err := factory.CreateRSI(14)
	if err != nil {
		t.Fatalf("Failed to create RSI: %v", err)
	}
	pipeline.AddIndicator(rsi)

	ema, err := factory.CreateEMA(10)
	if err != nil {
		t.Fatalf("Failed to create EMA: %v", err)
	}
	pipeline.AddIndicator(ema)

	// Create and add mock handler
	handler := NewMockHandler()
	pipeline.AddHandler(handler)

	// Test price processing
	ctx := context.Background()
	now := time.Now()

	testPrices := []float64{100, 105, 110, 115, 120}
	for _, price := range testPrices {
		err := pipeline.ProcessPrice(ctx, Price{
			Timestamp: now,
			Value:     price,
		})
		if err != nil {
			t.Fatalf("Failed to process price: %v", err)
		}
		now = now.Add(time.Minute)
	}

	// Verify handler received all prices
	receivedPrices := handler.GetPrices()
	if len(receivedPrices) != len(testPrices)*2 { // Each price is processed by 2 indicators
		t.Errorf("Expected %d prices, got %d", len(testPrices)*2, len(receivedPrices))
	}

	// Test reset
	pipeline.Reset()
}
