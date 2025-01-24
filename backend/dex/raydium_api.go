package dex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"solmeme-trader/dex/internal"
)

const (
	RaydiumAPIBaseURL = "https://api.raydium.io/v2"
)

type RaydiumAPIClient struct {
	httpClient *http.Client
}

func NewRaydiumAPIClient() *RaydiumAPIClient {
	return &RaydiumAPIClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *RaydiumAPIClient) GetPoolInfo(ctx context.Context, poolId string) (*internal.PoolInfo, error) {
	// Build URL
	url := fmt.Sprintf("%s/pool/%s", RaydiumAPIBaseURL, poolId)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var raydiumResp internal.RaydiumPoolResponse
	if err := json.Unmarshal(body, &raydiumResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Convert to common PoolInfo type
	return &internal.PoolInfo{
		ID:             raydiumResp.ID,
		BaseMint:       raydiumResp.BaseMint,
		QuoteMint:      raydiumResp.QuoteMint,
		LpMint:         raydiumResp.LpMint,
		BaseDecimals:   raydiumResp.BaseDecimals,
		QuoteDecimals:  raydiumResp.QuoteDecimals,
		LpDecimals:     raydiumResp.LpDecimals,
		Version:        raydiumResp.Version,
		ProgramId:      raydiumResp.ProgramId,
		BaseVault:      raydiumResp.BaseVault,
		QuoteVault:     raydiumResp.QuoteVault,
		Authority:      raydiumResp.Authority,
		OpenOrders:     raydiumResp.OpenOrders,
		TargetOrders:   raydiumResp.TargetOrders,
		BaseAmount:     raydiumResp.BaseAmount,
		QuoteAmount:    raydiumResp.QuoteAmount,
		LpSupply:       raydiumResp.LpSupply,
		LastPrice:      raydiumResp.LastPrice,
		Volume24h:      raydiumResp.Volume24h,
		Volume24hQuote: raydiumResp.Volume24hQuote,
		FeeRate:        raydiumResp.FeeRate,
		APR:            raydiumResp.APR,
		Status:         raydiumResp.Status,
		LiquidityUSD:   raydiumResp.LiquidityUSD,
		MarketPrice:    raydiumResp.MarketPrice,
		MarketPriceUSD: raydiumResp.MarketPriceUSD,
		Liquidity:      raydiumResp.LiquidityUSD, // Use USD liquidity as the common metric
	}, nil
}

func (c *RaydiumAPIClient) GetOrderBook(ctx context.Context, poolId string, limit int) (*internal.OrderBook, error) {
	// Build URL
	url := fmt.Sprintf("%s/orderbook/%s?limit=%d", RaydiumAPIBaseURL, poolId, limit)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var raydiumResp internal.RaydiumOrderBookResponse
	if err := json.Unmarshal(body, &raydiumResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &internal.OrderBook{
		Market: raydiumResp.Market,
		Asks:   raydiumResp.Asks,
		Bids:   raydiumResp.Bids,
	}, nil
}

func (c *RaydiumAPIClient) GetTokenInfo(ctx context.Context, tokenMint string) (*internal.TokenInfo, error) {
	// Build URL
	url := fmt.Sprintf("%s/token/%s", RaydiumAPIBaseURL, tokenMint)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var raydiumResp internal.RaydiumTokenResponse
	if err := json.Unmarshal(body, &raydiumResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &internal.TokenInfo{
		Symbol:         raydiumResp.Symbol,
		Name:           raydiumResp.Name,
		Mint:           raydiumResp.Mint,
		Decimals:       raydiumResp.Decimals,
		TotalSupply:    raydiumResp.TotalSupply,
		Price:          raydiumResp.Price,
		Volume24h:      raydiumResp.Volume24h,
		MarketCap:      raydiumResp.MarketCap,
		PriceChange24h: raydiumResp.PriceChange24h,
	}, nil
}

func (c *RaydiumAPIClient) GetLiquidity(ctx context.Context, poolId string) (float64, error) {
	pool, err := c.GetPoolInfo(ctx, poolId)
	if err != nil {
		return 0, err
	}
	return pool.LiquidityUSD, nil
}

func (c *RaydiumAPIClient) GetVolume24h(ctx context.Context, poolId string) (float64, error) {
	pool, err := c.GetPoolInfo(ctx, poolId)
	if err != nil {
		return 0, err
	}
	return pool.Volume24h, nil
}

func (c *RaydiumAPIClient) GetPriceImpact(ctx context.Context, poolId string, amount float64) (float64, error) {
	// Get current order book
	orderBook, err := c.GetOrderBook(ctx, poolId, 100)
	if err != nil {
		return 0, err
	}

	// Calculate price impact based on order book depth
	var totalLiquidity float64
	var weightedPrice float64
	remainingAmount := amount

	for _, level := range orderBook.Asks {
		if remainingAmount <= 0 {
			break
		}

		size := level.Size
		if size > remainingAmount {
			size = remainingAmount
		}

		totalLiquidity += size
		weightedPrice += level.Price * size
		remainingAmount -= size
	}

	if totalLiquidity == 0 {
		return 0, fmt.Errorf("insufficient liquidity")
	}

	averagePrice := weightedPrice / totalLiquidity
	marketPrice := orderBook.Asks[0].Price
	priceImpact := (averagePrice - marketPrice) / marketPrice * 100

	return priceImpact, nil
}

func (c *RaydiumAPIClient) GetMarketDepth(ctx context.Context, poolId string, limit int) (*internal.OrderBook, error) {
	return c.GetOrderBook(ctx, poolId, limit)
}
