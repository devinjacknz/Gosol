package strategy

import (
	"math"
	"sort"
)

// GridStrategy implements a grid trading strategy
type GridStrategy struct {
	upperPrice float64
	lowerPrice float64
	gridLevels int
	gridSize   float64
	positions  map[float64]float64 // price -> size
}

// NewGridStrategy creates a new grid trading strategy
func NewGridStrategy(upper, lower float64, levels int) *GridStrategy {
	return &GridStrategy{
		upperPrice: upper,
		lowerPrice: lower,
		gridLevels: levels,
		gridSize:   (upper - lower) / float64(levels-1),
		positions:  make(map[float64]float64),
	}
}

// CalculateGridLevels returns all price levels in the grid
func (g *GridStrategy) CalculateGridLevels() []float64 {
	levels := make([]float64, g.gridLevels)
	for i := 0; i < g.gridLevels; i++ {
		levels[i] = g.lowerPrice + float64(i)*g.gridSize
	}
	return levels
}

// EvaluateGridPosition determines if a trade should be made based on current price
func (g *GridStrategy) EvaluateGridPosition(currentPrice float64) (string, float64) {
	levels := g.CalculateGridLevels()

	// Find nearest grid levels
	var lowerLevel, upperLevel float64
	for i := 0; i < len(levels)-1; i++ {
		if currentPrice >= levels[i] && currentPrice <= levels[i+1] {
			lowerLevel = levels[i]
			upperLevel = levels[i+1]
			break
		}
	}

	// Calculate distances to nearest levels
	distToLower := math.Abs(currentPrice - lowerLevel)
	distToUpper := math.Abs(currentPrice - upperLevel)

	if distToLower < distToUpper {
		return "buy", lowerLevel
	}
	return "sell", upperLevel
}

// ArbitrageStrategy implements cross-exchange arbitrage
type ArbitrageStrategy struct {
	minProfitThreshold float64
	maxSlippage        float64
}

// NewArbitrageStrategy creates a new arbitrage strategy
func NewArbitrageStrategy(minProfit, maxSlip float64) *ArbitrageStrategy {
	return &ArbitrageStrategy{
		minProfitThreshold: minProfit,
		maxSlippage:        maxSlip,
	}
}

// FindArbitrageOpportunity looks for profitable arbitrage opportunities
func (a *ArbitrageStrategy) FindArbitrageOpportunity(prices map[string]float64) (string, string, float64) {
	if len(prices) < 2 {
		return "", "", 0
	}

	// Convert prices map to sorted slice
	type PriceEntry struct {
		exchange string
		price    float64
	}

	var priceList []PriceEntry
	for ex, price := range prices {
		priceList = append(priceList, PriceEntry{ex, price})
	}

	sort.Slice(priceList, func(i, j int) bool {
		return priceList[i].price < priceList[j].price
	})

	// Find largest price difference
	bestBuy := priceList[0]
	bestSell := priceList[len(priceList)-1]

	profitPerc := (bestSell.price - bestBuy.price) / bestBuy.price * 100

	if profitPerc > a.minProfitThreshold && profitPerc <= a.maxSlippage {
		return bestBuy.exchange, bestSell.exchange, profitPerc
	}

	return "", "", 0
}

// TrendFollowingStrategy implements trend following strategy
type TrendFollowingStrategy struct {
	shortPeriod   int
	longPeriod    int
	trendStrength float64
}

// NewTrendFollowingStrategy creates a new trend following strategy
func NewTrendFollowingStrategy(shortPeriod, longPeriod int, strength float64) *TrendFollowingStrategy {
	return &TrendFollowingStrategy{
		shortPeriod:   shortPeriod,
		longPeriod:    longPeriod,
		trendStrength: strength,
	}
}

// AnalyzeTrend determines the trend direction and strength
func (t *TrendFollowingStrategy) AnalyzeTrend(prices []float64) (string, float64) {
	if len(prices) < t.longPeriod {
		return "neutral", 0
	}

	// Calculate short and long moving averages
	shortMA := calculateMA(prices[len(prices)-t.shortPeriod:], t.shortPeriod)
	longMA := calculateMA(prices[len(prices)-t.longPeriod:], t.longPeriod)

	// Calculate trend strength
	strength := math.Abs(shortMA-longMA) / longMA * 100

	if strength >= t.trendStrength {
		if shortMA > longMA {
			return "uptrend", strength
		}
		return "downtrend", strength
	}

	return "neutral", strength
}

// ReversalStrategy implements price reversal strategy
type ReversalStrategy struct {
	rsiPeriod     int
	rsiOverbought float64
	rsiOversold   float64
}

// NewReversalStrategy creates a new reversal strategy
func NewReversalStrategy(period int, overbought, oversold float64) *ReversalStrategy {
	return &ReversalStrategy{
		rsiPeriod:     period,
		rsiOverbought: overbought,
		rsiOversold:   oversold,
	}
}

// FindReversalSignal looks for potential price reversals
func (r *ReversalStrategy) FindReversalSignal(prices []float64) string {
	if len(prices) < r.rsiPeriod {
		return "neutral"
	}

	rsi := calculateRSI(prices, r.rsiPeriod)

	if rsi >= r.rsiOverbought {
		return "sell"
	} else if rsi <= r.rsiOversold {
		return "buy"
	}

	return "neutral"
}

// Helper functions

func calculateMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	var sum float64
	for i := 0; i < period; i++ {
		sum += prices[len(prices)-1-i]
	}

	return sum / float64(period)
}

func calculateRSI(prices []float64, period int) float64 {
	if len(prices) <= period {
		return 50 // Neutral RSI
	}

	var gains, losses float64
	for i := 1; i < period+1; i++ {
		change := prices[len(prices)-i] - prices[len(prices)-i-1]
		if change >= 0 {
			gains += change
		} else {
			losses -= change
		}
	}

	if losses == 0 {
		return 100
	}

	rs := gains / losses
	return 100 - (100 / (1 + rs))
}
