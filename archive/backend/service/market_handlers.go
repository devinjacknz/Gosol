package service

import (
	"net/http"
	"strconv"
)

// MarketDataResponse represents a market data response
type MarketDataResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// handleGetMarketData handles market data requests
func (s *Service) handleGetMarketData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokenAddress := r.URL.Query().Get("token")
	if tokenAddress == "" {
		http.Error(w, "Token address is required", http.StatusBadRequest)
		return
	}

	marketData, err := s.dexClient.GetMarketData(r.Context(), tokenAddress)
	if err != nil {
		writeJSON(w, MarketDataResponse{
			Success: false,
			Error:   "Failed to get market data: " + err.Error(),
		})
		return
	}

	writeJSON(w, MarketDataResponse{
		Success: true,
		Data:    marketData,
	})
}

// handleGetOrderBook handles order book requests
func (s *Service) handleGetOrderBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokenAddress := r.URL.Query().Get("token")
	if tokenAddress == "" {
		http.Error(w, "Token address is required", http.StatusBadRequest)
		return
	}

	orderBook, err := s.dexClient.GetOrderBook(r.Context(), tokenAddress)
	if err != nil {
		writeJSON(w, MarketDataResponse{
			Success: false,
			Error:   "Failed to get order book: " + err.Error(),
		})
		return
	}

	writeJSON(w, MarketDataResponse{
		Success: true,
		Data:    orderBook,
	})
}

// handleGetQuote handles quote requests
func (s *Service) handleGetQuote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokenAddress := r.URL.Query().Get("token")
	if tokenAddress == "" {
		http.Error(w, "Token address is required", http.StatusBadRequest)
		return
	}

	amountStr := r.URL.Query().Get("amount")
	if amountStr == "" {
		http.Error(w, "Amount is required", http.StatusBadRequest)
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	price, err := s.dexClient.GetQuote(r.Context(), tokenAddress, amount)
	if err != nil {
		writeJSON(w, MarketDataResponse{
			Success: false,
			Error:   "Failed to get quote: " + err.Error(),
		})
		return
	}

	writeJSON(w, MarketDataResponse{
		Success: true,
		Data: map[string]interface{}{
			"price": price,
		},
	})
}
