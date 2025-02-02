package hyperliquid

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetMarkets returns all available markets
func (c *DefaultClient) GetMarkets(ctx context.Context) ([]Market, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v1/markets", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var markets []Market
	if err := json.NewDecoder(resp.Body).Decode(&markets); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return markets, nil
}

// GetOrderbook returns the current order book for a market
func (c *DefaultClient) GetOrderbook(ctx context.Context, symbol string) (*Orderbook, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v1/orderbook/%s", c.baseURL, symbol)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var orderbook Orderbook
	if err := json.NewDecoder(resp.Body).Decode(&orderbook); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return &orderbook, nil
}

// GetTrades returns recent trades for a market
func (c *DefaultClient) GetTrades(ctx context.Context, symbol string, limit int) ([]Trade, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v1/trades/%s?limit=%d", c.baseURL, symbol, limit)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var trades []Trade
	if err := json.NewDecoder(resp.Body).Decode(&trades); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return trades, nil
}

// GetFundingRate returns the current funding rate for a market
func (c *DefaultClient) GetFundingRate(ctx context.Context, symbol string) (*FundingRate, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v1/funding/%s", c.baseURL, symbol)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var fundingRate FundingRate
	if err := json.NewDecoder(resp.Body).Decode(&fundingRate); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return &fundingRate, nil
}

// SubscribeOrderbook subscribes to orderbook updates
func (c *DefaultClient) SubscribeOrderbook(symbol string, ch chan<- Orderbook) error {
	topic := fmt.Sprintf("orderbook:%s", symbol)
	return c.subscribe(topic, ch)
}

// SubscribeTrades subscribes to trade updates
func (c *DefaultClient) SubscribeTrades(symbol string, ch chan<- Trade) error {
	topic := fmt.Sprintf("trades:%s", symbol)
	return c.subscribe(topic, ch)
}

// UnsubscribeAll unsubscribes from all channels
func (c *DefaultClient) UnsubscribeAll() error {
	c.subMutex.Lock()
	defer c.subMutex.Unlock()

	msg := struct {
		Type string `json:"type"`
	}{
		Type: "unsubscribe_all",
	}
	return c.wsConn.WriteJSON(msg)
}
