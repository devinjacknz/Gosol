package trading

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"solmeme-trader/models"
)

type RiskManager struct {
	config     *RiskConfig
	state      *RiskState
	mu         sync.RWMutex
	executor   *TradeExecutor
}

type RiskConfig struct {
	MaxDrawdown       float64       `json:"max_drawdown"`        // Maximum allowed drawdown percentage
	MaxDailyLoss     float64       `json:"max_daily_loss"`      // Maximum daily loss percentage
	MaxPositionSize  float64       `json:"max_position_size"`   // Maximum position size as percentage of portfolio
	MinPositionSize  float64       `json:"min_position_size"`   // Minimum position size in SOL
	MaxDailyTrades   int           `json:"max_daily_trades"`    // Maximum number of trades per day
	RiskPerTrade     float64       `json:"risk_per_trade"`      // Risk per trade as percentage of portfolio
	StopLossBuffer   float64       `json:"stop_loss_buffer"`    // Additional buffer for stop loss
	TakeProfitRatio  float64       `json:"take_profit_ratio"`   // Take profit as ratio of risk
	CooldownPeriod   time.Duration `json:"cooldown_period"`     // Cooldown period after losses
	VolatilityLimit  float64       `json:"volatility_limit"`    // Maximum allowed volatility
	LiquidityLimit   float64       `json:"liquidity_limit"`     // Minimum required liquidity
	ConfidenceLimit  float64       `json:"confidence_limit"`    // Minimum required confidence for trades
}

type RiskState struct {
	CurrentDrawdown    float64
	DailyLoss         float64
	DailyTradeCount   int
	LastTradeTime     time.Time
	ConsecutiveLosses int
	InCooldown        bool
	CooldownUntil     time.Time
	DailyStats        map[string]*DailyStats // key: YYYY-MM-DD
}

type DailyStats struct {
	Date            time.Time
	TradeCount      int
	WinCount        int
	LossCount       int
	Volume          float64
	ProfitLoss      float64
	MaxDrawdown     float64
	StartBalance    float64
	EndBalance      float64
	LargestWin      float64
	LargestLoss     float64
	WinRate         float64
	AverageWin      float64
	AverageLoss     float64
}

func NewRiskManager(config *RiskConfig, executor *TradeExecutor) *RiskManager {
	return &RiskManager{
		config:   config,
		executor: executor,
		state: &RiskState{
			DailyStats: make(map[string]*DailyStats),
		},
	}
}

func (r *RiskManager) ValidateTradeSignal(ctx context.Context, signal *TradeSignal, marketData *models.MarketData) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if in cooldown
	if r.state.InCooldown && time.Now().Before(r.state.CooldownUntil) {
		return fmt.Errorf("trading suspended: in cooldown period until %v", r.state.CooldownUntil)
	}

	// Check daily trade limit
	today := time.Now().Format("2006-01-02")
	if stats, exists := r.state.DailyStats[today]; exists {
		if stats.TradeCount >= r.config.MaxDailyTrades {
			return fmt.Errorf("daily trade limit reached: %d trades", r.config.MaxDailyTrades)
		}
	}

	// Check drawdown limit
	if r.state.CurrentDrawdown <= -r.config.MaxDrawdown {
		return fmt.Errorf("maximum drawdown reached: %.2f%%", r.state.CurrentDrawdown*100)
	}

	// Check daily loss limit
	if r.state.DailyLoss <= -r.config.MaxDailyLoss {
		return fmt.Errorf("maximum daily loss reached: %.2f%%", r.state.DailyLoss*100)
	}

	// Validate market conditions
	if err := r.validateMarketConditions(marketData); err != nil {
		return fmt.Errorf("market conditions not met: %v", err)
	}

	// Validate signal confidence
	if signal.Confidence < r.config.ConfidenceLimit {
		return fmt.Errorf("signal confidence too low: %.2f < %.2f", signal.Confidence, r.config.ConfidenceLimit)
	}

	// Calculate and validate position size
	size, err := r.calculatePositionSize(signal, marketData)
	if err != nil {
		return fmt.Errorf("position size calculation failed: %v", err)
	}

	if size < r.config.MinPositionSize {
		return fmt.Errorf("position size too small: %.4f < %.4f", size, r.config.MinPositionSize)
	}

	// Set stop loss and take profit levels
	signal.StopLoss = r.calculateStopLoss(signal.Action, signal.ExpectedPrice, marketData)
	signal.TakeProfit = r.calculateTakeProfit(signal.Action, signal.ExpectedPrice, signal.StopLoss)

	return nil
}

