package dydx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// GetMarkets retrieves all available markets
func (c *DefaultClient) GetMarkets(ctx context.Context) ([]Market, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v3/markets", c.baseURL)
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

	var response struct {
		Markets map[string]Market `json:"markets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	markets := make([]Market, 0, len(response.Markets))
	for _, market := range response.Markets {
		markets = append(markets, market)
	}

	return markets, nil
}

// GetOrderbook retrieves the current order book for a market
func (c *DefaultClient) GetOrderbook(ctx context.Context, symbol string) (*Orderbook, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v3/orderbook/%s", c.baseURL, symbol)
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

	var response struct {
		Orderbook struct {
			Bids [][]string `json:"bids"`
			Asks [][]string `json:"asks"`
		} `json:"orderbook"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	orderbook := &Orderbook{
		Market: symbol,
		Bids:   make([]OrderbookLevel, len(response.Orderbook.Bids)),
		Asks:   make([]OrderbookLevel, len(response.Orderbook.Asks)),
	}

	for i, bid := range response.Orderbook.Bids {
		price, err := strconv.ParseFloat(bid[0], 64)
		if err != nil {
			return nil, fmt.Errorf("parse bid price error: %w", err)
		}
		size, err := strconv.ParseFloat(bid[1], 64)
		if err != nil {
			return nil, fmt.Errorf("parse bid size error: %w", err)
		}
		orderbook.Bids[i] = OrderbookLevel{
			Price: price,
			Size:  size,
		}
	}

	for i, ask := range response.Orderbook.Asks {
		price, err := strconv.ParseFloat(ask[0], 64)
		if err != nil {
			return nil, fmt.Errorf("parse ask price error: %w", err)
		}
		size, err := strconv.ParseFloat(ask[1], 64)
		if err != nil {
			return nil, fmt.Errorf("parse ask size error: %w", err)
		}
		orderbook.Asks[i] = OrderbookLevel{
			Price: price,
			Size:  size,
		}
	}

	return orderbook, nil
}

// GetTrades retrieves recent trades for a market
func (c *DefaultClient) GetTrades(ctx context.Context, symbol string, limit int) ([]Trade, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v3/trades/%s?limit=%d", c.baseURL, symbol, limit)
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

	var response struct {
		Trades []Trade `json:"trades"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return response.Trades, nil
}

// GetFundingRate retrieves the current funding rate for a market
func (c *DefaultClient) GetFundingRate(ctx context.Context, symbol string) (*FundingRate, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v3/funding-rates/%s/current", c.baseURL, symbol)
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

	var response struct {
		FundingRate FundingRate `json:"fundingRate"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return &response.FundingRate, nil
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
