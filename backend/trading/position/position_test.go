package position

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPositionManager(t *testing.T) {
	manager := NewManager()
	ctx := context.Background()

	t.Run("Open position", func(t *testing.T) {
		params := OpenPositionParams{
			Symbol:     "BTC/USD",
			Side:       Long,
			Size:       1.0,
			EntryPrice: 50000.0,
			Leverage:   10.0,
		}

		position, err := manager.OpenPosition(ctx, params)
		assert.NoError(t, err)
		assert.NotNil(t, position)
		assert.Equal(t, params.Symbol, position.Symbol)
		assert.Equal(t, params.Side, position.Side)
		assert.Equal(t, params.Size, position.Size)
		assert.Equal(t, params.EntryPrice, position.EntryPrice)
		assert.Equal(t, Open, position.Status)
	})

	t.Run("Open position with invalid parameters", func(t *testing.T) {
		invalidParams := []OpenPositionParams{
			{Symbol: "", Side: Long, Size: 1.0, EntryPrice: 50000.0, Leverage: 10.0},
			{Symbol: "BTC/USD", Side: Long, Size: 0, EntryPrice: 50000.0, Leverage: 10.0},
			{Symbol: "BTC/USD", Side: Long, Size: 1.0, EntryPrice: 0, Leverage: 10.0},
			{Symbol: "BTC/USD", Side: Long, Size: 1.0, EntryPrice: 50000.0, Leverage: 0},
		}

		for _, params := range invalidParams {
			position, err := manager.OpenPosition(ctx, params)
			assert.Error(t, err)
			assert.Nil(t, position)
		}
	})

	t.Run("Update position", func(t *testing.T) {
		// Open a position first
		openParams := OpenPositionParams{
			Symbol:     "ETH/USD",
			Side:       Long,
			Size:       2.0,
			EntryPrice: 3000.0,
			Leverage:   5.0,
		}

		position, err := manager.OpenPosition(ctx, openParams)
		assert.NoError(t, err)

		// Update position
		newPrice := 3500.0
		updateParams := UpdatePositionParams{
			CurrentPrice: &newPrice,
		}

		err = manager.UpdatePosition(ctx, position.ID, updateParams)
		assert.NoError(t, err)

		// Verify updates
		updatedPosition, err := manager.GetPosition(ctx, position.ID)
		assert.NoError(t, err)
		assert.Equal(t, newPrice, updatedPosition.CurrentPrice)
		assert.True(t, updatedPosition.UnrealizedPnL > 0) // Price increased for long position
	})

	t.Run("Close position", func(t *testing.T) {
		// Open a position first
		openParams := OpenPositionParams{
			Symbol:     "SOL/USD",
			Side:       Short,
			Size:       10.0,
			EntryPrice: 100.0,
			Leverage:   2.0,
		}

		position, err := manager.OpenPosition(ctx, openParams)
		assert.NoError(t, err)

		// Close position
		closePrice := 90.0
		err = manager.ClosePosition(ctx, position.ID, closePrice)
		assert.NoError(t, err)

		// Verify closure
		closedPosition, err := manager.GetPosition(ctx, position.ID)
		assert.NoError(t, err)
		assert.Equal(t, Closed, closedPosition.Status)
		assert.True(t, closedPosition.RealizedPnL > 0) // Price decreased for short position
	})

	t.Run("List positions", func(t *testing.T) {
		// Open multiple positions
		symbols := []string{"BTC/USD", "ETH/USD", "SOL/USD"}
		sides := []Side{Long, Short, Long}
		sizes := []float64{1.0, 2.0, 3.0}

		for i := range symbols {
			params := OpenPositionParams{
				Symbol:     symbols[i],
				Side:       sides[i],
				Size:       sizes[i],
				EntryPrice: 1000.0,
				Leverage:   10.0,
			}
			_, err := manager.OpenPosition(ctx, params)
			assert.NoError(t, err)
		}

		// Test different filters
		tests := []struct {
			name     string
			filter   PositionFilter
			expected int
		}{
			{
				name:     "No filter",
				filter:   PositionFilter{},
				expected: len(symbols),
			},
			{
				name: "Filter by symbol",
				filter: PositionFilter{
					Symbol: "BTC/USD",
				},
				expected: 1,
			},
			{
				name: "Filter by side",
				filter: PositionFilter{
					Side: &sides[0],
				},
				expected: 2, // Two long positions
			},
			{
				name: "Filter by min size",
				filter: PositionFilter{
					MinSize: &sizes[1], // 2.0
				},
				expected: 2, // Two positions >= 2.0
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				positions, err := manager.ListPositions(ctx, tt.filter)
				assert.NoError(t, err)
				assert.Len(t, positions, tt.expected)
			})
		}
	})
}

