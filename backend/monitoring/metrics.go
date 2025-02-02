package monitoring

import (
	"sync/atomic"
	"time"
)

// LLMMetrics tracks basic metrics for LLM operations
type LLMMetrics struct {
	RequestCount   atomic.Int64
	ErrorCount     atomic.Int64
	FallbackCount  atomic.Int64
	TotalTokens    atomic.Int64
	TotalLatencyMs atomic.Int64
}

var metrics = &LLMMetrics{}

// RecordLLMRequest records metrics for an LLM API request
func RecordLLMRequest(model, operation string, duration time.Duration, status string, tokens int) {
	metrics.RequestCount.Add(1)
	metrics.TotalLatencyMs.Add(duration.Milliseconds())
	metrics.TotalTokens.Add(int64(tokens))

	if status == "error" {
		metrics.ErrorCount.Add(1)
	}
}

// RecordLLMFallback records a fallback attempt
func RecordLLMFallback() {
	metrics.FallbackCount.Add(1)
}

// GetMetrics returns the current metrics
func GetMetrics() *LLMMetrics {
	return metrics
}
