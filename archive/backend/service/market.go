package service

import (
	"sync"
	"time"

	"github.com/your-username/gosol/backend/models"
)

// MarketDataService handles market data operations
type MarketDataService struct {
	// Cache for latest market data
	marketDataCache map[string]*models.MarketData
	klineCache      map[string]map[string][]*models.Kline // symbol -> interval -> klines
	orderBookCache  map[string]*models.OrderBook

	// Subscriptions
	subscribers   map[string][]chan *models.MarketData
	klineSubs     map[string][]chan *models.Kline
	orderBookSubs map[string][]chan *models.OrderBook
	tradeSubs     map[string][]chan *models.Trade

	// Mutex for thread safety
	mu sync.RWMutex

	// Cleanup interval
	cleanupInterval time.Duration
}

// NewMarketDataService creates a new market data service
func NewMarketDataService() *MarketDataService {
	service := &MarketDataService{
		marketDataCache: make(map[string]*models.MarketData),
		klineCache:      make(map[string]map[string][]*models.Kline),
		orderBookCache:  make(map[string]*models.OrderBook),
		subscribers:     make(map[string][]chan *models.MarketData),
		klineSubs:       make(map[string][]chan *models.Kline),
		orderBookSubs:   make(map[string][]chan *models.OrderBook),
		tradeSubs:       make(map[string][]chan *models.Trade),
		cleanupInterval: 24 * time.Hour,
	}

	go service.startCleanupRoutine()
	return service
}

// UpdateMarketData updates the market data and notifies subscribers
func (s *MarketDataService) UpdateMarketData(data *models.MarketData) {
	s.mu.Lock()
	s.marketDataCache[data.Symbol] = data
	subscribers := s.subscribers[data.Symbol]
	s.mu.Unlock()

	// Notify subscribers
	for _, ch := range subscribers {
		select {
		case ch <- data:
		default:
			// Skip if subscriber is not ready to receive
		}
	}
}

// Subscribe subscribes to market data updates for a symbol
func (s *MarketDataService) Subscribe(symbol string) chan *models.MarketData {
	ch := make(chan *models.MarketData, 100)

	s.mu.Lock()
	s.subscribers[symbol] = append(s.subscribers[symbol], ch)
	s.mu.Unlock()

	return ch
}

// Unsubscribe removes a subscription
func (s *MarketDataService) Unsubscribe(symbol string, ch chan *models.MarketData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	subs := s.subscribers[symbol]
	for i, sub := range subs {
		if sub == ch {
			s.subscribers[symbol] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}
}

// GetLatestMarketData returns the latest market data for a symbol
func (s *MarketDataService) GetLatestMarketData(symbol string) *models.MarketData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.marketDataCache[symbol]
}

// UpdateKline updates the kline data and notifies subscribers
func (s *MarketDataService) UpdateKline(kline *models.Kline) {
	s.mu.Lock()
	if _, ok := s.klineCache[kline.Symbol]; !ok {
		s.klineCache[kline.Symbol] = make(map[string][]*models.Kline)
	}

	// Add new kline and maintain history
	s.klineCache[kline.Symbol][kline.Interval] = append(s.klineCache[kline.Symbol][kline.Interval], kline)
	subscribers := s.klineSubs[kline.Symbol]
	s.mu.Unlock()

	// Notify subscribers
	for _, ch := range subscribers {
		select {
		case ch <- kline:
		default:
		}
	}
}

// GetKlines returns klines for a symbol and interval
func (s *MarketDataService) GetKlines(symbol, interval string, limit int) []*models.Kline {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if klines, ok := s.klineCache[symbol][interval]; ok {
		if len(klines) > limit {
			return klines[len(klines)-limit:]
		}
		return klines
	}
	return nil
}

// startCleanupRoutine starts a routine to clean up old data
func (s *MarketDataService) startCleanupRoutine() {
	ticker := time.NewTicker(s.cleanupInterval)
	for range ticker.C {
		s.cleanup()
	}
}

// cleanup removes old data from caches
func (s *MarketDataService) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Keep only recent klines (e.g., last 1000 per symbol/interval)
	const maxKlines = 1000
	for symbol, intervals := range s.klineCache {
		for interval, klines := range intervals {
			if len(klines) > maxKlines {
				s.klineCache[symbol][interval] = klines[len(klines)-maxKlines:]
			}
		}
	}
}
