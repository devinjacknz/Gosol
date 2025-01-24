package dex

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MarketService handles market data operations
type MarketService struct {
	jupiterClient *JupiterClient
	raydiumClient *RaydiumClient
	cache         map[string]*MarketData
	cacheMutex    sync.RWMutex
	cacheTTL      time.Duration
}

// NewMarketService creates a new market service instance
func NewMarketService(jupiterClient *JupiterClient, raydiumClient *RaydiumClient, cacheTTL time.Duration) *MarketService {
	return &MarketService{
		jupiterClient: jupiterClient,
		raydiumClient: raydiumClient,
		cache:         make(map[string]*MarketData),
		cacheTTL:      cacheTTL,
	}
}

// GetMarketData fetches market data for a token
func (s *MarketService) GetMarketData(ctx context.Context, tokenAddress string) (*MarketData, error) {
	// Check cache first
	s.cacheMutex.RLock()
	if data, ok := s.cache[tokenAddress]; ok {
		if time.Since(data.Timestamp) < s.cacheTTL {
			s.cacheMutex.RUnlock()
			return data, nil
		}
	}
	s.cacheMutex.RUnlock()

	// Fetch fresh data
	data, err := s.fetchMarketData(ctx, tokenAddress)
	if err != nil {
		return nil, err
	}

	// Update cache
	s.cacheMutex.Lock()
	s.cache[tokenAddress] = data
	s.cacheMutex.Unlock()

	return data, nil
}

// fetchMarketData fetches fresh market data from DEXes
func (s *MarketService) fetchMarketData(ctx context.Context, tokenAddress string) (*MarketData, error) {
	// Get Jupiter price data
	priceResp, err := s.jupiterClient.GetPrice(ctx, tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get Jupiter price: %w", err)
	}

	// Get Raydium liquidity data
	liquidityData, err := s.raydiumClient.GetLiquidity(ctx, tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get Raydium liquidity: %w", err)
	}

	// Get order book data
	orderBook, err := s.raydiumClient.GetOrderBook(ctx, tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get order book: %w", err)
	}

	// Calculate price impact
	priceImpact, err := s.calculatePriceImpact(ctx, tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate price impact: %w", err)
	}

	return &MarketData{
		TokenAddress: tokenAddress,
		Price:       priceResp.Data.MintsPrice[tokenAddress],
		Volume24h:   liquidityData.Volume24h,
		MarketCap:   liquidityData.TVL,
		Liquidity:   liquidityData.TokenAmount,
		PriceImpact: priceImpact,
		OrderBook:   orderBook,
		Timestamp:   time.Now(),
	}, nil
}

// calculatePriceImpact calculates the price impact for a standard trade size
func (s *MarketService) calculatePriceImpact(ctx context.Context, tokenAddress string) (float64, error) {
	// Use a standard trade size for price impact calculation (e.g., $1000 worth)
	standardAmount := "1000000000" // in lamports

	impact, err := s.jupiterClient.CalculatePriceImpact(ctx, tokenAddress, "So11111111111111111111111111111111111111112", standardAmount)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate price impact: %w", err)
	}

	return impact, nil
}

// GetMarketDepth fetches market depth data
func (s *MarketService) GetMarketDepth(ctx context.Context, tokenAddress string) (*MarketDepth, error) {
	orderBook, err := s.raydiumClient.GetOrderBook(ctx, tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get order book: %w", err)
	}

	depth := &MarketDepth{
		TokenAddress: tokenAddress,
		Timestamp:   time.Now(),
		Levels:      make([]struct {
			Price     float64 `json:"price"`
			Liquidity float64 `json:"liquidity"`
		}, len(orderBook.Bids) + len(orderBook.Asks)),
	}

	// Add bid levels
	for i, bid := range orderBook.Bids {
		depth.Levels[i] = struct {
			Price     float64 `json:"price"`
			Liquidity float64 `json:"liquidity"`
		}{
			Price:     bid.Price,
			Liquidity: bid.Size,
		}
	}

	// Add ask levels
	offset := len(orderBook.Bids)
	for i, ask := range orderBook.Asks {
		depth.Levels[offset+i] = struct {
			Price     float64 `json:"price"`
			Liquidity float64 `json:"liquidity"`
		}{
			Price:     ask.Price,
			Liquidity: ask.Size,
		}
	}

	return depth, nil
}

// GetMarketStats fetches market statistics
func (s *MarketService) GetMarketStats(ctx context.Context, tokenAddress string) (*MarketStats, error) {
	// Get current market data
	current, err := s.GetMarketData(ctx, tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get current market data: %w", err)
	}

	// Get 24h historical data from Raydium
	historical, err := s.raydiumClient.GetHistoricalData(ctx, tokenAddress, time.Hour*24)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}

	return &MarketStats{
		TokenAddress:     tokenAddress,
		Price24hChange:   (current.Price - historical.OpenPrice) / historical.OpenPrice * 100,
		Volume24hChange:  (current.Volume24h - historical.Volume) / historical.Volume * 100,
		HighPrice24h:     historical.HighPrice,
		LowPrice24h:      historical.LowPrice,
		NumTrades24h:     historical.NumTrades,
		AverageTradeSize: current.Volume24h / float64(historical.NumTrades),
		Timestamp:        time.Now(),
	}, nil
}
