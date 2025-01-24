package models

import "time"

// TechnicalIndicators contains calculated technical analysis values
type TechnicalIndicators struct {
	Timestamp       time.Time
	Price           float64
	SMA             float64
	EMA             float64
	RSI             float64
	MACD            MACD
	BollingerBands  BollingerBands
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
