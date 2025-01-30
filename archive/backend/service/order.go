package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/leonzhao/trading-system/backend/models"
	"github.com/leonzhao/trading-system/backend/repository"
)

var (
	ErrInvalidOrder      = errors.New("invalid order")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrOrderNotFound     = errors.New("order not found")
)

// OrderService handles order execution and management
type OrderService struct {
	db           *repository.Database
	marketData   *MarketDataService
	positions    map[string]map[string]*models.PositionRecord // userID -> symbol -> position
	activeOrders map[string]*models.OrderRecord               // orderID -> order
	mutex        sync.RWMutex
}

// NewOrderService creates a new order service
func NewOrderService(db *repository.Database, marketData *MarketDataService) *OrderService {
	return &OrderService{
		db:           db,
		marketData:   marketData,
		positions:    make(map[string]map[string]*models.PositionRecord),
		activeOrders: make(map[string]*models.OrderRecord),
	}
}

// CreateOrder creates a new order
func (s *OrderService) CreateOrder(ctx context.Context, order *models.OrderRecord) error {
	// Validate order
	if err := s.validateOrder(ctx, order); err != nil {
		return fmt.Errorf("invalid order: %w", err)
	}

	// Generate order ID
	order.OrderID = uuid.New().String()
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	order.Status = "pending"

	// Save to database
	if err := s.db.SaveOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	// Add to active orders
	s.mutex.Lock()
	s.activeOrders[order.OrderID] = order
	s.mutex.Unlock()

	// Start processing order
	go s.processOrder(order)

	return nil
}

// CancelOrder cancels an existing order
func (s *OrderService) CancelOrder(ctx context.Context, orderID string) error {
	s.mutex.Lock()
	order, exists := s.activeOrders[orderID]
	if !exists {
		s.mutex.Unlock()
		return fmt.Errorf("order not found: %s", orderID)
	}

	if order.Status != "pending" {
		s.mutex.Unlock()
		return fmt.Errorf("cannot cancel order in status: %s", order.Status)
	}

	order.Status = "cancelled"
	order.UpdatedAt = time.Now()
	delete(s.activeOrders, orderID)
	s.mutex.Unlock()

	// Update in database
	if err := s.db.UpdateOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}

// GetOrder retrieves an order by ID
func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*models.OrderRecord, error) {
	s.mutex.RLock()
	if order, exists := s.activeOrders[orderID]; exists {
		s.mutex.RUnlock()
		return order, nil
	}
	s.mutex.RUnlock()

	// If not in memory, get from database
	return s.db.GetOrder(ctx, orderID)
}

// GetUserOrders retrieves all orders for a user
func (s *OrderService) GetUserOrders(ctx context.Context, userID string, status string) ([]*models.OrderRecord, error) {
	return s.db.GetUserOrders(ctx, userID, status)
}

// GetPosition retrieves current position for a user and symbol
func (s *OrderService) GetPosition(ctx context.Context, userID, symbol string) (*models.PositionRecord, error) {
	s.mutex.RLock()
	if userPositions, exists := s.positions[userID]; exists {
		if position, exists := userPositions[symbol]; exists {
			s.mutex.RUnlock()
			return position, nil
		}
	}
	s.mutex.RUnlock()

	// If not in memory, get from database
	return s.db.GetPosition(ctx, userID, symbol)
}

// Internal methods

func (s *OrderService) validateOrder(ctx context.Context, order *models.OrderRecord) error {
	// Basic validation
	if order.UserID == "" || order.Symbol == "" || order.Amount <= 0 {
		return fmt.Errorf("missing required fields")
	}

	// Get current price
	currentPrice, err := s.marketData.GetLatestPrice(ctx, order.Symbol)
	if err != nil {
		return fmt.Errorf("failed to get current price: %w", err)
	}

	// Validate price for limit orders
	if order.Type == "limit" && order.Price <= 0 {
		return fmt.Errorf("invalid limit price")
	}

	// Check user position and risk limits
	position, err := s.GetPosition(ctx, order.UserID, order.Symbol)
	if err != nil && err != repository.ErrNotFound {
		return fmt.Errorf("failed to get position: %w", err)
	}

	// Calculate potential position size
	potentialSize := order.Amount
	if position != nil {
		if order.Side == "buy" {
			potentialSize += position.Amount
		} else {
			potentialSize = position.Amount - order.Amount
		}
	}

	// Check risk limits
	if err := s.checkRiskLimits(ctx, order.UserID, order.Symbol, potentialSize, currentPrice); err != nil {
		return fmt.Errorf("risk limit exceeded: %w", err)
	}

	return nil
}

