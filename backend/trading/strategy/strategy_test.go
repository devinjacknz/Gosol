package strategy

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"solmeme-trader/trading/analysis"
	"solmeme-trader/trading/risk"
)

// MockMarketAnalyzer mocks the market analyzer
type MockMarketAnalyzer struct {
	mock.Mock
}

func (m *MockMarketAnalyzer) AnalyzeTrend(timeframe string) (string, float64) {
	args := m.Called(timeframe)
	return args.String(0), args.Get(1).(float64)
}

func (m *MockMarketAnalyzer) AnalyzeVolatility(timeframe string) float64 {
	args := m.Called(timeframe)
	return args.Get(0).(float64)
}

func (m *MockMarketAnalyzer) GenerateReport(ctx context.Context) (*analysis.Report, error) {
	args := m.Called(ctx)
	return args.Get(0).(*analysis.Report), args.Error(1)
}

// MockRiskManager mocks the risk manager
type MockRiskManager struct {
	mock.Mock
}

func (m *MockRiskManager) CanOpenPosition(ctx context.Context, token string, size float64, price float64) error {
	args := m.Called(ctx, token, size, price)
	return args.Error(0)
}

func (m *MockRiskManager) OpenPosition(token string, size float64, price float64, side string) (*risk.Position, error) {
	args := m.Called(token, size, price, side)
	return args.Get(0).(*risk.Position), args.Error(1)
}

func (m *MockRiskManager) UpdatePosition(token string, currentPrice float64) (bool, error) {
	args := m.Called(token, currentPrice)
	return args.Bool(0), args.Error(1)
}

func TestNewStrategy(t *testing.T) {
	mockAnalyzer := new(MockMarketAnalyzer)
	mockRiskManager := new(MockRiskManager)

	strategy := NewStrategy(mockAnalyzer, mockRiskManager)
	assert.NotNil(t, strategy)
	assert.Equal(t, mockAnalyzer, strategy.analyzer)
	assert.Equal(t, mockRiskManager, strategy.riskManager)
}

func TestEvaluatePosition(t *testing.T) {
	ctx := context.Background()
	mockAnalyzer := new(MockMarketAnalyzer)
	mockRiskManager := new(MockRiskManager)
	strategy := NewStrategy(mockAnalyzer, mockRiskManager)

	testCases := []struct {
		name           string
		token          string
		currentPrice   float64
		trend          string
		trendStrength  float64
		volatility     float64
		expectDecision string
	}{
		{
			name:           "Strong bullish trend",
			token:          "SOL",
			currentPrice:   100.0,
			trend:          "bullish",
			trendStrength:  0.8,
			volatility:     0.2,
			expectDecision: "buy",
		},
		{
			name:           "Strong bearish trend",
			token:          "SOL",
			currentPrice:   100.0,
			trend:          "bearish",
			trendStrength:  0.7,
			volatility:     0.2,
			expectDecision: "sell",
		},
		{
			name:           "Neutral trend",
			token:          "SOL",
			currentPrice:   100.0,
			trend:          "neutral",
			trendStrength:  0.3,
			volatility:     0.2,
			expectDecision: "hold",
		},
		{
			name:           "High volatility",
			token:          "SOL",
			currentPrice:   100.0,
			trend:          "bullish",
			trendStrength:  0.6,
			volatility:     0.8,
			expectDecision: "hold",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockAnalyzer.On("AnalyzeTrend", "1d").Return(tc.trend, tc.trendStrength)
			mockAnalyzer.On("AnalyzeVolatility", "1d").Return(tc.volatility)

			decision, err := strategy.EvaluatePosition(ctx, tc.token, tc.currentPrice)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectDecision, decision)

			mockAnalyzer.AssertExpectations(t)
		})
	}
}

func TestExecuteTrade(t *testing.T) {
	ctx := context.Background()
	mockAnalyzer := new(MockMarketAnalyzer)
	mockRiskManager := new(MockRiskManager)
	strategy := NewStrategy(mockAnalyzer, mockRiskManager)

	testCases := []struct {
		name          string
		token         string
		decision      string
		currentPrice  float64
		size          float64
		expectSuccess bool
		expectError   bool
	}{
		{
			name:          "Successful buy",
			token:         "SOL",
			decision:      "buy",
			currentPrice:  100.0,
			size:          1.0,
			expectSuccess: true,
			expectError:   false,
		},
		{
			name:          "Successful sell",
			token:         "SOL",
			decision:      "sell",
			currentPrice:  100.0,
			size:          1.0,
			expectSuccess: true,
			expectError:   false,
		},
		{
			name:          "Hold - no trade",
			token:         "SOL",
			decision:      "hold",
			currentPrice:  100.0,
			size:          1.0,
			expectSuccess: false,
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.decision != "hold" {
				mockRiskManager.On("CanOpenPosition", ctx, tc.token, tc.size, tc.currentPrice).Return(nil)
				mockPosition := &risk.Position{
					TokenAddress: tc.token,
					EntryPrice:   tc.currentPrice,
					Size:         tc.size,
					Side:         tc.decision,
				}
				mockRiskManager.On("OpenPosition", tc.token, tc.size, tc.currentPrice, tc.decision).Return(mockPosition, nil)
			}

			success, err := strategy.ExecuteTrade(ctx, tc.token, tc.decision, tc.currentPrice, tc.size)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectSuccess, success)

			mockRiskManager.AssertExpectations(t)
		})
	}
}

func TestUpdatePositions(t *testing.T) {
	ctx := context.Background()
	mockAnalyzer := new(MockMarketAnalyzer)
	mockRiskManager := new(MockRiskManager)
	strategy := NewStrategy(mockAnalyzer, mockRiskManager)

	testCases := []struct {
		name         string
		token        string
		currentPrice float64
		shouldClose  bool
		expectError  bool
	}{
		{
			name:         "Keep position open",
			token:        "SOL",
			currentPrice: 100.0,
			shouldClose:  false,
			expectError:  false,
		},
		{
			name:         "Close position",
			token:        "SOL",
			currentPrice: 95.0,
			shouldClose:  true,
			expectError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRiskManager.On("UpdatePosition", tc.token, tc.currentPrice).Return(tc.shouldClose, nil)

			closed, err := strategy.UpdatePosition(ctx, tc.token, tc.currentPrice)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.shouldClose, closed)

			mockRiskManager.AssertExpectations(t)
		})
	}
}

func TestGenerateAnalysisReport(t *testing.T) {
	ctx := context.Background()
	mockAnalyzer := new(MockMarketAnalyzer)
	mockRiskManager := new(MockRiskManager)
	strategy := NewStrategy(mockAnalyzer, mockRiskManager)

	expectedReport := &analysis.Report{
		Trend:         "bullish",
		TrendStrength: 0.8,
		Volatility:    0.2,
		Support:       95.0,
		Resistance:    105.0,
		Timestamp:     time.Now(),
	}

	mockAnalyzer.On("GenerateReport", ctx).Return(expectedReport, nil)

	report, err := strategy.GenerateAnalysisReport(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedReport, report)

	mockAnalyzer.AssertExpectations(t)
}
