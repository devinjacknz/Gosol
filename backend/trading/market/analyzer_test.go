package market

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMarketAnalyzer_Analyze(t *testing.T) {
	tests := []struct {
		name         string
		data         []*PricePoint
		minPoints    int
		expectType   string
		expectSignal bool
	}{
		{
			name: "insufficient data points",
			data: []*PricePoint{
				{Price: 100, Volume: 1000, Timestamp: time.Now()},
			},
			minPoints:    2,
			expectType:   "hold",
			expectSignal: true,
		},
		{
			name: "strong uptrend low volatility",
			data: []*PricePoint{
				{Price: 100, Volume: 1000, Timestamp: time.Now().Add(-3 * time.Hour)},
				{Price: 102, Volume: 1000, Timestamp: time.Now().Add(-2 * time.Hour)},
				{Price: 104, Volume: 1000, Timestamp: time.Now().Add(-1 * time.Hour)},
				{Price: 106, Volume: 1000, Timestamp: time.Now()},
			},
			minPoints:    2,
			expectType:   "buy",
			expectSignal: true,
		},
		{
			name: "strong downtrend low volatility",
			data: []*PricePoint{
				{Price: 106, Volume: 1000, Timestamp: time.Now().Add(-3 * time.Hour)},
				{Price: 104, Volume: 1000, Timestamp: time.Now().Add(-2 * time.Hour)},
				{Price: 102, Volume: 1000, Timestamp: time.Now().Add(-1 * time.Hour)},
				{Price: 100, Volume: 1000, Timestamp: time.Now()},
			},
			minPoints:    2,
			expectType:   "sell",
			expectSignal: true,
		},
		{
			name: "high volatility no trend",
			data: []*PricePoint{
				{Price: 100, Volume: 1000, Timestamp: time.Now().Add(-3 * time.Hour)},
				{Price: 110, Volume: 1000, Timestamp: time.Now().Add(-2 * time.Hour)},
				{Price: 90, Volume: 1000, Timestamp: time.Now().Add(-1 * time.Hour)},
				{Price: 105, Volume: 1000, Timestamp: time.Now()},
			},
			minPoints:    2,
			expectType:   "hold",
			expectSignal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewMarketAnalyzer(1*time.Hour, tt.minPoints)
			signal, err := analyzer.Analyze(context.Background(), tt.data)

			if tt.expectSignal {
				assert.NoError(t, err)
				assert.NotNil(t, signal)
				assert.Equal(t, tt.expectType, signal.Type)
			} else {
				assert.Error(t, err)
				assert.Nil(t, signal)
			}
		})
	}
}

func TestMarketAnalyzer_CalculateVolatility(t *testing.T) {
	tests := []struct {
		name     string
		data     []*PricePoint
		expected float64
	}{
		{
			name: "low volatility",
			data: []*PricePoint{
				{Price: 100, Volume: 1000, Timestamp: time.Now().Add(-2 * time.Hour)},
				{Price: 101, Volume: 1000, Timestamp: time.Now().Add(-1 * time.Hour)},
				{Price: 102, Volume: 1000, Timestamp: time.Now()},
			},
			expected: 0.0001, // Approximately
		},
		{
			name: "high volatility",
			data: []*PricePoint{
				{Price: 100, Volume: 1000, Timestamp: time.Now().Add(-2 * time.Hour)},
				{Price: 120, Volume: 1000, Timestamp: time.Now().Add(-1 * time.Hour)},
				{Price: 90, Volume: 1000, Timestamp: time.Now()},
			},
			expected: 0.05, // Approximately
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewMarketAnalyzer(1*time.Hour, 2)
			volatility := analyzer.CalculateVolatility(tt.data)
			assert.InDelta(t, tt.expected, volatility, 0.01)
		})
	}
}

func TestMarketAnalyzer_CalculateTrend(t *testing.T) {
	tests := []struct {
		name     string
		data     []*PricePoint
		expected float64
	}{
		{
			name: "strong uptrend",
			data: []*PricePoint{
				{Price: 100, Volume: 1000, Timestamp: time.Now().Add(-2 * time.Hour)},
				{Price: 110, Volume: 1000, Timestamp: time.Now().Add(-1 * time.Hour)},
				{Price: 120, Volume: 1000, Timestamp: time.Now()},
			},
			expected: 0.8,
		},
		{
			name: "strong downtrend",
			data: []*PricePoint{
				{Price: 120, Volume: 1000, Timestamp: time.Now().Add(-2 * time.Hour)},
				{Price: 110, Volume: 1000, Timestamp: time.Now().Add(-1 * time.Hour)},
				{Price: 100, Volume: 1000, Timestamp: time.Now()},
			},
			expected: -0.8,
		},
		{
			name: "no trend",
			data: []*PricePoint{
				{Price: 100, Volume: 1000, Timestamp: time.Now().Add(-2 * time.Hour)},
				{Price: 101, Volume: 1000, Timestamp: time.Now().Add(-1 * time.Hour)},
				{Price: 100, Volume: 1000, Timestamp: time.Now()},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewMarketAnalyzer(1*time.Hour, 2)
			trend := analyzer.CalculateTrend(tt.data)
			assert.InDelta(t, tt.expected, trend, 0.2)
		})
	}
}
