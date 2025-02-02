package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Service represents the monitoring service
type Service struct {
	server *http.Server
}

// Config holds the configuration for the monitoring service
type Config struct {
	Port int
}

// NewService creates a new monitoring service
func NewService(cfg Config) *Service {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}

	return &Service{
		server: server,
	}
}

// Start starts the monitoring service
func (s *Service) Start() error {
	return s.server.ListenAndServe()
}

// Stop stops the monitoring service
func (s *Service) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// RecordIndicatorCalculation records the duration of an indicator calculation
func RecordIndicatorCalculation(indicatorName string, duration time.Duration) {
	IndicatorCalculationDuration.WithLabelValues(indicatorName).Observe(duration.Seconds())
}

// RecordIndicatorError records an indicator calculation error
func RecordIndicatorError(indicatorName, errorType string) {
	IndicatorCalculationErrors.WithLabelValues(indicatorName, errorType).Inc()
}

// RecordIndicatorValue records the current value of an indicator
func RecordIndicatorValue(indicatorName string, value float64) {
	IndicatorValues.WithLabelValues(indicatorName).Set(value)
}

// RecordPriceUpdate records a price update
func RecordPriceUpdate(symbol string) {
	PriceUpdates.WithLabelValues(symbol).Inc()
}

// RecordStorageOperation records the duration of a storage operation
func RecordStorageOperation(operation string, duration time.Duration) {
	StorageOperationDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordStorageError records a storage operation error
func RecordStorageError(operation, errorType string) {
	StorageOperationErrors.WithLabelValues(operation, errorType).Inc()
}

// RecordBatchProcessing records the duration and size of batch processing
func RecordBatchProcessing(indicatorName string, duration time.Duration, size int) {
	BatchProcessingDuration.WithLabelValues(indicatorName).Observe(duration.Seconds())
	BatchSize.WithLabelValues(indicatorName).Observe(float64(size))
}
