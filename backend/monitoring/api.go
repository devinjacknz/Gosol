package monitoring

import (
	"encoding/json"
	"net/http"
)

// API handles monitoring HTTP endpoints
type API struct {
	monitor IMonitor
}

// NewAPI creates a new monitoring API
func NewAPI(monitor IMonitor) *API {
	return &API{
		monitor: monitor,
	}
}

// RegisterRoutes registers HTTP routes
func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/metrics", a.handleMetrics)
	mux.HandleFunc("/api/health", a.handleHealth)
}

// handleMetrics handles metrics endpoint
func (a *API) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics := a.monitor.GetMetrics(r.Context())
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleHealth handles health check endpoint
func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	healthy, healthState := a.monitor.CheckHealth(r.Context())
	response := map[string]interface{}{
		"healthy": healthy,
		"state":   healthState,
	}

	w.Header().Set("Content-Type", "application/json")
	if !healthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
