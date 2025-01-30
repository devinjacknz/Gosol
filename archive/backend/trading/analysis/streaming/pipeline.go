package streaming

import (
	"sync"
	"time"
)

// 为Processor接口添加Name方法

// Pipeline 流式数据处理流水线
type Pipeline struct {
	windowSize time.Duration
	processors []Processor
}

// NewPipeline 创建新流水线
func NewPipeline(windowSize time.Duration) *Pipeline {
	return &Pipeline{
		windowSize: windowSize,
		processors: make([]Processor, 0),
	}
}

// AddProcessor 添加处理模块
func (p *Pipeline) AddProcessor(proc Processor) {
	p.processors = append(p.processors, proc)
}

// ProcessStream 处理数据流
func (p *Pipeline) ProcessStream(input <-chan float64) <-chan map[string]float64 {
	out := make(chan map[string]float64)
	
	go func() {
		defer close(out)
		
		// 创建处理器通道
		processorChans := make([]chan float64, len(p.processors))
		resultChans := make([]chan float64, len(p.processors))
		for i := range p.processors {
			processorChans[i] = make(chan float64)
			resultChans[i] = make(chan float64)
		}
		
		// 启动处理协程
		var wg sync.WaitGroup
		for i, proc := range p.processors {
			wg.Add(1)
			go func(idx int, p Processor) {
				defer wg.Done()
				for price := range processorChans[idx] {
					result, _ := p.Process(price)
					resultChans[idx] <- result
				}
				close(resultChans[idx])
			}(i, proc)
		}
		
		// 分发数据并收集结果
		go func() {
			for price := range input {
				// 分发到所有处理器
				for i := range processorChans {
					processorChans[i] <- price
				}
				
				// 收集所有处理结果
				results := make(map[string]float64)
				for i, proc := range p.processors {
					results[proc.Name()] = <-resultChans[i]
				}
				out <- results
			}
			
			// 关闭所有处理器输入通道
			for i := range processorChans {
				close(processorChans[i])
			}
		}()
		
		wg.Wait()
	}()
	
	return out
}
