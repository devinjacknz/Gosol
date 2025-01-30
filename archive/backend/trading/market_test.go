package trading

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDataProvider struct {
	mock.Mock
}

func (m *MockDataProvider) GetHistoricalData(ctx context.Context, tokenAddress string, period string) ([]PricePoint, error) {
	args := m.Called(ctx, tokenAddress, period)
	return args.Get(0).([]PricePoint), args.Error(1)
}

func TestMarketTrendAnalysis(t *testing.T) {
	mockData := new(MockDataProvider)
	analyzer := NewMarketAnalyzer(mockData)
	ctx := context.Background()

	tests := []struct {
		name          string
		pricePoints   []PricePoint
		expectedTrend string
		expectedConf  float64
	}{
		{
			name: "Strong Uptrend",
			pricePoints: []PricePoint{
				{Price: 100, Volume: 1000, Timestamp: time.Now().Add(-4 * time.Hour)},
				{Price: 105, Volume: 1200, Timestamp: time.Now().Add(-3 * time.Hour)},
				{Price: 110, Volume: 1500, Timestamp: time.Now().Add(-2 * time.Hour)},
				{Price: 118, Volume: 2000, Timestamp: time.Now().Add(-1 * time.Hour)},
				{Price: 125, Volume: 2500, Timestamp: time.Now()},
			},
			expectedTrend: "bullish",
			expectedConf:  0.85,
		},
		{
			name: "Strong Downtrend",
			pricePoints: []PricePoint{
				{Price: 100, Volume: 1000, Timestamp: time.Now().Add(-4 * time.Hour)},
				{Price: 95, Volume: 1200, Timestamp: time.Now().Add(-3 * time.Hour)},
				{Price: 88, Volume: 1500, Timestamp: time.Now().Add(-2 * time.Hour)},
				{Price: 82, Volume: 2000, Timestamp: time.Now().Add(-1 * time.Hour)},
				{Price: 75, Volume: 2500, Timestamp: time.Now()},
			},
			expectedTrend: "bearish",
			expectedConf:  0.80,
		},
		{
			name: "Sideways Market",
			pricePoints: []PricePoint{
				{Price: 100, Volume: 1000, Timestamp: time.Now().Add(-4 * time.Hour)},
				{Price: 102, Volume: 1100, Timestamp: time.Now().Add(-3 * time.Hour)},
				{Price: 99, Volume: 1050, Timestamp: time.Now().Add(-2 * time.Hour)},
				{Price: 101, Volume: 1200, Timestamp: time.Now().Add(-1 * time.Hour)},
				{Price: 100, Volume: 1100, Timestamp: time.Now()},
			},
			expectedTrend: "neutral",
			expectedConf:  0.60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockData.On("GetHistoricalData", ctx, "TestToken", "4h").Return(tt.pricePoints, nil)

			trend, conf, err := analyzer.AnalyzeTrend(ctx, "TestToken", "4h")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTrend, trend)
			assert.InDelta(t, tt.expectedConf, conf, 0.1)
		})
	}
}

func TestVolatilityAnalysis(t *testing.T) {
	mockData := new(MockDataProvider)
	analyzer := NewMarketAnalyzer(mockData)
	ctx := context.Background()

	tests := []struct {
		name               string
		pricePoints        []PricePoint
		expectedVolatility float64
		expectedRisk       string
	}{
		{
			name: "High Volatility",
			pricePoints: []PricePoint{
				{Price: 100, Volume: 1000, Timestamp: time.Now().Add(-4 * time.Hour)},
				{Price: 120, Volume: 1500, Timestamp: time.Now().Add(-3 * time.Hour)},
				{Price: 90, Volume: 2000, Timestamp: time.Now().Add(-2 * time.Hour)},
				{Price: 110, Volume: 1800, Timestamp: time.Now().Add(-1 * time.Hour)},
				{Price: 95, Volume: 2200, Timestamp: time.Now()},
			},
			expectedVolatility: 0.25,
			expectedRisk:       "high",
		},
		{
			name: "Low Volatility",
			pricePoints: []PricePoint{
				{Price: 100, Volume: 1000, Timestamp: time.Now().Add(-4 * time.Hour)},
				{Price: 101, Volume: 1100, Timestamp: time.Now().Add(-3 * time.Hour)},
				{Price: 99, Volume: 1050, Timestamp: time.Now().Add(-2 * time.Hour)},
				{Price: 100, Volume: 1200, Timestamp: time.Now().Add(-1 * time.Hour)},
				{Price: 101, Volume: 1150, Timestamp: time.Now()},
			},
			expectedVolatility: 0.05,
			expectedRisk:       "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockData.On("GetHistoricalData", ctx, "TestToken", "4h").Return(tt.pricePoints, nil)

			volatility, risk, err := analyzer.AnalyzeVolatility(ctx, "TestToken", "4h")
			assert.NoError(t, err)
			assert.InDelta(t, tt.expectedVolatility, volatility, 0.05)
			assert.Equal(t, tt.expectedRisk, risk)
		})
	}
}

