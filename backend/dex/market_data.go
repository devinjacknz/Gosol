package dex

import (
	"context"
	"sync"
	"time"

	"solmeme-trader/models"
)

// MarketDataService handles market data updates
type MarketDataService struct {
	client   *DexClient
	interval time.Duration
	cache    map[string]*models.MarketData
	lock     sync.RWMutex
}

// NewMarketDataService creates a new market data service
func NewMarketDataService(client *DexClient, interval time.Duration) *MarketDataService {
	return &MarketDataService{
		client:   client,
		interval: interval,
		cache:    make(map[string]*models.MarketData),
	}
}

// Start starts the market data service
func (s *MarketDataService) Start(ctx context.Context) error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := s.updateMarketData(ctx); err != nil {
				// Log error but continue
				continue
			}
		}
	}
}

// GetMarketData gets market data for a token
func (s *MarketDataService) GetMarketData(tokenAddress string) *models.MarketData {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.cache[tokenAddress]
}

// updateMarketData updates market data for all tokens
func (s *MarketDataService) updateMarketData(ctx context.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// TODO: Get list of tokens to monitor
	tokens := []string{"token1", "token2"}

	for _, token := range tokens {
		info, err := s.client.GetMarketInfo(ctx, token)
		if err != nil {
			return err
		}

		liquidity, err := s.client.GetLiquidity(ctx, token)
		if err != nil {
			return err
		}

		s.cache[token] = &models.MarketData{
			TokenAddress: token,
			Price:       info.LastPrice,
			Volume24h:   info.BaseVolume,
			MarketCap:   info.LastPrice * liquidity.TotalSupply,
			Liquidity:   liquidity.TVL,
			PriceImpact: 0, // TODO: Calculate from orderbook
			Timestamp:   info.Timestamp,
		}
	}

	return nil
}

// GetHistoricalData gets historical market data for a token
func (s *MarketDataService) GetHistoricalData(ctx context.Context, tokenAddress string, start, end time.Time) ([]*models.MarketData, error) {
	// TODO: Implement historical data retrieval
	return nil, nil
}
