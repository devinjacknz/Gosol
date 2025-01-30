package dex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// JupiterClient represents a client for interacting with Jupiter DEX
type JupiterClient struct {
	BaseURL    string
	httpClient *http.Client
}

// NewJupiterClient creates a new Jupiter client
func NewJupiterClient(baseURL string) *JupiterClient {
	return &JupiterClient{
		BaseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// GetPrice gets the price for a token
func (c *JupiterClient) GetPrice(ctx context.Context, tokenAddress string) (*PriceResponse, error) {
	endpoint := fmt.Sprintf("%s/v4/price?id=%s", c.BaseURL, tokenAddress)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get price: %w", err)
	}
	defer resp.Body.Close()

	var priceResp PriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&priceResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &priceResp, nil
}

// GetQuote gets a quote for a token swap
func (c *JupiterClient) GetQuote(ctx context.Context, req *QuoteRequest) (*QuoteResponse, error) {
	q := url.Values{}
	q.Add("inputMint", req.InputMint)
	q.Add("outputMint", req.OutputMint)
	q.Add("amount", fmt.Sprintf("%f", req.Amount))
	q.Add("slippageBps", fmt.Sprintf("%f", req.SlippageBps))

	endpoint := fmt.Sprintf("%s/v4/quote?%s", c.BaseURL, q.Encode())
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}
	defer resp.Body.Close()

	var quoteResp QuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&quoteResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &quoteResp, nil
}

// GetTokenInfo gets information about a token
func (c *JupiterClient) GetTokenInfo(ctx context.Context, tokenAddress string) (*TokenInfo, error) {
	endpoint := fmt.Sprintf("%s/v4/tokens/%s", c.BaseURL, tokenAddress)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get token info: %w", err)
	}
	defer resp.Body.Close()

	var tokenInfo TokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &tokenInfo, nil
}

// GetRoutes gets routes for a token swap
func (c *JupiterClient) GetRoutes(ctx context.Context, inputMint, outputMint string, amount float64) (*RouteMap, error) {
	q := url.Values{}
	q.Add("inputMint", inputMint)
	q.Add("outputMint", outputMint)
	q.Add("amount", strconv.FormatFloat(amount, 'f', -1, 64))

	endpoint := fmt.Sprintf("%s/v4/routes?%s", c.BaseURL, q.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get routes: %w", err)
	}
	defer resp.Body.Close()

	var routeMap RouteMap
	if err := json.NewDecoder(resp.Body).Decode(&routeMap); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &routeMap, nil
}
