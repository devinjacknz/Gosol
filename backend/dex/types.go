package dex

import (
	"time"
)

// QuoteResponse represents a DEX quote response
type QuoteResponse struct {
	InputAmount  float64 `json:"input_amount"`
	OutputAmount float64 `json:"output_amount"`
	Price        float64 `json:"price"`
	PriceImpact  float64 `json:"price_impact"`
	Fee          float64 `json:"fee"`
}

// OrderbookEntry represents an orderbook entry
type OrderbookEntry struct {
	Price  float64 `json:"price"`
	Amount float64 `json:"amount"`
}

// Orderbook represents a DEX orderbook
type Orderbook struct {
	Bids []OrderbookEntry `json:"bids"`
	Asks []OrderbookEntry `json:"asks"`
}

// MarketInfo represents market information from DEX
type MarketInfo struct {
	Address     string    `json:"address"`
	LastPrice   float64   `json:"last_price"`
	BaseVolume  float64   `json:"base_volume"`
	QuoteVolume float64   `json:"quote_volume"`
	Timestamp   time.Time `json:"timestamp"`
}

// LiquidityInfo represents liquidity information from DEX
type LiquidityInfo struct {
	TVL         float64 `json:"tvl"`
	TotalSupply float64 `json:"total_supply"`
}

// Trade represents a trade from DEX
type Trade struct {
	Price     float64   `json:"price"`
	Size      float64   `json:"size"`
	Side      string    `json:"side"`
	Timestamp time.Time `json:"timestamp"`
}

// TokenInfo represents token information
type TokenInfo struct {
	Address     string  `json:"address"`
	Symbol      string  `json:"symbol"`
	Name        string  `json:"name"`
	Decimals    int     `json:"decimals"`
	TotalSupply float64 `json:"total_supply"`
}

// PoolInfo represents liquidity pool information
type PoolInfo struct {
	Address      string  `json:"address"`
	Token0       string  `json:"token0"`
	Token1       string  `json:"token1"`
	Reserve0     float64 `json:"reserve0"`
	Reserve1     float64 `json:"reserve1"`
	TotalSupply  float64 `json:"total_supply"`
	SwapFee      float64 `json:"swap_fee"`
	ProtocolFee  float64 `json:"protocol_fee"`
	LPFee        float64 `json:"lp_fee"`
	TVL          float64 `json:"tvl"`
	Volume24h    float64 `json:"volume_24h"`
	APR          float64 `json:"apr"`
	PriceImpact  float64 `json:"price_impact"`
	Utilization  float64 `json:"utilization"`
	Volatility   float64 `json:"volatility"`
	Correlation  float64 `json:"correlation"`
	LastUpdated  time.Time `json:"last_updated"`
}
