package streaming

import (
	"math"
	"testing"
	"time"
)

func TestDataPipeline(t *testing.T) {
	pipeline := NewPipeline(1 * time.Minute)
	rsi, _ := NewRSI(14)
	pipeline.AddProcessor(rsi)
	ema, _ := NewEMA(5)
	pipeline.AddProcessor(ema)

	input := make(chan float64, 10)
	for i := 0; i < 10; i++ {
		input <- float64(i)
	}
	close(input)

	out := pipeline.ProcessStream(input)

	var results []map[string]float64
	for v := range out {
		results = append(results, v)
	}

	if len(results) != 10 {
		t.Errorf("Expected 10 processed values, got %d", len(results))
	}

	// 验证最后一个结果
	lastResult := results[len(results)-1]
	
	// 检查RSI值
	if rsiVal, ok := lastResult["RSI"]; !ok {
		t.Error("Missing RSI value in results")
	} else if rsiVal < 0 || rsiVal > 100 {
		t.Errorf("Invalid RSI value: %v", rsiVal)
	}
	
	// 检查EMA值
	if emaVal, ok := lastResult["EMA"]; !ok {
		t.Error("Missing EMA value in results")
	} else {
		// 预期EMA(5)计算结果为7.052025（根据精确计算）
		expectedEMA := 7.052025
		if math.Abs(emaVal - expectedEMA) > 0.0001 {
			t.Errorf("EMA calculation failed got %.6f, expected %.6f", emaVal, expectedEMA)
		}
	}
}
