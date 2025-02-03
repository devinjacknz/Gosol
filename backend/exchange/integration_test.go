package exchange

import (
	"context"
	"testing"
	"time"

	"github.com/devinjacknz/godydxhyber/backend/exchange/dydx"
	"github.com/devinjacknz/godydxhyber/backend/exchange/hyperliquid"
	"github.com/stretchr/testify/suite"
)

type ExchangeIntegrationSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *ExchangeIntegrationSuite) SetupSuite() {
	s.ctx = context.Background()
}

func (s *ExchangeIntegrationSuite) TestMarketDataFlow() {
	// Test market data flow for both exchanges
	s.Run("Hyperliquid Market Data", func() {
		client := newMockHyperliquidClient()

		// Test market list
		markets, err := client.GetMarkets(s.ctx)
		s.NoError(err)
		s.NotEmpty(markets)

		// Test orderbook
		orderbook, err := client.GetOrderbook(s.ctx, "BTC-USD")
		s.NoError(err)
		s.NotNil(orderbook)
		s.NotEmpty(orderbook.Bids)
		s.NotEmpty(orderbook.Asks)

		// Test trades
		trades, err := client.GetTrades(s.ctx, "BTC-USD", 10)
		s.NoError(err)
		s.Len(trades, 10)
	})

	s.Run("dYdX Market Data", func() {
		client := newMockDydxClient()

		// Test market list
		markets, err := client.GetMarkets(s.ctx)
		s.NoError(err)
		s.NotEmpty(markets)

		// Test orderbook
		orderbook, err := client.GetOrderbook(s.ctx, "BTC-USD")
		s.NoError(err)
		s.NotNil(orderbook)
		s.NotEmpty(orderbook.Bids)
		s.NotEmpty(orderbook.Asks)

		// Test trades
		trades, err := client.GetTrades(s.ctx, "BTC-USD", 10)
		s.NoError(err)
		s.Len(trades, 10)
	})
}

func (s *ExchangeIntegrationSuite) TestOrderFlow() {
	// Test order flow for both exchanges
	s.Run("Hyperliquid Order Flow", func() {
		client := newMockHyperliquidClient()

		// Create order
		order, err := client.CreateOrder(s.ctx, hyperliquid.CreateOrderRequest{
			Symbol: "BTC-USD",
			Side:   hyperliquid.OrderSideBuy,
			Type:   hyperliquid.OrderTypeLimit,
			Size:   1.0,
			Price:  50000.0,
		})
		s.NoError(err)
		s.NotNil(order)

		// Get order
		fetchedOrder, err := client.GetOrder(s.ctx, order.ID)
		s.NoError(err)
		s.Equal(order.ID, fetchedOrder.ID)

		// Cancel order
		err = client.CancelOrder(s.ctx, order.ID)
		s.NoError(err)

		// Verify cancelled
		fetchedOrder, err = client.GetOrder(s.ctx, order.ID)
		s.NoError(err)
		s.Equal("CANCELLED", fetchedOrder.Status)
	})

	s.Run("dYdX Order Flow", func() {
		client := newMockDydxClient()

		// Create order
		order, err := client.CreateOrder(s.ctx, dydx.CreateOrderRequest{
			Market: "BTC-USD",
			Side:   dydx.OrderSideBuy,
			Type:   dydx.OrderTypeLimit,
			Size:   1.0,
			Price:  50000.0,
		})
		s.NoError(err)
		s.NotNil(order)

		// Get order
		fetchedOrder, err := client.GetOrder(s.ctx, order.ID)
		s.NoError(err)
		s.Equal(order.ID, fetchedOrder.ID)

		// Cancel order
		err = client.CancelOrder(s.ctx, order.ID)
		s.NoError(err)

		// Verify cancelled
		fetchedOrder, err = client.GetOrder(s.ctx, order.ID)
		s.NoError(err)
		s.Equal("CANCELLED", fetchedOrder.Status)
	})
}

func (s *ExchangeIntegrationSuite) TestPositionFlow() {
	// Test position flow for both exchanges
	s.Run("Hyperliquid Position Flow", func() {
		client := newMockHyperliquidClient()

		// Get positions
		positions, err := client.GetPositions(s.ctx)
		s.NoError(err)
		s.Empty(positions)

		// Create position via order
		order, err := client.CreateOrder(s.ctx, hyperliquid.CreateOrderRequest{
			Symbol: "BTC-USD",
			Side:   hyperliquid.OrderSideBuy,
			Type:   hyperliquid.OrderTypeMarket,
			Size:   1.0,
		})
		s.NoError(err)
		s.NotNil(order)

		// Verify position created
		time.Sleep(time.Second) // Wait for order execution
		positions, err = client.GetPositions(s.ctx)
		s.NoError(err)
		s.Len(positions, 1)
		s.Equal("BTC-USD", positions[0].Symbol)
		s.Equal(1.0, positions[0].Size)
	})

	s.Run("dYdX Position Flow", func() {
		client := newMockDydxClient()

		// Get positions
		positions, err := client.GetPositions(s.ctx)
		s.NoError(err)
		s.Empty(positions)

		// Create position via order
		order, err := client.CreateOrder(s.ctx, dydx.CreateOrderRequest{
			Market: "BTC-USD",
			Side:   dydx.OrderSideBuy,
			Type:   dydx.OrderTypeMarket,
			Size:   1.0,
		})
		s.NoError(err)
		s.NotNil(order)

		// Verify position created
		time.Sleep(time.Second) // Wait for order execution
		positions, err = client.GetPositions(s.ctx)
		s.NoError(err)
		s.Len(positions, 1)
		s.Equal("BTC-USD", positions[0].Market)
		s.Equal(1.0, positions[0].Size)
	})
}

func TestExchangeIntegration(t *testing.T) {
	suite.Run(t, new(ExchangeIntegrationSuite))
}
