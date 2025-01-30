package models

import "time"

type MarketData struct {
	Symbol        string
	TokenAddress  string
	OpenPrice     float64
	ClosePrice    float64
	HighPrice     float64
	LowPrice      float64
	Volume        float64
	Volume24h     float64
	MarketCap     float64
	Liquidity     float64
	PriceImpact   float64
	OrderBook     struct {
		Bids [][]float64
		Asks [][]float64
	}
	Timestamp     time.Time
}
