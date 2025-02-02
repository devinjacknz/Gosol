package position

import (
	"context"
	"sync"
	"time"

"github.com/devinjacknz/godydxhyber/backend/trading/analysis/monitoring"
)

// Position represents a trading position
type Position struct {
	ID             string
	Symbol         string
	Side           Side
	EntryPrice     float64
	CurrentPrice   float64
	Size           float64
	OpenTime       time.Time
	LastUpdateTime time.Time
	StopLoss       *float64
	TakeProfit     *float64
	Status         PositionStatus
	UnrealizedPnL  float64
	RealizedPnL    float64
	Leverage       float64
	Margin         float64
	mu             sync.RWMutex
}

// Side represents the position side (long or short)
type Side int

const (
	Long Side = iota + 1
	Short
)

// PositionStatus represents the status of a position
type PositionStatus int

const (
	Open PositionStatus = iota + 1
	Closed
	Liquidated
)

// Manager manages trading positions
type Manager struct {
	positions map[string]*Position
	mu        sync.RWMutex
}

// NewManager creates a new position manager
func NewManager() *Manager {
	return &Manager{
		positions: make(map[string]*Position),
	}
}

// OpenPosition opens a new position
func (m *Manager) OpenPosition(ctx context.Context, params OpenPositionParams) (*Position, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		monitoring.RecordIndicatorCalculation("open_position", duration)
	}()

	if err := validateOpenParams(params); err != nil {
		monitoring.RecordIndicatorError("open_position", err.Error())
		return nil, err
	}

	position := &Position{
		ID:             generatePositionID(),
		Symbol:         params.Symbol,
		Side:           params.Side,
		EntryPrice:     params.EntryPrice,
		CurrentPrice:   params.EntryPrice,
		Size:           params.Size,
		OpenTime:       time.Now(),
		LastUpdateTime: time.Now(),
		StopLoss:       params.StopLoss,
		TakeProfit:     params.TakeProfit,
		Status:         Open,
		Leverage:       params.Leverage,
		Margin:         calculateMargin(params.Size, params.EntryPrice, params.Leverage),
	}

	m.mu.Lock()
	m.positions[position.ID] = position
	m.mu.Unlock()

	monitoring.RecordIndicatorValue("active_positions", float64(len(m.positions)))
	return position, nil
}

// ClosePosition closes an existing position
func (m *Manager) ClosePosition(ctx context.Context, id string, closePrice float64) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		monitoring.RecordIndicatorCalculation("close_position", duration)
	}()

	m.mu.Lock()
	defer m.mu.Unlock()

	position, exists := m.positions[id]
	if !exists {
		return ErrPositionNotFound
	}

	position.mu.Lock()
	defer position.mu.Unlock()

	if position.Status != Open {
		return ErrPositionAlreadyClosed
	}

	position.Status = Closed
	position.CurrentPrice = closePrice
	position.LastUpdateTime = time.Now()
	position.RealizedPnL = calculateRealizedPnL(position, closePrice)

	monitoring.RecordIndicatorValue("active_positions", float64(len(m.positions)-1))
	monitoring.RecordIndicatorValue("realized_pnl", position.RealizedPnL)

	return nil
}

// UpdatePosition updates position details
func (m *Manager) UpdatePosition(ctx context.Context, id string, params UpdatePositionParams) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		monitoring.RecordIndicatorCalculation("update_position", duration)
	}()

	m.mu.RLock()
	position, exists := m.positions[id]
	m.mu.RUnlock()

	if !exists {
		return ErrPositionNotFound
	}

	position.mu.Lock()
	defer position.mu.Unlock()

	if position.Status != Open {
		return ErrPositionAlreadyClosed
	}

	if params.CurrentPrice != nil {
		position.CurrentPrice = *params.CurrentPrice
		position.UnrealizedPnL = calculateUnrealizedPnL(position)
	}

	if params.StopLoss != nil {
		position.StopLoss = params.StopLoss
	}

	if params.TakeProfit != nil {
		position.TakeProfit = params.TakeProfit
	}

	position.LastUpdateTime = time.Now()

	monitoring.RecordIndicatorValue("unrealized_pnl", position.UnrealizedPnL)
	return nil
}

// GetPosition retrieves a position by ID
func (m *Manager) GetPosition(ctx context.Context, id string) (*Position, error) {
	m.mu.RLock()
	position, exists := m.positions[id]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrPositionNotFound
	}

	return position, nil
}

// ListPositions returns all positions
func (m *Manager) ListPositions(ctx context.Context, filter PositionFilter) ([]*Position, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	positions := make([]*Position, 0, len(m.positions))
	for _, pos := range m.positions {
		if matchesFilter(pos, filter) {
			positions = append(positions, pos)
		}
	}

	return positions, nil
}

// OpenPositionParams contains parameters for opening a position
type OpenPositionParams struct {
	Symbol     string
	Side       Side
	Size       float64
	EntryPrice float64
	StopLoss   *float64
	TakeProfit *float64
	Leverage   float64
}

// UpdatePositionParams contains parameters for updating a position
type UpdatePositionParams struct {
	CurrentPrice *float64
	StopLoss     *float64
	TakeProfit   *float64
}

// PositionFilter contains filters for listing positions
type PositionFilter struct {
	Symbol  string
	Side    *Side
	Status  *PositionStatus
	MinSize *float64
	MaxSize *float64
}

func calculateUnrealizedPnL(position *Position) float64 {
	if position.Side == Long {
		return (position.CurrentPrice - position.EntryPrice) * position.Size
	}
	return (position.EntryPrice - position.CurrentPrice) * position.Size
}

func calculateRealizedPnL(position *Position, closePrice float64) float64 {
	if position.Side == Long {
		return (closePrice - position.EntryPrice) * position.Size
	}
	return (position.EntryPrice - closePrice) * position.Size
}

func calculateMargin(size, price, leverage float64) float64 {
	return (size * price) / leverage
}

func matchesFilter(position *Position, filter PositionFilter) bool {
	if filter.Symbol != "" && position.Symbol != filter.Symbol {
		return false
	}
	if filter.Side != nil && position.Side != *filter.Side {
		return false
	}
	if filter.Status != nil && position.Status != *filter.Status {
		return false
	}
	if filter.MinSize != nil && position.Size < *filter.MinSize {
		return false
	}
	if filter.MaxSize != nil && position.Size > *filter.MaxSize {
		return false
	}
	return true
}

func validateOpenParams(params OpenPositionParams) error {
	if params.Symbol == "" {
		return ErrInvalidSymbol
	}
	if params.Size <= 0 {
		return ErrInvalidSize
	}
	if params.EntryPrice <= 0 {
		return ErrInvalidPrice
	}
	if params.Leverage <= 0 {
		return ErrInvalidLeverage
	}
	if params.StopLoss != nil && *params.StopLoss <= 0 {
		return ErrInvalidStopLoss
	}
	if params.TakeProfit != nil && *params.TakeProfit <= 0 {
		return ErrInvalidTakeProfit
	}
	return nil
}

func generatePositionID() string {
	return time.Now().Format("20060102150405.000") + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
