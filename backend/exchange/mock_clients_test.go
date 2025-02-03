package exchange

import (
	"context"
	"time"

	"github.com/devinjacknz/godydxhyber/backend/exchange/dydx"
	"github.com/devinjacknz/godydxhyber/backend/exchange/hyperliquid"
)

type mockHyperliquidClient struct{}

func newMockHyperliquidClient() *mockHyperliquidClient {
	return &mockHyperliquidClient{}
}

func (m *mockHyperliquidClient) GetMarkets(ctx context.Context) ([]hyperliquid.Market, error) {
	return []hyperliquid.Market{
		{
			Symbol:          "BTC-USD",
			BaseCurrency:    "BTC",
			QuoteCurrency:   "USD",
			MinSize:         0.001,
			PricePrecision:  8,
			SizePrecision:   4,
			MaxLeverage:     20,
			FundingInterval: 8,
		},
	}, nil
}

func (m *mockHyperliquidClient) GetOrderbook(ctx context.Context, symbol string) (*hyperliquid.Orderbook, error) {
	return &hyperliquid.Orderbook{
		Symbol: symbol,
		Bids: []hyperliquid.OrderbookLevel{
			{Price: 50000, Size: 1.0},
		},
		Asks: []hyperliquid.OrderbookLevel{
			{Price: 50100, Size: 1.0},
		},
		Time: time.Now(),
	}, nil
}

func (m *mockHyperliquidClient) GetTrades(ctx context.Context, symbol string, limit int) ([]hyperliquid.Trade, error) {
	trades := make([]hyperliquid.Trade, limit)
	for i := 0; i < limit; i++ {
		trades[i] = hyperliquid.Trade{
			ID:     string(rune(i)),
			Symbol:    symbol,
			Price:     50000.0,
			Size:      1.0,
			Side:      hyperliquid.OrderSideBuy,
			Timestamp: time.Now(),
		}
	}
	return trades, nil
}

func (m *mockHyperliquidClient) CreateOrder(ctx context.Context, req hyperliquid.CreateOrderRequest) (*hyperliquid.Order, error) {
	return &hyperliquid.Order{
		ID:            "mock-order-1",
		Symbol:        req.Symbol,
		Type:          req.Type,
		Side:          req.Side,
		Size:          req.Size,
		Price:         req.Price,
		Status:        hyperliquid.OrderStatusCreated,
		CreatedAt:     time.Now(),
	}, nil
}

func (m *mockHyperliquidClient) GetOrder(ctx context.Context, orderID string) (*hyperliquid.Order, error) {
	return &hyperliquid.Order{
		ID:         orderID,
		Symbol:     "BTC-USD",
		Type:      "LIMIT",
		Side:      "BUY",
		Size:      1.0,
		Price:     50000.0,
		Status:    "OPEN",
		CreatedAt: time.Now(),
	}, nil
}

func (m *mockHyperliquidClient) CancelOrder(ctx context.Context, orderID string) error {
	return nil
}

func (m *mockHyperliquidClient) GetPositions(ctx context.Context) ([]hyperliquid.Position, error) {
	return []hyperliquid.Position{
		{
			Symbol:           "BTC-USD",
			Side:             hyperliquid.OrderSideBuy,
			Size:             1.0,
			EntryPrice:       50000.0,
			CurrentPrice:     50100.0,
			Leverage:         10,
			LiquidationPrice: 45000.0,
			Status:           hyperliquid.PositionStatusOpen,
		},
	}, nil
}

type mockDydxClient struct{}

func newMockDydxClient() *mockDydxClient {
	return &mockDydxClient{}
}

func (m *mockDydxClient) GetMarkets(ctx context.Context) ([]dydx.Market, error) {
	return []dydx.Market{
		{
			Symbol:       "BTC-USD",
			BaseCurrency: "BTC",
			QuoteCurrency: "USD",
			MinOrderSize: 0.001,
			TickSize:    0.5,
			Status:      dydx.MarketStatusActive,
		},
	}, nil
}

func (m *mockDydxClient) GetOrderbook(ctx context.Context, symbol string) (*dydx.Orderbook, error) {
	return &dydx.Orderbook{
		Market: symbol,
		Bids: []dydx.OrderbookLevel{
			{Price: 50000, Size: 1.0},
		},
		Asks: []dydx.OrderbookLevel{
			{Price: 50100, Size: 1.0},
		},
		Time: time.Now(),
	}, nil
}

func (m *mockDydxClient) GetTrades(ctx context.Context, symbol string, limit int) ([]dydx.Trade, error) {
	trades := make([]dydx.Trade, limit)
	for i := 0; i < limit; i++ {
		trades[i] = dydx.Trade{
			ID:          string(rune(i)),
			Market:      symbol,
			Price:       50000.0,
			Size:        1.0,
			Side:        dydx.OrderSideBuy,
			Liquidation: false,
			Time:        time.Now(),
		}
	}
	return trades, nil
}

func (m *mockDydxClient) CreateOrder(ctx context.Context, req dydx.CreateOrderRequest) (*dydx.Order, error) {
	return &dydx.Order{
		ID:            "mock-order-1",
		Market:        req.Market,
		Type:         req.Type,
		Side:         req.Side,
		Size:         req.Size,
		Price:        req.Price,
		Status:       dydx.OrderStatusOpen,
		CreatedAt:    time.Now(),
	}, nil
}

func (m *mockDydxClient) GetOrder(ctx context.Context, orderID string) (*dydx.Order, error) {
	return &dydx.Order{
		ID:         orderID,
		Market:     "BTC-USD",
		Type:      dydx.OrderTypeLimit,
		Side:      dydx.OrderSideBuy,
		Size:      1.0,
		Price:     50000.0,
		Status:    dydx.OrderStatusPending,
		CreatedAt: time.Now(),
	}, nil
}

func (m *mockDydxClient) CancelOrder(ctx context.Context, orderID string) error {
	return nil
}

func (m *mockDydxClient) GetPositions(ctx context.Context) ([]dydx.Position, error) {
	return []dydx.Position{
		{
			Market:            "BTC-USD",
			Side:              dydx.PositionSideLong,
			Size:              1.0,
			EntryPrice:        50000.0,
			MarkPrice:         50100.0,
			LiquidationPrice:  45000.0,
			Leverage:          10,
			MaxLeverage:       20,
			InitialMargin:     5000.0,
			MaintenanceMargin: 2500.0,
			UnrealizedPnl:     100.0,
			RealizedPnl:       0.0,
		},
	}, nil
}