func (s *OrderService) processOrder(order *models.OrderRecord) {
	ctx := context.Background()

	// Get current price
	currentPrice, err := s.marketData.GetLatestPrice(ctx, order.Symbol)
	if err != nil {
		s.handleOrderError(order, fmt.Errorf("failed to get price: %w", err))
		return
	}

	// Check if order can be executed
	canExecute := false
	if order.Type == "market" {
		canExecute = true
		order.Price = currentPrice
	} else if order.Type == "limit" {
		if order.Side == "buy" && currentPrice <= order.Price {
			canExecute = true
		} else if order.Side == "sell" && currentPrice >= order.Price {
			canExecute = true
		}
	}

	if !canExecute {
		return // Keep order pending
	}

	// Execute order
	if err := s.executeOrder(ctx, order, currentPrice); err != nil {
		s.handleOrderError(order, err)
		return
	}

	// Update order status
	order.Status = "filled"
	order.Filled = order.Amount
	order.UpdatedAt = time.Now()

	// Update in database
	if err := s.db.UpdateOrder(ctx, order); err != nil {
		s.handleOrderError(order, fmt.Errorf("failed to update order: %w", err))
		return
	}

	// Remove from active orders
	s.mutex.Lock()
	delete(s.activeOrders, order.OrderID)
	s.mutex.Unlock()
}

func (s *OrderService) executeOrder(ctx context.Context, order *models.OrderRecord, price float64) error {
	// Update position
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Get or create user positions map
	userPositions, exists := s.positions[order.UserID]
	if !exists {
		userPositions = make(map[string]*models.PositionRecord)
		s.positions[order.UserID] = userPositions
	}

	// Get or create position
	position, exists := userPositions[order.Symbol]
	if !exists {
		position = &models.PositionRecord{
			UserID:    order.UserID,
			Symbol:    order.Symbol,
			Amount:    0,
			UpdatedAt: time.Now(),
		}
		userPositions[order.Symbol] = position
	}

	// Update position
	if order.Side == "buy" {
		position.Amount += order.Amount
		position.EntryPrice = (position.EntryPrice*position.Amount + price*order.Amount) /
			(position.Amount + order.Amount)
	} else {
		position.Amount -= order.Amount
		if position.Amount == 0 {
			position.EntryPrice = 0
		}
	}

	position.CurrentPrice = price
	position.UpdatedAt = time.Now()

	// Calculate PnL
	if position.Amount > 0 {
		position.UnrealizedPnL = (price - position.EntryPrice) * position.Amount
	}

	// Save position to database
	if err := s.db.SavePosition(ctx, position); err != nil {
		return fmt.Errorf("failed to save position: %w", err)
	}

	// Create trade record
	trade := &models.TradeRecord{
		UserID:    order.UserID,
		Symbol:    order.Symbol,
		Side:      order.Side,
		Price:     price,
		Amount:    order.Amount,
		Fee:       price * order.Amount * 0.001, // 0.1% fee
		Total:     price * order.Amount,
		Timestamp: time.Now(),
	}

	// Save trade to database
	if err := s.db.SaveTrade(ctx, trade); err != nil {
		return fmt.Errorf("failed to save trade: %w", err)
	}

	return nil
}

func (s *OrderService) handleOrderError(order *models.OrderRecord, err error) {
	ctx := context.Background()

	order.Status = "failed"
	order.UpdatedAt = time.Now()

	// Update in database
	if dbErr := s.db.UpdateOrder(ctx, order); dbErr != nil {
		// Log both errors
		fmt.Printf("Order error: %v, Update error: %v\n", err, dbErr)
		return
	}

	// Remove from active orders
	s.mutex.Lock()
	delete(s.activeOrders, order.OrderID)
	s.mutex.Unlock()

	// Log error
	fmt.Printf("Order failed: %v\n", err)
}

func (s *OrderService) checkRiskLimits(ctx context.Context, userID, symbol string,
	potentialSize float64, currentPrice float64) error {

	// Get risk limits
	limits, err := s.db.GetRiskLimit(ctx, userID, symbol)
	if err != nil {
		return fmt.Errorf("failed to get risk limits: %w", err)
	}

	// Check position size limit
	if potentialSize > limits.MaxPosition {
		return fmt.Errorf("position size exceeds limit: %.2f > %.2f",
			potentialSize, limits.MaxPosition)
	}

	// Check leverage limit
	leverage := (potentialSize * currentPrice) / 10000 // Assuming account size of 10000
	if leverage > limits.MaxLeverage {
		return fmt.Errorf("leverage exceeds limit: %.2f > %.2f",
			leverage, limits.MaxLeverage)
	}

	return nil
}
