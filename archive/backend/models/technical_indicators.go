package models

import "time"

// 交易信号类型
type SignalType int

const (
	Buy SignalType = iota + 1
	Sell
	Hold
)

// TechnicalIndicators contains calculated technical analysis values
type TechnicalIndicators struct {
	Timestamp       time.Time
	Price           float64
	SMA             float64
	EMA             float64
	RSI             float64
	MACD            MACD
	BollingerBands  BollingerBands
	Signal          SignalType // 新增信号类型字段
}

// MACD represents Moving Average Convergence Divergence values
type MACD struct {
	MACDLine   float64
	SignalLine float64
	Histogram  float64
}

// BollingerBands contains Bollinger Bands values
type BollingerBands struct {
	Upper  float64
	Middle float64
	Lower  float64
}
