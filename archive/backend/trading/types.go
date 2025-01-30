package trading

import (
	"context"
	"time"
)

type Trade struct {
	ID         string
	Pair       string
	Size       float64
	EntryPrice float64
	Side       string
	Timestamp  time.Time
}

type RiskConfig struct {
	MaxPositionSize    float64
	DailyLossLimit     float64
	LeverageMultiplier float64
	RiskPerTrade       float64
}

type DailyStats struct {
	DailyPnL     float64
	TradesCount  int
	VolumeTraded float64
	MaxDrawdown  float64
}

type RiskState struct {
	DailyStats map[string]*DailyStats
}

type MarketData struct {
	Pair      string
	Price     float64
	Volume    float64
	Timestamp time.Time
}

type Position struct {
	ID         string
	Pair       string 
	Size       float64
	EntryPrice float64
	Side       string
	Timestamp  time.Time
}

type TradeExecutor interface {
	ExecuteTrade(ctx context.Context, trade *Trade) error
	CancelOrder(orderID string) error
}
