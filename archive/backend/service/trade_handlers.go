package service

import (
	"encoding/json"
	"net/http"

	"github.com/leonzhao/trading-system/backend/models"
	"github.com/leonzhao/trading-system/backend/monitoring"
)

// TradeRequest represents a trade execution request
type TradeRequest struct {
	TokenAddress string  `json:"tokenAddress"`
	Type        string  `json:"type"`
	Side        string  `json:"side"`
	Amount      float64 `json:"amount"`
	Price       float64 `json:"price,omitempty"`
	MinOutput   float64 `json:"minOutput,omitempty"`
	SlippageBps float64 `json:"slippageBps"`
}

// TradeResponse represents a trade execution response
type TradeResponse struct {
	Success      bool    `json:"success"`
	TradeID      string  `json:"tradeId,omitempty"`
	Amount       float64 `json:"amount"`
	Price        float64 `json:"price"`
	Value        float64 `json:"value"`
	Fee          float64 `json:"fee"`
	TxHash       string  `json:"txHash,omitempty"`
	Error        string  `json:"error,omitempty"`
}

// handleExecuteTrade handles trade execution requests
func (s *Service) handleExecuteTrade(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req TradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.TokenAddress == "" || req.Amount <= 0 {
		http.Error(w, "Invalid trade parameters", http.StatusBadRequest)
		return
	}

	// Get market data to verify trade parameters
	marketData, err := s.dexClient.GetMarketData(r.Context(), req.TokenAddress)
	if err != nil {
		writeJSON(w, TradeResponse{
			Success: false,
			Error:   "Failed to get market data: " + err.Error(),
		})
		return
	}

	// Record trade attempt
	s.monitor.RecordEvent(r.Context(), monitoring.Event{
		Type:     monitoring.MetricTrading,
		Severity: monitoring.SeverityInfo,
		Message:  "Trade execution attempt",
		Details: map[string]interface{}{
			"tokenAddress": req.TokenAddress,
			"type":        req.Type,
			"side":        req.Side,
			"amount":      req.Amount,
			"price":       req.Price,
			"slippageBps": req.SlippageBps,
		},
	})

	// Create trade record
	trade := models.NewTrade(
		req.TokenAddress,
		models.TradeType(req.Type),
		models.TradeSide(req.Side),
		req.Amount,
		marketData.Price,
	)

	// Calculate fees
	trade.Fee = trade.CalculateFee()

	if err := s.repo.SaveTrade(r.Context(), trade); err != nil {
		writeJSON(w, TradeResponse{
			Success: false,
			Error:   "Failed to save trade: " + err.Error(),
		})
		return
	}

	// Record successful trade
	s.monitor.RecordEvent(r.Context(), monitoring.Event{
		Type:     monitoring.MetricTrading,
		Severity: monitoring.SeverityInfo,
		Message:  "Trade executed successfully",
		Details: map[string]interface{}{
			"tradeId":      trade.ID,
			"tokenAddress": trade.TokenAddress,
			"type":        trade.Type,
			"side":        trade.Side,
			"amount":      trade.Amount,
			"price":       trade.Price,
			"value":       trade.Value,
			"fee":         trade.Fee,
		},
	})

	// Return response
	writeJSON(w, TradeResponse{
		Success:  true,
		TradeID:  trade.ID,
		Amount:   trade.Amount,
		Price:    trade.Price,
		Value:    trade.Value,
		Fee:      trade.Fee,
		TxHash:   trade.TxHash,
	})
}
