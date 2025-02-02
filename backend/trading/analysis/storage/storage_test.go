package storage

import (
	"context"
	"testing"
	"time"

	"github.com/kwanRoshi/Gosol/backend/trading/analysis/streaming"
)

func TestMemoryStorage(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()

	// Create test data
	now := time.Now()
	testValues := []streaming.IndicatorValue{
		{
			Timestamp: now,
			Name:      "RSI(14)",
			Value:     70.5,
		},
		{
			Timestamp: now.Add(time.Minute),
			Name:      "RSI(14)",
			Value:     65.3,
		},
		{
			Timestamp: now.Add(2 * time.Minute),
			Name:      "RSI(14)",
			Value:     68.2,
		},
	}

	// Test storing values
	for _, value := range testValues {
		err := storage.Store(ctx, value)
		if err != nil {
			t.Fatalf("Failed to store value: %v", err)
		}
	}

	// Test querying values
	timeRange := TimeRange{
		Start: now,
		End:   now.Add(2 * time.Minute),
	}

	results, err := storage.Query(ctx, "RSI(14)", timeRange)
	if err != nil {
		t.Fatalf("Failed to query values: %v", err)
	}

	if len(results) != len(testValues) {
		t.Errorf("Expected %d results, got %d", len(testValues), len(results))
	}

	// Test getting latest value
	latest, err := storage.GetLatest(ctx, "RSI(14)")
	if err != nil {
		t.Fatalf("Failed to get latest value: %v", err)
	}

	expectedLatest := testValues[len(testValues)-1]
	if latest.Value != expectedLatest.Value {
		t.Errorf("Expected latest value %f, got %f", expectedLatest.Value, latest.Value)
	}

	// Test querying non-existent indicator
	_, err = storage.Query(ctx, "NonExistent", timeRange)
	if err == nil {
		t.Error("Expected error when querying non-existent indicator")
	}

	// Test getting latest for non-existent indicator
	_, err = storage.GetLatest(ctx, "NonExistent")
	if err == nil {
		t.Error("Expected error when getting latest for non-existent indicator")
	}
}

func TestStorageHandler(t *testing.T) {
	storage := NewMemoryStorage()
	handler := NewStorageHandler(storage)
	ctx := context.Background()

	// Test handling a value
	value := streaming.IndicatorValue{
		Timestamp: time.Now(),
		Name:      "Test",
		Value:     42.0,
	}

	err := handler.HandlePrice(ctx, value)
	if err != nil {
		t.Fatalf("Failed to handle price: %v", err)
	}

	// Verify the value was stored
	stored, err := storage.GetLatest(ctx, "Test")
	if err != nil {
		t.Fatalf("Failed to get stored value: %v", err)
	}

	if stored.Value != value.Value {
		t.Errorf("Expected value %f, got %f", value.Value, stored.Value)
	}
}
