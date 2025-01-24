package dex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	jupiterBaseURL  = "https://price.jup.ag/v4"
	jupiterQuoteURL = "https://quote-api.jup.ag/v4"
)

// JupiterClient handles interactions with Jupiter DEX API
type JupiterClient struct {
	httpClient *http.Client
}

// NewJupiterClient creates a new Jupiter API client
func NewJupiterClient() *JupiterClient {
	return &JupiterClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetPrice fetches the current price for a token
func (c *JupiterClient) GetPrice(ctx context.Context, mintAddress string) (*PriceResponse, error) {
	url := fmt.Sprintf("%s/price?ids=%s", jupiterBaseURL, mintAddress)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var priceResp PriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&priceResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &priceResp, nil
}

// GetQuote fetches a quote for a swap
func (c *JupiterClient) GetQuote(ctx context.Context, req QuoteRequest) (*QuoteResponse, error) {
	url := fmt.Sprintf("%s/quote", jupiterQuoteURL)
	
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := httpReq.URL.Query()
	q.Add("inputMint", req.InputMint)
	q.Add("outputMint", req.OutputMint)
	q.Add("amount", req.Amount)
	q.Add("slippageBps", fmt.Sprintf("%d", req.SlippageBps))
	httpReq.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var quoteResp QuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&quoteResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &quoteResp, nil
}

// GetRouteMap fetches the available trading routes
func (c *JupiterClient) GetRouteMap(ctx context.Context) (*RouteMap, error) {
	url := fmt.Sprintf("%s/indexed-route-map", jupiterQuoteURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var routeMap RouteMap
	if err := json.NewDecoder(resp.Body).Decode(&routeMap); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &routeMap, nil
}

// HasRoute checks if a trading route exists between two tokens
func (c *JupiterClient) HasRoute(ctx context.Context, inputMint, outputMint string) (bool, error) {
	routeMap, err := c.GetRouteMap(ctx)
	if err != nil {
		return false, err
	}

	// Find indices of input and output mints
	var inputIdx, outputIdx int
	found := false
	for i, mint := range routeMap.Data.MintKeys {
		if mint == inputMint {
			inputIdx = i
			found = true
			break
		}
	}
	if !found {
		return false, nil
	}

	found = false
	for i, mint := range routeMap.Data.MintKeys {
		if mint == outputMint {
			outputIdx = i
			found = true
			break
		}
	}
	if !found {
		return false, nil
	}

	// Check if there's a route between the mints
	routes, ok := routeMap.Data.Routes[fmt.Sprintf("%d", inputIdx)]
	if !ok {
		return false, nil
	}

	for _, route := range routes {
		if route == fmt.Sprintf("%d", outputIdx) {
			return true, nil
		}
	}

	return false, nil
}

// CalculatePriceImpact calculates the price impact for a given trade
func (c *JupiterClient) CalculatePriceImpact(ctx context.Context, inputMint string, outputMint string, amount string) (float64, error) {
	quote, err := c.GetQuote(ctx, QuoteRequest{
		InputMint:   inputMint,
		OutputMint:  outputMint,
		Amount:      amount,
		SlippageBps: 100, // 1% slippage
	})
	if err != nil {
		return 0, err
	}

	return quote.Data.PriceImpactPct, nil
}

// GetBestRoute finds the optimal trading route for a swap
func (c *JupiterClient) GetBestRoute(ctx context.Context, inputMint string, outputMint string, amount string) (string, error) {
	quote, err := c.GetQuote(ctx, QuoteRequest{
		InputMint:   inputMint,
		OutputMint:  outputMint,
		Amount:      amount,
		SlippageBps: 100, // 1% slippage
	})
	if err != nil {
		return "", err
	}

	if len(quote.Data.MarketInfos) == 0 {
		return "", fmt.Errorf("no routes available")
	}

	// Return the ID of the first market info (Jupiter already sorts by best route)
	return quote.Data.MarketInfos[0].ID, nil
}
