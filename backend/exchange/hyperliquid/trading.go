package hyperliquid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// CreateOrder creates a new order
func (c *DefaultClient) CreateOrder(ctx context.Context, req CreateOrderRequest) (*Order, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request error: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/order", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("create order failed: %s (code: %d)", errResp.Message, errResp.Code)
	}

	var order Order
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return &order, nil
}

// CancelOrder cancels an existing order
func (c *DefaultClient) CancelOrder(ctx context.Context, orderID string) error {
	if err := c.limiter.Wait(ctx); err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/order/%s", c.baseURL, orderID)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		return fmt.Errorf("cancel order failed: %s (code: %d)", errResp.Message, errResp.Code)
	}

	return nil
}

// GetOrder retrieves an order by ID
func (c *DefaultClient) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v1/order/%s", c.baseURL, orderID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}

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
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("get order failed: %s (code: %d)", errResp.Message, errResp.Code)
	}

	var order Order
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return &order, nil
}

// GetOpenOrders retrieves all open orders
func (c *DefaultClient) GetOpenOrders(ctx context.Context) ([]Order, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v1/orders/open", c.baseURL)
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
		var errResp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("get open orders failed: %s (code: %d)", errResp.Message, errResp.Code)
	}

	var orders []Order
	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return orders, nil
}

// Error response types
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error codes
const (
	ErrCodeInvalidRequest     = 400
	ErrCodeUnauthorized       = 401
	ErrCodeInsufficientFunds  = 402
	ErrCodeOrderNotFound      = 404
	ErrCodeRateLimitExceeded  = 429
	ErrCodeInternalError      = 500
	ErrCodeServiceUnavailable = 503
)
