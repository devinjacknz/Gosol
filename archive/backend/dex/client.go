package dex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// DexClient defines the interface for DEX operations
type DexClient interface {
	// Market data operations
	GetMarketData(ctx context.Context, tokenAddress string) (*MarketData, error)
	GetOrderBook(ctx context.Context, tokenAddress string) (*OrderBook, error)
	GetQuote(ctx context.Context, tokenAddress string, amount float64) (float64, error)

	// Trading operations
	ExecuteTrade(ctx context.Context, tokenAddress string, amount float64, side string) error
	CancelTrade(ctx context.Context, tradeID string) error

	// Health check
	Ping(ctx context.Context) error
}

// BaseClient provides common functionality for DEX clients
type BaseClient struct {
	baseURL     string
	httpClient  *http.Client
	apiKey      string
	rateLimit   time.Duration
	lastRequest time.Time
	mu          sync.Mutex
}

// NewBaseClient creates a new base DEX client
func NewBaseClient(baseURL, apiKey string) *BaseClient {
	return &BaseClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		rateLimit: 100 * time.Millisecond, // 10 requests per second
	}
}

// checkRateLimit enforces rate limiting
func (c *BaseClient) checkRateLimit() {
	c.mu.Lock()
	defer c.mu.Unlock()

	timeSinceLastRequest := time.Since(c.lastRequest)
	if timeSinceLastRequest < c.rateLimit {
		time.Sleep(c.rateLimit - timeSinceLastRequest)
	}
	c.lastRequest = time.Now()
}

// doRequest performs an HTTP request with rate limiting
func (c *BaseClient) doRequest(req *http.Request) (*http.Response, error) {
	c.checkRateLimit()
	
	req.Header.Set("X-API-Key", c.apiKey)
	if req.Method == http.MethodPost || req.Method == http.MethodPut {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp, nil
}

// get performs a GET request
func (c *BaseClient) get(ctx context.Context, endpoint string) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	return c.doRequest(req)
}

// post performs a POST request
func (c *BaseClient) post(ctx context.Context, endpoint string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return c.doRequest(req)
}

// delete performs a DELETE request
func (c *BaseClient) delete(ctx context.Context, endpoint string) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	return c.doRequest(req)
}
