package models

import (
	"fmt"
	"time"
)

// TradeStatus represents the status of a trade
type TradeStatus string

const (
	TradePending   TradeStatus = "pending"
	TradeExecuted  TradeStatus = "executed"
	TradeFailed    TradeStatus = "failed"
	TradeCancelled TradeStatus = "cancelled"
)

// TradeType represents the type of trade
type TradeType string

const (
	TradeTypeMarket TradeType = "market"
	TradeTypeLimit  TradeType = "limit"
	TradeTypeStop   TradeType = "stop"
)

// TradeSide represents the side of a trade
type TradeSide string

const (
	TradeSideBuy  TradeSide = "buy"
	TradeSideSell TradeSide = "sell"
)

// Trade represents a trade in the system
type Trade struct {
	ID            string      `bson:"_id,omitempty" json:"id"`
	TokenAddress  string      `bson:"token_address" json:"tokenAddress"`
	Type          TradeType   `bson:"type" json:"type"`
	Side          TradeSide   `bson:"side" json:"side"`
	Amount        float64     `bson:"amount" json:"amount"`
	Price         float64     `bson:"price" json:"price"`
	Value         float64     `bson:"value" json:"value"`
	Fee           float64     `bson:"fee" json:"fee"`
	Status        TradeStatus `bson:"status" json:"status"`
	TxHash        string      `bson:"tx_hash,omitempty" json:"txHash,omitempty"`
	ErrorMessage  string      `bson:"error_message,omitempty" json:"errorMessage,omitempty"`
	Timestamp     time.Time   `bson:"timestamp" json:"timestamp"`
	UpdateTime    time.Time   `bson:"update_time" json:"updateTime"`
}

// TradeFilter represents filters for querying trades
type TradeFilter struct {
	TokenAddress string       `json:"tokenAddress,omitempty"`
	Type         []TradeType  `json:"type,omitempty"`
	Side         []TradeSide  `json:"side,omitempty"`
	Status       []TradeStatus `json:"status,omitempty"`
	StartTime    *time.Time   `json:"startTime,omitempty"`
	EndTime      *time.Time   `json:"endTime,omitempty"`
	MinAmount    *float64     `json:"minAmount,omitempty"`
	MaxAmount    *float64     `json:"maxAmount,omitempty"`
}

// TradeStats represents trade statistics
type TradeStats struct {
	TotalTrades      int       `json:"totalTrades"`
	SuccessfulTrades int       `json:"successfulTrades"`
	FailedTrades     int       `json:"failedTrades"`
	TotalVolume      float64   `json:"totalVolume"`
	TotalFees        float64   `json:"totalFees"`
	AverageAmount    float64   `json:"averageAmount"`
	AverageFee       float64   `json:"averageFee"`
	LastTradeTime    time.Time `json:"lastTradeTime"`
}

// NewTrade creates a new trade
func NewTrade(tokenAddress string, tradeType TradeType, side TradeSide, amount, price float64) *Trade {
	now := time.Now()
	return &Trade{
		TokenAddress: tokenAddress,
		Type:        tradeType,
		Side:        side,
		Amount:      amount,
		Price:       price,
		Value:       amount * price,
		Status:      TradePending,
		Timestamp:   now,
		UpdateTime:  now,
	}
}

// CalculateFee calculates the fee for the trade
func (t *Trade) CalculateFee() float64 {
	// Example fee calculation (0.1%)
	return t.Value * 0.001
}

// UpdateStatus updates the trade status
func (t *Trade) UpdateStatus(status TradeStatus) {
	t.Status = status
	t.UpdateTime = time.Now()
}

// SetError sets an error message and updates the status to failed
func (t *Trade) SetError(err error) {
	t.Status = TradeFailed
	t.ErrorMessage = err.Error()
	t.UpdateTime = time.Now()
}

// IsComplete returns true if the trade is in a final state
func (t *Trade) IsComplete() bool {
	return t.Status == TradeExecuted || t.Status == TradeFailed || t.Status == TradeCancelled
}

// IsPending returns true if the trade is pending
func (t *Trade) IsPending() bool {
	return t.Status == TradePending
}

// String returns a string representation of the trade
func (t *Trade) String() string {
	return fmt.Sprintf(
		"Trade{ID: %s, Token: %s, Type: %s, Side: %s, Amount: %.2f, Price: %.2f, Status: %s}",
		t.ID, t.TokenAddress, t.Type, t.Side, t.Amount, t.Price, t.Status,
	)
}
