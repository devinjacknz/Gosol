package trading

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMarketService struct {
	mock.Mock
}

func (m *MockMarketService) GetPrice(ctx context.Context, tokenAddress string) (float64, error) {
	args := m.Called(ctx, tokenAddress)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockMarketService) GetVolume(ctx context.Context, tokenAddress string) (float64, error) {
	args := m.Called(ctx, tokenAddress)
	return args.Get(0).(float64), args.Error(1)
}

func TestTradeExecution(t *testing.T) {
	mockMarket := new(MockMarketService)
	ctx := context.Background()

	tests := []struct {
		name          string
		tokenAddress  string
		action        string
		amount        float64
		marketPrice   float64
		expectedError error
	}{
		{
			name:         "Successful Buy Trade",
			tokenAddress: "TokenXYZ",
			action:       "buy",
			amount:       1.0,
			marketPrice:  100.0,
		},
		{
			name:         "Successful Sell Trade",
			tokenAddress: "TokenXYZ",
			action:       "sell",
			amount:       0.5,
			marketPrice:  150.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMarket.On("GetPrice", ctx, tt.tokenAddress).Return(tt.marketPrice, nil)

			trade := &Trade{
				TokenAddress: tt.tokenAddress,
				Action:       tt.action,
				Amount:       tt.amount,
				Price:        tt.marketPrice,
				Timestamp:    time.Now(),
			}

			err := ExecuteTrade(ctx, trade, mockMarket)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRiskManagement(t *testing.T) {
	tests := []struct {
		name           string
		position       Position
		stopLoss       float64
		takeProfit     float64
		currentPrice   float64
		expectedAction string
	}{
		{
			name: "Stop Loss Triggered",
			position: Position{
				EntryPrice: 100.0,
				Size:       1.0,
				Side:       "long",
			},
			stopLoss:       0.05, // 5%
			currentPrice:   94.0,
			expectedAction: "sell",
		},
		{
			name: "Take Profit Triggered",
			position: Position{
				EntryPrice: 100.0,
				Size:       1.0,
				Side:       "long",
			},
			takeProfit:     0.10, // 10%
			currentPrice:   111.0,
			expectedAction: "sell",
		},
		{
			name: "Hold Position",
			position: Position{
				EntryPrice: 100.0,
				Size:       1.0,
				Side:       "long",
			},
			stopLoss:       0.05,
			takeProfit:     0.10,
			currentPrice:   102.0,
			expectedAction: "hold",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := NewRiskManager(tt.stopLoss, tt.takeProfit)
			action := rm.CheckPosition(tt.position, tt.currentPrice)
			assert.Equal(t, tt.expectedAction, action)
		})
	}
}

func TestProfitCalculation(t *testing.T) {
	tests := []struct {
		name          string
		trades        []Trade
		expectedPnL   float64
		expectedCount int
	}{
		{
			name: "Profitable Trades",
			trades: []Trade{
				{Action: "buy", Amount: 1.0, Price: 100.0},
				{Action: "sell", Amount: 1.0, Price: 110.0},
				{Action: "buy", Amount: 0.5, Price: 105.0},
				{Action: "sell", Amount: 0.5, Price: 115.0},
			},
			expectedPnL:   15.0,
			expectedCount: 2,
		},
		{
			name: "Loss Making Trades",
			trades: []Trade{
				{Action: "buy", Amount: 1.0, Price: 100.0},
				{Action: "sell", Amount: 1.0, Price: 90.0},
			},
			expectedPnL:   -10.0,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calculator := NewProfitCalculator()
			for _, trade := range tt.trades {
				calculator.AddTrade(trade)
			}

			pnl := calculator.GetTotalPnL()
			assert.InDelta(t, tt.expectedPnL, pnl, 0.001)
			assert.Equal(t, tt.expectedCount, calculator.GetTradeCount())
		})
	}
}

func TestMarketAnalysis(t *testing.T) {
	mockMarket := new(MockMarketService)
	ctx := context.Background()

	tests := []struct {
		name         string
		tokenAddress string
		priceData    []float64
		volumeData   []float64
		expected     MarketSignal
	}{
		{
			name:         "Bullish Signal",
			tokenAddress: "TokenXYZ",
			priceData:    []float64{100, 101, 103, 106, 110},
			volumeData:   []float64{1000, 1200, 1500, 1800, 2000},
			expected: MarketSignal{
				Trend:      "bullish",
				Strength:   0.8,
				Confidence: 0.75,
			},
		},
		{
			name:         "Bearish Signal",
			tokenAddress: "TokenXYZ",
			priceData:    []float64{100, 98, 95, 92, 88},
			volumeData:   []float64{1000, 1100, 1300, 1400, 1600},
			expected: MarketSignal{
				Trend:      "bearish",
				Strength:   0.7,
				Confidence: 0.65,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewMarketAnalyzer(mockMarket)
			signal, err := analyzer.AnalyzeMarket(ctx, tt.tokenAddress)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Trend, signal.Trend)
			assert.InDelta(t, tt.expected.Strength, signal.Strength, 0.1)
			assert.InDelta(t, tt.expected.Confidence, signal.Confidence, 0.1)
		})
	}
}
