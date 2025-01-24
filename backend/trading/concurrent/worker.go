package concurrent

import (
	"context"
	"fmt"
	"time"

	"solmeme-trader/monitoring"
)

// Worker handles task execution
type Worker struct {
	id      int
	tasks   chan Task
	results chan *TaskResult
	monitor *monitoring.Monitor
}

// NewWorker creates a new worker
func NewWorker(id int, tasks chan Task, results chan *TaskResult, monitor *monitoring.Monitor) *Worker {
	return &Worker{
		id:      id,
		tasks:   tasks,
		results: results,
		monitor: monitor,
	}
}

// Start starts the worker
func (w *Worker) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case task := <-w.tasks:
				if task == nil {
					continue
				}

				// Record task start
				w.monitor.RecordEvent(ctx, monitoring.Event{
					Type:      monitoring.MetricProcessing,
					Severity:  monitoring.SeverityInfo,
					Message:   fmt.Sprintf("Worker %d starting task %s", w.id, task.GetName()),
					Timestamp: time.Now(),
				})

				// Execute task
				startTime := time.Now()
				err := task.Execute(ctx)

				// Record result
				result := NewTaskResult(task, err, startTime)
				w.results <- result

				// Record task completion
				severity := monitoring.SeverityInfo
				if err != nil {
					severity = monitoring.SeverityError
				}

				w.monitor.RecordEvent(ctx, monitoring.Event{
					Type:     monitoring.MetricProcessing,
					Severity: severity,
					Message: fmt.Sprintf("Worker %d completed task %s in %s",
						w.id, task.GetName(), time.Since(startTime)),
					Details: map[string]interface{}{
						"error":    err,
						"duration": time.Since(startTime).String(),
					},
					Timestamp: time.Now(),
				})
			}
		}
	}()
}
