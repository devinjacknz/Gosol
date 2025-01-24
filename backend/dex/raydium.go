package dex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// RaydiumClient represents a client for interacting with Raydium DEX
type RaydiumClient struct {
	BaseURL    string
	httpClient *http.Client
}

// NewRaydiumClient creates a new Raydium client
func NewRaydiumClient(baseURL string) *RaydiumClient {
	return &RaydiumClient{
		BaseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// GetLiquidity gets liquidity information for a token
func (c *RaydiumClient) GetLiquidity(ctx context.Context, tokenAddress string) (*LiquidityInfo, error) {
	endpoint := fmt.Sprintf("%s/v4/liquidity/%s", c.BaseURL, tokenAddress)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get liquidity: %w", err)
	}
	defer resp.Body.Close()

	var liquidity LiquidityInfo
	if err := json.NewDecoder(resp.Body).Decode(&liquidity); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &liquidity, nil
}

// GetOrderBook gets the order book for a token
func (c *RaydiumClient) GetOrderBook(ctx context.Context, tokenAddress string) (*OrderBook, error) {
	endpoint := fmt.Sprintf("%s/v4/orderbook/%s", c.BaseURL, tokenAddress)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get order book: %w", err)
	}
	defer resp.Body.Close()

	var orderBook OrderBook
	if err := json.NewDecoder(resp.Body).Decode(&orderBook); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &orderBook, nil
}

// GetHistoricalData gets historical data for a token
func (c *RaydiumClient) GetHistoricalData(ctx context.Context, tokenAddress string, period time.Duration) (*HistoricalData, error) {
	endpoint := fmt.Sprintf("%s/v4/historical/%s?period=%d", c.BaseURL, tokenAddress, int64(period.Seconds()))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}
	defer resp.Body.Close()

	var historical HistoricalData
	if err := json.NewDecoder(resp.Body).Decode(&historical); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &historical, nil
}

// GetPoolInfo gets information about a liquidity pool
func (c *RaydiumClient) GetPoolInfo(ctx context.Context, poolAddress string) (*PoolInfo, error) {
	endpoint := fmt.Sprintf("%s/v4/pools/%s", c.BaseURL, poolAddress)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool info: %w", err)
	}
	defer resp.Body.Close()

	var poolInfo PoolInfo
	if err := json.NewDecoder(resp.Body).Decode(&poolInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &poolInfo, nil
}

// GetRecentTrades gets recent trades for a token
func (c *RaydiumClient) GetRecentTrades(ctx context.Context, tokenAddress string, limit int) ([]Trade, error) {
	endpoint := fmt.Sprintf("%s/v4/trades/%s?limit=%d", c.BaseURL, tokenAddress, limit)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent trades: %w", err)
	}
	defer resp.Body.Close()

	var trades []Trade
	if err := json.NewDecoder(resp.Body).Decode(&trades); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return trades, nil
}
