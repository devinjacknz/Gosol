package market

import (
	"context"
	"math"
)

// LiquidityAnalyzer analyzes market liquidity
type LiquidityAnalyzer struct{}

// NewLiquidityAnalyzer creates a new liquidity analyzer
func NewLiquidityAnalyzer() *LiquidityAnalyzer {
	return &LiquidityAnalyzer{}
}

// LiquidityAnalysis contains liquidity analysis results
type LiquidityAnalysis struct {
	MarketDepth      float64
	BidAskSpread     float64
	OrderBookBalance float64
	ImpactCost       float64
	Slippage         float64
	TimeToFill       float64
}

// Analyze performs liquidity analysis
func (la *LiquidityAnalyzer) Analyze(ctx context.Context, data MarketData) (*LiquidityAnalysis, error) {
	if len(data.OrderBook.Bids) == 0 || len(data.OrderBook.Asks) == 0 {
		return nil, ErrInvalidOrderBook
	}

	// Calculate market depth
	marketDepth := calculateMarketDepth(data.OrderBook)

	// Calculate bid-ask spread
	bidAskSpread := calculateBidAskSpread(data.OrderBook)

	// Calculate order book balance
	orderBookBalance := calculateOrderBookBalance(data.OrderBook)

	// Calculate impact cost
	impactCost := calculateImpactCost(data.OrderBook)

	// Calculate average slippage
	slippage := calculateSlippage(data.OrderBook)

	// Estimate time to fill
	timeToFill := estimateTimeToFill(data.OrderBook, data.Volumes)

	return &LiquidityAnalysis{
		MarketDepth:      marketDepth,
		BidAskSpread:     bidAskSpread,
		OrderBookBalance: orderBookBalance,
		ImpactCost:       impactCost,
		Slippage:         slippage,
		TimeToFill:       timeToFill,
	}, nil
}

func calculateMarketDepth(orderBook OrderBook) float64 {
	bidDepth := 0.0
	askDepth := 0.0

	for _, bid := range orderBook.Bids {
		bidDepth += bid.Price * bid.Amount
	}

	for _, ask := range orderBook.Asks {
		askDepth += ask.Price * ask.Amount
	}

	return bidDepth + askDepth
}

func calculateBidAskSpread(orderBook OrderBook) float64 {
	if len(orderBook.Bids) == 0 || len(orderBook.Asks) == 0 {
		return 0
	}

	bestBid := orderBook.Bids[0].Price
	bestAsk := orderBook.Asks[0].Price

	if bestBid == 0 {
		return 0
	}

	return (bestAsk - bestBid) / bestBid * 100
}

func calculateOrderBookBalance(orderBook OrderBook) float64 {
	bidVolume := 0.0
	askVolume := 0.0

	for _, bid := range orderBook.Bids {
		bidVolume += bid.Amount
	}

	for _, ask := range orderBook.Asks {
		askVolume += ask.Amount
	}

	if bidVolume+askVolume == 0 {
		return 0
	}

	return (bidVolume - askVolume) / (bidVolume + askVolume)
}

func calculateImpactCost(orderBook OrderBook) float64 {
	if len(orderBook.Bids) == 0 || len(orderBook.Asks) == 0 {
		return 0
	}

	// Calculate impact cost for a standard order size (e.g., 1% of market depth)
	marketDepth := calculateMarketDepth(orderBook)
	standardSize := marketDepth * 0.01

	// Calculate weighted average price for the standard size
	remainingSize := standardSize
	totalCost := 0.0

	for _, ask := range orderBook.Asks {
		if remainingSize <= 0 {
			break
		}

		size := math.Min(remainingSize, ask.Amount)
		totalCost += size * ask.Price
		remainingSize -= size
	}

	if standardSize == 0 {
		return 0
	}

	averagePrice := totalCost / standardSize
	midPrice := (orderBook.Bids[0].Price + orderBook.Asks[0].Price) / 2

	return (averagePrice - midPrice) / midPrice * 100
}

func calculateSlippage(orderBook OrderBook) float64 {
	if len(orderBook.Bids) == 0 || len(orderBook.Asks) == 0 {
		return 0
	}

	// Calculate average slippage for different order sizes
	sizes := []float64{0.001, 0.005, 0.01} // 0.1%, 0.5%, 1% of market depth
	marketDepth := calculateMarketDepth(orderBook)

	totalSlippage := 0.0
	for _, sizePercent := range sizes {
		size := marketDepth * sizePercent
		slippage := calculateSlippageForSize(orderBook, size)
		totalSlippage += slippage
	}

	return totalSlippage / float64(len(sizes))
}

func calculateSlippageForSize(orderBook OrderBook, size float64) float64 {
	if size == 0 {
		return 0
	}

	remainingSize := size
	totalCost := 0.0

	for _, ask := range orderBook.Asks {
		if remainingSize <= 0 {
			break
		}

		executedSize := math.Min(remainingSize, ask.Amount)
		totalCost += executedSize * ask.Price
		remainingSize -= executedSize
	}

	averagePrice := totalCost / size
	bestPrice := orderBook.Asks[0].Price

	return (averagePrice - bestPrice) / bestPrice * 100
}

func estimateTimeToFill(orderBook OrderBook, volumes []float64) float64 {
	if len(volumes) == 0 {
		return 0
	}

	// Calculate average volume per second
	averageVolume := 0.0
	for _, volume := range volumes {
		averageVolume += volume
	}
	averageVolume /= float64(len(volumes))

	if averageVolume == 0 {
		return 0
	}

	// Calculate total order book volume
	totalVolume := 0.0
	for _, level := range orderBook.Bids {
		totalVolume += level.Amount
	}
	for _, level := range orderBook.Asks {
		totalVolume += level.Amount
	}

	// Estimate time to fill in seconds
	return totalVolume / averageVolume
}
