package order

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOrderManager(t *testing.T) {
	manager := NewOrderManager()
	ctx := context.Background()

	t.Run("Create order", func(t *testing.T) {
		price := 50000.0
		params := CreateOrderParams{
			Symbol:        "BTC/USD",
			Type:          Limit,
			Side:          Buy,
			Price:         &price,
			Size:          1.0,
			ClientOrderID: "test-order-1",
		}

		order, err := manager.CreateOrder(ctx, params)
		assert.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, params.Symbol, order.Symbol)
		assert.Equal(t, params.Type, order.Type)
		assert.Equal(t, params.Side, order.Side)
		assert.Equal(t, params.Price, order.Price)
		assert.Equal(t, params.Size, order.Size)
		assert.Equal(t, params.Size, order.RemainingSize)
		assert.Equal(t, Created, order.Status)
	})

	t.Run("Create order with invalid parameters", func(t *testing.T) {
		invalidParams := []CreateOrderParams{
			{Symbol: "", Type: Market, Side: Buy, Size: 1.0},
			{Symbol: "BTC/USD", Type: Market, Side: Buy, Size: 0},
			{Symbol: "BTC/USD", Type: Limit, Side: Buy, Size: 1.0}, // Missing price for limit order
		}

		for _, params := range invalidParams {
			order, err := manager.CreateOrder(ctx, params)
			assert.Error(t, err)
			assert.Nil(t, order)
		}
	})

	t.Run("Update order status", func(t *testing.T) {
		// Create a new order
		price := 3000.0
		params := CreateOrderParams{
			Symbol:        "ETH/USD",
			Type:          Limit,
			Side:          Sell,
			Price:         &price,
			Size:          2.0,
			ClientOrderID: "test-order-2",
		}

		order, err := manager.CreateOrder(ctx, params)
		assert.NoError(t, err)

		// Test valid status transitions
		validTransitions := []OrderStatus{Pending, PartiallyFilled, Filled}
		for _, status := range validTransitions {
			err = manager.UpdateOrderStatus(ctx, order.ID, status)
			assert.NoError(t, err)

			updatedOrder, err := manager.GetOrder(ctx, order.ID)
			assert.NoError(t, err)
			assert.Equal(t, status, updatedOrder.Status)
		}

		// Test invalid status transition
		err = manager.UpdateOrderStatus(ctx, order.ID, Created)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidStatusTransition, err)
	})

	t.Run("Cancel order", func(t *testing.T) {
		// Create a new order
		price := 100.0
		params := CreateOrderParams{
			Symbol:        "SOL/USD",
			Type:          Limit,
			Side:          Buy,
			Price:         &price,
			Size:          10.0,
			ClientOrderID: "test-order-3",
		}

		order, err := manager.CreateOrder(ctx, params)
		assert.NoError(t, err)

		// Cancel the order
		err = manager.CancelOrder(ctx, order.ID)
		assert.NoError(t, err)

		// Verify cancellation
		cancelledOrder, err := manager.GetOrder(ctx, order.ID)
		assert.NoError(t, err)
		assert.Equal(t, Cancelled, cancelledOrder.Status)

		// Try to cancel again
		err = manager.CancelOrder(ctx, order.ID)
		assert.Error(t, err)
		assert.Equal(t, ErrOrderNotCancellable, err)
	})

	t.Run("Update filled size", func(t *testing.T) {
		// Create a new order
		price := 200.0
		params := CreateOrderParams{
			Symbol:        "DOT/USD",
			Type:          Limit,
			Side:          Buy,
			Price:         &price,
			Size:          5.0,
			ClientOrderID: "test-order-4",
		}

		order, err := manager.CreateOrder(ctx, params)
		assert.NoError(t, err)

		// Update status to pending
		err = manager.UpdateOrderStatus(ctx, order.ID, Pending)
		assert.NoError(t, err)

		// Test partial fill
		err = manager.UpdateFilledSize(ctx, order.ID, 2.0)
		assert.NoError(t, err)

		partialOrder, err := manager.GetOrder(ctx, order.ID)
		assert.NoError(t, err)
		assert.Equal(t, PartiallyFilled, partialOrder.Status)
		assert.Equal(t, 2.0, partialOrder.FilledSize)
		assert.Equal(t, 3.0, partialOrder.RemainingSize)

		// Test complete fill
		err = manager.UpdateFilledSize(ctx, order.ID, 3.0)
		assert.NoError(t, err)

		filledOrder, err := manager.GetOrder(ctx, order.ID)
		assert.NoError(t, err)
		assert.Equal(t, Filled, filledOrder.Status)
		assert.Equal(t, 5.0, filledOrder.FilledSize)
		assert.Equal(t, 0.0, filledOrder.RemainingSize)

		// Test overfill
		err = manager.UpdateFilledSize(ctx, order.ID, 1.0)
		assert.Error(t, err)
		assert.Equal(t, ErrOrderNotFillable, err)
	})

	t.Run("List orders", func(t *testing.T) {
		// Create multiple orders
		symbols := []string{"BTC/USD", "ETH/USD", "SOL/USD"}
		sides := []OrderSide{Buy, Sell, Buy}
		types := []OrderType{Market, Limit, StopLoss}
		prices := []float64{60000.0, 4000.0, 200.0}

		for i := range symbols {
			var price *float64
			if types[i] != Market {
				price = &prices[i]
			}

			params := CreateOrderParams{
				Symbol:        symbols[i],
				Type:          types[i],
				Side:          sides[i],
				Price:         price,
				Size:          1.0,
				ClientOrderID: "test-order-list-" + symbols[i],
			}

			_, err := manager.CreateOrder(ctx, params)
			assert.NoError(t, err)
		}

		// Test different filters
		tests := []struct {
			name     string
			filter   OrderFilter
			expected int
		}{
			{
				name:     "No filter",
				filter:   OrderFilter{},
				expected: len(symbols),
			},
			{
				name: "Filter by symbol",
				filter: OrderFilter{
					Symbol: "BTC/USD",
				},
				expected: 1,
			},
			{
				name: "Filter by side",
				filter: OrderFilter{
					Side: &sides[0],
				},
				expected: 2, // Two buy orders
			},
			{
				name: "Filter by type",
				filter: OrderFilter{
					Type: &types[1],
				},
				expected: 1, // One limit order
			},
			{
				name: "Filter by time range",
				filter: OrderFilter{
					StartTime: &time.Time{},
					EndTime:   &time.Time{},
				},
				expected: len(symbols),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				orders, err := manager.ListOrders(ctx, tt.filter)
				assert.NoError(t, err)
				assert.Len(t, orders, tt.expected)
			})
		}
	})
}

