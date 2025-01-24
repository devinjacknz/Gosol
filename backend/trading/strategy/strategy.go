package strategy

import (
	"context"
	"fmt"
	"time"

	"solmeme-trader/dex"
	"solmeme-trader/models"
)

// Signal represents a trading signal
type Signal struct {
	TokenAddress string    `json:"token_address"`
	Action       string    `json:"action"` // "buy", "sell", "hold"
	Price        float64   `json:"price"`
	Size         float64   `json:"size"`
	Confidence   float64   `json:"confidence"` // 0-1
	Reason       string    `json:"reason"`
	Timestamp    time.Time `json:"timestamp"`
}

// StrategyConfig defines strategy parameters
type StrategyConfig struct {
	MinConfidence     float64 `json:"min_confidence"`      // Minimum confidence level to execute trade
	MinVolume         float64 `json:"min_volume"`          // Minimum 24h volume in SOL
	MaxPriceImpact    float64 `json:"max_price_impact"`    // Maximum price impact percentage
	MinLiquidity      float64 `json:"min_liquidity"`       // Minimum liquidity in SOL
	TrendPeriod       string  `json:"trend_period"`        // Time period for trend analysis
	VolatilityPeriod  string  `json:"volatility_period"`   // Time period for volatility calculation
	RSIPeriod         int     `json:"rsi_period"`          // Period for RSI calculation
	RSIOverbought     float64 `json:"rsi_overbought"`      // RSI overbought level
	RSIOversold       float64 `json:"rsi_oversold"`        // RSI oversold level
	EMAPeriod         int     `json:"ema_period"`          // Period for EMA calculation
	StopLoss          float64 `json:"stop_loss"`           // Stop loss percentage
	TakeProfit        float64 `json:"take_profit"`         // Take profit percentage
	TrailingStop      float64 `json:"trailing_stop"`       // Trailing stop percentage
	MaxOpenPositions  int     `json:"max_open_positions"`  // Maximum number of open positions
	MaxPositionSize   float64 `json:"max_position_size"`   // Maximum position size in SOL
	RiskPerTrade      float64 `json:"risk_per_trade"`      // Risk per trade percentage
	InitialCapital    float64 `json:"initial_capital"`     // Initial capital in SOL
	MaxDrawdown       float64 `json:"max_drawdown"`        // Maximum drawdown percentage
}

// Strategy defines the interface for trading strategies
type Strategy interface {
	// Initialize sets up the strategy with configuration and dependencies
	Initialize(config StrategyConfig) error

	// Analyze analyzes market data and generates trading signals
	Analyze(ctx context.Context, marketData *dex.MarketData) (*Signal, error)

	// ValidateSignal validates a trading signal before execution
	ValidateSignal(ctx context.Context, signal *Signal) error

	// OnTradeExecuted is called after a trade is executed
	OnTradeExecuted(position *models.Position) error

	// OnPositionClosed is called when a position is closed
	OnPositionClosed(position *models.Position, pnl float64) error

	// GetStats returns strategy performance statistics
	GetStats() map[string]interface{}
}

// BaseStrategy provides common functionality for strategies
type BaseStrategy struct {
	config     StrategyConfig
	positions  map[string]*models.Position
	signals    []*Signal
	stats      map[string]interface{}
}

// NewBaseStrategy creates a new base strategy
func NewBaseStrategy() *BaseStrategy {
	return &BaseStrategy{
		positions: make(map[string]*models.Position),
		signals:   make([]*Signal, 0),
		stats:    make(map[string]interface{}),
	}
}

// Initialize initializes the base strategy
func (s *BaseStrategy) Initialize(config StrategyConfig) error {
	s.config = config
	return nil
}

// ValidateSignal performs basic signal validation
func (s *BaseStrategy) ValidateSignal(ctx context.Context, signal *Signal) error {
	if signal.Confidence < s.config.MinConfidence {
		return fmt.Errorf("signal confidence %f below minimum %f", signal.Confidence, s.config.MinConfidence)
	}

	if len(s.positions) >= s.config.MaxOpenPositions {
		return fmt.Errorf("maximum open positions reached: %d", s.config.MaxOpenPositions)
	}

	return nil
}

// OnTradeExecuted updates strategy state after trade execution
func (s *BaseStrategy) OnTradeExecuted(position *models.Position) error {
	s.positions[position.TokenAddress] = position
	return nil
}

// OnPositionClosed updates strategy state after position closure
func (s *BaseStrategy) OnPositionClosed(position *models.Position, pnl float64) error {
	delete(s.positions, position.TokenAddress)

	// Update statistics
	s.stats["total_trades"] = s.stats["total_trades"].(int) + 1
	if pnl > 0 {
		s.stats["winning_trades"] = s.stats["winning_trades"].(int) + 1
	} else {
		s.stats["losing_trades"] = s.stats["losing_trades"].(int) + 1
	}
	s.stats["total_pnl"] = s.stats["total_pnl"].(float64) + pnl

	return nil
}

// GetStats returns strategy statistics
func (s *BaseStrategy) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})
	for k, v := range s.stats {
		stats[k] = v
	}
	return stats
}
