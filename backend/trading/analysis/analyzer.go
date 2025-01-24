package analysis

import (
	"context"
	"math"
	"sort"
	"time"
)

// MarketData represents a single market data point
type MarketData struct {
	Price     float64   `json:"price"`
	Volume    float64   `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

// Trade represents a completed trade
type Trade struct {
	EntryPrice float64   `json:"entry_price"`
	ExitPrice  float64   `json:"exit_price"`
	Size       float64   `json:"size"`
	Profit     float64   `json:"profit"`
	Timestamp  time.Time `json:"timestamp"`
}

// TradingMetrics represents calculated trading performance metrics
type TradingMetrics struct {
	TotalTrades      int     `json:"total_trades"`
	ProfitableTrades int     `json:"profitable_trades"`
	WinRate          float64 `json:"win_rate"`
	ProfitFactor     float64 `json:"profit_factor"`
	TotalProfit      float64 `json:"total_profit"`
	AverageWin       float64 `json:"average_win"`
	AverageLoss      float64 `json:"average_loss"`
	SharpeRatio      float64 `json:"sharpe_ratio"`
	MaxDrawdown      float64 `json:"max_drawdown"`
}

// Report represents a comprehensive market analysis report
type Report struct {
	Trend           string    `json:"trend"`
	TrendStrength   float64   `json:"trend_strength"`
	Volatility      float64   `json:"volatility"`
	Support         float64   `json:"support"`
	Resistance      float64   `json:"resistance"`
	PredictedPrice  float64   `json:"predicted_price"`
	Confidence      float64   `json:"confidence"`
	RecommendedSize float64   `json:"recommended_size"`
	Timestamp       time.Time `json:"timestamp"`
}

// MarketAnalyzer provides market analysis functionality
type MarketAnalyzer struct {
	historicalData []MarketData
}

// NewMarketAnalyzer creates a new market analyzer instance
func NewMarketAnalyzer() *MarketAnalyzer {
	return &MarketAnalyzer{
		historicalData: make([]MarketData, 0),
	}
}

// AddMarketData adds a new market data point to the historical data
func (a *MarketAnalyzer) AddMarketData(data MarketData) {
	a.historicalData = append(a.historicalData, data)
}

// CalculateMetrics calculates trading performance metrics
func (a *MarketAnalyzer) CalculateMetrics(trades []Trade) TradingMetrics {
	metrics := TradingMetrics{}

	if len(trades) == 0 {
		return metrics
	}

	var totalWins float64
	var totalLosses float64
	var winCount int
	var lossCount int

	for _, trade := range trades {
		if trade.Profit > 0 {
			totalWins += trade.Profit
			winCount++
		} else {
			totalLosses -= trade.Profit
			lossCount++
		}
	}

	metrics.TotalTrades = len(trades)
	metrics.ProfitableTrades = winCount
	metrics.WinRate = float64(winCount) / float64(metrics.TotalTrades)
	metrics.ProfitFactor = totalWins / totalLosses
	metrics.TotalProfit = totalWins - totalLosses

	if winCount > 0 {
		metrics.AverageWin = totalWins / float64(winCount)
	}
	if lossCount > 0 {
		metrics.AverageLoss = totalLosses / float64(lossCount)
	}

	metrics.SharpeRatio = a.calculateSharpeRatio(trades)
	metrics.MaxDrawdown = a.calculateMaxDrawdown(trades)

	return metrics
}

// AnalyzeTrend analyzes market trend and returns trend direction and strength
func (a *MarketAnalyzer) AnalyzeTrend(timeframe string) (string, float64) {
	if len(a.historicalData) < 2 {
		return "neutral", 0.0
	}

	// Calculate moving averages
	shortMA := a.calculateMA(10) // 10-period MA
	longMA := a.calculateMA(30)  // 30-period MA

	if len(shortMA) < 2 || len(longMA) < 2 {
		return "neutral", 0.0
	}

	// Calculate trend direction and strength
	lastShortMA := shortMA[len(shortMA)-1]
	lastLongMA := longMA[len(longMA)-1]
	prevShortMA := shortMA[len(shortMA)-2]
	prevLongMA := longMA[len(longMA)-2]

	// Calculate trend strength based on MA crossover and price movement
	strength := math.Abs((lastShortMA - lastLongMA) / lastLongMA)

	if lastShortMA > lastLongMA && prevShortMA > prevLongMA {
		return "bullish", strength
	} else if lastShortMA < lastLongMA && prevShortMA < prevLongMA {
		return "bearish", strength
	}

	return "neutral", strength
}

// AnalyzeVolatility calculates market volatility
func (a *MarketAnalyzer) AnalyzeVolatility(timeframe string) float64 {
	if len(a.historicalData) < 2 {
		return 0.0
	}

	// Calculate price returns
	returns := make([]float64, len(a.historicalData)-1)
	for i := 1; i < len(a.historicalData); i++ {
		returns[i-1] = (a.historicalData[i].Price - a.historicalData[i-1].Price) / a.historicalData[i-1].Price
	}

	// Calculate standard deviation of returns
	var mean float64
	for _, r := range returns {
		mean += r
	}
	mean /= float64(len(returns))

	var variance float64
	for _, r := range returns {
		variance += math.Pow(r-mean, 2)
	}
	variance /= float64(len(returns))

	return math.Sqrt(variance)
}

// GenerateReport generates a comprehensive market analysis report
func (a *MarketAnalyzer) GenerateReport(ctx context.Context) (*Report, error) {
	if len(a.historicalData) < 30 {
		return nil, nil
	}

	trend, strength := a.AnalyzeTrend("1d")
	volatility := a.AnalyzeVolatility("1d")
	support, resistance := a.calculateSupportResistance()
	predictedPrice, confidence := a.predictNextPrice()

	return &Report{
		Trend:           trend,
		TrendStrength:   strength,
		Volatility:      volatility,
		Support:         support,
		Resistance:      resistance,
		PredictedPrice:  predictedPrice,
		Confidence:      confidence,
		RecommendedSize: a.calculateRecommendedSize(volatility),
		Timestamp:       time.Now(),
	}, nil
}

// calculateMA calculates moving average for the specified period
func (a *MarketAnalyzer) calculateMA(period int) []float64 {
	if len(a.historicalData) < period {
		return nil
	}

	ma := make([]float64, len(a.historicalData)-period+1)
	for i := 0; i <= len(a.historicalData)-period; i++ {
		sum := 0.0
		for j := 0; j < period; j++ {
			sum += a.historicalData[i+j].Price
		}
		ma[i] = sum / float64(period)
	}

	return ma
}

// calculateMaxDrawdown calculates the maximum drawdown from trade history
func (a *MarketAnalyzer) calculateMaxDrawdown(trades []Trade) float64 {
	if len(trades) == 0 {
		return 0.0
	}

	peak := trades[0].Profit
	maxDrawdown := 0.0
	runningTotal := trades[0].Profit

	for i := 1; i < len(trades); i++ {
		runningTotal += trades[i].Profit
		if runningTotal > peak {
			peak = runningTotal
		}
		drawdown := (peak - runningTotal) / peak
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown
}

// calculateSharpeRatio calculates the Sharpe ratio from trade history
func (a *MarketAnalyzer) calculateSharpeRatio(trades []Trade) float64 {
	if len(trades) < 2 {
		return 0.0
	}

	var totalReturn float64
	var returns []float64

	for _, trade := range trades {
		returns = append(returns, trade.Profit)
		totalReturn += trade.Profit
	}

	averageReturn := totalReturn / float64(len(trades))

	var variance float64
	for _, r := range returns {
		variance += math.Pow(r-averageReturn, 2)
	}
	variance /= float64(len(trades) - 1)

	// Assuming risk-free rate is 0 for simplicity
	return averageReturn / math.Sqrt(variance)
}

// calculateSupportResistance calculates support and resistance levels
func (a *MarketAnalyzer) calculateSupportResistance() (float64, float64) {
	if len(a.historicalData) < 10 {
		return 0.0, 0.0
	}

	prices := make([]float64, len(a.historicalData))
	for i, data := range a.historicalData {
		prices[i] = data.Price
	}
	sort.Float64s(prices)

	// Use percentiles for support and resistance
	supportIdx := int(float64(len(prices)) * 0.1)    // 10th percentile
	resistanceIdx := int(float64(len(prices)) * 0.9) // 90th percentile

	return prices[supportIdx], prices[resistanceIdx]
}

// predictNextPrice predicts the next price and returns confidence level
func (a *MarketAnalyzer) predictNextPrice() (float64, float64) {
	if len(a.historicalData) < 30 {
		return 0.0, 0.0
	}

	// Simple prediction using weighted moving average
	weights := []float64{0.4, 0.3, 0.2, 0.1}
	periods := []int{5, 10, 20, 30}

	var predictedPrice float64
	var totalWeight float64

	for i, period := range periods {
		ma := a.calculateMA(period)
		if len(ma) > 0 {
			predictedPrice += ma[len(ma)-1] * weights[i]
			totalWeight += weights[i]
		}
	}

	if totalWeight > 0 {
		predictedPrice /= totalWeight
	}

	// Calculate confidence based on volatility and trend strength
	_, trendStrength := a.AnalyzeTrend("1d")
	volatility := a.AnalyzeVolatility("1d")

	// Higher trend strength and lower volatility = higher confidence
	confidence := trendStrength * (1 - volatility)
	confidence = math.Max(0.1, math.Min(0.9, confidence))

	return predictedPrice, confidence
}

// calculateRecommendedSize calculates recommended position size based on volatility
func (a *MarketAnalyzer) calculateRecommendedSize(volatility float64) float64 {
	// Base size inversely proportional to volatility
	baseSize := 1.0 / (1.0 + volatility*10)

	// Adjust based on trend strength
	_, trendStrength := a.AnalyzeTrend("1d")
	return baseSize * (1.0 + trendStrength)
}
