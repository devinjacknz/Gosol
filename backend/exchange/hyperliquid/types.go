package hyperliquid

import "time"

// Market represents a trading market
type Market struct {
	Symbol          string  `json:"symbol"`
	BaseCurrency    string  `json:"baseCurrency"`
	QuoteCurrency   string  `json:"quoteCurrency"`
	MinSize         float64 `json:"minSize"`
	PricePrecision  int     `json:"pricePrecision"`
	SizePrecision   int     `json:"sizePrecision"`
	MaxLeverage     int     `json:"maxLeverage"`
	FundingInterval int     `json:"fundingInterval"`
}

// Orderbook represents the order book
type Orderbook struct {
	Symbol string           `json:"symbol"`
	Bids   []OrderbookLevel `json:"bids"`
	Asks   []OrderbookLevel `json:"asks"`
	Time   time.Time        `json:"time"`
}

// OrderbookLevel represents a price level in the order book
type OrderbookLevel struct {
	Price float64 `json:"price"`
	Size  float64 `json:"size"`
}

// Trade represents a trade
type Trade struct {
	ID        string    `json:"id"`
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Size      float64   `json:"size"`
	Side      string    `json:"side"`
	Timestamp time.Time `json:"timestamp"`
}

// FundingRate represents the funding rate
type FundingRate struct {
	Symbol    string    `json:"symbol"`
	Rate      float64   `json:"rate"`
	Timestamp time.Time `json:"timestamp"`
}

// Order represents an order
type Order struct {
	ID            string    `json:"id"`
	ClientOrderID string    `json:"clientOrderId"`
	Symbol        string    `json:"symbol"`
	Type          string    `json:"type"`
	Side          string    `json:"side"`
	Price         float64   `json:"price"`
	StopPrice     float64   `json:"stopPrice,omitempty"`
	Size          float64   `json:"size"`
	FilledSize    float64   `json:"filledSize"`
	RemainingSize float64   `json:"remainingSize"`
	Leverage      int       `json:"leverage"`
	ReduceOnly    bool      `json:"reduceOnly"`
	PostOnly      bool      `json:"postOnly"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	ExpiresAt     time.Time `json:"expiresAt,omitempty"`
}

// CreateOrderRequest represents a request to create an order
type CreateOrderRequest struct {
	Symbol        string  `json:"symbol"`
	Type          string  `json:"type"`
	Side          string  `json:"side"`
	Price         float64 `json:"price,omitempty"`
	StopPrice     float64 `json:"stopPrice,omitempty"`
	Size          float64 `json:"size"`
	Leverage      int     `json:"leverage"`
	ReduceOnly    bool    `json:"reduceOnly"`
	PostOnly      bool    `json:"postOnly"`
	ClientOrderID string  `json:"clientOrderId"`
}

// Position represents a trading position
type Position struct {
	Symbol           string    `json:"symbol"`
	Side             string    `json:"side"`
	EntryPrice       float64   `json:"entryPrice"`
	CurrentPrice     float64   `json:"currentPrice"`
	Size             float64   `json:"size"`
	Leverage         int       `json:"leverage"`
	LiquidationPrice float64   `json:"liquidationPrice"`
	MarginUsed       float64   `json:"marginUsed"`
	UnrealizedPnL    float64   `json:"unrealizedPnl"`
	RealizedPnL      float64   `json:"realizedPnl"`
	FundingFee       float64   `json:"fundingFee"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// Balance represents account balance
type Balance struct {
	Currency          string  `json:"currency"`
	Available         float64 `json:"available"`
	Total             float64 `json:"total"`
	MarginUsed        float64 `json:"marginUsed"`
	UnrealizedPnL     float64 `json:"unrealizedPnl"`
	RealizedPnL       float64 `json:"realizedPnl"`
	TotalFundingFee   float64 `json:"totalFundingFee"`
	TotalTradingFee   float64 `json:"totalTradingFee"`
	MaintenanceMargin float64 `json:"maintenanceMargin"`
	InitialMargin     float64 `json:"initialMargin"`
}

// Constants for order types
const (
	OrderTypeMarket     = "MARKET"
	OrderTypeLimit      = "LIMIT"
	OrderTypeStopMarket = "STOP_MARKET"
	OrderTypeStopLimit  = "STOP_LIMIT"
)

// Constants for order sides
const (
	OrderSideBuy  = "BUY"
	OrderSideSell = "SELL"
)

// Constants for order status
const (
	OrderStatusCreated         = "CREATED"
	OrderStatusPending         = "PENDING"
	OrderStatusPartiallyFilled = "PARTIALLY_FILLED"
	OrderStatusFilled          = "FILLED"
	OrderStatusCancelled       = "CANCELLED"
	OrderStatusRejected        = "REJECTED"
	OrderStatusExpired         = "EXPIRED"
)

// Constants for position status
const (
	PositionStatusOpen       = "OPEN"
	PositionStatusClosed     = "CLOSED"
	PositionStatusLiquidated = "LIQUIDATED"
)