func (r *RiskManager) validateMarketConditions(marketData *models.MarketData) error {
	// Check liquidity
	if marketData.Liquidity < r.config.LiquidityLimit {
		return fmt.Errorf("insufficient liquidity: %.2f < %.2f", marketData.Liquidity, r.config.LiquidityLimit)
	}

	// Check volatility
	if marketData.PriceImpact > r.config.VolatilityLimit {
		return fmt.Errorf("volatility too high: %.2f > %.2f", marketData.PriceImpact, r.config.VolatilityLimit)
	}

	return nil
}

func (r *RiskManager) calculatePositionSize(signal *TradeSignal, marketData *models.MarketData) (float64, error) {
	// Get account balance
	balance, err := r.executor.getWalletBalance()
	if err != nil {
		return 0, fmt.Errorf("failed to get wallet balance: %v", err)
	}

	// Calculate risk amount
	riskAmount := balance * r.config.RiskPerTrade

	// Calculate position size based on stop loss
	stopLoss := r.calculateStopLoss(signal.Action, signal.ExpectedPrice, marketData)
	stopDistance := math.Abs(signal.ExpectedPrice - stopLoss)
	if stopDistance == 0 {
		return 0, fmt.Errorf("invalid stop loss: same as entry price")
	}

	size := riskAmount / stopDistance

	// Apply maximum position size limit
	maxSize := balance * r.config.MaxPositionSize
	if size > maxSize {
		size = maxSize
	}

	return size, nil
}

func (r *RiskManager) calculateStopLoss(action string, entryPrice float64, marketData *models.MarketData) float64 {
	// Calculate technical stop loss based on market conditions
	var technicalStop float64
	if action == "BUY" {
		technicalStop = entryPrice * (1 - r.config.StopLossBuffer)
		// Consider support levels from market data
		if marketData.OrderBook != nil && len(marketData.OrderBook.Bids) > 0 {
			supportLevel := marketData.OrderBook.Bids[0].Price
			if supportLevel < technicalStop {
				technicalStop = supportLevel
			}
		}
	} else {
		technicalStop = entryPrice * (1 + r.config.StopLossBuffer)
		// Consider resistance levels from market data
		if marketData.OrderBook != nil && len(marketData.OrderBook.Asks) > 0 {
			resistanceLevel := marketData.OrderBook.Asks[0].Price
			if resistanceLevel > technicalStop {
				technicalStop = resistanceLevel
			}
		}
	}

	return technicalStop
}

func (r *RiskManager) calculateTakeProfit(action string, entryPrice, stopLoss float64) float64 {
	riskDistance := math.Abs(entryPrice - stopLoss)
	if action == "BUY" {
		return entryPrice + (riskDistance * r.config.TakeProfitRatio)
	}
	return entryPrice - (riskDistance * r.config.TakeProfitRatio)
}

func (r *RiskManager) UpdateStats(trade *models.Trade) {
	r.mu.Lock()
	defer r.mu.Unlock()

	date := trade.Timestamp.Format("2006-01-02")
	stats, exists := r.state.DailyStats[date]
	if !exists {
		stats = &DailyStats{
			Date:         trade.Timestamp,
			StartBalance: 0, // Should be set when first trade of the day is made
		}
		r.state.DailyStats[date] = stats
	}

	// Update trade counts
	stats.TradeCount++
	stats.Volume += trade.Amount

	// Calculate P&L
	var pnl float64
	if trade.Status == models.TradeStatusCompleted {
		if trade.Type == "BUY" {
			pnl = (trade.ClosePrice - trade.Price) * trade.Amount
		} else {
			pnl = (trade.Price - trade.ClosePrice) * trade.Amount
		}

		// Update win/loss stats
		if pnl > 0 {
			stats.WinCount++
			if pnl > stats.LargestWin {
				stats.LargestWin = pnl
			}
			r.state.ConsecutiveLosses = 0
		} else {
			stats.LossCount++
			if pnl < stats.LargestLoss {
				stats.LargestLoss = pnl
			}
			r.state.ConsecutiveLosses++

			// Check if cooldown should be triggered
			if r.state.ConsecutiveLosses >= 3 {
				r.state.InCooldown = true
				r.state.CooldownUntil = time.Now().Add(r.config.CooldownPeriod)
			}
		}

		// Update P&L stats
		stats.ProfitLoss += pnl
		r.state.DailyLoss = stats.ProfitLoss / stats.StartBalance

		// Update drawdown
		if stats.ProfitLoss < stats.MaxDrawdown {
			stats.MaxDrawdown = stats.ProfitLoss
			r.state.CurrentDrawdown = stats.MaxDrawdown / stats.StartBalance
		}

		// Calculate win rate and averages
		if stats.TradeCount > 0 {
			stats.WinRate = float64(stats.WinCount) / float64(stats.TradeCount)
			if stats.WinCount > 0 {
				stats.AverageWin = stats.LargestWin / float64(stats.WinCount)
			}
			if stats.LossCount > 0 {
				stats.AverageLoss = stats.LargestLoss / float64(stats.LossCount)
			}
		}
	}
}
