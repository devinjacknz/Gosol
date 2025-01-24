package analysis

import (
	"math"
	"testing"
	"solmeme-trader/models"
)

func TestCalculateMACD(t *testing.T) {
	tests := []struct {
		name     string
		prices   []float64
		wantMACD models.MACD
		wantValid bool
	}{
		{
			name: "valid data",
			prices: []float64{45.0, 46.2, 47.5, 48.1, 47.8, 49.2, 50.5, 51.3, 50.9, 52.0,
				53.2, 52.8, 54.1, 55.0, 54.5, 56.2, 55.8, 57.1, 58.0, 59.2,
				60.5, 61.3, 62.0, 63.2, 64.1, 63.8, 65.0, 66.2, 67.5, 68.0,
				68.5, 69.1, 70.0, 71.2, 72.5, 73.0, 74.2, 75.5, 76.0, 77.2},
			wantValid: true,
			wantMACD: models.MACD{
				MACDLine:   5.86098207,
				SignalLine: 5.67830368,
				Histogram:  0.18267839,
			},
		},
		{
			name: "insufficient data",
			prices: []float64{45.0, 46.2, 47.5},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateMACD(tt.prices)
			if tt.wantValid {
				if math.IsNaN(got.MACDLine) || math.IsNaN(got.SignalLine) {
					t.Errorf("CalculateMACD() returned invalid values for valid input")
				}
				// Verify precision
				if got.MACDLine != math.Round(got.MACDLine*1e8)/1e8 {
					t.Errorf("MACDLine precision incorrect, got: %v", got.MACDLine)
				}
				if got.Histogram != math.Round((got.MACDLine - got.SignalLine)*1e8)/1e8 {
					t.Errorf("Histogram calculation precision error")
				}
				// 调试输出实际计算结果
				t.Logf("MACDLine: %.8f, SignalLine: %.8f, Histogram: %.8f", 
					got.MACDLine, 
					got.SignalLine,
					got.Histogram)
				
				// 验证实际计算结果
				if math.Abs(got.MACDLine - tt.wantMACD.MACDLine) > 1e-8 {
					t.Errorf("MACDLine mismatch\nWant: %.8f\nGot : %.8f", 
						tt.wantMACD.MACDLine, got.MACDLine)
				}
				if math.Abs(got.SignalLine - tt.wantMACD.SignalLine) > 1e-8 {
					t.Errorf("SignalLine mismatch\nWant: %.8f\nGot : %.8f", 
						tt.wantMACD.SignalLine, got.SignalLine)
				}
			} else {
				if got.MACDLine != 0 || got.SignalLine != 0 {
					t.Errorf("CalculateMACD() should return zero values for invalid input")
				}
			}
		})
	}
}

// 保留原有的Bollinger Bands和RSI测试用例不变
func TestCalculateBollingerBands(t *testing.T) {
	prices := []float64{
		45.0, 46.2, 47.5, 48.1, 47.8, 49.2, 50.5, 51.3, 50.9, 52.0,
		53.2, 52.8, 54.1, 55.0, 54.5, 56.2, 55.8, 57.1, 58.0, 59.2,
	}

	t.Run("standard case", func(t *testing.T) {
		period := 20
		bb := CalculateBollingerBands(prices, period)
		
		if math.IsNaN(bb.Upper) || math.IsNaN(bb.Middle) || math.IsNaN(bb.Lower) {
			t.Fatalf("Bollinger Bands contains NaN values")
		}
		
		// Verify middle band is SMA
		sma := CalculateSMA(prices, period)
		if bb.Middle != sma {
			t.Errorf("Middle band mismatch. Expected SMA: %v, got: %v", sma, bb.Middle)
		}
		
		// Verify standard deviation calculation
		window := prices[len(prices)-period:]
		var sum float64
		for _, price := range window {
			sum += math.Pow(price - sma, 2)
		}
		expectedStdDev := math.Sqrt(sum / float64(period - 1))
		actualStdDev := (bb.Upper - bb.Middle) / 2
		if math.Abs(actualStdDev - expectedStdDev) > 1e-8 {
			t.Errorf("Standard deviation calculation error. Expected: %v, got: %v", 
				expectedStdDev, actualStdDev)
		}
	})

	t.Run("invalid period", func(t *testing.T) {
		bb := CalculateBollingerBands(prices, 1)
		if bb.Upper != 0 || bb.Middle != 0 || bb.Lower != 0 {
			t.Errorf("Expected zero values for invalid period")
		}
	})
}

func TestCalculateRSI(t *testing.T) {
	// Test data from typical RSI example
	prices := []float64{
		44.34, 44.09, 44.15, 43.61, 44.33, 44.83, 45.10, 45.42, 45.84,
		46.08, 45.89, 46.03, 45.61, 46.28, 46.28, 46.00, 46.03, 46.41,
		46.22, 45.64, 46.21, 46.25, 45.71, 46.45, 45.78, 45.35, 44.03,
		44.18, 44.22, 44.57, 43.42, 42.66, 43.13,
	}

	t.Run("14-period RSI", func(t *testing.T) {
		rsi := CalculateRSI(prices, 14)
		expected := 70.53 // Known value for this dataset
		if math.Abs(rsi - expected) > 0.1 {
			t.Errorf("RSI calculation error. Expected: %.2f, got: %.2f", expected, rsi)
		}
	})

	t.Run("clamp values", func(t *testing.T) {
		allGains := make([]float64, 30)
		for i := range allGains {
			allGains[i] = 100.0 + float64(i)
		}
		rsi := CalculateRSI(allGains, 14)
		if rsi != 100.0 {
			t.Errorf("RSI should clamp to 100.0 for all gains, got: %v", rsi)
		}

		allLosses := make([]float64, 30)
		for i := range allLosses {
			allLosses[i] = 100.0 - float64(i)
		}
		rsi = CalculateRSI(allLosses, 14)
		if rsi != 0.0 {
			t.Errorf("RSI should clamp to 0.0 for all losses, got: %v", rsi)
		}
	})
}
