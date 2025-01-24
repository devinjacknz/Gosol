package analysis

import "fmt"

// RSI 计算相对强弱指数
func RSI(prices []float64, period int) []float64 {
    gains := make([]float64, len(prices))
    losses := make([]float64, len(prices))
    
    for i := 1; i < len(prices); i++ {
        change := prices[i] - prices[i-1]
        if change > 0 {
            gains[i] = change
        } else {
            losses[i] = -change
        }
    }
    
    avgGain, _ := movingAverage(gains[1:], period)
    avgLoss, _ := movingAverage(losses[1:], period)
    
    rs := avgGain / avgLoss
    return []float64{100 - (100 / (1 + rs))}
}

// MACD 计算移动平均收敛散度 
func MACD(prices []float64, fastPeriod, slowPeriod, signalPeriod int) ([]float64, []float64, []float64) {
    emaFast := EMA(prices, fastPeriod)
    emaSlow := EMA(prices, slowPeriod)
    
    macdLine := make([]float64, len(prices))
    for i := range emaFast {
        macdLine[i] = emaFast[i] - emaSlow[i]
    }
    
    signalLine := EMA(macdLine, signalPeriod)
    histogram := make([]float64, len(macdLine))
    for i := range macdLine {
        histogram[i] = macdLine[i] - signalLine[i]
    }
    
    return macdLine, signalLine, histogram
}

// EMA 计算指数移动平均
func EMA(prices []float64, period int) []float64 {
    ema := make([]float64, len(prices))
    multiplier := 2.0 / (float64(period) + 1)
    
    // 第一个EMA是SMA
    sma, _ := movingAverage(prices[:period], period)
    ema[period-1] = sma
    
    for i := period; i < len(prices); i++ {
        ema[i] = (prices[i]-ema[i-1])*multiplier + ema[i-1]
    }
    return ema[period-1:]
}

func movingAverage(data []float64, period int) (float64, error) {
    if len(data) < period {
        return 0, fmt.Errorf("数据长度不足：需要 %d 个周期，当前 %d", period, len(data))
    }
    
    sum := 0.0
    for _, v := range data[:period] {
        sum += v
    }
    return sum / float64(period), nil
}
