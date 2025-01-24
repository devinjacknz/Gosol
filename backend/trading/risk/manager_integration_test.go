package risk

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRiskManagerIntegration(t *testing.T) {
	ctx := context.Background()

	// Initialize risk manager with test configuration
	config := RiskConfig{
		MaxPositionSize: 100.0,
		MaxDailyLoss:   10.0,  // 10%
		MaxDrawdown:    20.0,  // 20%
		StopLoss:      5.0,   // 5%
		TrailingStop:  2.0,   // 2%
		RiskPerTrade:  2.0,   // 2%
		MaxDailyTrades: 10,
		MinLiquidity:  1000.0,
		MaxSlippage:   1.0,   // 1%
	}

	initialPortfolio := 10000.0
	rm := NewRiskManager(config, initialPortfolio)

	t.Run("Position Management Flow", func(t *testing.T) {
		// Test opening a position
		token := "SOL"
		size := 10.0
		price := 100.0

		// Check if can open position
		err := rm.CanOpenPosition(ctx, token, size, price)
		assert.NoError(t, err)

		// Open position
		position, err := rm.OpenPosition(token, size, price, "long")
		assert.NoError(t, err)
		assert.NotNil(t, position)
		assert.Equal(t, token, position.TokenAddress)
		assert.Equal(t, price, position.EntryPrice)
		assert.Equal(t, size, position.Size)

		// Update position with price movement
		shouldClose, err := rm.UpdatePosition(token, 102.0)
		assert.NoError(t, err)
		assert.False(t, shouldClose)

		// Test trailing stop adjustment
		currentPosition := rm.GetPositions()[token]
		assert.Greater(t, currentPosition.StopLoss, position.StopLoss)

		// Test stop loss hit
		shouldClose, err = rm.UpdatePosition(token, 95.0)
		assert.NoError(t, err)
		assert.True(t, shouldClose)

		// Close position
		pnl, err := rm.ClosePosition(token, 95.0)
		assert.NoError(t, err)
		assert.Less(t, pnl, 0.0)
	})

	t.Run("Risk Limits", func(t *testing.T) {
		// Test position size limit
		token := "SOL"
		size := config.MaxPositionSize + 1
		price := 100.0

		err := rm.CanOpenPosition(ctx, token, size, price)
		assert.Error(t, err)

		// Test daily trade limit
		size = 10.0
		for i := 0; i < config.MaxDailyTrades+1; i++ {
			position, err := rm.OpenPosition(token, size, price, "long")
			if i < config.MaxDailyTrades {
				assert.NoError(t, err)
				assert.NotNil(t, position)
				_, err = rm.ClosePosition(token, price)
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		}
	})

	t.Run("Daily Statistics", func(t *testing.T) {
		token := "SOL"
		today := time.Now()

		// Execute some trades
		trades := []struct {
			size      float64
			entry     float64
			exit      float64
			side      string
			expectPnL float64
		}{
			{10.0, 100.0, 105.0, "long", 50.0},
			{10.0, 105.0, 103.0, "short", 20.0},
			{10.0, 103.0, 101.0, "long", -20.0},
		}

		for _, trade := range trades {
			position, err := rm.OpenPosition(token, trade.size, trade.entry, trade.side)
			assert.NoError(t, err)
			assert.NotNil(t, position)

			pnl, err := rm.ClosePosition(token, trade.exit)
			assert.NoError(t, err)
			assert.InDelta(t, trade.expectPnL, pnl, 0.01)
		}

		// Check daily statistics
		stats, err := rm.GetDailyStats(today)
		assert.NoError(t, err)
		assert.Equal(t, 3, stats.TotalTrades)
		assert.Equal(t, 2, stats.WinningTrades)
		assert.Equal(t, 1, stats.LosingTrades)
		assert.Greater(t, stats.RealizedPnL, 0.0)
	})

	t.Run("Portfolio Management", func(t *testing.T) {
		token := "SOL"
		initialBalance := rm.portfolioSize

		// Execute a series of trades to test portfolio management
		trades := []struct {
			size  float64
			entry float64
			exit  float64
			side  string
		}{
			{10.0, 100.0, 105.0, "long"},  // +50
			{10.0, 105.0, 102.0, "long"},  // -30
			{10.0, 102.0, 106.0, "long"},  // +40
		}

		var totalPnL float64
		for _, trade := range trades {
			position, err := rm.OpenPosition(token, trade.size, trade.entry, trade.side)
			assert.NoError(t, err)
			assert.NotNil(t, position)

			pnl, err := rm.ClosePosition(token, trade.exit)
			assert.NoError(t, err)
			totalPnL += pnl
		}

		// Verify portfolio size updated correctly
		expectedBalance := initialBalance + totalPnL
		assert.InDelta(t, expectedBalance, rm.portfolioSize, 0.01)
	})

	t.Run("Edge Cases", func(t *testing.T) {
		token := "SOL"

		// Test closing non-existent position
		_, err := rm.ClosePosition(token, 100.0)
		assert.Error(t, err)

		// Test updating non-existent position
		_, err = rm.UpdatePosition(token, 100.0)
		assert.Error(t, err)

		// Test invalid position size
		err = rm.CanOpenPosition(ctx, token, -1.0, 100.0)
		assert.Error(t, err)

		// Test zero price
		err = rm.CanOpenPosition(ctx, token, 10.0, 0.0)
		assert.Error(t, err)
	})
}) 