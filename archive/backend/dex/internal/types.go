package internal

import (
	"time"
)

// Common types shared across DEX implementations

type QuoteResponse struct {
	InputMint    string
	OutputMint   string
	InAmount     uint64
	OutAmount    string
	Price        float64
	PriceImpact  float64
	Source       string // "jupiter" or "raydium"
	Route        interface{}
}

type MarketDepth struct {
	Asks []PriceLevel
	Bids []PriceLevel
}

type PriceLevel struct {
	Price     float64
	Size      float64
	Source    string
	Liquidity float64
}

type OrderBook struct {
	Market string
	Asks   []OrderItem
	Bids   []OrderItem
}

type OrderItem struct {
	Price     float64
	Size      float64
	Quantity  float64
	Liquidity float64
}

type PoolInfo struct {
	ID              string
	BaseMint        string
	QuoteMint       string
	LpMint          string
	BaseDecimals    int
	QuoteDecimals   int
	LpDecimals      int
	Version         int
	ProgramId       string
	BaseVault       string
	QuoteVault      string
	Authority       string
	OpenOrders      string
	TargetOrders    string
	BaseAmount      float64
	QuoteAmount     float64
	LpSupply        float64
	LastPrice       float64
	Volume24h       float64
	Volume24hQuote  float64
	FeeRate         float64
	APR             float64
	Status          string
	Liquidity       float64
	LiquidityUSD    float64
	MarketPrice     float64
	MarketPriceUSD  float64
}

type TokenInfo struct {
	Symbol         string
	Name           string
	Mint           string
	Decimals       int
	TotalSupply    float64
	Price          float64
	Volume24h      float64
	MarketCap      float64
	Liquidity      float64
	PriceChange24h float64
}

// Jupiter-specific types
type JupiterQuoteResponse struct {
	InAmount            string   `json:"inAmount"`
	OutAmount           string   `json:"outAmount"`
	PriceImpactPct     float64  `json:"priceImpactPct"`
	MarketInfos        []Market `json:"marketInfos"`
	SlippageBps        int      `json:"slippageBps"`
	OtherAmountThreshold string  `json:"otherAmountThreshold"`
}

type Market struct {
	ID            string  `json:"id"`
	Label         string  `json:"label"`
	InputMint     string  `json:"inputMint"`
	OutputMint    string  `json:"outputMint"`
	NotEnoughLiquidity bool `json:"notEnoughLiquidity"`
	InAmount      string  `json:"inAmount"`
	OutAmount     string  `json:"outAmount"`
	MinInAmount   string  `json:"minInAmount"`
	MinOutAmount  string  `json:"minOutAmount"`
	PriceImpactPct float64 `json:"priceImpactPct"`
	LpFee         struct {
		Amount float64 `json:"amount"`
		Mint   string  `json:"mint"`
	} `json:"lpFee"`
	PlatformFee struct {
		Amount float64 `json:"amount"`
		Mint   string  `json:"mint"`
	} `json:"platformFee"`
}

type JupiterPriceResponse struct {
	Data struct {
		ID            string  `json:"id"`
		MintsPrice    float64 `json:"mintsPrice"`
		Price         float64 `json:"price"`
		PriceChange   float64 `json:"priceChange"`
		Volume24h     float64 `json:"volume24h"`
		MarketCap     float64 `json:"marketCap"`
		Liquidity     float64 `json:"liquidity"`
		LiquidityBN   string  `json:"liquidityBN"`
		LastTradeTime int64   `json:"lastTradeTime"`
	} `json:"data"`
}

// Raydium-specific types
type RaydiumPoolResponse struct {
	ID              string  `json:"id"`
	BaseMint        string  `json:"baseMint"`
	QuoteMint       string  `json:"quoteMint"`
	LpMint          string  `json:"lpMint"`
	BaseDecimals    int     `json:"baseDecimals"`
	QuoteDecimals   int     `json:"quoteDecimals"`
	LpDecimals      int     `json:"lpDecimals"`
	Version         int     `json:"version"`
	ProgramId       string  `json:"programId"`
	BaseVault       string  `json:"baseVault"`
	QuoteVault      string  `json:"quoteVault"`
	Authority       string  `json:"authority"`
	OpenOrders      string  `json:"openOrders"`
	TargetOrders    string  `json:"targetOrders"`
	BaseAmount      float64 `json:"baseAmount"`
	QuoteAmount     float64 `json:"quoteAmount"`
	LpSupply        float64 `json:"lpSupply"`
	LastPrice       float64 `json:"lastPrice"`
	Volume24h       float64 `json:"volume24h"`
	Volume24hQuote  float64 `json:"volume24hQuote"`
	FeeRate         float64 `json:"feeRate"`
	APR             float64 `json:"apr"`
	Status          string  `json:"status"`
	LiquidityUSD    float64 `json:"liquidityUSD"`
	MarketPrice     float64 `json:"marketPrice"`
	MarketPriceUSD  float64 `json:"marketPriceUSD"`
}

type RaydiumOrderBookResponse struct {
	Market string      `json:"market"`
	Asks   []OrderItem `json:"asks"`
	Bids   []OrderItem `json:"bids"`
}

type RaydiumTokenResponse struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Mint          string  `json:"mint"`
	Decimals      int     `json:"decimals"`
	TotalSupply   float64 `json:"totalSupply"`
	Price         float64 `json:"price"`
	PriceChange24h float64 `json:"priceChange24h"`
	Volume24h     float64 `json:"volume24h"`
	MarketCap     float64 `json:"marketCap"`
}

type CacheEntry struct {
	Data      interface{}
	Timestamp time.Time
}
