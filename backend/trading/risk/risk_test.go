package risk

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRiskManager(t *testing.T) {
	manager := NewRiskManager()
	ctx := context.Background()

	t.Run("Position limit checks", func(t *testing.T) {
		// Set position limit
		err := manager.UpdatePositionLimit(ctx, "BTC/USD", 1000000.0)
		assert.NoError(t, err)

		// Test within limit
		params := PositionLimitParams{
			Symbol:        "BTC/USD",
			Size:          1.0,
			CurrentPrice:  50000.0,
			TotalPosition: 15.0,
		}

		check, err := manager.CheckPositionLimit(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, Pass, check.Status)
		assert.Equal(t, Low, check.Level)

		// Test near limit (warning)
		params.TotalPosition = 18.0
		check, err = manager.CheckPositionLimit(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, Warning, check.Status)
		assert.Equal(t, High, check.Level)

		// Test exceeding limit
		params.TotalPosition = 25.0
		check, err = manager.CheckPositionLimit(ctx, params)
		assert.Error(t, err)
		assert.Equal(t, ErrPositionLimitExceeded, err)
		assert.Equal(t, Violation, check.Status)
		assert.Equal(t, Critical, check.Level)

		// Test invalid limit
		err = manager.UpdatePositionLimit(ctx, "ETH/USD", -1.0)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidLimit, err)
	})

	t.Run("Exposure limit checks", func(t *testing.T) {
		// Set exposure limit
		err := manager.UpdateExposureLimit(ctx, 2.0) // 2x leverage
		assert.NoError(t, err)

		// Test within limit
		params := ExposureLimitParams{
			TotalExposure:     500000.0,
			AdditionalAmount:  50000.0,
			CollateralBalance: 300000.0,
		}

		check, err := manager.CheckExposureLimit(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, Pass, check.Status)
		assert.Equal(t, Low, check.Level)

		// Test near limit (warning)
		params.TotalExposure = 500000.0
		check, err = manager.CheckExposureLimit(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, Warning, check.Status)
		assert.Equal(t, High, check.Level)

		// Test exceeding limit
		params.TotalExposure = 700000.0
		check, err = manager.CheckExposureLimit(ctx, params)
		assert.Error(t, err)
		assert.Equal(t, ErrExposureLimitExceeded, err)
		assert.Equal(t, Violation, check.Status)
		assert.Equal(t, Critical, check.Level)

		// Test invalid limit
		err = manager.UpdateExposureLimit(ctx, 0.0)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidLimit, err)
	})

	t.Run("Drawdown checks", func(t *testing.T) {
		// Set drawdown limit
		err := manager.UpdateDrawdownLimit(ctx, 0.20) // 20% max drawdown
		assert.NoError(t, err)

		// Test within limit
		params := DrawdownParams{
			CurrentEquity: 90000.0,
			PeakEquity:    100000.0,
			TimeWindow:    24 * time.Hour,
		}

		check, err := manager.CheckDrawdown(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, Pass, check.Status)
		assert.Equal(t, Low, check.Level)

		// Test near limit (warning)
		params.CurrentEquity = 82000.0
		check, err = manager.CheckDrawdown(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, Warning, check.Status)
		assert.Equal(t, High, check.Level)

		// Test exceeding limit
		params.CurrentEquity = 75000.0
		check, err = manager.CheckDrawdown(ctx, params)
		assert.Error(t, err)
		assert.Equal(t, ErrDrawdownLimitExceeded, err)
		assert.Equal(t, Violation, check.Status)
		assert.Equal(t, Critical, check.Level)

		// Test invalid limit
		err = manager.UpdateDrawdownLimit(ctx, 1.5)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidLimit, err)
	})

	t.Run("Volatility checks", func(t *testing.T) {
		// Set volatility thresholds
		thresholds := VolatilityThresholds{
			LowThreshold:      0.15,
			MediumThreshold:   0.30,
			HighThreshold:     0.50,
			CriticalThreshold: 0.75,
		}
		err := manager.UpdateVolatilityThresholds(ctx, thresholds)
		assert.NoError(t, err)

		// Test low volatility
		params := VolatilityParams{
			Symbol:            "BTC/USD",
			CurrentVolatility: 0.10,
			TimeWindow:        24 * time.Hour,
		}

		check, err := manager.CheckVolatility(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, Pass, check.Status)
		assert.Equal(t, Low, check.Level)

		// Test medium volatility
		params.CurrentVolatility = 0.35
		check, err = manager.CheckVolatility(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, Warning, check.Status)
		assert.Equal(t, Medium, check.Level)

		// Test high volatility
		params.CurrentVolatility = 0.60
		check, err = manager.CheckVolatility(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, Warning, check.Status)
		assert.Equal(t, High, check.Level)

		// Test critical volatility
		params.CurrentVolatility = 0.80
		check, err = manager.CheckVolatility(ctx, params)
		assert.Error(t, err)
		assert.Equal(t, ErrVolatilityTooHigh, err)
		assert.Equal(t, Violation, check.Status)
		assert.Equal(t, Critical, check.Level)

		// Test invalid thresholds
		invalidThresholds := VolatilityThresholds{
			LowThreshold:      0.30,
			MediumThreshold:   0.20,
			HighThreshold:     0.10,
			CriticalThreshold: 0.05,
		}
		err = manager.UpdateVolatilityThresholds(ctx, invalidThresholds)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidThresholds, err)
	})

	t.Run("Risk history", func(t *testing.T) {
		// Create some risk checks
		symbols := []string{"BTC/USD", "ETH/USD", "SOL/USD"}
		types := []RiskType{PositionRisk, ExposureRisk, DrawdownRisk}
		levels := []RiskLevel{Low, Medium, High}

		for i := range symbols {
			params := PositionLimitParams{
				Symbol:        symbols[i],
				Size:          1.0,
				CurrentPrice:  50000.0,
				TotalPosition: 10.0,
			}

			err := manager.UpdatePositionLimit(ctx, symbols[i], 1000000.0)
			assert.NoError(t, err)

			_, err = manager.CheckPositionLimit(ctx, params)
			assert.NoError(t, err)
		}

		// Test different filters
		tests := []struct {
			name     string
			filter   RiskHistoryFilter
			expected int
		}{
			{
				name:     "No filter",
				filter:   RiskHistoryFilter{},
				expected: len(symbols),
			},
			{
				name: "Filter by symbol",
				filter: RiskHistoryFilter{
					Symbol: "BTC/USD",
				},
				expected: 1,
			},
			{
				name: "Filter by type",
				filter: RiskHistoryFilter{
					Type: &types[0],
				},
				expected: len(symbols),
			},
			{
				name: "Filter by level",
				filter: RiskHistoryFilter{
					Level: &levels[0],
				},
				expected: len(symbols),
			},
			{
				name: "Filter by time range",
				filter: RiskHistoryFilter{
					StartTime: &time.Time{},
					EndTime:   &time.Time{},
				},
				expected: len(symbols),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				checks, err := manager.GetRiskHistory(ctx, tt.filter)
				assert.NoError(t, err)
				assert.Len(t, checks, tt.expected)
			})
		}
	})
}

func TestConcurrency(t *testing.T) {
	manager := NewRiskManager()
	ctx := context.Background()
	numOperations := 100

	t.Run("Concurrent position limit checks", func(t *testing.T) {
		err := manager.UpdatePositionLimit(ctx, "BTC/USD", 1000000.0)
		assert.NoError(t, err)

		done := make(chan bool)
		for i := 0; i < numOperations; i++ {
			go func() {
				params := PositionLimitParams{
					Symbol:        "BTC/USD",
					Size:          1.0,
					CurrentPrice:  50000.0,
					TotalPosition: 10.0,
				}

				_, err := manager.CheckPositionLimit(ctx, params)
				assert.NoError(t, err)
				done <- true
			}()
		}

		for i := 0; i < numOperations; i++ {
			<-done
		}
	})

	t.Run("Concurrent limit updates", func(t *testing.T) {
		done := make(chan bool)
		for i := 0; i < numOperations; i++ {
			go func(idx int) {
				symbol := "TEST/USD"
				limit := float64(1000000 + idx)
				err := manager.UpdatePositionLimit(ctx, symbol, limit)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		for i := 0; i < numOperations; i++ {
			<-done
		}
	})
}
