package models

import "time"

// MarketData represents real-time market data
type MarketData struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Volume    float64   `json:"volume"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Open      float64   `json:"open"`
	Close     float64   `json:"close"`
	Timestamp time.Time `json:"timestamp"`
}

// Kline represents candlestick data
type Kline struct {
	Symbol    string    `json:"symbol"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
	Interval  string    `json:"interval"` // 1m, 5m, 15m, 1h, etc.
}

// OrderBook represents market depth
type OrderBook struct {
	Symbol    string          `json:"symbol"`
	Bids      []OrderBookItem `json:"bids"`
	Asks      []OrderBookItem `json:"asks"`
	Timestamp time.Time       `json:"timestamp"`
}

// OrderBookItem represents a single order book entry
type OrderBookItem struct {
	Price  float64 `json:"price"`
	Amount float64 `json:"amount"`
}

// Trade represents a market trade
type Trade struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Amount    float64   `json:"amount"`
	Side      string    `json:"side"` // buy or sell
	Timestamp time.Time `json:"timestamp"`
}
