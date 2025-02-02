package dydx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// CreateOrder creates a new order
func (c *DefaultClient) CreateOrder(ctx context.Context, req CreateOrderRequest) (*Order, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	// Generate signature
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signature := c.sign(fmt.Sprintf("%s%s%s%s", req.Market, req.Side, timestamp, req.Type))

	// Add authentication headers
	headers := map[string]string{
		"DYDX-SIGNATURE":  signature,
		"DYDX-API-KEY":    c.apiKey,
		"DYDX-TIMESTAMP":  timestamp,
		"DYDX-PASSPHRASE": c.passphrase,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request error: %w", err)
	}

	url := fmt.Sprintf("%s/v3/orders", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}

	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limit exceeded")
	}

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
			return nil, fmt.Errorf("create order failed: %s", errResp.Errors[0].Msg)
		}
		return nil, fmt.Errorf("create order failed with status code: %d", resp.StatusCode)
	}

	var response struct {
		Order Order `json:"order"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return &response.Order, nil
}

// CancelOrder cancels an existing order
func (c *DefaultClient) CancelOrder(ctx context.Context, orderID string) error {
	if err := c.limiter.Wait(ctx); err != nil {
		return err
	}

	// Generate signature
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signature := c.sign(fmt.Sprintf("%s%s", orderID, timestamp))

	// Add authentication headers
	headers := map[string]string{
		"DYDX-SIGNATURE":  signature,
		"DYDX-API-KEY":    c.apiKey,
		"DYDX-TIMESTAMP":  timestamp,
		"DYDX-PASSPHRASE": c.passphrase,
	}

	url := fmt.Sprintf("%s/v3/orders/%s", c.baseURL, orderID)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("rate limit exceeded")
	}

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
			return fmt.Errorf("cancel order failed: %s", errResp.Errors[0].Msg)
		}
		return fmt.Errorf("cancel order failed with status code: %d", resp.StatusCode)
	}

	return nil
}

// GetOrder retrieves an order by ID
func (c *DefaultClient) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v3/orders/%s", c.baseURL, orderID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}

	// Add authentication headers
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signature := c.sign(fmt.Sprintf("%s%s", orderID, timestamp))
	req.Header.Set("DYDX-SIGNATURE", signature)
	req.Header.Set("DYDX-API-KEY", c.apiKey)
	req.Header.Set("DYDX-TIMESTAMP", timestamp)
	req.Header.Set("DYDX-PASSPHRASE", c.passphrase)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("order not found: %s", orderID)
	}

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
			return nil, fmt.Errorf("get order failed: %s", errResp.Errors[0].Msg)
		}
		return nil, fmt.Errorf("get order failed with status code: %d", resp.StatusCode)
	}

	var response struct {
		Order Order `json:"order"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return &response.Order, nil
}

// GetOpenOrders retrieves all open orders
func (c *DefaultClient) GetOpenOrders(ctx context.Context) ([]Order, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v3/orders?status=OPEN", c.baseURL)
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
			return nil, fmt.Errorf("get open orders failed: %s", errResp.Errors[0].Msg)
		}
		return nil, fmt.Errorf("get open orders failed with status code: %d", resp.StatusCode)
	}

	var response struct {
		Orders []Order `json:"orders"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return response.Orders, nil
}
