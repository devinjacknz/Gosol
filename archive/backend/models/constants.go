package models

// Position side constants
const (
	PositionSideLong  = "long"
	PositionSideShort = "short"
)

// Position status constants
const (
	PositionStatusPending    = "pending"
	PositionStatusOpen       = "open"
	PositionStatusClosed     = "closed"
	PositionStatusLiquidated = "liquidated"
	PositionStatusCancelled  = "cancelled"
)

// Trade status constants
const (
	TradeStatusPending    = "pending"
	TradeStatusCompleted  = "completed"
	TradeStatusFailed     = "failed"
	TradeStatusCancelled  = "cancelled"
)

// Order type constants
const (
	OrderTypeMarket = "market"
	OrderTypeLimit  = "limit"
	OrderTypeStop   = "stop"
)

// Order side constants
const (
	OrderSideBuy  = "buy"
	OrderSideSell = "sell"
)

// Risk level constants
const (
	RiskLevelLow    = "low"
	RiskLevelMedium = "medium"
	RiskLevelHigh   = "high"
)

// Market status constants
const (
	MarketStatusActive    = "active"
	MarketStatusInactive  = "inactive"
	MarketStatusHalted    = "halted"
	MarketStatusDelisted  = "delisted"
)

// Token status constants
const (
	TokenStatusActive    = "active"
	TokenStatusInactive  = "inactive"
	TokenStatusBlocked   = "blocked"
	TokenStatusDelisted  = "delisted"
)

// Wallet status constants
const (
	WalletStatusActive    = "active"
	WalletStatusInactive  = "inactive"
	WalletStatusBlocked   = "blocked"
)

// Transaction status constants
const (
	TxStatusPending    = "pending"
	TxStatusConfirmed  = "confirmed"
	TxStatusFailed     = "failed"
	TxStatusCancelled  = "cancelled"
)

// Error codes
const (
	ErrorCodeInvalidInput      = "invalid_input"
	ErrorCodeInsufficientFunds = "insufficient_funds"
	ErrorCodeMarketClosed      = "market_closed"
	ErrorCodeTokenBlocked      = "token_blocked"
	ErrorCodeWalletBlocked     = "wallet_blocked"
	ErrorCodeTxFailed         = "tx_failed"
	ErrorCodeInternalError    = "internal_error"
)

// Event types
const (
	EventTypeMarketData  = "market_data"
	EventTypeTradeSignal = "trade_signal"
	EventTypeTrade       = "trade"
	EventTypePosition    = "position"
	EventTypeWallet      = "wallet"
	EventTypeSystem      = "system"
)

// Severity levels
const (
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityError    = "error"
	SeverityCritical = "critical"
)
