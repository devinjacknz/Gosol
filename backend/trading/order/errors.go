package order

import "errors"

var (
	// ErrOrderNotFound is returned when an order is not found
	ErrOrderNotFound = errors.New("order not found")

	// ErrOrderNotCancellable is returned when an order cannot be cancelled
	ErrOrderNotCancellable = errors.New("order not cancellable")

	// ErrOrderNotFillable is returned when an order cannot be filled
	ErrOrderNotFillable = errors.New("order not fillable")

	// ErrInvalidSymbol is returned when the symbol is invalid
	ErrInvalidSymbol = errors.New("invalid symbol")

	// ErrInvalidSize is returned when the order size is invalid
	ErrInvalidSize = errors.New("invalid order size")

	// ErrInvalidPrice is returned when the price is invalid
	ErrInvalidPrice = errors.New("invalid price")

	// ErrInvalidStopPrice is returned when the stop price is invalid
	ErrInvalidStopPrice = errors.New("invalid stop price")

	// ErrInvalidFilledSize is returned when the filled size is invalid
	ErrInvalidFilledSize = errors.New("invalid filled size")

	// ErrInvalidStatusTransition is returned when the status transition is invalid
	ErrInvalidStatusTransition = errors.New("invalid status transition")

	// ErrDuplicateClientOrderID is returned when the client order ID is already in use
	ErrDuplicateClientOrderID = errors.New("duplicate client order ID")

	// ErrInsufficientBalance is returned when there is insufficient balance
	ErrInsufficientBalance = errors.New("insufficient balance")

	// ErrOrderLimitExceeded is returned when the order limit is exceeded
	ErrOrderLimitExceeded = errors.New("order limit exceeded")

	// ErrMarketClosed is returned when the market is closed
	ErrMarketClosed = errors.New("market closed")
)
