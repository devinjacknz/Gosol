package streaming

import (
	"errors"
	"fmt"
)

type Processor interface {
	Initialize(int) error
	Process(float64) (float64, error)
	Current() float64
	Name() string
}

type RSI struct {
	windowSize     int
	window         []float64
	currentAvgGain float64
	currentAvgLoss float64
	initialized    bool
}

func (r *RSI) Initialize(windowSize int) error {
	if windowSize < 2 { // 最小窗口大小改为2
		return fmt.Errorf("window size must be ≥2, got %d", windowSize)
	}
	r.windowSize = windowSize
	r.window = make([]float64, 0, windowSize+1)
	return nil
}

func (r *RSI) Process(price float64) (float64, error) {
	// 初始化阶段处理
	if len(r.window) == 0 {
		r.window = append(r.window, price)
		return 0, nil
	}

	// 维护滚动窗口（保持windowSize+1个最新价格）
	r.window = append(r.window, price)
	if len(r.window) > r.windowSize+1 {
		r.window = r.window[1:]
	}

	// 当有足够数据时标记初始化完成
	if !r.initialized && len(r.window) >= r.windowSize+1 {
		r.initialized = true
	}

	// 在未初始化完成前返回中间状态
	if !r.initialized {
		return 0, nil
	}

	var avgGain, avgLoss float64
	if len(r.window) == r.windowSize+1 {
		// 初始计算
		for i := 0; i < r.windowSize; i++ {
			delta := r.window[i+1] - r.window[i]
			if delta > 0 {
				avgGain += delta
			} else {
				avgLoss -= delta
			}
		}
		r.currentAvgGain = avgGain / float64(r.windowSize)
		r.currentAvgLoss = avgLoss / float64(r.windowSize)
	} else {
		// 更新计算
		// 使用精确的指数平滑计算
		delta := price - r.window[len(r.window)-2]
		smoothing := 1.0 / float64(r.windowSize)
		
		if delta > 0 {
			r.currentAvgGain = (1.0 - smoothing)*r.currentAvgGain + smoothing*delta
			r.currentAvgLoss = (1.0 - smoothing)*r.currentAvgLoss
		} else {
			r.currentAvgGain = (1.0 - smoothing)*r.currentAvgGain
			r.currentAvgLoss = (1.0 - smoothing)*r.currentAvgLoss - smoothing*delta
		}
	}

	if r.currentAvgLoss == 0 {
		return 100.0, nil
	}
	rs := r.currentAvgGain / r.currentAvgLoss
	return 100.0 - (100.0 / (1 + rs)), nil
}

func (r *RSI) Current() float64 {
	if r.currentAvgLoss == 0 {
		return 100.0
	}
	rs := r.currentAvgGain / r.currentAvgLoss
	return 100.0 - (100.0 / (1 + rs))
}

func (r *RSI) Name() string {
	return "RSI"
}

type EMA struct {
	period      int
	alpha       float64
	current     float64
	initialized bool
}

func (e *EMA) Initialize(period int) error {
	if period < 1 {
		return errors.New("period must be ≥1")
	}
	e.period = period
	e.alpha = 2.0 / (float64(period) + 1.0) // 精确浮点数计算
	return nil
}

func (e *EMA) Process(price float64) (float64, error) {
	if price < 0 {
		return 0, errors.New("price cannot be negative")
	}

	if !e.initialized {
		// 初始化阶段使用第一个price作为初始值（SMA方式）
		e.current = price
		e.initialized = true
		return e.current, nil
	}

	// 精确EMA计算公式：EMA = (price * alpha) + (previous_EMA * (1 - alpha))
	e.current = (price * e.alpha) + (e.current * (1 - e.alpha))
	return e.current, nil
}

func (e *EMA) Current() float64 {
	return e.current
}

func (e *EMA) Name() string {
	return "EMA"
}

func NewRSI(windowSize int) (*RSI, error) {
	rsi := &RSI{}
	if err := rsi.Initialize(windowSize); err != nil {
		return nil, err
	}
	return rsi, nil
}

func NewEMA(period int) (*EMA, error) {
	ema := &EMA{}
	if err := ema.Initialize(period); err != nil {
		return nil, err
	}
	return ema, nil
}
