package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/leonzhao/trading-system/backend/models"
	"github.com/leonzhao/trading-system/backend/repository"
)

// MarketDataService handles real-time market data processing and distribution
type MarketDataService struct {
	db           *repository.Database
	wsService    *WebSocketService
	cache        map[string]*models.MarketDataCache
	orderBooks   map[string]*models.OrderBookCache
	subscribers  map[string][]chan *models.MarketDataRecord
	mutex        sync.RWMutex
	cleanupTimer *time.Timer
}

// NewMarketDataService creates a new market data service
func NewMarketDataService(db *repository.Database, ws *WebSocketService) *MarketDataService {
	service := &MarketDataService{
		db:          db,
		wsService:   ws,
		cache:       make(map[string]*models.MarketDataCache),
		orderBooks:  make(map[string]*models.OrderBookCache),
		subscribers: make(map[string][]chan *models.MarketDataRecord),
	}

	// Start cleanup timer
	service.cleanupTimer = time.NewTimer(24 * time.Hour)
	go service.cleanupRoutine()

	return service
}

// ProcessMarketData handles incoming market data
func (s *MarketDataService) ProcessMarketData(ctx context.Context, data *models.MarketDataRecord) error {
	// Save to database
	if err := s.db.SaveMarketData(ctx, data); err != nil {
		return fmt.Errorf("failed to save market data: %w", err)
	}

	// Update cache
	s.mutex.Lock()
	s.cache[data.Symbol] = &models.MarketDataCache{
		Symbol:    data.Symbol,
		Data:      string(mustMarshal(data)),
		UpdatedAt: time.Now(),
	}
	s.mutex.Unlock()

	// Notify subscribers
	s.notifySubscribers(data.Symbol, data)

	// Broadcast through WebSocket
	message := map[string]interface{}{
		"type":   "market_data",
		"symbol": data.Symbol,
		"data":   data,
	}
	s.wsService.Broadcast(mustMarshal(message))

	return nil
}

// ProcessOrderBook handles incoming order book updates
func (s *MarketDataService) ProcessOrderBook(ctx context.Context, symbol string, bids, asks [][2]float64) error {
	orderBook := &models.OrderBookCache{
		Symbol:    symbol,
		Data:      string(mustMarshal(map[string]interface{}{"bids": bids, "asks": asks})),
		UpdatedAt: time.Now(),
	}

	// Update cache
	s.mutex.Lock()
	s.orderBooks[symbol] = orderBook
	s.mutex.Unlock()

	// Save to database
	if err := s.db.SaveOrderBookCache(ctx, orderBook); err != nil {
		return fmt.Errorf("failed to save order book: %w", err)
	}

	// Broadcast through WebSocket
	message := map[string]interface{}{
		"type":   "order_book",
		"symbol": symbol,
		"bids":   bids,
		"asks":   asks,
	}
	s.wsService.Broadcast(mustMarshal(message))

	return nil
}

// Subscribe adds a subscriber for market data updates
func (s *MarketDataService) Subscribe(symbol string) chan *models.MarketDataRecord {
	ch := make(chan *models.MarketDataRecord, 100)

	s.mutex.Lock()
	if _, ok := s.subscribers[symbol]; !ok {
		s.subscribers[symbol] = make([]chan *models.MarketDataRecord, 0)
	}
	s.subscribers[symbol] = append(s.subscribers[symbol], ch)
	s.mutex.Unlock()

	return ch
}

// Unsubscribe removes a subscriber
func (s *MarketDataService) Unsubscribe(symbol string, ch chan *models.MarketDataRecord) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if subs, ok := s.subscribers[symbol]; ok {
		for i, sub := range subs {
			if sub == ch {
				s.subscribers[symbol] = append(subs[:i], subs[i+1:]...)
				close(ch)
				break
			}
		}
	}
}

// GetLatestPrice returns the latest price for a symbol
func (s *MarketDataService) GetLatestPrice(ctx context.Context, symbol string) (float64, error) {
	s.mutex.RLock()
	if cache, ok := s.cache[symbol]; ok {
		s.mutex.RUnlock()
		var data models.MarketDataRecord
		if err := json.Unmarshal([]byte(cache.Data), &data); err != nil {
			return 0, fmt.Errorf("failed to unmarshal cached data: %w", err)
		}
		return data.Price, nil
	}
	s.mutex.RUnlock()

	// If not in cache, get from database
	records, err := s.db.GetMarketData(ctx, symbol, time.Now().Add(-1*time.Hour), time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to get market data: %w", err)
	}
	if len(records) == 0 {
		return 0, fmt.Errorf("no market data available for symbol: %s", symbol)
	}
	return records[len(records)-1].Price, nil
}

// GetOrderBook returns the current order book for a symbol
func (s *MarketDataService) GetOrderBook(ctx context.Context, symbol string) ([][2]float64, [][2]float64, error) {
	s.mutex.RLock()
	if ob, ok := s.orderBooks[symbol]; ok {
		s.mutex.RUnlock()
		var data struct {
			Bids [][2]float64 `json:"bids"`
			Asks [][2]float64 `json:"asks"`
		}
		if err := json.Unmarshal([]byte(ob.Data), &data); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal order book: %w", err)
		}
		return data.Bids, data.Asks, nil
	}
	s.mutex.RUnlock()

	// If not in cache, get from database
	ob, err := s.db.GetOrderBookCache(ctx, symbol)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get order book: %w", err)
	}
	var data struct {
		Bids [][2]float64 `json:"bids"`
		Asks [][2]float64 `json:"asks"`
	}
	if err := json.Unmarshal([]byte(ob.Data), &data); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal order book: %w", err)
	}
	return data.Bids, data.Asks, nil
}

// Internal helper functions

func (s *MarketDataService) notifySubscribers(symbol string, data *models.MarketDataRecord) {
	s.mutex.RLock()
	if subs, ok := s.subscribers[symbol]; ok {
		for _, ch := range subs {
			select {
			case ch <- data:
			default:
				// Channel is full, skip this update
			}
		}
	}
	s.mutex.RUnlock()
}

func (s *MarketDataService) cleanupRoutine() {
	for {
		<-s.cleanupTimer.C

		ctx := context.Background()
		before := time.Now().Add(-7 * 24 * time.Hour) // Keep 7 days of data

		if err := s.db.CleanupOldData(ctx, before); err != nil {
			// Log error but continue
			fmt.Printf("Failed to cleanup old data: %v\n", err)
		}

		s.cleanupTimer.Reset(24 * time.Hour)
	}
}

func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal json: %v", err))
	}
	return data
}
