package batch

import (
	"fmt"
	"math"
)

type MACDResult struct {
	MACD   float64
	Signal float64
	Hist   float64
}

// CalculateMACD 实现自主MACD计算逻辑
func CalculateMACD(prices []float64, fastPeriod, slowPeriod, signalPeriod int) ([]float64, []float64, []float64, error) {
	macdResults, err := ConvertHistoricalMACD(prices, fastPeriod, slowPeriod, signalPeriod)
	if err != nil {
		return nil, nil, nil, err
	}
	
	macd := make([]float64, len(macdResults))
	signal := make([]float64, len(macdResults))
	hist := make([]float64, len(macdResults))
	
	for i, result := range macdResults {
		macd[i] = result.MACD
		signal[i] = result.Signal 
		hist[i] = result.Hist
	}
	
	return macd, signal, hist, nil
}

func ConvertHistoricalMACD(data []float64, fastPeriod, slowPeriod, signalPeriod int) ([]MACDResult, error) {
	// 参数验证
	if fastPeriod <= 0 || slowPeriod <= 0 || signalPeriod <= 0 {
		return nil, fmt.Errorf("invalid parameters: periods must be positive (fast=%d, slow=%d, signal=%d)", 
			fastPeriod, slowPeriod, signalPeriod)
	}
	if fastPeriod >= slowPeriod {
		return nil, fmt.Errorf("invalid parameters: fast period (%d) must be less than slow period (%d)", 
			fastPeriod, slowPeriod)
	}
	
	// 数据验证
	if len(data) == 0 {
		return nil, fmt.Errorf("empty input data")
	}
	for i, v := range data {
		if math.IsNaN(v) {
			return nil, fmt.Errorf("NaN value detected at index %d", i)
		}
	}

	minDataLength := slowPeriod + signalPeriod
	if len(data) < minDataLength {
		return nil, fmt.Errorf("insufficient data: need at least %d values (slow=%d + signal=%d), got %d", 
			minDataLength, slowPeriod, signalPeriod, len(data))
	}

	// 计算完整EMA序列
	fastEMAs := calculateEMA(data, fastPeriod)
	slowEMAs := calculateEMA(data, slowPeriod)
	
	// 修正数据对齐逻辑（增加1个偏移量）
	startIdx := max(fastPeriod, slowPeriod) - 1
	
	// 计算有效数据长度（考虑信号计算所需空间）
	validLength := len(data) - startIdx - signalPeriod
	if validLength <= 0 {
		return nil, fmt.Errorf("insufficient data for MACD calculation")
	}
	
	// 对齐EMA数据（仅保留有效数据段）
	alignedFast := fastEMAs[startIdx : startIdx+validLength+(signalPeriod-1)]
	alignedSlow := slowEMAs[startIdx : startIdx+validLength+(signalPeriod-1)]
	
	// 计算MACD差值（保留signalPeriod-1个额外值用于信号计算）
	macdValues := make([]float64, validLength+(signalPeriod-1))
	for i := 0; i < len(macdValues); i++ {
		macdValues[i] = alignedFast[i] - alignedSlow[i]
	}

	// 计算信号线EMA（完整保留计算结果）
	signalEMAs := calculateEMA(macdValues, signalPeriod)
	
	// 对齐有效结果（去除前signalPeriod-1个无效值）
	finalStart := signalPeriod - 1
	if len(signalEMAs) < finalStart + validLength {
		return nil, fmt.Errorf("insufficient data for signal calculation")
	}
	
	macdValues = macdValues[finalStart : finalStart+validLength]
	signalEMAs = signalEMAs[finalStart : finalStart+validLength]
	
	// 生成最终结果（使用对齐后的macdValues长度）
	results := make([]MACDResult, len(macdValues))
	for i := 0; i < len(macdValues); i++ {
		results[i] = MACDResult{
			MACD:   macdValues[i],
			Signal: signalEMAs[i],
			Hist:   macdValues[i] - signalEMAs[i],
		}
	}
	return results, nil
}

// 修正EMA计算（严格遵循标准公式）
func calculateEMA(data []float64, period int) []float64 {
	if len(data) < period {
		return make([]float64, len(data)) // 返回全NaN数组
	}

	emas := make([]float64, len(data))
	for i := 0; i < period-1; i++ {
		emas[i] = math.NaN()
	}

	// 计算初始SMA（使用前period个价格的平均值）
	sma := 0.0
	for _, v := range data[:period] {
		sma += v
	}
	sma /= float64(period)
	emas[period-1] = sma

	// 计算EMA乘数
	multiplier := 2.0 / (float64(period) + 1.0)

	// 严格遵循标准EMA公式
	for i := period; i < len(data); i++ {
		currentPrice := data[i]
		prevEMA := emas[i-1]
		emas[i] = (currentPrice * multiplier) + (prevEMA * (1 - multiplier))
	}

	return emas
}
