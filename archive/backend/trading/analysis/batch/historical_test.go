package batch_test

import (
	"testing"
	"math/rand"
	"github.com/leonzhao/trading-system/backend/trading/analysis/batch"
)

func generateTestData(n int) []float64 {
	data := make([]float64, n)
	base := rand.Float64() * 100
	for i := range data {
		base += rand.NormFloat64() * 0.5
		data[i] = base
	}
	return data
}

func TestHistoricalConversion(t *testing.T) {
	testCases := []struct {
		name          string
		input         []float64
		expectedLen   int
		expectValid   bool
	}{
		{"充足数据", generateTestData(100), 66, true},     // 100-26-9+1=66
		{"边界数据", generateTestData(35), 1, true},      // 35-26-9+1=1
		{"不足数据", generateTestData(34), 0, false},    // 需要至少35个数据点
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := batch.ConvertHistoricalMACD(tc.input, 12, 26, 9)
			if err != nil && tc.expectValid {
				t.Fatalf("预期成功但失败: %v", err)
			}
			
			if tc.expectValid {
				if err != nil {
					t.Fatalf("预期成功但失败: %v", err)
				}
				if len(result) != tc.expectedLen {
					t.Errorf("MACD结果数量不符，期望%d，实际%d", tc.expectedLen, len(result))
				}
				// 验证最后一个值
				if len(result) > 0 {
					last := result[len(result)-1]
					if last.MACD == 0 || last.Signal == 0 {
						t.Error("发现无效的零值")
					}
				}
			} else {
				if err == nil {
					t.Error("预期错误但成功")
				}
			}
		})
	}
}

func TestInvalidInput(t *testing.T) {
	_, err := batch.ConvertHistoricalMACD([]float64{}, 12, 26, 9)
	if err == nil {
		t.Error("空输入时应返回错误")
	}
}
