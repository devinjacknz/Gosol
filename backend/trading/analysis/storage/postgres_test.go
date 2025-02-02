package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/kwanRoshi/Gosol/backend/trading/analysis/streaming"
)

func getTestPostgresConnStr() string {
	// Use environment variable or default to a test database
	connStr := os.Getenv("TEST_POSTGRES_URL")
	if connStr == "" {
		// Use URL format
		connStr = "postgresql://postgres:postgres@localhost:5432/gosol_test?sslmode=disable"
	}
	return connStr
}

func TestPostgresStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL integration test in short mode")
	}

	storage, err := NewPostgresStorage(getTestPostgresConnStr())
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL storage: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()

	// Create test data
	now := time.Now().UTC().Truncate(time.Microsecond)
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

	// Verify timestamps and values
	for i, result := range results {
		expected := testValues[i]
		if !result.Timestamp.Equal(expected.Timestamp) {
			t.Errorf("Result %d: expected timestamp %v, got %v", i, expected.Timestamp, result.Timestamp)
		}
		if result.Value != expected.Value {
			t.Errorf("Result %d: expected value %f, got %f", i, expected.Value, result.Value)
		}
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

func TestPostgresStorageRaceConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL race condition test in short mode")
	}

	storage, err := NewPostgresStorage(getTestPostgresConnStr())
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL storage: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(i int) {
			value := streaming.IndicatorValue{
				Timestamp: now.Add(time.Duration(i) * time.Second),
				Name:      "ConcurrentTest",
				Value:     float64(i),
			}
			err := storage.Store(ctx, value)
			if err != nil {
				t.Errorf("Failed to store value concurrently: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all values were stored
	timeRange := TimeRange{
		Start: now,
		End:   now.Add(10 * time.Second),
	}
	results, err := storage.Query(ctx, "ConcurrentTest", timeRange)
	if err != nil {
		t.Fatalf("Failed to query concurrent values: %v", err)
	}

	if len(results) != 10 {
		t.Errorf("Expected 10 results from concurrent writes, got %d", len(results))
	}
}
