package hyperliquid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Account represents account information
type Account struct {
	Email           string    `json:"email"`
	AccountID       string    `json:"accountId"`
	AccountType     string    `json:"accountType"`
	MarginType      string    `json:"marginType"`
	TradingEnabled  bool      `json:"tradingEnabled"`
	CreatedAt       time.Time `json:"createdAt"`
	LastLoginAt     time.Time `json:"lastLoginAt"`
	TradingFeeTier  int       `json:"tradingFeeTier"`
	MakerFeeRate    float64   `json:"makerFeeRate"`
	TakerFeeRate    float64   `json:"takerFeeRate"`
	WithdrawEnabled bool      `json:"withdrawEnabled"`
}

// APIKey represents API key information
type APIKey struct {
	KeyID       string    `json:"keyId"`
	Name        string    `json:"name"`
	Permissions []string  `json:"permissions"`
	IpWhitelist []string  `json:"ipWhitelist"`
	CreatedAt   time.Time `json:"createdAt"`
	LastUsedAt  time.Time `json:"lastUsedAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

// GetAccount retrieves account information
func (c *DefaultClient) GetAccount(ctx context.Context) (*Account, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v1/account", c.baseURL)
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
		return nil, fmt.Errorf("get account failed: %s (code: %d)", errResp.Message, errResp.Code)
	}

	var account Account
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return &account, nil
}

// GetAPIKeys retrieves all API keys
func (c *DefaultClient) GetAPIKeys(ctx context.Context) ([]APIKey, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v1/api-keys", c.baseURL)
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
		return nil, fmt.Errorf("get api keys failed: %s (code: %d)", errResp.Message, errResp.Code)
	}

	var keys []APIKey
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return keys, nil
}

// CreateAPIKey creates a new API key
func (c *DefaultClient) CreateAPIKey(ctx context.Context, name string, permissions []string, ipWhitelist []string) (*APIKey, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	body, err := json.Marshal(struct {
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
		IpWhitelist []string `json:"ipWhitelist"`
	}{
		Name:        name,
		Permissions: permissions,
		IpWhitelist: ipWhitelist,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request error: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/api-keys", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

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
		return nil, fmt.Errorf("create api key failed: %s (code: %d)", errResp.Message, errResp.Code)
	}

	var key APIKey
	if err := json.NewDecoder(resp.Body).Decode(&key); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return &key, nil
}

// DeleteAPIKey deletes an API key
func (c *DefaultClient) DeleteAPIKey(ctx context.Context, keyID string) error {
	if err := c.limiter.Wait(ctx); err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/api-keys/%s", c.baseURL, keyID)
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
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		return fmt.Errorf("delete api key failed: %s (code: %d)", errResp.Message, errResp.Code)
	}

	return nil
}

// Account types
const (
	AccountTypeSpot    = "SPOT"
	AccountTypeFutures = "FUTURES"
	AccountTypeMargin  = "MARGIN"
)

// API key permissions
const (
	PermissionRead     = "READ"
	PermissionTrade    = "TRADE"
	PermissionWithdraw = "WITHDRAW"
	PermissionTransfer = "TRANSFER"
)
