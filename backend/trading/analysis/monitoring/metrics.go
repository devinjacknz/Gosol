package monitoring

import "github.com/prometheus/client_golang/prometheus"

var (
	IndicatorCalculationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "indicator_calculation_duration_seconds",
			Help: "Duration of indicator calculations",
		},
		[]string{"indicator"},
	)

	IndicatorCalculationErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "indicator_calculation_errors_total",
			Help: "Total number of indicator calculation errors",
		},
		[]string{"indicator", "error_type"},
	)

	IndicatorValues = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "indicator_values",
			Help: "Current values of indicators",
		},
		[]string{"indicator"},
	)

	PriceUpdates = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "price_updates_total",
			Help: "Total number of price updates",
		},
		[]string{"symbol"},
	)

	StorageOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "storage_operation_duration_seconds",
			Help: "Duration of storage operations",
		},
		[]string{"operation"},
	)

	StorageOperationErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "storage_operation_errors_total",
			Help: "Total number of storage operation errors",
		},
		[]string{"operation", "error_type"},
	)

	BatchProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "batch_processing_duration_seconds",
			Help: "Duration of batch processing",
		},
		[]string{"indicator"},
	)

	BatchSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "batch_size",
			Help: "Size of processed batches",
		},
		[]string{"indicator"},
	)
)

func init() {
	prometheus.MustRegister(
		IndicatorCalculationDuration,
		IndicatorCalculationErrors,
		IndicatorValues,
		PriceUpdates,
		StorageOperationDuration,
		StorageOperationErrors,
		BatchProcessingDuration,
		BatchSize,
	)
}
