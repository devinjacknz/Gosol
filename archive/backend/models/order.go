package models

import "time"

// OrderType represents the type of order
type OrderType string

const (
	OrderTypeMarket OrderType = "MARKET"
	OrderTypeLimit  OrderType = "LIMIT"
	OrderTypeStop   OrderType = "STOP"
)

// OrderSide represents the side of order
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// OrderStatus represents the status of order
type OrderStatus string

const (
	OrderStatusNew      OrderStatus = "NEW"
	OrderStatusPartial  OrderStatus = "PARTIAL"
	OrderStatusFilled   OrderStatus = "FILLED"
	OrderStatusCanceled OrderStatus = "CANCELED"
	OrderStatusRejected OrderStatus = "REJECTED"
)

// Order represents a trading order
type Order struct {
	ID            string      `json:"id"`
	UserID        string      `json:"userId"`
	Symbol        string      `json:"symbol"`
	Type          OrderType   `json:"type"`
	Side          OrderSide   `json:"side"`
	Price         float64     `json:"price"`
	Amount        float64     `json:"amount"`
	FilledAmount  float64     `json:"filledAmount"`
	Status        OrderStatus `json:"status"`
	CreatedAt     time.Time   `json:"createdAt"`
	UpdatedAt     time.Time   `json:"updatedAt"`
	StopPrice     *float64    `json:"stopPrice,omitempty"`
	ClientOrderID string      `json:"clientOrderId,omitempty"`
}

// OrderFill represents a fill of an order
type OrderFill struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"orderId"`
	Price     float64   `json:"price"`
	Amount    float64   `json:"amount"`
	Fee       float64   `json:"fee"`
	CreatedAt time.Time `json:"createdAt"`
}

// Position represents a trading position
type Position struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Symbol    string    `json:"symbol"`
	Side      OrderSide `json:"side"`
	Amount    float64   `json:"amount"`
	AvgPrice  float64   `json:"avgPrice"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// OrderRequest represents an order creation request
type OrderRequest struct {
	Symbol        string    `json:"symbol"`
	Type          OrderType `json:"type"`
	Side          OrderSide `json:"side"`
	Price         float64   `json:"price,omitempty"`
	Amount        float64   `json:"amount"`
	StopPrice     *float64  `json:"stopPrice,omitempty"`
	ClientOrderID string    `json:"clientOrderId,omitempty"`
}
