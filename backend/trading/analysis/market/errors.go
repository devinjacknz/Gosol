package market

import "errors"

var (
	// ErrInsufficientData is returned when there is not enough data for analysis
	ErrInsufficientData = errors.New("insufficient data for analysis")

	// ErrInvalidPeriod is returned when an invalid period is specified
	ErrInvalidPeriod = errors.New("invalid period specified")

	// ErrInvalidPrice is returned when a price value is invalid
	ErrInvalidPrice = errors.New("invalid price value")

	// ErrInvalidVolume is returned when a volume value is invalid
	ErrInvalidVolume = errors.New("invalid volume value")

	// ErrInvalidOrderBook is returned when order book data is invalid
	ErrInvalidOrderBook = errors.New("invalid order book data")

	// ErrMarketClosed is returned when the market is closed
	ErrMarketClosed = errors.New("market is closed")

	// ErrDataUnavailable is returned when required data is unavailable
	ErrDataUnavailable = errors.New("required data is unavailable")
)
