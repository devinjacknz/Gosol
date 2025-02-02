package hyperliquid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetPositions retrieves all positions
func (c *DefaultClient) GetPositions(ctx context.Context) ([]Position, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v1/positions", c.baseURL)
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
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("get positions failed: %s (code: %d)", errResp.Message, errResp.Code)
	}

	var positions []Position
	if err := json.NewDecoder(resp.Body).Decode(&positions); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return positions, nil
}

// GetBalance retrieves account balance
func (c *DefaultClient) GetBalance(ctx context.Context) (*Balance, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v1/balance", c.baseURL)
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
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("get balance failed: %s (code: %d)", errResp.Message, errResp.Code)
	}

	var balance Balance
	if err := json.NewDecoder(resp.Body).Decode(&balance); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return &balance, nil
}

// GetLeverage retrieves current leverage for a symbol
func (c *DefaultClient) GetLeverage(ctx context.Context, symbol string) (int, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return 0, err
	}

	url := fmt.Sprintf("%s/api/v1/leverage/%s", c.baseURL, symbol)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("create request error: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		return 0, fmt.Errorf("get leverage failed: %s (code: %d)", errResp.Message, errResp.Code)
	}

	var response struct {
		Leverage int `json:"leverage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("decode response error: %w", err)
	}

	return response.Leverage, nil
}

// SetLeverage sets leverage for a symbol
func (c *DefaultClient) SetLeverage(ctx context.Context, symbol string, leverage int) error {
	if err := c.limiter.Wait(ctx); err != nil {
		return err
	}

	body, err := json.Marshal(struct {
		Leverage int `json:"leverage"`
	}{
		Leverage: leverage,
	})
	if err != nil {
		return fmt.Errorf("marshal request error: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/leverage/%s", c.baseURL, symbol)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		return fmt.Errorf("set leverage failed: %s (code: %d)", errResp.Message, errResp.Code)
	}

	return nil
}

// SubscribePositions subscribes to position updates
func (c *DefaultClient) SubscribePositions(ch chan<- Position) error {
	return c.subscribe("positions", ch)
}

// Position update types
const (
	PositionUpdateTypeNew        = "NEW"
	PositionUpdateTypeModified   = "MODIFIED"
	PositionUpdateTypeLiquidated = "LIQUIDATED"
	PositionUpdateTypeClosed     = "CLOSED"
)

// Position margin types
const (
	MarginTypeIsolated = "ISOLATED"
	MarginTypeCross    = "CROSS"
)

// Position side types
const (
	PositionSideLong  = "LONG"
	PositionSideShort = "SHORT"
)