func TestSupportResistanceAnalysis(t *testing.T) {
	mockData := new(MockDataProvider)
	analyzer := NewMarketAnalyzer(mockData)
	ctx := context.Background()

	tests := []struct {
		name               string
		pricePoints        []PricePoint
		expectedSupport    float64
		expectedResistance float64
	}{
		{
			name: "Clear Support and Resistance",
			pricePoints: []PricePoint{
				{Price: 100, Volume: 1000},
				{Price: 105, Volume: 1200},
				{Price: 95, Volume: 1500},
				{Price: 110, Volume: 1800},
				{Price: 98, Volume: 2000},
				{Price: 108, Volume: 1900},
				{Price: 96, Volume: 2200},
			},
			expectedSupport:    95.0,
			expectedResistance: 110.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockData.On("GetHistoricalData", ctx, "TestToken", "1d").Return(tt.pricePoints, nil)

			support, resistance, err := analyzer.FindSupportResistance(ctx, "TestToken", "1d")
			assert.NoError(t, err)
			assert.InDelta(t, tt.expectedSupport, support, 1.0)
			assert.InDelta(t, tt.expectedResistance, resistance, 1.0)
		})
	}
}

func TestVolumeAnalysis(t *testing.T) {
	mockData := new(MockDataProvider)
	analyzer := NewMarketAnalyzer(mockData)
	ctx := context.Background()

	tests := []struct {
		name             string
		pricePoints      []PricePoint
		expectedSignal   string
		expectedStrength float64
	}{
		{
			name: "Rising Volume with Price",
			pricePoints: []PricePoint{
				{Price: 100, Volume: 1000},
				{Price: 105, Volume: 1500},
				{Price: 110, Volume: 2000},
				{Price: 115, Volume: 2500},
				{Price: 120, Volume: 3000},
			},
			expectedSignal:   "strong_buy",
			expectedStrength: 0.9,
		},
		{
			name: "Falling Volume with Price",
			pricePoints: []PricePoint{
				{Price: 100, Volume: 3000},
				{Price: 95, Volume: 2500},
				{Price: 90, Volume: 2000},
				{Price: 85, Volume: 1500},
				{Price: 80, Volume: 1000},
			},
			expectedSignal:   "strong_sell",
			expectedStrength: 0.85,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockData.On("GetHistoricalData", ctx, "TestToken", "1d").Return(tt.pricePoints, nil)

			signal, strength, err := analyzer.AnalyzeVolume(ctx, "TestToken", "1d")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedSignal, signal)
			assert.InDelta(t, tt.expectedStrength, strength, 0.1)
		})
	}
}

func TestMarketSentimentAnalysis(t *testing.T) {
	mockData := new(MockDataProvider)
	analyzer := NewMarketAnalyzer(mockData)
	ctx := context.Background()

	tests := []struct {
		name               string
		pricePoints        []PricePoint
		expectedSentiment  string
		expectedConfidence float64
	}{
		{
			name: "Bullish Sentiment",
			pricePoints: []PricePoint{
				{Price: 100, Volume: 1000, BuyVolume: 800},
				{Price: 105, Volume: 1500, BuyVolume: 1200},
				{Price: 110, Volume: 2000, BuyVolume: 1700},
			},
			expectedSentiment:  "bullish",
			expectedConfidence: 0.85,
		},
		{
			name: "Bearish Sentiment",
			pricePoints: []PricePoint{
				{Price: 100, Volume: 1000, BuyVolume: 300},
				{Price: 95, Volume: 1500, BuyVolume: 400},
				{Price: 90, Volume: 2000, BuyVolume: 500},
			},
			expectedSentiment:  "bearish",
			expectedConfidence: 0.80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockData.On("GetHistoricalData", ctx, "TestToken", "1d").Return(tt.pricePoints, nil)

			sentiment, confidence, err := analyzer.AnalyzeSentiment(ctx, "TestToken", "1d")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedSentiment, sentiment)
			assert.InDelta(t, tt.expectedConfidence, confidence, 0.1)
		})
	}
}

func TestPriceTargetPrediction(t *testing.T) {
	mockData := new(MockDataProvider)
	analyzer := NewMarketAnalyzer(mockData)
	ctx := context.Background()

	tests := []struct {
		name           string
		pricePoints    []PricePoint
		expectedTarget float64
		expectedConf   float64
	}{
		{
			name: "Upward Target",
			pricePoints: []PricePoint{
				{Price: 100, Volume: 1000},
				{Price: 105, Volume: 1200},
				{Price: 110, Volume: 1500},
				{Price: 115, Volume: 1800},
				{Price: 120, Volume: 2000},
			},
			expectedTarget: 125.0,
			expectedConf:   0.75,
		},
		{
			name: "Downward Target",
			pricePoints: []PricePoint{
				{Price: 100, Volume: 1000},
				{Price: 95, Volume: 1200},
				{Price: 90, Volume: 1500},
				{Price: 85, Volume: 1800},
				{Price: 80, Volume: 2000},
			},
			expectedTarget: 75.0,
			expectedConf:   0.70,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockData.On("GetHistoricalData", ctx, "TestToken", "1d").Return(tt.pricePoints, nil)

			target, conf, err := analyzer.PredictPriceTarget(ctx, "TestToken", "1d")
			assert.NoError(t, err)
			assert.InDelta(t, tt.expectedTarget, target, 5.0)
			assert.InDelta(t, tt.expectedConf, conf, 0.1)
		})
	}
}
