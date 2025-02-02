package streaming

import (
	"context"
	"testing"
	"time"
)

func TestRSI(t *testing.T) {
	rsi, err := NewRSI(14)
	if err != nil {
		t.Fatalf("Failed to create RSI: %v", err)
	}

	// Test initialization
	if rsi.Name() != "RSI(14)" {
		t.Errorf("Expected name RSI(14), got %s", rsi.Name())
	}

	ctx := context.Background()
	now := time.Now()

	// Test first value
	value, err := rsi.Update(ctx, Price{Timestamp: now, Value: 100})
	if err != nil {
		t.Fatalf("Failed to update RSI: %v", err)
	}
	if value.Value != 50 {
		t.Errorf("Expected initial RSI value 50, got %f", value.Value)
	}

	// Test reset
	rsi.Reset()
	if rsi.initialized {
		t.Error("RSI should not be initialized after reset")
	}
}

func TestEMA(t *testing.T) {
	ema, err := NewEMA(10)
	if err != nil {
		t.Fatalf("Failed to create EMA: %v", err)
	}

	// Test initialization
	if ema.Name() != "EMA(10)" {
		t.Errorf("Expected name EMA(10), got %s", ema.Name())
	}

	ctx := context.Background()
	now := time.Now()

	// Test first value
	value, err := ema.Update(ctx, Price{Timestamp: now, Value: 100})
	if err != nil {
		t.Fatalf("Failed to update EMA: %v", err)
	}
	if value.Value != 100 {
		t.Errorf("Expected initial EMA value 100, got %f", value.Value)
	}

	// Test subsequent value
	value, err = ema.Update(ctx, Price{Timestamp: now.Add(time.Minute), Value: 200})
	if err != nil {
		t.Fatalf("Failed to update EMA: %v", err)
	}
	if value.Value <= 100 {
		t.Errorf("Expected EMA value > 100, got %f", value.Value)
	}
}

func TestMACD(t *testing.T) {
	macd, err := NewMACD(12, 26, 9)
	if err != nil {
		t.Fatalf("Failed to create MACD: %v", err)
	}

	// Test initialization
	expectedName := "MACD(12,26,9)"
	if macd.Name() != expectedName {
		t.Errorf("Expected name %s, got %s", expectedName, macd.Name())
	}

	ctx := context.Background()
	now := time.Now()

	// Test updates
	prices := []float64{100, 105, 110, 115, 120}
	for _, price := range prices {
		_, err := macd.Update(ctx, Price{Timestamp: now, Value: price})
		if err != nil {
			t.Fatalf("Failed to update MACD: %v", err)
		}
		now = now.Add(time.Minute)
	}

	// Test reset
	macd.Reset()
	if macd.initialized {
		t.Error("MACD should not be initialized after reset")
	}
}

func TestIndicatorFactory(t *testing.T) {
	factory := NewIndicatorFactory()

	// Test RSI creation
	rsi, err := factory.CreateRSI(14)
	if err != nil {
		t.Fatalf("Failed to create RSI: %v", err)
	}
	if rsi.Name() != "RSI(14)" {
		t.Errorf("Expected RSI name RSI(14), got %s", rsi.Name())
	}

	// Test EMA creation
	ema, err := factory.CreateEMA(10)
	if err != nil {
		t.Fatalf("Failed to create EMA: %v", err)
	}
	if ema.Name() != "EMA(10)" {
		t.Errorf("Expected EMA name EMA(10), got %s", ema.Name())
	}

	// Test MACD creation
	macd, err := factory.CreateMACD(12, 26, 9)
	if err != nil {
		t.Fatalf("Failed to create MACD: %v", err)
	}
	if macd.Name() != "MACD(12,26,9)" {
		t.Errorf("Expected MACD name MACD(12,26,9), got %s", macd.Name())
	}
}
