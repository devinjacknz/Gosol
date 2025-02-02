package dydx

import "time"

// Market represents a trading market
type Market struct {
	Symbol            string  `json:"market"`
	BaseCurrency      string  `json:"baseCurrency"`
	QuoteCurrency     string  `json:"quoteCurrency"`
	MinOrderSize      float64 `json:"minOrderSize"`
	TickSize          float64 `json:"tickSize"`
	StepSize          float64 `json:"stepSize"`
	MaxLeverage       int     `json:"maxLeverage"`
	InitialMargin     float64 `json:"initialMargin"`
	MaintenanceMargin float64 `json:"maintenanceMargin"`
	FundingInterval   int     `json:"fundingInterval"`
	Status            string  `json:"status"`
}

// Orderbook represents the order book
type Orderbook struct {
	Market string           `json:"market"`
	Bids   []OrderbookLevel `json:"bids"`
	Asks   []OrderbookLevel `json:"asks"`
	Time   time.Time        `json:"time"`
}

// OrderbookLevel represents a price level in the order book
type OrderbookLevel struct {
	Price     float64 `json:"price"`
	Size      float64 `json:"size"`
	NumOrders int     `json:"numOrders"`
}

// Trade represents a trade
type Trade struct {
	ID          string    `json:"id"`
	Market      string    `json:"market"`
	Price       float64   `json:"price"`
	Size        float64   `json:"size"`
	Side        string    `json:"side"`
	Liquidation bool      `json:"liquidation"`
	Time        time.Time `json:"time"`
}

// FundingRate represents the funding rate
type FundingRate struct {
	Market string    `json:"market"`
	Rate   float64   `json:"rate"`
	Price  float64   `json:"price"`
	Time   time.Time `json:"time"`
}

// Order represents an order
type Order struct {
	ID            string    `json:"id"`
	ClientID      string    `json:"clientId"`
	Market        string    `json:"market"`
	Type          string    `json:"type"`
	Side          string    `json:"side"`
	Price         float64   `json:"price"`
	TriggerPrice  float64   `json:"triggerPrice,omitempty"`
	Size          float64   `json:"size"`
	FilledSize    float64   `json:"filledSize"`
	RemainingSize float64   `json:"remainingSize"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	ExpiresAt     time.Time `json:"expiresAt,omitempty"`
	PostOnly      bool      `json:"postOnly"`
	ReduceOnly    bool      `json:"reduceOnly"`
}

// CreateOrderRequest represents a request to create an order
type CreateOrderRequest struct {
	Market       string  `json:"market"`
	Side         string  `json:"side"`
	Type         string  `json:"type"`
	Size         float64 `json:"size"`
	Price        float64 `json:"price,omitempty"`
	TriggerPrice float64 `json:"triggerPrice,omitempty"`
	PostOnly     bool    `json:"postOnly"`
	ReduceOnly   bool    `json:"reduceOnly"`
	ClientID     string  `json:"clientId"`
	ExpiresAt    int64   `json:"expiresAt,omitempty"`
}

// Position represents a position
type Position struct {
	Market            string    `json:"market"`
	Side              string    `json:"side"`
	Size              float64   `json:"size"`
	EntryPrice        float64   `json:"entryPrice"`
	MarkPrice         float64   `json:"markPrice"`
	LiquidationPrice  float64   `json:"liquidationPrice"`
	Leverage          int       `json:"leverage"`
	MaxLeverage       int       `json:"maxLeverage"`
	InitialMargin     float64   `json:"initialMargin"`
	MaintenanceMargin float64   `json:"maintenanceMargin"`
	UnrealizedPnl     float64   `json:"unrealizedPnl"`
	RealizedPnl       float64   `json:"realizedPnl"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// Account represents account information
type Account struct {
	ID               string    `json:"id"`
	StarkKey         string    `json:"starkKey"`
	Equity           float64   `json:"equity"`
	FreeCollateral   float64   `json:"freeCollateral"`
	QuoteBalance     float64   `json:"quoteBalance"`
	PendingDeposits  float64   `json:"pendingDeposits"`
	PendingWithdraws float64   `json:"pendingWithdraws"`
	CreatedAt        time.Time `json:"createdAt"`
	ActivePositions  int       `json:"activePositions"`
	OpenOrders       int       `json:"openOrders"`
}

// Balance represents account balance
type Balance struct {
	Currency          string  `json:"currency"`
	Balance           float64 `json:"balance"`
	Available         float64 `json:"available"`
	InitialMargin     float64 `json:"initialMargin"`
	MaintenanceMargin float64 `json:"maintenanceMargin"`
	UnrealizedPnl     float64 `json:"unrealizedPnl"`
	RealizedPnl       float64 `json:"realizedPnl"`
}

// Constants for order types
const (
	OrderTypeMarket     = "MARKET"
	OrderTypeLimit      = "LIMIT"
	OrderTypeStopMarket = "STOP_MARKET"
	OrderTypeStopLimit  = "STOP_LIMIT"
	OrderTypeTakeProfit = "TAKE_PROFIT"
)

// Constants for order sides
const (
	OrderSideBuy  = "BUY"
	OrderSideSell = "SELL"
)

// Constants for order status
const (
	OrderStatusPending     = "PENDING"
	OrderStatusOpen        = "OPEN"
	OrderStatusFilled      = "FILLED"
	OrderStatusCanceled    = "CANCELED"
	OrderStatusUntriggered = "UNTRIGGERED"
	OrderStatusTriggered   = "TRIGGERED"
)

// Constants for position side
const (
	PositionSideLong  = "LONG"
	PositionSideShort = "SHORT"
)

// Constants for market status
const (
	MarketStatusActive   = "ACTIVE"
	MarketStatusPaused   = "PAUSED"
	MarketStatusClosed   = "CLOSED"
	MarketStatusDelisted = "DELISTED"
)
