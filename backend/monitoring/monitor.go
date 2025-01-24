package monitoring

import (
	"context"
	"sync"
	"time"
)

// EventType represents the type of event
type EventType string

// EventSeverity represents the severity level of an event
type EventSeverity string

const (
	// Event types
	MetricTrading     EventType = "trading"
	MetricSystem      EventType = "system"
	MetricPerformance EventType = "performance"
	MetricMarketData  EventType = "market_data"
	MetricProcessing  EventType = "processing"
	MetricTaskCompletion EventType = "task_completion"

	// Event severities
	SeverityInfo     EventSeverity = "info"
	SeverityWarning  EventSeverity = "warning"
	SeverityError    EventSeverity = "error"
	SeverityCritical EventSeverity = "critical"
)

// Event represents a monitoring event
type Event struct {
	Type      EventType     `json:"type"`
	Severity  EventSeverity `json:"severity"`
	Message   string        `json:"message"`
	Details   interface{}   `json:"details,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

// Metrics represents system metrics
type Metrics struct {
	TotalTrades      int64
	SuccessfulTrades int64
	FailedTrades     int64
	TotalVolume      float64
	AverageLatency   float64
	mu               sync.RWMutex
}

// HealthState represents the system health state
type HealthState struct {
	Status    string
	LastCheck time.Time
	mu        sync.RWMutex
}

// IMonitor defines the interface for system monitoring
type IMonitor interface {
	// Event recording
	RecordEvent(ctx context.Context, event Event)
	RecordMetric(ctx context.Context, name string, value float64, tags map[string]string)
	RecordMarketData(ctx context.Context, tokenAddress string, price float64, volume float64)

	// Health checks
	CheckHealth(ctx context.Context) (bool, map[string]bool)
	SetHealth(status string)

	// Event handling
	AddEventHandler(eventType EventType, handler EventHandler)
	GetEvents() []Event
	GetEventsByType(eventType EventType) []Event

	// Metrics
	GetMetrics() Metrics
	GetSuccessRate() float64
	GetAverageVolume() float64

	// Lifecycle
	Reset()
	Close()
}

// Monitor handles system monitoring and metrics
type Monitor struct {
	events      []Event
	eventsChan  chan Event
	metrics     Metrics
	healthState HealthState
	handlers    map[EventType][]EventHandler
	mu          sync.RWMutex
}

// EventHandler represents a function that handles events
type EventHandler func(Event)

// NewMonitor creates a new monitor
func NewMonitor() *Monitor {
	m := &Monitor{
		events:     make([]Event, 0),
		eventsChan: make(chan Event, 1000),
		handlers:   make(map[EventType][]EventHandler),
		metrics: Metrics{
			mu: sync.RWMutex{},
		},
		healthState: HealthState{
			Status:    "healthy",
			LastCheck: time.Now(),
			mu:        sync.RWMutex{},
		},
	}

	go m.processEvents()
	return m
}

// processEvents processes events from the channel
func (m *Monitor) processEvents() {
	for event := range m.eventsChan {
		m.mu.Lock()
		m.events = append(m.events, event)
		m.mu.Unlock()

		// Update metrics
		m.UpdateMetrics(event)

		// Call handlers for this event type
		if handlers, ok := m.handlers[event.Type]; ok {
			for _, handler := range handlers {
				handler(event)
			}
		}
	}
}

// RecordEvent records a monitoring event
func (m *Monitor) RecordEvent(ctx context.Context, event Event) {
	event.Timestamp = time.Now()
	select {
	case m.eventsChan <- event:
	default:
		// Channel is full, log error
		m.recordError("Event channel full, dropping event")
	}
}

// RecordMetric records a metric event
func (m *Monitor) RecordMetric(ctx context.Context, name string, value float64, tags map[string]string) {
	m.RecordEvent(ctx, Event{
		Type:     MetricSystem,
		Severity: SeverityInfo,
		Message:  "Metric recorded",
		Details: map[string]interface{}{
			"name":  name,
			"value": value,
			"tags":  tags,
		},
	})
}

// RecordMarketData records market data
func (m *Monitor) RecordMarketData(ctx context.Context, tokenAddress string, price float64, volume float64) {
	m.RecordEvent(ctx, Event{
		Type:     MetricMarketData,
		Severity: SeverityInfo,
		Message:  "Market data updated",
		Details: map[string]interface{}{
			"tokenAddress": tokenAddress,
			"price":       price,
			"volume":      volume,
		},
	})
}

// CheckHealth checks the system health
func (m *Monitor) CheckHealth(ctx context.Context) (bool, map[string]bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	isHealthy := m.healthState.Status == "healthy"
	subsystems := map[string]bool{
		"database":    isHealthy,
		"api":        isHealthy,
		"processing": isHealthy,
	}

	return isHealthy, subsystems
}

// SetHealth sets the system health status
func (m *Monitor) SetHealth(status string) {
	m.UpdateHealthState(status)
}

// Reset resets all metrics and events
func (m *Monitor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.events = make([]Event, 0)
	m.metrics = Metrics{
		mu: sync.RWMutex{},
	}
	m.healthState = HealthState{
		Status:    "healthy",
		LastCheck: time.Now(),
		mu:        sync.RWMutex{},
	}
}

// AddEventHandler adds a handler for a specific event type
func (m *Monitor) AddEventHandler(eventType EventType, handler EventHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.handlers[eventType]; !ok {
		m.handlers[eventType] = make([]EventHandler, 0)
	}
	m.handlers[eventType] = append(m.handlers[eventType], handler)
}

// GetEvents returns all recorded events
func (m *Monitor) GetEvents() []Event {
	m.mu.RLock()
	defer m.mu.RUnlock()

	events := make([]Event, len(m.events))
	copy(events, m.events)
	return events
}

// GetEventsByType returns events of a specific type
func (m *Monitor) GetEventsByType(eventType EventType) []Event {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var events []Event
	for _, event := range m.events {
		if event.Type == eventType {
			events = append(events, event)
		}
	}
	return events
}

// GetMetrics returns the current metrics
func (m *Monitor) GetMetrics() Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.metrics
}

// GetSuccessRate returns the trade success rate
func (m *Monitor) GetSuccessRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.metrics.TotalTrades == 0 {
		return 0
	}
	return float64(m.metrics.SuccessfulTrades) / float64(m.metrics.TotalTrades)
}

// GetAverageVolume returns the average trade volume
func (m *Monitor) GetAverageVolume() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.metrics.TotalTrades == 0 {
		return 0
	}
	return m.metrics.TotalVolume / float64(m.metrics.TotalTrades)
}

// UpdateMetrics updates system metrics based on an event
func (m *Monitor) UpdateMetrics(event Event) {
	if event.Type != MetricTrading {
		return
	}

	details, ok := event.Details.(map[string]interface{})
	if !ok {
		return
	}

	m.metrics.mu.Lock()
	defer m.metrics.mu.Unlock()

	switch event.Message {
	case "Trade execution attempt":
		m.metrics.TotalTrades++
	case "Trade executed successfully":
		m.metrics.SuccessfulTrades++
		if volume, ok := details["value"].(float64); ok {
			m.metrics.TotalVolume += volume
		}
	case "Trade execution failed":
		m.metrics.FailedTrades++
	}
}

// UpdateHealthState updates the system health state
func (m *Monitor) UpdateHealthState(status string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.healthState.Status = status
	m.healthState.LastCheck = time.Now()
}

// recordError records an internal monitor error
func (m *Monitor) recordError(message string) {
	event := Event{
		Type:      MetricSystem,
		Severity:  SeverityError,
		Message:   message,
		Timestamp: time.Now(),
	}

	m.mu.Lock()
	m.events = append(m.events, event)
	m.mu.Unlock()
}

// Close closes the monitor and its event channel
func (m *Monitor) Close() {
	close(m.eventsChan)
}

// Ensure Monitor implements IMonitor
var _ IMonitor = (*Monitor)(nil)

// NewTestEvent creates a new event for testing
func NewTestEvent(eventType EventType, severity EventSeverity, message string) Event {
	return Event{
		Type:      eventType,
		Severity:  severity,
		Message:   message,
		Details:   map[string]interface{}{},
		Timestamp: time.Now(),
	}
}
