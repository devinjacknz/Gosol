package risk

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/leonzhao/trading-system/backend/models"
)

// RiskMetrics represents portfolio risk metrics
type RiskMetrics struct {
	ValueAtRisk    float64   // 95% VaR
	Volatility     float64   // Historical volatility
	SharpeRatio    float64   // Risk-adjusted return
	Beta           float64   // Market correlation
	MaxDrawdown    float64   // Maximum historical drawdown
	WinRate        float64   // Percentage of winning trades
	ProfitFactor   float64   // Gross profits / gross losses
	LastUpdated    time.Time
}

// Portfolio manages trading positions and risk
type Portfolio struct {
	positions     map[string]*models.Position
	balance       float64
	initialValue  float64
	highWaterMark float64
	dailyStats    map[string]*models.DailyStats
	riskMetrics   RiskMetrics
	returns       []float64 // Historical returns for risk calculation
	mutex         sync.RWMutex
}

// NewPortfolio creates a new portfolio manager
func NewPortfolio(initialBalance float64) *Portfolio {
	return &Portfolio{
		positions:     make(map[string]*models.Position),
		balance:       initialBalance,
		initialValue:  initialBalance,
		highWaterMark: initialBalance,
		dailyStats:    make(map[string]*models.DailyStats),
	}
}

// AddPosition adds a new position to the portfolio
func (p *Portfolio) AddPosition(ctx context.Context, position *models.Position) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Validate position
	if position.Size <= 0 {
		return NewInvalidPositionError("invalid size", map[string]interface{}{
			"size": position.Size,
		})
	}

	if position.EntryPrice <= 0 {
		return NewInvalidPositionError("invalid entry price", map[string]interface{}{
			"price": position.EntryPrice,
		})
	}

	// Check if we can afford the position
	cost := position.Size * position.EntryPrice
	if cost > p.balance {
		return NewInsufficientBalanceError(cost, p.balance)
	}

	// Add position
	p.positions[position.TokenAddress] = position
	p.balance -= cost

	// Update daily stats
	dateKey := position.OpenTime.Format("2006-01-02")
	stats, exists := p.dailyStats[dateKey]
	if !exists {
		stats = NewDailyStats(position.OpenTime, p.balance)
		p.dailyStats[dateKey] = stats
	}

	return nil
}

// ClosePosition closes a position and realizes PnL
func (p *Portfolio) ClosePosition(ctx context.Context, tokenAddress string, closePrice float64) (float64, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	position, exists := p.positions[tokenAddress]
	if !exists {
		return 0, NewPositionNotFoundError(tokenAddress)
	}

	// Calculate PnL
	var pnl float64
	if position.Side == models.PositionSideLong {
		pnl = position.Size * (closePrice - position.EntryPrice)
	} else {
		pnl = position.Size * (position.EntryPrice - closePrice)
	}

	// Update balance and stats
	p.balance += position.Size*closePrice + pnl
	delete(p.positions, tokenAddress)

	// Update high water mark
	portfolioValue := p.GetPortfolioValue()
	if portfolioValue > p.highWaterMark {
		p.highWaterMark = portfolioValue
	}

	// Update daily stats
	dateKey := time.Now().Format("2006-01-02")
	stats, exists := p.dailyStats[dateKey]
	if !exists {
		stats = NewDailyStats(time.Now(), p.balance)
		p.dailyStats[dateKey] = stats
	}
	UpdateStats(stats, pnl, position.Commission, position.Size*closePrice)

	return pnl, nil
}

// UpdatePositions updates all position prices and risk metrics
func (p *Portfolio) UpdatePositions(ctx context.Context, prices map[string]float64) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for tokenAddress, position := range p.positions {
		price, ok := prices[tokenAddress]
		if !ok {
			continue
		}

		position.UpdatePrice(price)

		// Check stop loss
		if position.ShouldLiquidate() {
			if _, err := p.ClosePosition(ctx, tokenAddress, price); err != nil {
				return fmt.Errorf("failed to close position at stop loss: %w", err)
			}
		}
	}

	// Update daily stats
	dateKey := time.Now().Format("2006-01-02")
	stats, exists := p.dailyStats[dateKey]
	if !exists {
		stats = NewDailyStats(time.Now(), p.balance)
		p.dailyStats[dateKey] = stats
	}

	currentValue := p.GetPortfolioValue()
	if currentValue > stats.HighWaterMark {
		stats.HighWaterMark = currentValue
	}

	// Calculate daily return
	prevValue := stats.OpenValue
	if prevValue > 0 {
		dailyReturn := (currentValue - prevValue) / prevValue
		p.returns = append(p.returns, dailyReturn)
		
		// Keep last 252 trading days (1 year) of returns
		if len(p.returns) > 252 {
			p.returns = p.returns[1:]
		}
	}

	// Update risk metrics
	p.updateRiskMetrics()

	drawdown := (stats.HighWaterMark - currentValue) / stats.HighWaterMark * 100
	UpdateDrawdown(stats, drawdown)

	return nil
}

