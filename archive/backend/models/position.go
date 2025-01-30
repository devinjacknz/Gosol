package models

import (
	"fmt"
	"time"
)

// Position represents a trading position
type Position struct {
	ID            string    `bson:"_id,omitempty" json:"id"`
	TokenAddress  string    `bson:"token_address" json:"tokenAddress"`
	Side          string    `bson:"side" json:"side"`
	EntryPrice    float64   `bson:"entry_price" json:"entryPrice"`
	CurrentPrice  float64   `bson:"current_price" json:"currentPrice"`
	Size          float64   `bson:"size" json:"size"`
	Amount        float64   `bson:"amount" json:"amount"`
	Leverage      float64   `bson:"leverage" json:"leverage"`
	Value         float64   `bson:"value" json:"value"`
	Commission    float64   `bson:"commission" json:"commission"`
	UnrealizedPnL float64   `bson:"unrealized_pnl" json:"unrealizedPnL"`
	RealizedPnL   float64   `bson:"realized_pnl" json:"realizedPnL"`
	StopLoss      float64   `bson:"stop_loss" json:"stopLoss"`
	TakeProfit    float64   `bson:"take_profit" json:"takeProfit"`
	Status        string    `bson:"status" json:"status"`
	OpenTime      time.Time `bson:"open_time" json:"openTime"`
	CloseTime     time.Time `bson:"close_time,omitempty" json:"closeTime,omitempty"`
	LastUpdated   time.Time `bson:"last_updated" json:"lastUpdated"`
}

// UpdateValue updates the position's value and PnL
func (p *Position) UpdateValue(currentPrice float64) {
	p.CurrentPrice = currentPrice
	p.Value = p.Size * currentPrice
	if p.Side == "long" {
		p.UnrealizedPnL = (currentPrice - p.EntryPrice) * p.Size
	} else {
		p.UnrealizedPnL = (p.EntryPrice - currentPrice) * p.Size
	}
	p.LastUpdated = time.Now()
}

// ShouldLiquidate checks if the position should be liquidated
func (p *Position) ShouldLiquidate(liquidationThreshold float64) bool {
	if p.Side == "long" {
		return p.CurrentPrice <= p.EntryPrice*(1-liquidationThreshold)
	}
	return p.CurrentPrice >= p.EntryPrice*(1+liquidationThreshold)
}

// GetValue returns the current value of the position
func (p *Position) GetValue() float64 {
	return p.Value
}

// UpdatePrice updates the current price of the position
func (p *Position) UpdatePrice(price float64) {
	p.CurrentPrice = price
	p.UpdateValue(price)
}

// PositionFilter represents filters for querying positions
type PositionFilter struct {
	TokenAddress string     `json:"tokenAddress,omitempty"`
	Side        string     `json:"side,omitempty"`
	Status      string     `json:"status,omitempty"`
	StartTime   *time.Time `json:"startTime,omitempty"`
	EndTime     *time.Time `json:"endTime,omitempty"`
}

// PositionStats represents position statistics
type PositionStats struct {
	TotalPositions     int       `json:"totalPositions"`
	OpenPositions      int       `json:"openPositions"`
	ClosedPositions    int       `json:"closedPositions"`
	TotalValue         float64   `json:"totalValue"`
	UnrealizedPnL      float64   `json:"unrealizedPnL"`
	RealizedPnL        float64   `json:"realizedPnL"`
	AverageHoldingTime float64   `json:"averageHoldingTime"`
	LastUpdated        time.Time `json:"lastUpdated"`
}

// TradeSignal represents a trading signal
type TradeSignalOld struct {
	TokenAddress string    `json:"tokenAddress"`
	Type        string    `json:"type"`
	Side        string    `json:"side"`
	Price       float64   `json:"price"`
	Amount      float64   `json:"amount"`
	Size        float64   `json:"size"`
	TargetPrice float64   `json:"targetPrice"`
	Action      string    `json:"action"`
	Confidence  float64   `json:"confidence"`
	Timestamp   time.Time `json:"timestamp"`
}

// TradeSignalMessage represents a trade signal message
type TradeSignalMessage struct {
	Signal    *TradeSignal `json:"signal"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// TradeExecutor represents a trade executor
type TradeExecutor interface {
	ExecuteTrade(signal *TradeSignal) error
	GetOpenPositions() ([]*Position, error)
	ClosePosition(positionID string) error
}

// String returns a string representation of the position
func (p *Position) String() string {
	return fmt.Sprintf(
		"Position{ID: %s, Token: %s, Side: %s, Size: %.2f, Entry: %.2f, Current: %.2f, PnL: %.2f}",
		p.ID, p.TokenAddress, p.Side, p.Size, p.EntryPrice, p.CurrentPrice, p.UnrealizedPnL,
	)
}
