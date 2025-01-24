package models

import "time"

// AnalysisResult represents the result of market analysis
type AnalysisResult struct {
	ID           string    `json:"id" bson:"_id,omitempty"`
	TokenAddress string    `json:"tokenAddress" bson:"tokenAddress"`
	Timestamp    time.Time `json:"timestamp" bson:"timestamp"`

	// Analysis results
	Signal       string  `json:"signal" bson:"signal"`           // buy, sell, hold
	Confidence   float64 `json:"confidence" bson:"confidence"`   // 0-1 scale
	TrendScore   float64 `json:"trendScore" bson:"trendScore"`  // -1 to 1 scale
	RiskScore    float64 `json:"riskScore" bson:"riskScore"`    // 0-1 scale
	
	// Market conditions
	MarketTrend  string  `json:"marketTrend" bson:"marketTrend"`   // bullish, bearish, neutral
	Volatility   float64 `json:"volatility" bson:"volatility"`     // historical volatility
	Momentum     float64 `json:"momentum" bson:"momentum"`         // momentum indicator
	
	// Technical indicators
	RSI          float64 `json:"rsi" bson:"rsi"`
	MACD         float64 `json:"macd" bson:"macd"`
	MACDSignal   float64 `json:"macdSignal" bson:"macdSignal"`
	MACDHist     float64 `json:"macdHist" bson:"macdHist"`
	
	// Price metrics
	CurrentPrice float64 `json:"currentPrice" bson:"currentPrice"`
	PriceTarget  float64 `json:"priceTarget" bson:"priceTarget"`
	StopLoss     float64 `json:"stopLoss" bson:"stopLoss"`
	
	// Volume analysis
	VolumeProfile float64 `json:"volumeProfile" bson:"volumeProfile"`
	Liquidity    float64 `json:"liquidity" bson:"liquidity"`
	
	// Additional factors
	SentimentScore float64 `json:"sentimentScore" bson:"sentimentScore"`
	NewsImpact     string  `json:"newsImpact" bson:"newsImpact"`
	
	// Metadata
	UpdatedAt     time.Time                 `json:"updatedAt" bson:"updatedAt"`
	Metadata      map[string]interface{}    `json:"metadata,omitempty" bson:"metadata,omitempty"`
	Indicators    map[string]interface{}    `json:"indicators,omitempty" bson:"indicators,omitempty"`
}

// NewAnalysisResult creates a new analysis result instance
func NewAnalysisResult(tokenAddress string) *AnalysisResult {
	now := time.Now()
	return &AnalysisResult{
		TokenAddress: tokenAddress,
		Timestamp:    now,
		UpdatedAt:    now,
		Metadata:     make(map[string]interface{}),
		Indicators:   make(map[string]interface{}),
	}
}
