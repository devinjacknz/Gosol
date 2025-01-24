package risk

import (
	"time"
)

// Position represents a trading position
type Position struct {
	ID           string    `json:"id"`
	TokenAddress string    `json:"token_address"`
	Side         string    `json:"side"` // "long" or "short"
	EntryPrice   float64   `json:"entry_price"`
	CurrentPrice float64   `json:"current_price"`
	Size         float64   `json:"size"`
	Leverage     float64   `json:"leverage"`
	PnL          float64   `json:"pnl"`
	OpenTime     time.Time `json:"open_time"`
	UpdateTime   time.Time `json:"update_time"`
	StopLoss     *float64  `json:"stop_loss,omitempty"`
	TakeProfit   *float64  `json:"take_profit,omitempty"`
}

// RiskConfig defines risk management configuration
type RiskConfig struct {
	MaxPositions    int     `json:"max_positions"`
	MaxPositionSize float64 `json:"max_position_size"`
	MaxLeverage     float64 `json:"max_leverage"`
	MaxDrawdown     float64 `json:"max_drawdown"`
	InitialCapital  float64 `json:"initial_capital"`
	RiskPerTrade    float64 `json:"risk_per_trade"`
	StopLoss        float64 `json:"stop_loss"`
	TakeProfit      float64 `json:"take_profit"`
}