func TestConcurrency(t *testing.T) {
	manager := NewOrderManager()
	ctx := context.Background()
	numOperations := 100

	t.Run("Concurrent order creation", func(t *testing.T) {
		done := make(chan bool)
		price := 50000.0

		for i := 0; i < numOperations; i++ {
			go func(idx int) {
				params := CreateOrderParams{
					Symbol:        "BTC/USD",
					Type:          Limit,
					Side:          Buy,
					Price:         &price,
					Size:          1.0,
					ClientOrderID: "test-concurrent-" + string(rune(idx)),
				}

				_, err := manager.CreateOrder(ctx, params)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		for i := 0; i < numOperations; i++ {
			<-done
		}

		orders, err := manager.ListOrders(ctx, OrderFilter{})
		assert.NoError(t, err)
		assert.Len(t, orders, numOperations)
	})

	t.Run("Concurrent order updates", func(t *testing.T) {
		price := 3000.0
		params := CreateOrderParams{
			Symbol:        "ETH/USD",
			Type:          Limit,
			Side:          Buy,
			Price:         &price,
			Size:          1.0,
			ClientOrderID: "test-concurrent-updates",
		}

		order, err := manager.CreateOrder(ctx, params)
		assert.NoError(t, err)

		err = manager.UpdateOrderStatus(ctx, order.ID, Pending)
		assert.NoError(t, err)

		done := make(chan bool)
		for i := 0; i < numOperations; i++ {
			go func() {
				err := manager.UpdateFilledSize(ctx, order.ID, 0.01)
				if err != nil && err != ErrOrderNotFillable {
					assert.NoError(t, err)
				}
				done <- true
			}()
		}

		for i := 0; i < numOperations; i++ {
			<-done
		}

		updatedOrder, err := manager.GetOrder(ctx, order.ID)
		assert.NoError(t, err)
		assert.True(t, updatedOrder.FilledSize > 0)
	})
}
