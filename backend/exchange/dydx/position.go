package dydx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// GetPositions retrieves all positions
func (c *DefaultClient) GetPositions(ctx context.Context) ([]Position, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v3/positions", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}

	// Add authentication headers
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signature := c.sign(timestamp)
	req.Header.Set("DYDX-SIGNATURE", signature)
	req.Header.Set("DYDX-API-KEY", c.apiKey)
	req.Header.Set("DYDX-TIMESTAMP", timestamp)
	req.Header.Set("DYDX-PASSPHRASE", c.passphrase)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Errors []struct {
				Msg string `json:"msg"`
			} `json:"errors"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		if len(errResp.Errors) > 0 {
			return nil, fmt.Errorf("get positions failed: %s", errResp.Errors[0].Msg)
		}
		return nil, fmt.Errorf("get positions failed with status code: %d", resp.StatusCode)
	}

	var response struct {
		Positions []Position `json:"positions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return response.Positions, nil
}

// GetLeverage retrieves current leverage for a symbol
func (c *DefaultClient) GetLeverage(ctx context.Context, symbol string) (int, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return 0, err
	}

	url := fmt.Sprintf("%s/v3/configs/leverage/%s", c.baseURL, symbol)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("create request error: %w", err)
	}

	// Add authentication headers
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signature := c.sign(fmt.Sprintf("%s%s", symbol, timestamp))
	req.Header.Set("DYDX-SIGNATURE", signature)
	req.Header.Set("DYDX-API-KEY", c.apiKey)
	req.Header.Set("DYDX-TIMESTAMP", timestamp)
	req.Header.Set("DYDX-PASSPHRASE", c.passphrase)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Errors []struct {
				Msg string `json:"msg"`
			} `json:"errors"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		if len(errResp.Errors) > 0 {
			return 0, fmt.Errorf("get leverage failed: %s", errResp.Errors[0].Msg)
		}
		return 0, fmt.Errorf("get leverage failed with status code: %d", resp.StatusCode)
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

	url := fmt.Sprintf("%s/v3/configs/leverage/%s", c.baseURL, symbol)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}

	// Add authentication headers
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signature := c.sign(fmt.Sprintf("%s%d%s", symbol, leverage, timestamp))
	req.Header.Set("DYDX-SIGNATURE", signature)
	req.Header.Set("DYDX-API-KEY", c.apiKey)
	req.Header.Set("DYDX-TIMESTAMP", timestamp)
	req.Header.Set("DYDX-PASSPHRASE", c.passphrase)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Errors []struct {
				Msg string `json:"msg"`
			} `json:"errors"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		if len(errResp.Errors) > 0 {
			return fmt.Errorf("set leverage failed: %s", errResp.Errors[0].Msg)
		}
		return fmt.Errorf("set leverage failed with status code: %d", resp.StatusCode)
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
