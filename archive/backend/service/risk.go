package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/your-username/gosol/backend/models"
)

var (
	ErrRiskLimitExceeded = errors.New("risk limit exceeded")
	ErrInvalidRiskLimit  = errors.New("invalid risk limit")
)

// RiskLimit represents a risk control limit
type RiskLimit struct {
	UserID         string    `json:"userId"`
	Symbol         string    `json:"symbol"`
	MaxPosition    float64   `json:"maxPosition"`    // Maximum position size
	MaxLeverage    float64   `json:"maxLeverage"`    // Maximum leverage
	MaxDrawdown    float64   `json:"maxDrawdown"`    // Maximum drawdown percentage
	DailyLossLimit float64   `json:"dailyLossLimit"` // Maximum daily loss
	UpdatedAt      time.Time `json:"updatedAt"`
}

// RiskMetrics represents current risk metrics
type RiskMetrics struct {
	UserID          string    `json:"userId"`
	Symbol          string    `json:"symbol"`
	CurrentPosition float64   `json:"currentPosition"`
	CurrentLeverage float64   `json:"currentLeverage"`
	DailyPnL        float64   `json:"dailyPnL"`
	Drawdown        float64   `json:"drawdown"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// RiskService handles risk control
type RiskService struct {
	// Dependencies
	orderService *OrderService
	marketData   *MarketDataService

	// Risk limits and metrics
	riskLimits  map[string]*RiskLimit   // userID_symbol -> limit
	riskMetrics map[string]*RiskMetrics // userID_symbol -> metrics

	// Daily PnL tracking
	dailyPnL map[string]float64 // userID_symbol -> daily PnL

	// Mutex for thread safety
	mu sync.RWMutex
}

// NewRiskService creates a new risk service
func NewRiskService(orderService *OrderService, marketData *MarketDataService) *RiskService {
	service := &RiskService{
		orderService: orderService,
		marketData:   marketData,
		riskLimits:   make(map[string]*RiskLimit),
		riskMetrics:  make(map[string]*RiskMetrics),
		dailyPnL:     make(map[string]float64),
	}

	go service.startDailyReset()
	return service
}

// SetRiskLimit sets risk limits for a user and symbol
func (s *RiskService) SetRiskLimit(ctx context.Context, limit *RiskLimit) error {
	if err := s.validateRiskLimit(limit); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s_%s", limit.UserID, limit.Symbol)
	limit.UpdatedAt = time.Now()
	s.riskLimits[key] = limit

	return nil
}

// CheckOrderRisk checks if an order violates risk limits
func (s *RiskService) CheckOrderRisk(ctx context.Context, order *models.Order) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := fmt.Sprintf("%s_%s", order.UserID, order.Symbol)
	limit := s.riskLimits[key]
	if limit == nil {
		return nil // No limits set
	}

	// Check position limit
	position, err := s.orderService.GetPosition(ctx, order.UserID, order.Symbol)
	if err != nil {
		return err
	}

	newPosition := position.Amount
	if order.Side == models.OrderSideBuy {
		newPosition += order.Amount
	} else {
		newPosition -= order.Amount
	}

	if abs(newPosition) > limit.MaxPosition {
		return fmt.Errorf("%w: position size exceeds limit", ErrRiskLimitExceeded)
	}

	// Check daily loss limit
	if s.dailyPnL[key] < -limit.DailyLossLimit {
		return fmt.Errorf("%w: daily loss limit exceeded", ErrRiskLimitExceeded)
	}

	return nil
}

// UpdateRiskMetrics updates risk metrics for a user and symbol
func (s *RiskService) UpdateRiskMetrics(ctx context.Context, userID, symbol string) error {
	position, err := s.orderService.GetPosition(ctx, userID, symbol)
	if err != nil {
		return err
	}

	marketData := s.marketData.GetLatestMarketData(symbol)
	if marketData == nil {
		return errors.New("no market data available")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s_%s", userID, symbol)
	metrics := &RiskMetrics{
		UserID:          userID,
		Symbol:          symbol,
		CurrentPosition: position.Amount,
		CurrentLeverage: calculateLeverage(position, marketData),
		DailyPnL:        s.dailyPnL[key],
		Drawdown:        calculateDrawdown(position, marketData),
		UpdatedAt:       time.Now(),
	}

	s.riskMetrics[key] = metrics

	// Check if any risk limits are breached
	if limit := s.riskLimits[key]; limit != nil {
		if metrics.CurrentLeverage > limit.MaxLeverage {
			return fmt.Errorf("%w: leverage exceeds limit", ErrRiskLimitExceeded)
		}
		if metrics.Drawdown > limit.MaxDrawdown {
			return fmt.Errorf("%w: drawdown exceeds limit", ErrRiskLimitExceeded)
		}
	}

	return nil
}

// GetRiskMetrics returns risk metrics for a user and symbol
func (s *RiskService) GetRiskMetrics(ctx context.Context, userID, symbol string) *RiskMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := fmt.Sprintf("%s_%s", userID, symbol)
	return s.riskMetrics[key]
}

// validateRiskLimit validates risk limit values
func (s *RiskService) validateRiskLimit(limit *RiskLimit) error {
	if limit.MaxPosition <= 0 {
		return fmt.Errorf("%w: max position must be positive", ErrInvalidRiskLimit)
	}
	if limit.MaxLeverage <= 0 {
		return fmt.Errorf("%w: max leverage must be positive", ErrInvalidRiskLimit)
	}
	if limit.MaxDrawdown <= 0 {
		return fmt.Errorf("%w: max drawdown must be positive", ErrInvalidRiskLimit)
	}
	if limit.DailyLossLimit <= 0 {
		return fmt.Errorf("%w: daily loss limit must be positive", ErrInvalidRiskLimit)
	}
	return nil
}

// startDailyReset starts a routine to reset daily PnL
func (s *RiskService) startDailyReset() {
	ticker := time.NewTicker(24 * time.Hour)
	for range ticker.C {
		s.resetDailyPnL()
	}
}

// resetDailyPnL resets daily PnL tracking
func (s *RiskService) resetDailyPnL() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dailyPnL = make(map[string]float64)
}

// Helper functions

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func calculateLeverage(position *models.Position, marketData *models.MarketData) float64 {
	if position == nil || marketData == nil {
		return 0
	}
	// Implement leverage calculation based on your business logic
	return abs(position.Amount * marketData.Price / 100000) // Example calculation
}

func calculateDrawdown(position *models.Position, marketData *models.MarketData) float64 {
	if position == nil || marketData == nil {
		return 0
	}
	// Implement drawdown calculation based on your business logic
	unrealizedPnL := (marketData.Price - position.AvgPrice) * position.Amount
	return -unrealizedPnL / (position.AvgPrice * abs(position.Amount)) * 100
}
