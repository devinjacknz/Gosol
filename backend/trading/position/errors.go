package position

import "errors"

var (
	// ErrPositionNotFound is returned when a position is not found
	ErrPositionNotFound = errors.New("position not found")

	// ErrPositionAlreadyClosed is returned when trying to modify a closed position
	ErrPositionAlreadyClosed = errors.New("position already closed")

	// ErrInvalidSymbol is returned when the symbol is invalid
	ErrInvalidSymbol = errors.New("invalid symbol")

	// ErrInvalidSize is returned when the position size is invalid
	ErrInvalidSize = errors.New("invalid position size")

	// ErrInvalidPrice is returned when the price is invalid
	ErrInvalidPrice = errors.New("invalid price")

	// ErrInvalidLeverage is returned when the leverage is invalid
	ErrInvalidLeverage = errors.New("invalid leverage")

	// ErrInvalidStopLoss is returned when the stop loss is invalid
	ErrInvalidStopLoss = errors.New("invalid stop loss")

	// ErrInvalidTakeProfit is returned when the take profit is invalid
	ErrInvalidTakeProfit = errors.New("invalid take profit")

	// ErrInsufficientMargin is returned when there is insufficient margin
	ErrInsufficientMargin = errors.New("insufficient margin")

	// ErrPositionLimitExceeded is returned when position limit is exceeded
	ErrPositionLimitExceeded = errors.New("position limit exceeded")

	// ErrExposureLimitExceeded is returned when exposure limit is exceeded
	ErrExposureLimitExceeded = errors.New("exposure limit exceeded")

	// ErrInvalidPositionSide is returned when the position side is invalid
	ErrInvalidPositionSide = errors.New("invalid position side")
)
