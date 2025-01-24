package monitoring

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

// TestContext returns a context with timeout for testing
func TestContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// TestRequest creates a test HTTP request with test context
func TestRequest(t *testing.T, method, path string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	return req.WithContext(TestContext(t))
}

// TestMonitor creates a mock monitor with common expectations
func TestMonitor(t *testing.T) *MockMonitor {
	mockMonitor := NewMockMonitor()

	// Setup common expectations
	mockMonitor.On("GetMetrics", mock.MatchedBy(func(ctx context.Context) bool {
		return ctx != nil
	})).Maybe().Return(map[string]float64{})

	mockMonitor.On("CheckHealth", mock.MatchedBy(func(ctx context.Context) bool {
		return ctx != nil
	})).Maybe().Return(true, map[string]bool{})

	mockMonitor.On("RecordEvent", mock.MatchedBy(func(ctx context.Context) bool {
		return ctx != nil
	}), mock.AnythingOfType("Event")).Maybe().Return()

	mockMonitor.On("RecordMetric", mock.MatchedBy(func(ctx context.Context) bool {
		return ctx != nil
	}), mock.AnythingOfType("string"), mock.AnythingOfType("float64")).Maybe().Return()

	mockMonitor.On("RecordMarketData", mock.MatchedBy(func(ctx context.Context) bool {
		return ctx != nil
	}), mock.Anything, mock.AnythingOfType("time.Duration")).Maybe().Return()

	mockMonitor.On("SetHealth", mock.AnythingOfType("string"), mock.AnythingOfType("bool")).Maybe().Return()

	mockMonitor.On("Reset").Maybe().Return()

	return mockMonitor
}

// TestEvent creates a test event
func TestEvent(eventType string, severity string) Event {
	return Event{
		Type:      eventType,
		Severity:  severity,
		Message:   "Test event",
		Details:   map[string]interface{}{"test": true},
		Timestamp: time.Now(),
	}
}

// TestAPI creates a test API with mock monitor
func TestAPI(t *testing.T) (*API, *MockMonitor) {
	mockMonitor := TestMonitor(t)
	api := NewAPI(mockMonitor)
	return api, mockMonitor
}

// TestServer creates a test HTTP server with API routes
func TestServer(t *testing.T, api *API) *httptest.Server {
	mux := http.NewServeMux()
	api.RegisterRoutes(mux)
	return httptest.NewServer(mux)
}

// AssertJSONResponse asserts JSON response headers
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder) {
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type: application/json, got %s", contentType)
	}
}

// AssertStatusCode asserts HTTP status code
func AssertStatusCode(t *testing.T, expected int, actual int) {
	if expected != actual {
		t.Errorf("Expected status code %d, got %d", expected, actual)
	}
}
