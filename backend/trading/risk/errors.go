package risk

import (
	"fmt"
)

// Error types
const (
	ErrInsufficientBalance = "insufficient_balance"
	ErrInvalidAmount      = "invalid_amount"
	ErrInvalidPrice       = "invalid_price"
	ErrInvalidToken       = "invalid_token"
	ErrMaxPositions       = "max_positions_reached"
	ErrMaxDrawdown        = "max_drawdown_reached"
	ErrMaxDailyLoss      = "max_daily_loss_reached"
	ErrInsufficientLiquidity = "insufficient_liquidity"
	ErrHighPriceImpact    = "high_price_impact"
	ErrInvalidPosition    = "invalid_position"
	ErrPositionNotFound   = "position_not_found"
	ErrTradeNotAllowed    = "trade_not_allowed"
)

// RiskError represents a risk management error
type RiskError struct {
	Type    string
	Message string
	Details map[string]interface{}
}

// Error implements the error interface
func (e *RiskError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// NewRiskError creates a new risk error
func NewRiskError(errType string, message string, details map[string]interface{}) *RiskError {
	return &RiskError{
		Type:    errType,
		Message: message,
		Details: details,
	}
}

// IsRiskError checks if an error is a RiskError
func IsRiskError(err error) bool {
	_, ok := err.(*RiskError)
	return ok
}

// GetRiskErrorType returns the type of a RiskError
func GetRiskErrorType(err error) string {
	if riskErr, ok := err.(*RiskError); ok {
		return riskErr.Type
	}
	return ""
}

// Error constructors for common risk errors

// NewInsufficientBalanceError creates a new insufficient balance error
func NewInsufficientBalanceError(required, available float64) *RiskError {
	return NewRiskError(
		ErrInsufficientBalance,
		fmt.Sprintf("insufficient balance: required %.2f, available %.2f", required, available),
		map[string]interface{}{
			"required":  required,
			"available": available,
		},
	)
}

// NewMaxPositionsError creates a new max positions error
func NewMaxPositionsError(current, max int) *RiskError {
	return NewRiskError(
		ErrMaxPositions,
		fmt.Sprintf("maximum positions reached: %d/%d", current, max),
		map[string]interface{}{
			"current": current,
			"max":     max,
		},
	)
}

// NewMaxDrawdownError creates a new max drawdown error
func NewMaxDrawdownError(current, max float64) *RiskError {
	return NewRiskError(
		ErrMaxDrawdown,
		fmt.Sprintf("maximum drawdown reached: %.2f%% > %.2f%%", current, max),
		map[string]interface{}{
			"current": current,
			"max":     max,
		},
	)
}

// NewMaxDailyLossError creates a new max daily loss error
func NewMaxDailyLossError(current, max float64) *RiskError {
	return NewRiskError(
		ErrMaxDailyLoss,
		fmt.Sprintf("maximum daily loss reached: %.2f > %.2f", current, max),
		map[string]interface{}{
			"current": current,
			"max":     max,
		},
	)
}

// NewInsufficientLiquidityError creates a new insufficient liquidity error
func NewInsufficientLiquidityError(required, available float64, token string) *RiskError {
	return NewRiskError(
		ErrInsufficientLiquidity,
		fmt.Sprintf("insufficient liquidity for %s: required %.2f, available %.2f", token, required, available),
		map[string]interface{}{
			"token":     token,
			"required":  required,
			"available": available,
		},
	)
}

// NewHighPriceImpactError creates a new high price impact error
func NewHighPriceImpactError(impact, max float64) *RiskError {
	return NewRiskError(
		ErrHighPriceImpact,
		fmt.Sprintf("price impact too high: %.2f%% > %.2f%%", impact, max),
		map[string]interface{}{
			"impact": impact,
			"max":    max,
		},
	)
}

// NewInvalidPositionError creates a new invalid position error
func NewInvalidPositionError(reason string, details map[string]interface{}) *RiskError {
	return NewRiskError(
		ErrInvalidPosition,
		fmt.Sprintf("invalid position: %s", reason),
		details,
	)
}

// NewPositionNotFoundError creates a new position not found error
func NewPositionNotFoundError(id string) *RiskError {
	return NewRiskError(
		ErrPositionNotFound,
		fmt.Sprintf("position not found: %s", id),
		map[string]interface{}{
			"position_id": id,
		},
	)
}

// NewTradeNotAllowedError creates a new trade not allowed error
func NewTradeNotAllowedError(reason string, details map[string]interface{}) *RiskError {
	return NewRiskError(
		ErrTradeNotAllowed,
		fmt.Sprintf("trade not allowed: %s", reason),
		details,
	)
}
