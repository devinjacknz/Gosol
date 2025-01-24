package strategy

import (
	"context"
	"fmt"
	"time"

	"solmeme-trader/dex"
	"solmeme-trader/models"
	"solmeme-trader/trading/analysis"
)

// MomentumStrategy implements a momentum-based trading strategy
type MomentumStrategy struct {
	*BaseStrategy
	historicalPrices map[string][]float64
	lastUpdate      map[string]time.Time
}

// NewMomentumStrategy creates a new momentum strategy instance
func NewMomentumStrategy() *MomentumStrategy {
	return &MomentumStrategy{
		BaseStrategy:      NewBaseStrategy(),
		historicalPrices: make(map[string][]float64),
		lastUpdate:      make(map[string]time.Time),
	}
}

// Initialize initializes the momentum strategy
func (s *MomentumStrategy) Initialize(config StrategyConfig) error {
	if err := s.BaseStrategy.Initialize(config); err != nil {
		return err
	}

	// Initialize strategy-specific statistics
	s.stats["total_trades"] = 0
	s.stats["winning_trades"] = 0
	s.stats["losing_trades"] = 0
	s.stats["total_pnl"] = 0.0
	s.stats["win_rate"] = 0.0
	s.stats["avg_profit"] = 0.0
	s.stats["avg_loss"] = 0.0
	s.stats["max_drawdown"] = 0.0

	return nil
}

// Analyze analyzes market data and generates trading signals
func (s *MomentumStrategy) Analyze(ctx context.Context, marketData *dex.MarketData) (*Signal, error) {
	// Update historical prices
	prices, ok := s.historicalPrices[marketData.TokenAddress]
	if !ok {
		prices = make([]float64, 0)
	}
	prices = append(prices, marketData.Price)
	if len(prices) > 100 { // Keep last 100 prices
		prices = prices[1:]
	}
	s.historicalPrices[marketData.TokenAddress] = prices

	// Skip if not enough data
	if len(prices) < s.config.RSIPeriod {
		return nil, nil
	}

	// Calculate technical indicators
	rsi := analysis.CalculateRSI(prices, s.config.RSIPeriod)
	trend := analysis.CalculateTrend(prices, s.config.EMAPeriod)
	volatility := analysis.CalculateVolatility(prices)

	// Check volume and liquidity requirements
	if marketData.Volume24h < s.config.MinVolume {
		return nil, fmt.Errorf("insufficient volume: %f < %f", marketData.Volume24h, s.config.MinVolume)
	}
	if marketData.Liquidity < s.config.MinLiquidity {
		return nil, fmt.Errorf("insufficient liquidity: %f < %f", marketData.Liquidity, s.config.MinLiquidity)
	}

	// Check if we already have a position
	if position, exists := s.positions[marketData.TokenAddress]; exists {
		// Check exit conditions
		if s.shouldExit(position, marketData, rsi, trend, volatility) {
			return &Signal{
				TokenAddress: marketData.TokenAddress,
				Action:      "sell",
				Price:       marketData.Price,
				Size:        position.Size,
				Confidence:  0.8,
				Reason:     "Exit conditions met",
				Timestamp:   time.Now(),
			}, nil
		}
		return nil, nil // Hold position
	}

	// Check entry conditions
	if s.shouldEnter(marketData, rsi, trend, volatility) {
		// Calculate position size based on risk parameters
		size := s.calculatePositionSize(marketData)
		return &Signal{
			TokenAddress: marketData.TokenAddress,
			Action:      "buy",
			Price:       marketData.Price,
			Size:        size,
			Confidence:  0.7,
			Reason:     "Entry conditions met",
			Timestamp:   time.Now(),
		}, nil
	}

	return nil, nil // No signal
}

// shouldExit determines if we should exit a position
func (s *MomentumStrategy) shouldExit(position *models.Position, marketData *dex.MarketData, rsi float64, trend string, volatility float64) bool {
	// Exit on RSI overbought
	if rsi > s.config.RSIOverbought {
		return true
	}

	// Exit on trend reversal
	if position.Side == models.PositionSideLong && trend == "bearish" {
		return true
	}

	// Exit on high volatility
	if volatility > 0.1 { // 10% volatility threshold
		return true
	}

	// Exit on price impact concerns
	if marketData.PriceImpact > s.config.MaxPriceImpact {
		return true
	}

	return false
}

// shouldEnter determines if we should enter a position
func (s *MomentumStrategy) shouldEnter(marketData *dex.MarketData, rsi float64, trend string, volatility float64) bool {
	// Check RSI oversold condition
	if rsi < s.config.RSIOversold {
		return false
	}

	// Check trend
	if trend != "bullish" {
		return false
	}

	// Check volatility
	if volatility > 0.1 { // 10% volatility threshold
		return false
	}

	// Check price impact
	if marketData.PriceImpact > s.config.MaxPriceImpact {
		return false
	}

	return true
}

// calculatePositionSize determines the position size based on risk parameters
func (s *MomentumStrategy) calculatePositionSize(marketData *dex.MarketData) float64 {
	// Calculate base position size using Kelly Criterion
	winRate := s.getWinRate()
	avgWin := s.getAverageWin()
	avgLoss := s.getAverageLoss()
	
	var kellyFraction float64
	if avgLoss != 0 {
		kellyFraction = (winRate*avgWin - (1-winRate)*avgLoss) / avgWin
	} else {
		kellyFraction = 0.5 // Default to 50% if no loss data
	}

	// Limit position size based on configuration
	maxSize := s.config.MaxPositionSize
	riskBasedSize := s.portfolioValue() * (s.config.RiskPerTrade / 100)
	
	// Take the minimum of Kelly size and risk-based size
	size := min(kellyFraction*s.portfolioValue(), riskBasedSize)
	
	// Ensure we don't exceed maximum position size
	return min(size, maxSize)
}

// getWinRate calculates the strategy's win rate
func (s *MomentumStrategy) getWinRate() float64 {
	totalTrades := s.stats["total_trades"].(int)
	if totalTrades == 0 {
		return 0.5 // Default to 50% if no trades
	}
	return float64(s.stats["winning_trades"].(int)) / float64(totalTrades)
}

// getAverageWin calculates the average winning trade
func (s *MomentumStrategy) getAverageWin() float64 {
	winningTrades := s.stats["winning_trades"].(int)
	if winningTrades == 0 {
		return 0
	}
	return s.stats["avg_profit"].(float64)
}

// getAverageLoss calculates the average losing trade
func (s *MomentumStrategy) getAverageLoss() float64 {
	losingTrades := s.stats["losing_trades"].(int)
	if losingTrades == 0 {
		return 0
	}
	return s.stats["avg_loss"].(float64)
}

// portfolioValue returns the current portfolio value
func (s *MomentumStrategy) portfolioValue() float64 {
	// This should be implemented to return actual portfolio value
	// For now, return a placeholder value
	return s.config.InitialCapital
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
