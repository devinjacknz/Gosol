package analysis

import (
	"math"
	"testing"
	"github.com/leonzhao/trading-system/backend/models"
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
		{
			name: "empty prices",
			prices: []float64{},
			wantValid: false,
		},
		{
			name: "single price",
			prices: []float64{45.0},
			wantValid: false,
		},
		{
			name: "NaN values",
			prices: []float64{45.0, 46.2, math.NaN(), 48.1},
			wantValid: false,
		},
		{
			name: "infinity values",
			prices: []float64{45.0, 46.2, math.Inf(1), 48.1},
			wantValid: false,
		},
		{
			name: "negative prices",
			prices: []float64{-45.0, -46.2, -47.5},
			wantValid: false,
		},
		{
			name: "zero prices",
			prices: []float64{0, 0, 0},
			wantValid: false,
		},
		{
			name: "large price range",
			prices: []float64{1e-10, 1e10, 1e20},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMACD, gotHistogram, err := MACD(tt.prices, 12, 26, 9)
			if tt.wantValid {
				if math.IsNaN(gotMACD) || math.IsNaN(gotHistogram) {
					t.Errorf("CalculateMACD() returned invalid values for valid input")
				}
				// Verify precision
				if gotMACD != math.Round(gotMACD*1e8)/1e8 {
					t.Errorf("MACD precision incorrect, got: %v", gotMACD)
				}
				if gotHistogram != math.Round(gotHistogram*1e8)/1e8 {
					t.Errorf("Histogram precision error")
				}
				// 调试输出实际计算结果
				t.Logf("MACD: %.8f, Histogram: %.8f", 
					gotMACD, 
					gotHistogram)
				
				// 验证实际计算结果
				if math.Abs(gotMACD - tt.wantMACD.MACDLine) > 1e-8 {
					t.Errorf("MACD mismatch\nWant: %.8f\nGot : %.8f", 
						tt.wantMACD.MACDLine, gotMACD)
				}
				if math.Abs(gotHistogram - tt.wantMACD.Histogram) > 1e-8 {
					t.Errorf("Histogram mismatch\nWant: %.8f\nGot : %.8f", 
						tt.wantMACD.Histogram, gotHistogram)
				}
			} else {
				if err == nil {
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
		bb, err := BollingerBands(prices, period, 2)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if math.IsNaN(bb.Upper) || math.IsNaN(bb.Middle) || math.IsNaN(bb.Lower) {
			t.Fatalf("Bollinger Bands contains NaN values")
		}
		
		// Verify middle band is SMA
		sma, _ := SMA(prices, period)
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
		bb, err := BollingerBands(prices, 1, 2)
		if err == nil {
			t.Errorf("Expected error for invalid period")
		}
		if bb.Upper != 0 || bb.Middle != 0 || bb.Lower != 0 {
			t.Errorf("Expected zero values for invalid period")
		}
	})

	t.Run("empty prices", func(t *testing.T) {
		bb, err := BollingerBands([]float64{}, 20, 2)
		if err == nil {
			t.Errorf("Expected error for empty prices")
		}
		if bb.Upper != 0 || bb.Middle != 0 || bb.Lower != 0 {
			t.Errorf("Expected zero values for empty prices")
		}
	})

	t.Run("NaN values", func(t *testing.T) {
		bb, err := BollingerBands([]float64{45.0, 46.2, math.NaN()}, 3, 2)
		if err == nil {
			t.Errorf("Expected error for NaN prices")
		}
		if bb.Upper != 0 || bb.Middle != 0 || bb.Lower != 0 {
			t.Errorf("Expected zero values for NaN prices")
		}
	})

	t.Run("infinity values", func(t *testing.T) {
		bb, err := BollingerBands([]float64{45.0, 46.2, math.Inf(1)}, 3, 2)
		if err == nil {
			t.Errorf("Expected error for infinity prices")
		}
		if bb.Upper != 0 || bb.Middle != 0 || bb.Lower != 0 {
			t.Errorf("Expected zero values for infinity prices")
		}
	})

	t.Run("negative prices", func(t *testing.T) {
		bb, err := BollingerBands([]float64{-45.0, -46.2, -47.5}, 3, 2)
		if err == nil {
			t.Errorf("Expected error for negative prices")
		}
		if bb.Upper != 0 || bb.Middle != 0 || bb.Lower != 0 {
			t.Errorf("Expected zero values for negative prices")
		}
	})

	t.Run("zero prices", func(t *testing.T) {
		bb, err := BollingerBands([]float64{0, 0, 0}, 3, 2)
		if err == nil {
			t.Errorf("Expected error for zero prices")
		}
		if bb.Upper != 0 || bb.Middle != 0 || bb.Lower != 0 {
			t.Errorf("Expected zero values for zero prices")
		}
	})

	t.Run("large price range", func(t *testing.T) {
		bb, _ := BollingerBands([]float64{1e-10, 1e10, 1e20}, 3, 2)
		if math.IsNaN(bb.Upper) || math.IsNaN(bb.Middle) || math.IsNaN(bb.Lower) {
			t.Fatalf("Bollinger Bands contains NaN values for large price range")
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
		rsi, _ := RSI(prices, 14)
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
		rsi, _ := RSI(allGains, 14)
		if rsi != 100.0 {
			t.Errorf("RSI should clamp to 100.0 for all gains, got: %v", rsi)
		}

		allLosses := make([]float64, 30)
		for i := range allLosses {
			allLosses[i] = 100.0 - float64(i)
		}
		rsi, _ = RSI(allLosses, 14)
		if rsi != 0.0 {
			t.Errorf("RSI should clamp to 0.0 for all losses, got: %v", rsi)
		}
	})

	t.Run("empty prices", func(t *testing.T) {
		rsi, err := RSI([]float64{}, 14)
		if err == nil {
			t.Errorf("Expected error for empty prices")
		}
		if rsi != 0 {
			t.Errorf("Expected RSI 0 for empty prices, got: %v", rsi)
		}
	})

	t.Run("single price", func(t *testing.T) {
		rsi, err := RSI([]float64{45.0}, 14)
		if err == nil {
			t.Errorf("Expected error for single price")
		}
		if rsi != 0 {
			t.Errorf("Expected RSI 0 for single price, got: %v", rsi)
		}
	})

	t.Run("NaN values", func(t *testing.T) {
		rsi, err := RSI([]float64{45.0, 46.2, math.NaN()}, 14)
		if err == nil {
			t.Errorf("Expected error for NaN prices")
		}
		if rsi != 0 {
			t.Errorf("Expected RSI 0 for NaN prices, got: %v", rsi)
		}
	})

	t.Run("infinity values", func(t *testing.T) {
		rsi, err := RSI([]float64{45.0, 46.2, math.Inf(1)}, 14)
		if err == nil {
			t.Errorf("Expected error for infinity prices")
		}
		if rsi != 0 {
			t.Errorf("Expected RSI 0 for infinity prices, got: %v", rsi)
		}
	})

	t.Run("negative prices", func(t *testing.T) {
		rsi, err := RSI([]float64{-45.0, -46.2, -47.5}, 14)
		if err == nil {
			t.Errorf("Expected error for negative prices")
		}
		if rsi != 0 {
			t.Errorf("Expected RSI 0 for negative prices, got: %v", rsi)
		}
	})

	t.Run("zero prices", func(t *testing.T) {
		rsi, err := RSI([]float64{0, 0, 0}, 14)
		if err == nil {
			t.Errorf("Expected error for zero prices")
		}
		if rsi != 0 {
			t.Errorf("Expected RSI 0 for zero prices, got: %v", rsi)
		}
	})

	t.Run("large price range", func(t *testing.T) {
		rsi, _ := RSI([]float64{1e-10, 1e10, 1e20}, 14)
		if math.IsNaN(rsi) {
			t.Errorf("RSI should handle large price range without NaN")
		}
	})
}
