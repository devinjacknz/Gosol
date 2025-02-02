package order

import (
	"context"
	"sync"
	"time"

	"github.com/leonzhao/gosol/backend/trading/analysis/monitoring"
)

// OrderType represents the type of order
type OrderType int

const (
	Market OrderType = iota + 1
	Limit
	StopLoss
	TakeProfit
)

// OrderStatus represents the status of an order
type OrderStatus int

const (
	Created OrderStatus = iota + 1
	Pending
	PartiallyFilled
	Filled
	Cancelled
	Rejected
	Expired
)

// OrderSide represents the side of an order (buy/sell)
type OrderSide int

const (
	Buy OrderSide = iota + 1
	Sell
)

// Order represents a trading order
type Order struct {
	ID            string
	Symbol        string
	Type          OrderType
	Side          OrderSide
	Price         *float64 // nil for market orders
	StopPrice     *float64 // for stop loss/take profit orders
	Size          float64
	FilledSize    float64
	RemainingSize float64
	Status        OrderStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ExpiresAt     *time.Time
	ClientOrderID string
	mu            sync.RWMutex
}

// OrderManager interface defines the contract for order management
type OrderManager interface {
	CreateOrder(ctx context.Context, params CreateOrderParams) (*Order, error)
	CancelOrder(ctx context.Context, orderID string) error
	GetOrder(ctx context.Context, orderID string) (*Order, error)
	ListOrders(ctx context.Context, filter OrderFilter) ([]*Order, error)
	UpdateOrderStatus(ctx context.Context, orderID string, status OrderStatus) error
	UpdateFilledSize(ctx context.Context, orderID string, filledSize float64) error
}

// CreateOrderParams contains parameters for creating an order
type CreateOrderParams struct {
	Symbol        string
	Type          OrderType
	Side          OrderSide
	Price         *float64
	StopPrice     *float64
	Size          float64
	ClientOrderID string
	ExpiresAt     *time.Time
}

// OrderFilter contains filters for listing orders
type OrderFilter struct {
	Symbol    string
	Type      *OrderType
	Side      *OrderSide
	Status    *OrderStatus
	StartTime *time.Time
	EndTime   *time.Time
}

// DefaultOrderManager implements OrderManager interface
type DefaultOrderManager struct {
	orders map[string]*Order
	mu     sync.RWMutex
}

// NewOrderManager creates a new order manager instance
func NewOrderManager() OrderManager {
	return &DefaultOrderManager{
		orders: make(map[string]*Order),
	}
}

// CreateOrder creates a new order
func (m *DefaultOrderManager) CreateOrder(ctx context.Context, params CreateOrderParams) (*Order, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		monitoring.RecordIndicatorCalculation("create_order", duration)
	}()

	if err := validateCreateParams(params); err != nil {
		monitoring.RecordIndicatorError("create_order", err.Error())
		return nil, err
	}

	order := &Order{
		ID:            generateOrderID(),
		Symbol:        params.Symbol,
		Type:          params.Type,
		Side:          params.Side,
		Price:         params.Price,
		StopPrice:     params.StopPrice,
		Size:          params.Size,
		RemainingSize: params.Size,
		Status:        Created,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		ExpiresAt:     params.ExpiresAt,
		ClientOrderID: params.ClientOrderID,
	}

	m.mu.Lock()
	m.orders[order.ID] = order
	m.mu.Unlock()

	monitoring.RecordIndicatorValue("active_orders", float64(len(m.orders)))
	return order, nil
}

// CancelOrder cancels an existing order
func (m *DefaultOrderManager) CancelOrder(ctx context.Context, orderID string) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		monitoring.RecordIndicatorCalculation("cancel_order", duration)
	}()

	m.mu.Lock()
	defer m.mu.Unlock()

	order, exists := m.orders[orderID]
	if !exists {
		return ErrOrderNotFound
	}

	order.mu.Lock()
	defer order.mu.Unlock()

	if !isOrderCancellable(order.Status) {
		return ErrOrderNotCancellable
	}

	order.Status = Cancelled
	order.UpdatedAt = time.Now()

	monitoring.RecordIndicatorValue("cancelled_orders", 1)
	return nil
}

