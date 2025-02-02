package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kwanRoshi/Gosol/backend/trading/analysis/streaming"
)

// TimeRange represents a time range for querying data
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// Storage defines the interface for storing and retrieving indicator values
type Storage interface {
	// Store stores an indicator value
	Store(ctx context.Context, value streaming.IndicatorValue) error
	// Query retrieves indicator values within a time range
	Query(ctx context.Context, indicatorName string, timeRange TimeRange) ([]streaming.IndicatorValue, error)
	// GetLatest retrieves the latest value for an indicator
	GetLatest(ctx context.Context, indicatorName string) (*streaming.IndicatorValue, error)
}

// MemoryStorage implements in-memory storage for indicator values
type MemoryStorage struct {
	values map[string][]streaming.IndicatorValue
	mu     sync.RWMutex
}

// NewMemoryStorage creates a new in-memory storage instance
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		values: make(map[string][]streaming.IndicatorValue),
	}
}

// Store stores an indicator value
func (s *MemoryStorage) Store(ctx context.Context, value streaming.IndicatorValue) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.values[value.Name]; !exists {
		s.values[value.Name] = make([]streaming.IndicatorValue, 0)
	}
	s.values[value.Name] = append(s.values[value.Name], value)
	return nil
}

// Query retrieves indicator values within a time range
func (s *MemoryStorage) Query(ctx context.Context, indicatorName string, timeRange TimeRange) ([]streaming.IndicatorValue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	values, exists := s.values[indicatorName]
	if !exists {
		return nil, fmt.Errorf("no data found for indicator %s", indicatorName)
	}

	results := make([]streaming.IndicatorValue, 0)
	for _, value := range values {
		if (value.Timestamp.Equal(timeRange.Start) || value.Timestamp.After(timeRange.Start)) &&
			(value.Timestamp.Equal(timeRange.End) || value.Timestamp.Before(timeRange.End)) {
			results = append(results, value)
		}
	}

	return results, nil
}

// GetLatest retrieves the latest value for an indicator
func (s *MemoryStorage) GetLatest(ctx context.Context, indicatorName string) (*streaming.IndicatorValue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	values, exists := s.values[indicatorName]
	if !exists || len(values) == 0 {
		return nil, fmt.Errorf("no data found for indicator %s", indicatorName)
	}

	latest := values[len(values)-1]
	return &latest, nil
}

// StorageHandler implements PriceHandler for storing indicator values
type StorageHandler struct {
	storage Storage
}

// NewStorageHandler creates a new storage handler
func NewStorageHandler(storage Storage) *StorageHandler {
	return &StorageHandler{
		storage: storage,
	}
}

// HandlePrice implements PriceHandler.HandlePrice
func (h *StorageHandler) HandlePrice(ctx context.Context, value streaming.IndicatorValue) error {
	return h.storage.Store(ctx, value)
}
