package dex

import (
	"context"
	"time"
)

// MarketDataService handles market data operations
type MarketDataService struct {
	raydiumClient  *RaydiumClient
	jupiterClient  *JupiterClient
}

// NewMarketDataService creates a new market data service
func NewMarketDataService(raydiumClient *RaydiumClient, jupiterClient *JupiterClient) *MarketDataService {
	return &MarketDataService{
		raydiumClient:  raydiumClient,
		jupiterClient:  jupiterClient,
	}
}

// GetMarketData gets market data for a token
func (s *MarketDataService) GetMarketData(ctx context.Context, tokenAddress string) (*MarketData, error) {
	// Get price data from Jupiter
	priceResp, err := s.jupiterClient.GetPrice(ctx, tokenAddress)
	if err != nil {
		return nil, err
	}

	// Get liquidity data from Raydium
	liquidityData, err := s.raydiumClient.GetLiquidity(ctx, tokenAddress)
	if err != nil {
		return nil, err
	}

	// Get order book from Raydium
	orderBook, err := s.raydiumClient.GetOrderBook(ctx, tokenAddress)
	if err != nil {
		return nil, err
	}

	// Calculate market data
	marketData := &MarketData{
		TokenAddress: tokenAddress,
		Price:       priceResp.Data.Price,
		Volume24h:   liquidityData.Volume24h,
		Change24h:   calculatePriceChange(priceResp.Data.Price, liquidityData.TokenAmount),
		MarketCap:   calculateMarketCap(priceResp.Data.Price, liquidityData.TokenAmount),
		Liquidity:   liquidityData.TVL,
		PriceImpact: calculatePriceImpact(orderBook),
		OrderBook:   orderBook,
		MintsPrice:  priceResp.Data.MintsPrice,
		Timestamp:   time.Now(),
	}

	return marketData, nil
}

// GetMarketDepth gets market depth for a token
func (s *MarketDataService) GetMarketDepth(ctx context.Context, tokenAddress string) (*MarketDepth, error) {
	// Get order book from Raydium
	orderBook, err := s.raydiumClient.GetOrderBook(ctx, tokenAddress)
	if err != nil {
		return nil, err
	}

	// Create depth levels
	levels := make([]DepthLevel, 0, len(orderBook.Bids)+len(orderBook.Asks))

	// Add bid levels
	for _, bid := range orderBook.Bids {
		levels = append(levels, DepthLevel{
			Price:     bid.Price,
			Liquidity: bid.Amount,
			Size:      bid.Size,
		})
	}

	// Add ask levels
	for _, ask := range orderBook.Asks {
		levels = append(levels, DepthLevel{
			Price:     ask.Price,
			Liquidity: ask.Amount,
			Size:      ask.Size,
		})
	}

	depth := &MarketDepth{
		TokenAddress: tokenAddress,
		Bids:        orderBook.Bids,
		Asks:        orderBook.Asks,
		Levels:      levels,
		Timestamp:   time.Now(),
	}

	return depth, nil
}

// GetMarketStats gets market statistics for a token
func (s *MarketDataService) GetMarketStats(ctx context.Context, tokenAddress string) (*MarketStats, error) {
	// Get historical data from Raydium
	historical, err := s.raydiumClient.GetHistoricalData(ctx, tokenAddress, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	// Calculate statistics
	stats := &MarketStats{
		TokenAddress:     tokenAddress,
		Price24hHigh:    historical.HighPrice,
		Price24hLow:     historical.LowPrice,
		Volume24h:       historical.Volume,
		Price24hChange:  calculatePriceChange(historical.OpenPrice, historical.Data[len(historical.Data)-1].Price),
		Volume24hChange: calculateVolumeChange(historical.Volume, historical.Data[len(historical.Data)-1].Volume),
		NumTrades24h:    historical.NumTrades,
		AverageTradeSize: historical.Volume / float64(historical.NumTrades),
		Timestamp:       time.Now(),
	}

	return stats, nil
}

// Helper functions

func calculatePriceChange(oldPrice, newPrice float64) float64 {
	if oldPrice == 0 {
		return 0
	}
	return (newPrice - oldPrice) / oldPrice * 100
}

func calculateVolumeChange(oldVolume, newVolume float64) float64 {
	if oldVolume == 0 {
		return 0
	}
	return (newVolume - oldVolume) / oldVolume * 100
}

func calculateMarketCap(price, totalSupply float64) float64 {
	return price * totalSupply
}

func calculatePriceImpact(orderBook *OrderBook) float64 {
	if len(orderBook.Asks) == 0 {
		return 0
	}
	bestAsk := orderBook.Asks[0].Price
	if bestAsk == 0 {
		return 0
	}
	return (orderBook.Asks[len(orderBook.Asks)-1].Price - bestAsk) / bestAsk * 100
}
