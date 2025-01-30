package concurrent

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/leonzhao/trading-system/backend/monitoring"
)

// Pool represents a worker pool
type Pool struct {
	workers       []*Worker
	tasks         chan Task
	results       chan *TaskResult
	wg            sync.WaitGroup
	monitor       *monitoring.Monitor
	activeWorkers int32
	queueSize     int32
	mutex         sync.RWMutex
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

// Start starts the worker pool with monitoring
func (p *Pool) Start(ctx context.Context) {
	// Start workers
	for _, worker := range p.workers {
		p.wg.Add(1)
		go func(w *Worker) {
			defer p.wg.Done()
			defer p.decrementActiveWorkers()
			p.incrementActiveWorkers()
			w.Start(ctx)
		}(worker)
	}

	// Start result collector
	go p.collectResults(ctx)
}

// Submit submits a task to the pool with error handling
func (p *Pool) Submit(task Task) error {
	select {
	case p.tasks <- task:
		atomic.AddInt32(&p.queueSize, 1)
		return nil
	default:
		return fmt.Errorf("task queue is full")
	}
}

// QueueSize returns the current size of the task queue
func (p *Pool) QueueSize() int {
	return int(atomic.LoadInt32(&p.queueSize))
}

// ActiveWorkers returns the number of currently active workers
func (p *Pool) ActiveWorkers() int {
	return int(atomic.LoadInt32(&p.activeWorkers))
}

// incrementActiveWorkers increments the active workers count
func (p *Pool) incrementActiveWorkers() {
	atomic.AddInt32(&p.activeWorkers, 1)
}

// decrementActiveWorkers decrements the active workers count
func (p *Pool) decrementActiveWorkers() {
	atomic.AddInt32(&p.activeWorkers, -1)
}

// decrementQueueSize decrements the queue size
func (p *Pool) decrementQueueSize() {
	atomic.AddInt32(&p.queueSize, -1)
}

// Wait waits for all tasks to complete
func (p *Pool) Wait() {
	p.wg.Wait()
}

// collectResults collects and processes task results
func (p *Pool) collectResults(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case result := <-p.results:
			if result == nil {
				continue
			}

			p.decrementQueueSize()

			// Record task completion metrics
			severity := monitoring.SeverityInfo
			if result.Error != nil {
				severity = monitoring.SeverityError
			}

			p.monitor.RecordEvent(ctx, monitoring.Event{
				Type:     monitoring.MetricTaskCompletion,
				Severity: severity,
				Message:  "Task completed",
				Details: map[string]interface{}{
					"task":      result.Task.GetName(),
					"type":      result.Task.GetType(),
					"error":     result.Error,
					"duration":  result.Duration.String(),
					"priority":  result.Task.GetPriority(),
					"retries":   result.Task.GetRetryCount(),
					"queueSize": p.QueueSize(),
					"workers":   p.ActiveWorkers(),
				},
				Timestamp: result.EndTime,
			})

			// Handle task retry if needed
			if result.Error != nil && result.Task.ShouldRetry() {
				go func(task Task) {
					if err := p.Submit(task); err != nil {
						p.monitor.RecordEvent(ctx, monitoring.Event{
							Type:     monitoring.MetricTaskCompletion,
							Severity: monitoring.SeverityError,
							Message:  "Task retry failed",
							Details: map[string]interface{}{
								"task":  task.GetName(),
								"error": err.Error(),
							},
							Timestamp: result.EndTime,
						})
					}
				}(result.Task)
			}
		}
	}
}