// GetOrder retrieves an order by ID
func (m *DefaultOrderManager) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	m.mu.RLock()
	order, exists := m.orders[orderID]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrOrderNotFound
	}

	return order, nil
}

// ListOrders returns orders based on filter criteria
func (m *DefaultOrderManager) ListOrders(ctx context.Context, filter OrderFilter) ([]*Order, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	orders := make([]*Order, 0, len(m.orders))
	for _, order := range m.orders {
		if matchesOrderFilter(order, filter) {
			orders = append(orders, order)
		}
	}

	return orders, nil
}

// UpdateOrderStatus updates the status of an order
func (m *DefaultOrderManager) UpdateOrderStatus(ctx context.Context, orderID string, status OrderStatus) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		monitoring.RecordIndicatorCalculation("update_order_status", duration)
	}()

	m.mu.RLock()
	order, exists := m.orders[orderID]
	m.mu.RUnlock()

	if !exists {
		return ErrOrderNotFound
	}

	order.mu.Lock()
	defer order.mu.Unlock()

	if !isValidStatusTransition(order.Status, status) {
		return ErrInvalidStatusTransition
	}

	order.Status = status
	order.UpdatedAt = time.Now()

	monitoring.RecordIndicatorValue("order_status_updates", 1)
	return nil
}

// UpdateFilledSize updates the filled size of an order
func (m *DefaultOrderManager) UpdateFilledSize(ctx context.Context, orderID string, filledSize float64) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		monitoring.RecordIndicatorCalculation("update_filled_size", duration)
	}()

	m.mu.RLock()
	order, exists := m.orders[orderID]
	m.mu.RUnlock()

	if !exists {
		return ErrOrderNotFound
	}

	order.mu.Lock()
	defer order.mu.Unlock()

	if order.Status != Pending && order.Status != PartiallyFilled {
		return ErrOrderNotFillable
	}

	if filledSize > order.RemainingSize {
		return ErrInvalidFilledSize
	}

	order.FilledSize += filledSize
	order.RemainingSize -= filledSize
	order.UpdatedAt = time.Now()

	if order.RemainingSize == 0 {
		order.Status = Filled
	} else {
		order.Status = PartiallyFilled
	}

	monitoring.RecordIndicatorValue("filled_size", filledSize)
	return nil
}

func validateCreateParams(params CreateOrderParams) error {
	if params.Symbol == "" {
		return ErrInvalidSymbol
	}
	if params.Size <= 0 {
		return ErrInvalidSize
	}
	if params.Type == Limit && params.Price == nil {
		return ErrInvalidPrice
	}
	if (params.Type == StopLoss || params.Type == TakeProfit) && params.StopPrice == nil {
		return ErrInvalidStopPrice
	}
	return nil
}

func isOrderCancellable(status OrderStatus) bool {
	return status == Created || status == Pending || status == PartiallyFilled
}

func isValidStatusTransition(from, to OrderStatus) bool {
	validTransitions := map[OrderStatus][]OrderStatus{
		Created:         {Pending, Rejected, Cancelled},
		Pending:         {PartiallyFilled, Filled, Cancelled, Rejected, Expired},
		PartiallyFilled: {Filled, Cancelled, Expired},
	}

	if transitions, exists := validTransitions[from]; exists {
		for _, validTo := range transitions {
			if to == validTo {
				return true
			}
		}
	}
	return false
}

func matchesOrderFilter(order *Order, filter OrderFilter) bool {
	if filter.Symbol != "" && order.Symbol != filter.Symbol {
		return false
	}
	if filter.Type != nil && order.Type != *filter.Type {
		return false
	}
	if filter.Side != nil && order.Side != *filter.Side {
		return false
	}
	if filter.Status != nil && order.Status != *filter.Status {
		return false
	}
	if filter.StartTime != nil && order.CreatedAt.Before(*filter.StartTime) {
		return false
	}
	if filter.EndTime != nil && order.CreatedAt.After(*filter.EndTime) {
		return false
	}
	return true
}

func generateOrderID() string {
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
