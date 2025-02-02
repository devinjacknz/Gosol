package dydx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// GetAccount retrieves account information
func (c *DefaultClient) GetAccount(ctx context.Context) (*Account, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v3/accounts", c.baseURL)
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
			return nil, fmt.Errorf("get account failed: %s", errResp.Errors[0].Msg)
		}
		return nil, fmt.Errorf("get account failed with status code: %d", resp.StatusCode)
	}

	var response struct {
		Account Account `json:"account"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return &response.Account, nil
}

// GetBalance retrieves account balance
func (c *DefaultClient) GetBalance(ctx context.Context) (*Balance, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v3/accounts/balance", c.baseURL)
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
			return nil, fmt.Errorf("get balance failed: %s", errResp.Errors[0].Msg)
		}
		return nil, fmt.Errorf("get balance failed with status code: %d", resp.StatusCode)
	}

	var response struct {
		Balance Balance `json:"balance"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return &response.Balance, nil
}

// Account status types
const (
	AccountStatusActive     = "ACTIVE"
	AccountStatusDisabled   = "DISABLED"
	AccountStatusLocked     = "LOCKED"
	AccountStatusLiquidated = "LIQUIDATED"
)

// Account types
const (
	AccountTypeSpot    = "SPOT"
	AccountTypeFutures = "FUTURES"
	AccountTypeMargin  = "MARGIN"
)

// Account permission types
const (
	PermissionTrade    = "TRADE"
	PermissionTransfer = "TRANSFER"
	PermissionWithdraw = "WITHDRAW"
)
