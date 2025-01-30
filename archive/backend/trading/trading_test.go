package trading

import (
	"context"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/leonzhao/trading-system/backend/models"
	"github.com/leonzhao/trading-system/backend/trading/types"
)

type MockExecutor struct {
	mock.Mock
}

func (m *MockExecutor) ExecuteOrder(order interface{}) error {
	args := m.Called(order)
	return args.Error(0)
}

func TestRiskManager(t *testing.T) {
	tests := []struct {
		name        string
		config      *types.RiskConfig
		trade       *models.Trade
		marketData  *models.MarketData
		expectError bool
	}{
		{
			name: "valid market buy",
			config: &types.RiskConfig{
				MaxPositionSize:  1000.0,
				LeverageLimit:    10,
				DailyLossLimit:   0.05,
				VolatilityWindow: 30,
			},
			trade: &models.Trade{
				TokenAddress: "SOL/USDC",
				Type:         models.TradeTypeMarket,
				Side:         models.TradeSideBuy,
				Amount:       100.0,
				Price:        95.0,
			},
			marketData: &models.MarketData{
				ClosePrice: 95.0,
				Volume:     100000.0,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := new(MockExecutor)
			rm := NewRiskManager(tt.config, mockExec)
			
			// ValidateTrade should accept *models.Trade
			err := rm.ValidateTradeSignal(context.Background(), tt.trade, tt.marketData)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPositionSizing(t *testing.T) {
	rm := NewRiskManager(&types.RiskConfig{
		MaxPositionSize:  500.0,
		LeverageLimit:    5,
	}, new(MockExecutor))

	trade := &models.Trade{
		TokenAddress: "ETH/USDC",
		Type:         models.TradeTypeLimit,
		Side:         models.TradeSideBuy,
		Amount:       200.0,
		Price:        2000.0,
	}

	marketData := &models.MarketData{
		ClosePrice: 2000.0,
	}

	size, err := rm.calculatePositionSize(trade, marketData)
	assert.NoError(t, err)
	assert.InDelta(t, 0.1, size, 0.001)
}
