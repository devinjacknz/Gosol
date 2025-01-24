package monitoring

import (
	"encoding/json"
	"net/http"
)

// API provides HTTP endpoints for monitoring
type API struct {
	monitor IMonitor
}

// NewAPI creates a new monitoring API
func NewAPI(monitor IMonitor) *API {
	return &API{monitor: monitor}
}

// RegisterRoutes registers monitoring API routes
func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/monitoring/health", a.handleHealth)
	mux.HandleFunc("/api/v1/monitoring/metrics", a.handleMetrics)
	mux.HandleFunc("/api/v1/monitoring/events", a.handleEvents)
}

// handleHealth handles health check requests
func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	healthy, subsystems := a.monitor.CheckHealth(r.Context())
	json.NewEncoder(w).Encode(map[string]interface{}{
		"healthy":    healthy,
		"subsystems": subsystems,
	})
}

// handleMetrics handles metrics requests
func (a *API) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics := a.monitor.GetMetrics()
	json.NewEncoder(w).Encode(metrics)
}

// handleEvents handles event requests
func (a *API) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	events := a.monitor.GetEvents()
	json.NewEncoder(w).Encode(events)
}