func TestPnLCalculations(t *testing.T) {
	t.Run("Long position PnL", func(t *testing.T) {
		position := &Position{
			Side:       Long,
			EntryPrice: 100.0,
			Size:       1.0,
		}

		// Profit scenario
		position.CurrentPrice = 120.0
		profit := calculateUnrealizedPnL(position)
		assert.Equal(t, 20.0, profit)

		// Loss scenario
		position.CurrentPrice = 80.0
		loss := calculateUnrealizedPnL(position)
		assert.Equal(t, -20.0, loss)
	})

	t.Run("Short position PnL", func(t *testing.T) {
		position := &Position{
			Side:       Short,
			EntryPrice: 100.0,
			Size:       1.0,
		}

		// Profit scenario
		position.CurrentPrice = 80.0
		profit := calculateUnrealizedPnL(position)
		assert.Equal(t, 20.0, profit)

		// Loss scenario
		position.CurrentPrice = 120.0
		loss := calculateUnrealizedPnL(position)
		assert.Equal(t, -20.0, loss)
	})
}

func TestMarginCalculations(t *testing.T) {
	tests := []struct {
		name     string
		size     float64
		price    float64
		leverage float64
		expected float64
	}{
		{
			name:     "Basic calculation",
			size:     1.0,
			price:    100.0,
			leverage: 10.0,
			expected: 10.0,
		},
		{
			name:     "High leverage",
			size:     1.0,
			price:    100.0,
			leverage: 100.0,
			expected: 1.0,
		},
		{
			name:     "Large position",
			size:     10.0,
			price:    100.0,
			leverage: 10.0,
			expected: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			margin := calculateMargin(tt.size, tt.price, tt.leverage)
			assert.Equal(t, tt.expected, margin)
		})
	}
}

func TestConcurrency(t *testing.T) {
	manager := NewManager()
	ctx := context.Background()
	numOperations := 100

	// Concurrent position opening
	t.Run("Concurrent open positions", func(t *testing.T) {
		done := make(chan bool)
		for i := 0; i < numOperations; i++ {
			go func(idx int) {
				params := OpenPositionParams{
					Symbol:     "BTC/USD",
					Side:       Long,
					Size:       1.0,
					EntryPrice: 50000.0,
					Leverage:   10.0,
				}
				_, err := manager.OpenPosition(ctx, params)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		for i := 0; i < numOperations; i++ {
			<-done
		}

		positions, err := manager.ListPositions(ctx, PositionFilter{})
		assert.NoError(t, err)
		assert.Len(t, positions, numOperations)
	})

	// Concurrent position updates
	t.Run("Concurrent position updates", func(t *testing.T) {
		position, err := manager.OpenPosition(ctx, OpenPositionParams{
			Symbol:     "ETH/USD",
			Side:       Long,
			Size:       1.0,
			EntryPrice: 3000.0,
			Leverage:   10.0,
		})
		assert.NoError(t, err)

		done := make(chan bool)
		for i := 0; i < numOperations; i++ {
			go func(idx int) {
				price := float64(3000 + idx)
				err := manager.UpdatePosition(ctx, position.ID, UpdatePositionParams{
					CurrentPrice: &price,
				})
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		for i := 0; i < numOperations; i++ {
			<-done
		}
	})
}
