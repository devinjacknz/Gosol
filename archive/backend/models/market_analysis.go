package models

import "time"

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
}
