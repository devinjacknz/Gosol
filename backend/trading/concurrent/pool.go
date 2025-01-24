package concurrent

import (
	"context"
	"sync"

	"solmeme-trader/monitoring"
)

// Pool represents a worker pool
type Pool struct {
	workers []*Worker
	tasks   chan Task
	results chan *TaskResult
	wg      sync.WaitGroup
	monitor *monitoring.Monitor
}

// NewPool creates a new worker pool
func NewPool(numWorkers int, monitor *monitoring.Monitor) *Pool {
	p := &Pool{
		tasks:   make(chan Task, numWorkers*2),
		results: make(chan *TaskResult, numWorkers*2),
		monitor: monitor,
	}

	// Create workers
	p.workers = make([]*Worker, numWorkers)
	for i := 0; i < numWorkers; i++ {
		p.workers[i] = NewWorker(i, p.tasks, p.results, monitor)
	}

	return p
}

// Start starts the worker pool
func (p *Pool) Start(ctx context.Context) {
	// Start workers
	for _, worker := range p.workers {
		p.wg.Add(1)
		go func(w *Worker) {
			defer p.wg.Done()
			w.Start(ctx)
		}(worker)
	}

	// Start result collector
	go p.collectResults(ctx)
}

// Submit submits a task to the pool
func (p *Pool) Submit(task Task) {
	p.tasks <- task
}

// Wait waits for all tasks to complete
func (p *Pool) Wait() {
	p.wg.Wait()
}

// collectResults collects task results
func (p *Pool) collectResults(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case result := <-p.results:
			if result == nil {
				continue
			}

			// Record task metrics
			p.monitor.RecordEvent(ctx, monitoring.Event{
				Type:      monitoring.MetricTaskCompletion,
				Severity:  monitoring.SeverityInfo,
				Message:   "Task completed",
				Details: map[string]interface{}{
					"task":     result.Task.GetName(),
					"type":     result.Task.GetType(),
					"error":    result.Error,
					"duration": result.Duration.String(),
				},
				Timestamp: result.EndTime,
			})
		}
	}
}
