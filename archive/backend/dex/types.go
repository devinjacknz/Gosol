package dex

import "time"

// MarketData represents market data for a token
type MarketData struct {
	TokenAddress  string    `json:"token_address"`
	Price         float64   `json:"price"`
	Volume24h     float64   `json:"volume_24h"`
	Change24h     float64   `json:"change_24h"`
	MarketCap     float64   `json:"market_cap"`
	Liquidity     float64   `json:"liquidity"`
	PriceImpact   float64   `json:"price_impact"`
	OrderBook     *OrderBook `json:"order_book"`
	MintsPrice    map[string]float64 `json:"mints_price"`
	Timestamp     time.Time `json:"timestamp"`
}

// MarketDepth represents market depth data
type MarketDepth struct {
	TokenAddress string          `json:"token_address"`
	Bids        []OrderBookItem  `json:"bids"`
	Asks        []OrderBookItem  `json:"asks"`
	Levels      []DepthLevel    `json:"levels"`
	Timestamp   time.Time       `json:"timestamp"`
}

// DepthLevel represents a depth level
type DepthLevel struct {
	Price     float64 `json:"price"`
	Liquidity float64 `json:"liquidity"`
	Size      float64 `json:"size"`
}

// OrderBookItem represents an order book item
type OrderBookItem struct {
	Price  float64 `json:"price"`
	Amount float64 `json:"amount"`
	Size   float64 `json:"size"`
}

// OrderBook represents the full order book
type OrderBook struct {
	Bids []OrderBookItem `json:"bids"`
	Asks []OrderBookItem `json:"asks"`
}

// MarketStats represents market statistics
type MarketStats struct {
	TokenAddress     string    `json:"token_address"`
	Price24hHigh    float64   `json:"price_24h_high"`
	Price24hLow     float64   `json:"price_24h_low"`
	Volume24h       float64   `json:"volume_24h"`
	Price24hChange  float64   `json:"price_24h_change"`
	Volume24hChange float64   `json:"volume_24h_change"`
	HighPrice24h    float64   `json:"high_price_24h"`
	LowPrice24h     float64   `json:"low_price_24h"`
	NumTrades24h    int       `json:"num_trades_24h"`
	AverageTradeSize float64   `json:"average_trade_size"`
	Timestamp       time.Time `json:"timestamp"`
}

// PriceResponse represents a price response from Jupiter
type PriceResponse struct {
	Data struct {
		Price      float64            `json:"price"`
		Timestamp  time.Time          `json:"timestamp"`
		MintsPrice map[string]float64 `json:"mints_price"`
	} `json:"data"`
}

// QuoteRequest represents a quote request for Jupiter
type QuoteRequest struct {
	InputMint    string  `json:"inputMint"`
	OutputMint   string  `json:"outputMint"`
	Amount       float64 `json:"amount"`
	SlippageBps  float64 `json:"slippageBps"`
}

// QuoteResponse represents a quote response from Jupiter
type QuoteResponse struct {
	Data QuoteData `json:"data"`
	InputAmount  float64 `json:"inputAmount"`
	OutputAmount float64 `json:"outputAmount"`
	Price       float64 `json:"price"`
	PriceImpact float64 `json:"priceImpact"`
	Fee         float64 `json:"fee"`
	MarketInfos []MarketInfo `json:"marketInfos"`
}

// QuoteData represents quote data
type QuoteData struct {
	InAmount       float64   `json:"inAmount"`
	OutAmount      float64   `json:"outAmount"`
	Price          float64  `json:"price"`
	PriceImpactPct float64  `json:"priceImpactPct"`
	Routes         []Route   `json:"routes"`
	MarketInfos    []MarketInfo `json:"marketInfos"`
}

// RouteMap represents a route map from Jupiter
type RouteMap struct {
	Data struct {
		InAmount    float64  `json:"inAmount"`
		OutAmount   float64  `json:"outAmount"`
		PriceImpact float64  `json:"priceImpact"`
		Routes      []Route  `json:"routes"`
		MintKeys    []string `json:"mintKeys"`
	} `json:"data"`
}

// Route represents a trading route
type Route struct {
	Address     string  `json:"address"`
	Percentage  float64 `json:"percentage"`
	PriceImpact float64 `json:"priceImpact"`
}

// MarketInfo represents market information
type MarketInfo struct {
	ID            string    `json:"id"`
	Address       string    `json:"address"`
	Name          string    `json:"name"`
	BaseToken     string    `json:"baseToken"`
	QuoteToken    string    `json:"quoteToken"`
	BaseDecimals  int       `json:"baseDecimals"`
	QuoteDecimals int       `json:"quoteDecimals"`
	LastPrice     float64   `json:"lastPrice"`
	BaseVolume    float64   `json:"baseVolume"`
	QuoteVolume   float64   `json:"quoteVolume"`
	Timestamp     time.Time `json:"timestamp"`
}

// LiquidityInfo represents liquidity information
type LiquidityInfo struct {
	TokenAddress string  `json:"tokenAddress"`
	Amount       float64 `json:"amount"`
	Value        float64 `json:"value"`
	TVL          float64 `json:"tvl"`
	TotalSupply  float64 `json:"totalSupply"`
	TokenAmount  float64 `json:"tokenAmount"`
	Volume24h    float64 `json:"volume24h"`
}

// Trade represents a trade
type Trade struct {
	ID          string    `json:"id"`
	Market      string    `json:"market"`
	Side        string    `json:"side"`
	Price       float64   `json:"price"`
	Amount      float64   `json:"amount"`
	Size        float64   `json:"size"`
	Value       float64   `json:"value"`
	Timestamp   time.Time `json:"timestamp"`
}

// PoolInfo represents pool information
type PoolInfo struct {
	Address      string    `json:"address"`
	TokenA       string    `json:"tokenA"`
	TokenB       string    `json:"tokenB"`
	Token0       string    `json:"token0"`
	Token1       string    `json:"token1"`
	ReserveA     float64   `json:"reserveA"`
	ReserveB     float64   `json:"reserveB"`
	Reserve0     float64   `json:"reserve0"`
	Reserve1     float64   `json:"reserve1"`
	TotalSupply  float64   `json:"totalSupply"`
	SwapFee      float64   `json:"swapFee"`
	ProtocolFee  float64   `json:"protocolFee"`
	LPFee        float64   `json:"lpFee"`
	TVL          float64   `json:"tvl"`
	Volume24h    float64   `json:"volume24h"`
	APR          float64   `json:"apr"`
	PriceImpact  float64   `json:"priceImpact"`
	Utilization  float64   `json:"utilization"`
	Volatility   float64   `json:"volatility"`
	Correlation  float64   `json:"correlation"`
	LastUpdated  time.Time `json:"lastUpdated"`
}

// TokenInfo represents token information
type TokenInfo struct {
	Address     string  `json:"address"`
	Symbol      string  `json:"symbol"`
	Name        string  `json:"name"`
	Decimals    int     `json:"decimals"`
	TotalSupply float64 `json:"totalSupply"`
}

// HistoricalData represents historical market data
type HistoricalData struct {
	TokenAddress string        `json:"tokenAddress"`
	Data        []PricePoint  `json:"data"`
	OpenPrice   float64       `json:"openPrice"`
	HighPrice   float64       `json:"highPrice"`
	LowPrice    float64       `json:"lowPrice"`
	Volume      float64       `json:"volume"`
	NumTrades   int          `json:"numTrades"`
	Timestamp   time.Time     `json:"timestamp"`
}

// PricePoint represents a historical price point
type PricePoint struct {
	Price     float64   `json:"price"`
	Volume    float64   `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}
