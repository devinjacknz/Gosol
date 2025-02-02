package middleware

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestStats stores request performance metrics
type RequestStats struct {
	Path           string        `json:"path"`
	Method         string        `json:"method"`
	StartTime      time.Time     `json:"start_time"`
	Duration       time.Duration `json:"duration"`
	StatusCode     int          `json:"status_code"`
	BytesWritten   int64        `json:"bytes_written"`
	ClientIP       string       `json:"client_ip"`
	UserAgent      string       `json:"user_agent"`
	NumGoroutines  int          `json:"num_goroutines"`
	MemoryAlloc    uint64       `json:"memory_alloc"`
	MemoryTotal    uint64       `json:"memory_total"`
	NumGC          uint32       `json:"num_gc"`
	TraceID        string       `json:"trace_id"`
	Error          string       `json:"error,omitempty"`
}

// DebugMiddleware adds request tracing and performance monitoring
func DebugMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate trace ID
		traceID := uuid.New().String()
		c.Set("trace_id", traceID)

		// Start timer
		start := time.Now()

		// Get initial memory stats
		var memStatsBefore runtime.MemStats
		runtime.ReadMemStats(&memStatsBefore)

		// Process request
		c.Next()

		// Get final memory stats
		var memStatsAfter runtime.MemStats
		runtime.ReadMemStats(&memStatsAfter)

		// Collect request stats
		stats := RequestStats{
			Path:           c.Request.URL.Path,
			Method:         c.Request.Method,
			StartTime:      start,
			Duration:       time.Since(start),
			StatusCode:     c.Writer.Status(),
			BytesWritten:   int64(c.Writer.Size()),
			ClientIP:       c.ClientIP(),
			UserAgent:      c.Request.UserAgent(),
			NumGoroutines:  runtime.NumGoroutine(),
			MemoryAlloc:    memStatsAfter.Alloc - memStatsBefore.Alloc,
			MemoryTotal:    memStatsAfter.TotalAlloc - memStatsBefore.TotalAlloc,
			NumGC:          memStatsAfter.NumGC - memStatsBefore.NumGC,
			TraceID:        traceID,
		}

		// Get error if any
		if len(c.Errors) > 0 {
			stats.Error = c.Errors.String()
		}

		// Store stats in context and log them
		c.Set("request_stats", stats)
		
		// Log request stats
		fmt.Printf("REQUEST STATS [%s] %s %s - %v\n",
			stats.TraceID,
			stats.Method,
			stats.Path,
			stats.Duration,
		)

		// Store stats in memory for the /debug/stats endpoint
		requestStats := c.MustGet("request_stats").(RequestStats)
		lastRequestStats = &requestStats

		// Log slow requests (>500ms)
		if stats.Duration > 500*time.Millisecond {
			c.Set("slow_request", true)
			fmt.Printf("SLOW REQUEST [%s] %s %s - %v\n",
				stats.TraceID,
				stats.Method,
				stats.Path,
				stats.Duration,
			)
		}
	}
}

// Global variable to store last request stats
var lastRequestStats *RequestStats

// RequestStatsEndpoint returns debug information about requests
func RequestStatsEndpoint(c *gin.Context) {
	if lastRequestStats == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No request stats found"})
		return
	}

	c.JSON(http.StatusOK, lastRequestStats)
}

// MemoryStatsEndpoint returns current memory statistics
func MemoryStatsEndpoint(c *gin.Context) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.JSON(http.StatusOK, gin.H{
		"alloc":        memStats.Alloc,
		"total_alloc":  memStats.TotalAlloc,
		"sys":          memStats.Sys,
		"num_gc":       memStats.NumGC,
		"goroutines":   runtime.NumGoroutine(),
		"cpu_threads":  runtime.NumCPU(),
		"go_version":   runtime.Version(),
		"gc_pause_ns":  memStats.PauseNs[(memStats.NumGC+255)%256],
		"heap_objects": memStats.HeapObjects,
	})
}

// StackTraceEndpoint returns the current goroutine stack traces
func StackTraceEndpoint(c *gin.Context) {
	buf := make([]byte, 1<<20)
	n := runtime.Stack(buf, true)
	c.String(http.StatusOK, string(buf[:n]))
}

// PprofEndpoint enables runtime profiling
func PprofEndpoint(c *gin.Context) {
	c.String(http.StatusOK, "Profiling enabled at /debug/pprof/")
}

// RecoverMiddleware recovers from panics and logs them
func RecoverMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get stack trace
				buf := make([]byte, 1<<16)
				n := runtime.Stack(buf, false)
				stackTrace := string(buf[:n])

				// Log panic
				fmt.Printf("PANIC [%s] %s\nError: %v\nStack Trace:\n%s\n",
					c.GetString("trace_id"),
					c.Request.URL.Path,
					err,
					stackTrace,
				)

				// Return error response
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":       "Internal Server Error",
					"trace_id":    c.GetString("trace_id"),
					"stack_trace": stackTrace,
				})
			}
		}()

		c.Next()
	}
}
