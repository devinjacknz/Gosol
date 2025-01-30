package market

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockAnalyzer is a mock implementation of the Analyzer interface
type MockAnalyzer struct {
	mock.Mock
}

// Analyze mocks the Analyze method
func (m *MockAnalyzer) Analyze(ctx context.Context, data []*PricePoint) (*MarketSignal, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*MarketSignal), args.Error(1)
}

// CalculateVolatility mocks the CalculateVolatility method
func (m *MockAnalyzer) CalculateVolatility(data []*PricePoint) float64 {
	args := m.Called(data)
	return args.Get(0).(float64)
}

// CalculateTrend mocks the CalculateTrend method
func (m *MockAnalyzer) CalculateTrend(data []*PricePoint) float64 {
	args := m.Called(data)
	return args.Get(0).(float64)
}

// NewMockAnalyzer creates a new mock analyzer
func NewMockAnalyzer() *MockAnalyzer {
	return &MockAnalyzer{}
}

// ExpectAnalyze sets up expectations for the Analyze method
func (m *MockAnalyzer) ExpectAnalyze(ctx context.Context, data []*PricePoint, signal *MarketSignal, err error) *mock.Call {
	return m.On("Analyze", ctx, data).Return(signal, err)
}

// ExpectVolatility sets up expectations for the CalculateVolatility method
func (m *MockAnalyzer) ExpectVolatility(data []*PricePoint, volatility float64) *mock.Call {
	return m.On("CalculateVolatility", data).Return(volatility)
}

// ExpectTrend sets up expectations for the CalculateTrend method
func (m *MockAnalyzer) ExpectTrend(data []*PricePoint, trend float64) *mock.Call {
	return m.On("CalculateTrend", data).Return(trend)
}
