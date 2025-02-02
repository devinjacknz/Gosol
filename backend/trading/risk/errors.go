package risk

import "errors"

var (
	// ErrLimitNotSet is returned when a limit is not set
	ErrLimitNotSet = errors.New("limit not set")

	// ErrInvalidLimit is returned when a limit value is invalid
	ErrInvalidLimit = errors.New("invalid limit value")

	// ErrPositionLimitExceeded is returned when position limit is exceeded
	ErrPositionLimitExceeded = errors.New("position limit exceeded")

	// ErrExposureLimitExceeded is returned when exposure limit is exceeded
	ErrExposureLimitExceeded = errors.New("exposure limit exceeded")

	// ErrDrawdownLimitExceeded is returned when drawdown limit is exceeded
	ErrDrawdownLimitExceeded = errors.New("drawdown limit exceeded")

	// ErrVolatilityTooHigh is returned when volatility exceeds critical threshold
	ErrVolatilityTooHigh = errors.New("volatility too high")

	// ErrInvalidThresholds is returned when volatility thresholds are invalid
	ErrInvalidThresholds = errors.New("invalid volatility thresholds")

	// ErrInvalidTimeWindow is returned when time window is invalid
	ErrInvalidTimeWindow = errors.New("invalid time window")

	// ErrInsufficientData is returned when there is insufficient data for calculation
	ErrInsufficientData = errors.New("insufficient data for calculation")

	// ErrInvalidSymbol is returned when the symbol is invalid
	ErrInvalidSymbol = errors.New("invalid symbol")
)
