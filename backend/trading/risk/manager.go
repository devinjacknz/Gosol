package risk

import (
	"context"
	"fmt"
	"time"

	"solmeme-trader/models"
)

// RiskManager handles trading risk management
type RiskManager struct {
	config     RiskConfig
	positions  map[string]*Position // Map of position ID to position
	totalValue float64             // Total portfolio value
}

// NewRiskManager creates a new risk manager
func NewRiskManager(config RiskConfig) *RiskManager {
	return &RiskManager{
		config:     config,
		positions:  make(map[string]*Position),
		totalValue: config.InitialCapital,
	}
}

// CanOpenPosition checks if a new position can be opened
func (m *RiskManager) CanOpenPosition(ctx context.Context, tokenAddress string, size float64, marketData *models.MarketData) error {
	// Check number of positions
	if len(m.positions) >= m.config.MaxPositions {
		return fmt.Errorf("maximum number of positions (%d) reached", m.config.MaxPositions)
	}

	// Check position size
	if size > m.config.MaxPositionSize {
		return fmt.Errorf("position size %f exceeds maximum allowed %f", size, m.config.MaxPositionSize)
	}

	// Check if enough capital available
	if size > m.totalValue*m.config.RiskPerTrade {
		return fmt.Errorf("position size %f exceeds risk per trade limit", size)
	}

	return nil
}

// OpenPosition opens a new position
func (m *RiskManager) OpenPosition(ctx context.Context, position *Position, marketData *models.MarketData) error {
	// Validate position
	if err := m.ValidatePosition(ctx, position); err != nil {
		return err
	}

	// Set default stop loss and take profit if not set
	if position.StopLoss == nil {
		stopLoss := m.GetDefaultStopLoss(position.Side, position.EntryPrice)
		position.StopLoss = &stopLoss
	}
	if position.TakeProfit == nil {
		takeProfit := m.GetDefaultTakeProfit(position.Side, position.EntryPrice)
		position.TakeProfit = &takeProfit
	}

	// Add position
	m.positions[position.ID] = position
	return nil
}

// ClosePosition closes a position
func (m *RiskManager) ClosePosition(positionID string) error {
	position, exists := m.positions[positionID]
	if !exists {
		return fmt.Errorf("position %s not found", positionID)
	}

	// Update total value
	m.totalValue += position.PnL
	delete(m.positions, positionID)
	return nil
}

// ValidatePosition validates a position against risk rules
func (m *RiskManager) ValidatePosition(ctx context.Context, position *Position) error {
	// Check position size
	if position.Size > m.config.MaxPositionSize {
		return fmt.Errorf("position size %f exceeds maximum allowed %f", position.Size, m.config.MaxPositionSize)
	}

	// Check leverage
	if position.Leverage > m.config.MaxLeverage {
		return fmt.Errorf("leverage %f exceeds maximum allowed %f", position.Leverage, m.config.MaxLeverage)
	}

	// Check drawdown
	if position.PnL < 0 && -position.PnL/position.Size > m.config.MaxDrawdown {
		return fmt.Errorf("drawdown %f exceeds maximum allowed %f", -position.PnL/position.Size, m.config.MaxDrawdown)
	}

	return nil
}

// CheckStopLoss checks if stop loss is triggered
func (m *RiskManager) CheckStopLoss(position *Position) bool {
	if position.StopLoss == nil {
		return false
	}

	if position.Side == models.PositionSideLong {
		return position.CurrentPrice <= *position.StopLoss
	}
	return position.CurrentPrice >= *position.StopLoss
}

// CheckTakeProfit checks if take profit is triggered
func (m *RiskManager) CheckTakeProfit(position *Position) bool {
	if position.TakeProfit == nil {
		return false
	}

	if position.Side == models.PositionSideLong {
		return position.CurrentPrice >= *position.TakeProfit
	}
	return position.CurrentPrice <= *position.TakeProfit
}

// CalculatePnL calculates position PnL
func (m *RiskManager) CalculatePnL(position *Position) float64 {
	if position.Side == models.PositionSideLong {
		return (position.CurrentPrice - position.EntryPrice) * position.Size * position.Leverage
	}
	return (position.EntryPrice - position.CurrentPrice) * position.Size * position.Leverage
}

// GetDefaultStopLoss gets default stop loss price
func (m *RiskManager) GetDefaultStopLoss(side string, entryPrice float64) float64 {
	if side == models.PositionSideLong {
		return entryPrice * (1 - m.config.StopLoss)
	}
	return entryPrice * (1 + m.config.StopLoss)
}

// GetDefaultTakeProfit gets default take profit price
func (m *RiskManager) GetDefaultTakeProfit(side string, entryPrice float64) float64 {
	if side == models.PositionSideLong {
		return entryPrice * (1 + m.config.TakeProfit)
	}
	return entryPrice * (1 - m.config.TakeProfit)
}

// UpdatePosition updates position with current market data
func (m *RiskManager) UpdatePosition(position *Position, currentPrice float64) {
	position.CurrentPrice = currentPrice
	position.PnL = m.CalculatePnL(position)
	position.UpdateTime = time.Now()
}

// GetTotalValue gets total portfolio value
func (m *RiskManager) GetTotalValue() float64 {
	return m.totalValue
}

// GetPositions gets all positions
func (m *RiskManager) GetPositions() map[string]*Position {
	return m.positions
}

// GetPosition gets a position by ID
func (m *RiskManager) GetPosition(positionID string) (*Position, bool) {
	position, exists := m.positions[positionID]
	return position, exists
}