// updateRiskMetrics updates portfolio risk metrics
func (p *Portfolio) updateRiskMetrics() {
	if len(p.returns) < 2 {
		return
	}

	// Calculate volatility
	volatility := calculateVolatility(p.returns)
	
	// Calculate VaR using historical simulation
	valueAtRisk := calculateVaR(p.returns, 0.95)
	
	// Calculate Sharpe ratio (assuming 2% risk-free rate)
	avgReturn := calculateMean(p.returns)
	sharpeRatio := (avgReturn - 0.02) / volatility
	
	// Calculate win rate
	winRate := calculateWinRate(p.returns)
	
	// Calculate profit factor
	profitFactor := calculateProfitFactor(p.returns)

	p.riskMetrics = RiskMetrics{
		ValueAtRisk:    valueAtRisk,
		Volatility:     volatility,
		SharpeRatio:    sharpeRatio,
		MaxDrawdown:    p.GetDrawdown(),
		WinRate:        winRate,
		ProfitFactor:   profitFactor,
		LastUpdated:    time.Now(),
	}
}

// GetRiskMetrics returns current risk metrics
func (p *Portfolio) GetRiskMetrics() RiskMetrics {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.riskMetrics
}

// Helper functions for risk calculations

func calculateVolatility(returns []float64) float64 {
	if len(returns) < 2 {
		return 0
	}
	
	mean := calculateMean(returns)
	var sumSquaredDiff float64
	for _, r := range returns {
		diff := r - mean
		sumSquaredDiff += diff * diff
	}
	
	variance := sumSquaredDiff / float64(len(returns)-1)
	return math.Sqrt(variance)
}

func calculateVaR(returns []float64, confidence float64) float64 {
	if len(returns) < 2 {
		return 0
	}
	
	sorted := make([]float64, len(returns))
	copy(sorted, returns)
	sort.Float64s(sorted)
	
	index := int(float64(len(sorted)) * (1 - confidence))
	return -sorted[index]
}

func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateWinRate(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}
	
	wins := 0
	for _, r := range returns {
		if r > 0 {
			wins++
		}
	}
	return float64(wins) / float64(len(returns))
}

func calculateProfitFactor(returns []float64) float64 {
	var grossProfit, grossLoss float64
	
	for _, r := range returns {
		if r > 0 {
			grossProfit += r
		} else {
			grossLoss -= r
		}
	}
	
	if grossLoss == 0 {
		return 0
	}
	return grossProfit / grossLoss
}

// GetPosition returns a position by token address
func (p *Portfolio) GetPosition(tokenAddress string) (*models.Position, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	position, exists := p.positions[tokenAddress]
	if !exists {
		return nil, NewPositionNotFoundError(tokenAddress)
	}

	return position, nil
}

// GetPositions returns all open positions
func (p *Portfolio) GetPositions() map[string]*models.Position {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	positions := make(map[string]*models.Position)
	for k, v := range p.positions {
		positions[k] = v
	}
	return positions
}

// GetPortfolioValue returns the total portfolio value including open positions
func (p *Portfolio) GetPortfolioValue() float64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	value := p.balance
	for _, position := range p.positions {
		value += position.Value()
	}
	return value
}

// GetDrawdown returns the current drawdown percentage
func (p *Portfolio) GetDrawdown() float64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	currentValue := p.GetPortfolioValue()
	if currentValue >= p.highWaterMark {
		return 0
	}

	return (p.highWaterMark - currentValue) / p.highWaterMark * 100
}

// GetROI returns the total return on investment percentage
func (p *Portfolio) GetROI() float64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	currentValue := p.GetPortfolioValue()
	return (currentValue - p.initialValue) / p.initialValue * 100
}

// GetDailyStats returns statistics for a specific day
func (p *Portfolio) GetDailyStats(date time.Time) (*models.DailyStats, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	dateKey := date.Format("2006-01-02")
	stats, exists := p.dailyStats[dateKey]
	if !exists {
		return nil, fmt.Errorf("no stats found for date %s", dateKey)
	}

	return stats, nil
}
