package models

import "time"

// MarketData represents market data for a token
type MarketData struct {
	ID           string    `json:"id" bson:"_id,omitempty"`
	TokenAddress string    `json:"tokenAddress" bson:"tokenAddress"`
	Price        float64   `json:"price" bson:"price"`
	Volume24h    float64   `json:"volume24h" bson:"volume24h"`
	Change24h    float64   `json:"change24h" bson:"change24h"`
	MarketCap    float64   `json:"marketCap" bson:"marketCap"`
	Liquidity    float64   `json:"liquidity" bson:"liquidity"`
	PriceImpact  float64   `json:"priceImpact" bson:"priceImpact"`
	Timestamp    time.Time `json:"timestamp" bson:"timestamp"`
}

// MarketAnalysis represents market analysis data
type MarketAnalysis struct {
	ID           string    `json:"id" bson:"_id,omitempty"`
	TokenAddress string    `json:"tokenAddress" bson:"tokenAddress"`
	Timestamp    time.Time `json:"timestamp" bson:"timestamp"`
	
	// Price metrics
	Price          float64 `json:"price" bson:"price"`
	PriceChange24h float64 `json:"priceChange24h" bson:"priceChange24h"`
	PriceChange1h  float64 `json:"priceChange1h" bson:"priceChange1h"`
	
	// Volume metrics
	Volume24h     float64 `json:"volume24h" bson:"volume24h"`
	VolumeChange  float64 `json:"volumeChange" bson:"volumeChange"`
	
	// Technical indicators
	RSI14        float64 `json:"rsi14" bson:"rsi14"`
	MACD         float64 `json:"macd" bson:"macd"`
	MACDSignal   float64 `json:"macdSignal" bson:"macdSignal"`
	MACDHist     float64 `json:"macdHist" bson:"macdHist"`
	EMA20        float64 `json:"ema20" bson:"ema20"`
	SMA50        float64 `json:"sma50" bson:"sma50"`
	SMA200       float64 `json:"sma200" bson:"sma200"`
	
	// Market sentiment
	BuyPressure  float64 `json:"buyPressure" bson:"buyPressure"`
	SellPressure float64 `json:"sellPressure" bson:"sellPressure"`
	
	// Liquidity metrics
	Liquidity    float64 `json:"liquidity" bson:"liquidity"`
	Slippage     float64 `json:"slippage" bson:"slippage"`
	
	// Market signals
	Signals      []string `json:"signals" bson:"signals"`
	TrendScore   float64  `json:"trendScore" bson:"trendScore"`
	
	// Risk metrics
	Volatility   float64 `json:"volatility" bson:"volatility"`
	RiskScore    float64 `json:"riskScore" bson:"riskScore"`
	
	// Additional metadata
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// NewMarketData creates a new market data instance
func NewMarketData(tokenAddress string, price float64) *MarketData {
	return &MarketData{
		TokenAddress: tokenAddress,
		Price:       price,
		Timestamp:   time.Now(),
	}
}

// NewMarketAnalysis creates a new market analysis instance
func NewMarketAnalysis(tokenAddress string) *MarketAnalysis {
	now := time.Now()
	return &MarketAnalysis{
		TokenAddress: tokenAddress,
		Timestamp:    now,
		UpdatedAt:    now,
		Signals:      make([]string, 0),
		Metadata:     make(map[string]interface{}),
	}
// DailyStats represents daily trading statistics
type DailyStats struct {
	Date                time.Time `json:"date"`
	TotalTrades         int       `json:"totalTrades"`
	TotalVolume         float64   `json:"totalVolume"`
	AveragePrice        float64   `json:"averagePrice"`
	HighPrice           float64   `json:"highPrice"`
	LowPrice            float64   `json:"lowPrice"`
	OpenPrice           float64   `json:"openPrice"`
	ClosePrice          float64   `json:"closePrice"`
	PriceChange         float64   `json:"priceChange"`
	VolumeChange        float64   `json:"volumeChange"`
	HighWaterMark       float64   `json:"highWaterMark"`
	StartBalance        float64   `json:"startBalance"`
	EndBalance          float64   `json:"endBalance"`
	RealizedPnL         float64   `json:"realizedPnL"`
	Commissions         float64   `json:"commissions"`
	WinningTrades       int       `json:"winningTrades"`
	LosingTrades        int       `json:"losingTrades"`
	LargestWin          float64   `json:"largestWin"`
	LargestLoss         float64   `json:"largestLoss"`
	AverageWin          float64   `json:"averageWin"`
	AverageLoss         float64   `json:"averageLoss"`
	WinRate             float64   `json:"winRate"`
	ProfitFactor        float64   `json:"profitFactor"`
	SharpeRatio         float64   `json:"sharpeRatio"`
	MaxDrawdown         float64   `json:"maxDrawdown"`
	MaxConsecWins       int       `json:"maxConsecWins"`
	MaxConsecLosses     int       `json:"maxConsecLosses"`
	CurrentConsecWins   int       `json:"currentConsecWins"`
	CurrentConsecLosses int       `json:"currentConsecLosses"`
}
