package monitoring

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPI_HandleMetrics(t *testing.T) {
	api, mockMonitor := TestAPI(t)

	testMetrics := map[string]float64{
		"test_metric":     42.0,
		"another_metric": 123.0,
	}

	// Setup expectations
	mockMonitor.ExpectGetMetrics(TestContext(t), testMetrics).Return(testMetrics)

	// Create test request
	req := TestRequest(t, http.MethodGet, "/api/metrics")
	w := httptest.NewRecorder()

	// Handle request
	api.handleMetrics(w, req)

	// Check response
	AssertStatusCode(t, http.StatusOK, w.Code)
	AssertJSONResponse(t, w)

	var response map[string]float64
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, testMetrics, response)

	// Verify expectations
	mockMonitor.AssertExpectations(t)
}

func TestAPI_HandleHealth(t *testing.T) {
	tests := []struct {
		name           string
		healthStates   map[string]bool
		healthy        bool
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "all healthy",
			healthStates: map[string]bool{
				"component1": true,
				"component2": true,
			},
			healthy:        true,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"healthy": true,
				"state": map[string]interface{}{
					"component1": true,
					"component2": true,
				},
			},
		},
		{
			name: "partially unhealthy",
			healthStates: map[string]bool{
				"component1": true,
				"component2": false,
			},
			healthy:        false,
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody: map[string]interface{}{
				"healthy": false,
				"state": map[string]interface{}{
					"component1": true,
					"component2": false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api, mockMonitor := TestAPI(t)

			// Setup expectations
			mockMonitor.ExpectCheckHealth(TestContext(t), tt.healthy, tt.healthStates).
				Return(tt.healthy, tt.healthStates)

			// Create test request
			req := TestRequest(t, http.MethodGet, "/api/health")
			w := httptest.NewRecorder()

			// Handle request
			api.handleHealth(w, req)

			// Check response
			AssertStatusCode(t, tt.expectedStatus, w.Code)
			AssertJSONResponse(t, w)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			// Verify expectations
			mockMonitor.AssertExpectations(t)
		})
	}
}

func TestAPI_MethodNotAllowed(t *testing.T) {
	api, _ := TestAPI(t)

	tests := []struct {
		name     string
		method   string
		endpoint string
	}{
		{
			name:     "POST metrics",
			method:   http.MethodPost,
			endpoint: "/api/metrics",
		},
		{
			name:     "PUT metrics",
			method:   http.MethodPut,
			endpoint: "/api/metrics",
		},
		{
			name:     "POST health",
			method:   http.MethodPost,
			endpoint: "/api/health",
		},
		{
			name:     "PUT health",
			method:   http.MethodPut,
			endpoint: "/api/health",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := TestRequest(t, tt.method, tt.endpoint)
			w := httptest.NewRecorder()

			// Handle request based on endpoint
			if tt.endpoint == "/api/metrics" {
				api.handleMetrics(w, req)
			} else {
				api.handleHealth(w, req)
			}

			AssertStatusCode(t, http.StatusMethodNotAllowed, w.Code)
		})
	}
}

func TestAPI_RegisterRoutes(t *testing.T) {
	api, mockMonitor := TestAPI(t)
	server := TestServer(t, api)
	defer server.Close()

	// Test metrics endpoint
	testMetrics := map[string]float64{"test": 42.0}
	mockMonitor.ExpectGetMetrics(TestContext(t), testMetrics).Return(testMetrics)

	resp, err := http.Get(server.URL + "/api/metrics")
	assert.NoError(t, err)
	AssertStatusCode(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Test health endpoint
	healthStates := map[string]bool{"test": true}
	mockMonitor.ExpectCheckHealth(TestContext(t), true, healthStates).Return(true, healthStates)

	resp, err = http.Get(server.URL + "/api/health")
	assert.NoError(t, err)
	AssertStatusCode(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Verify all expectations
	mockMonitor.AssertExpectations(t)
}
