package batch

import (
	"context"
	"testing"
	"time"

	"github.com/kwanRoshi/Gosol/backend/trading/analysis/streaming"
)

func TestBatchAdapter(t *testing.T) {
	// Create a streaming indicator
	factory := streaming.NewIndicatorFactory()
	rsi, err := factory.CreateRSI(14)
	if err != nil {
		t.Fatalf("Failed to create RSI: %v", err)
	}

	// Create batch adapter
	adapter := NewBatchAdapter(rsi)

	// Test adapter name
	expectedName := "Batch(RSI(14))"
	if adapter.Name() != expectedName {
		t.Errorf("Expected name %s, got %s", expectedName, adapter.Name())
	}

	// Create test data
	now := time.Now()
	prices := []BatchPrice{
		{
			Timestamp: now,
			Open:      100,
			High:      105,
			Low:       98,
			Close:     102,
			Volume:    1000,
		},
		{
			Timestamp: now.Add(time.Minute),
			Open:      102,
			High:      108,
			Low:       101,
			Close:     107,
			Volume:    1200,
		},
		{
			Timestamp: now.Add(2 * time.Minute),
			Open:      107,
			High:      110,
			Low:       105,
			Close:     109,
			Volume:    1100,
		},
	}

	// Process batch
	ctx := context.Background()
	results, err := adapter.ProcessBatch(ctx, prices)
	if err != nil {
		t.Fatalf("Failed to process batch: %v", err)
	}

	// Verify results
	if len(results) != len(prices) {
		t.Errorf("Expected %d results, got %d", len(prices), len(results))
	}

	// Verify each result has correct name and timestamp
	for i, result := range results {
		if result.Name != "RSI(14)" {
			t.Errorf("Result %d: expected name RSI(14), got %s", i, result.Name)
		}
		if !result.Timestamp.Equal(prices[i].Timestamp) {
			t.Errorf("Result %d: timestamp mismatch", i)
		}
	}

	// Test with empty batch
	emptyResults, err := adapter.ProcessBatch(ctx, []BatchPrice{})
	if err != nil {
		t.Fatalf("Failed to process empty batch: %v", err)
	}
	if len(emptyResults) != 0 {
		t.Errorf("Expected empty results, got %d results", len(emptyResults))
	}

	// Test GetLastValue with no data
	_, err = adapter.GetLastValue()
	if err == nil {
		t.Error("Expected error when getting last value with no data")
	}
}
