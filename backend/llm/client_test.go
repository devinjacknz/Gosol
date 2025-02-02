package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	primaryModel := &Model{
		Type:    LocalOllama,
		Name:    ModelLlama3,
		BaseURL: "http://localhost:11434",
	}
	fallbackModel := &Model{
		Type:    DeepSeekAPI,
		Name:    ModelDeepSeekR1,
		BaseURL: "http://api.deepseek.com",
		APIKey:  "test-key",
	}

	client := NewClient(primaryModel, fallbackModel)
	require.NotNil(t, client)

	defaultClient, ok := client.(*DefaultClient)
	require.True(t, ok)
	assert.Equal(t, primaryModel, defaultClient.primaryModel)
	assert.Equal(t, fallbackModel, defaultClient.fallbackModel)
	assert.NotNil(t, defaultClient.httpClient)
	assert.NotNil(t, defaultClient.ollamaClient)
	assert.NotNil(t, defaultClient.rateLimiter)
}

func TestGetSetModel(t *testing.T) {
	client := NewClient(&Model{}, &Model{}).(*DefaultClient)

	newModel := &Model{
		Type:    LocalOllama,
		Name:    ModelLlama3,
		BaseURL: "http://localhost:11434",
	}

	err := client.SetModel(newModel)
	require.NoError(t, err)
	assert.Equal(t, newModel, client.GetModel())

	err = client.SetModel(nil)
	require.Error(t, err)
}

func TestGenerate_PrimaryModel(t *testing.T) {
	// Mock Ollama server
	ollamaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/generate", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var reqBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)
		assert.Equal(t, "llama2", reqBody["model"])

		response := map[string]interface{}{
			"response": "Test response",
			"metrics": map[string]interface{}{
				"tokens":         42,
				"total_duration": 1.5,
			},
			"done": true,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer ollamaServer.Close()

	client := NewClient(
		&Model{
			Type:    LocalOllama,
			Name:    ModelLlama3,
			BaseURL: ollamaServer.URL,
		},
		&Model{},
	)

	ctx := context.Background()
	resp, err := client.Generate(ctx, "test prompt")
	require.NoError(t, err)
	assert.Equal(t, "Test response", resp.Text)
	assert.Equal(t, ModelLlama3, resp.ModelUsed)
	assert.Equal(t, 42, resp.TokenCount)
	assert.Contains(t, resp.Metadata, "total_duration")
}

func TestGenerate_FallbackModel(t *testing.T) {
	// Mock servers
	ollamaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ollamaServer.Close()

	deepseekServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		response := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"content": "Fallback response",
					},
				},
			},
			"usage": map[string]interface{}{
				"total_tokens": 24,
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer deepseekServer.Close()

	client := NewClient(
		&Model{
			Type:    LocalOllama,
			Name:    ModelLlama3,
			BaseURL: ollamaServer.URL,
		},
		&Model{
			Type:    DeepSeekAPI,
			Name:    ModelDeepSeekR1,
			BaseURL: deepseekServer.URL,
			APIKey:  "test-key",
		},
	)

	ctx := context.Background()
	resp, err := client.Generate(ctx, "test prompt")
	require.NoError(t, err)
	assert.Equal(t, "Fallback response", resp.Text)
	assert.Equal(t, ModelDeepSeekR1, resp.ModelUsed)
	assert.Equal(t, 24, resp.TokenCount)
}

func TestGenerate_BothModelsFail(t *testing.T) {
	// Mock servers that both fail
	ollamaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ollamaServer.Close()

	deepseekServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer deepseekServer.Close()

	client := NewClient(
		&Model{
			Type:    LocalOllama,
			Name:    ModelLlama3,
			BaseURL: ollamaServer.URL,
		},
		&Model{
			Type:    DeepSeekAPI,
			Name:    ModelDeepSeekR1,
			BaseURL: deepseekServer.URL,
			APIKey:  "test-key",
		},
	)

	ctx := context.Background()
	_, err := client.Generate(ctx, "test prompt")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "both models failed")
}

func TestStream_PrimaryModel(t *testing.T) {
	streamResponses := []string{
		`{"response":"Part 1","done":false,"metrics":{"tokens":10}}`,
		`{"response":"Part 2","done":false,"metrics":{"tokens":20}}`,
		`{"response":"Part 3","done":true,"metrics":{"tokens":30}}`,
	}

	ollamaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/generate", r.URL.Path)
		
		flusher, ok := w.(http.Flusher)
		require.True(t, ok)

		for _, resp := range streamResponses {
			_, err := fmt.Fprintln(w, resp)
			require.NoError(t, err)
			flusher.Flush()
		}
	}))
	defer ollamaServer.Close()

	client := NewClient(
		&Model{
			Type:    LocalOllama,
			Name:    ModelLlama3,
			BaseURL: ollamaServer.URL,
		},
		&Model{},
	)

	ctx := context.Background()
	stream, err := client.Stream(ctx, "test prompt")
	require.NoError(t, err)

	var collectedResponses []string
	for resp := range stream {
		collectedResponses = append(collectedResponses, resp.Text)
	}

	assert.Equal(t, []string{"Part 1", "Part 2", "Part 3"}, collectedResponses)
}

func TestStream_FallbackModel(t *testing.T) {
	ollamaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ollamaServer.Close()

	deepseekResponses := []string{
		`data: {"choices":[{"delta":{"content":"Part 1"}}]}`,
		`data: {"choices":[{"delta":{"content":"Part 2"}}]}`,
		`data: {"choices":[{"delta":{"content":"Part 3"}}]}`,
		`data: [DONE]`,
	}

	deepseekServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		
		flusher, ok := w.(http.Flusher)
		require.True(t, ok)

		for _, resp := range deepseekResponses {
			_, err := fmt.Fprintln(w, resp)
			require.NoError(t, err)
			flusher.Flush()
		}
	}))
	defer deepseekServer.Close()

	client := NewClient(
		&Model{
			Type:    LocalOllama,
			Name:    ModelLlama3,
			BaseURL: ollamaServer.URL,
		},
		&Model{
			Type:    DeepSeekAPI,
			Name:    ModelDeepSeekR1,
			BaseURL: deepseekServer.URL,
			APIKey:  "test-key",
		},
	)

	ctx := context.Background()
	stream, err := client.Stream(ctx, "test prompt")
	require.NoError(t, err)

	var collectedResponses []string
	for resp := range stream {
		collectedResponses = append(collectedResponses, resp.Text)
	}

	assert.Equal(t, []string{"Part 1", "Part 2", "Part 3"}, collectedResponses)
}

func TestStream_BothModelsFail(t *testing.T) {
	ollamaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ollamaServer.Close()

	deepseekServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer deepseekServer.Close()

	client := NewClient(
		&Model{
			Type:    LocalOllama,
			Name:    ModelLlama3,
			BaseURL: ollamaServer.URL,
		},
		&Model{
			Type:    DeepSeekAPI,
			Name:    ModelDeepSeekR1,
			BaseURL: deepseekServer.URL,
			APIKey:  "test-key",
		},
	)

	ctx := context.Background()
	stream, err := client.Stream(ctx, "test prompt")
	require.NoError(t, err)

	var collectedResponses []string
	for resp := range stream {
		collectedResponses = append(collectedResponses, resp.Text)
	}

	assert.Empty(t, collectedResponses)
}

func TestRateLimiting(t *testing.T) {
	ollamaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"response": "Test response",
			"metrics": map[string]interface{}{
				"tokens": 1,
			},
			"done": true,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer ollamaServer.Close()

	client := NewClient(
		&Model{
			Type:    LocalOllama,
			Name:    ModelLlama3,
			BaseURL: ollamaServer.URL,
		},
		&Model{},
	)

	ctx := context.Background()
	start := time.Now()

	// Make multiple requests to test rate limiting
	for i := 0; i < 3; i++ {
		_, err := client.Generate(ctx, "test prompt")
		require.NoError(t, err)
	}

	duration := time.Since(start)
	assert.True(t, duration >= 200*time.Millisecond, "Rate limiting should space out requests")
}

func TestContextCancellation(t *testing.T) {
	primaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer primaryServer.Close()

	fallbackServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer fallbackServer.Close()

	client := NewClient(
		&Model{
			Type:    LocalOllama,
			Name:    ModelLlama3,
			BaseURL: primaryServer.URL,
		},
		&Model{
			Type:    DeepSeekAPI,
			Name:    ModelDeepSeekR1,
			BaseURL: fallbackServer.URL,
			APIKey:  "test-key",
		},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.Generate(ctx, "test prompt")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deadline exceeded")
}

func TestUnsupportedModelType(t *testing.T) {
	client := NewClient(
		&Model{
			Type: ModelType(999), // Invalid type
			Name: "invalid",
			BaseURL: "http://localhost:11434",
		},
		&Model{
			Type: ModelType(999), // Invalid type
			Name: "invalid",
			BaseURL: "http://localhost:11434",
		},
	)

	ctx := context.Background()
	_, err := client.Generate(ctx, "test prompt")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported model type")
}

func TestDeepSeekErrorResponse(t *testing.T) {
	ollamaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ollamaServer.Close()

	deepseekServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{}, // Empty choices
		})
	}))
	defer deepseekServer.Close()

	client := NewClient(
		&Model{
			Type:    LocalOllama,
			Name:    ModelLlama3,
			BaseURL: ollamaServer.URL,
		},
		&Model{
			Type:    DeepSeekAPI,
			Name:    ModelDeepSeekR1,
			BaseURL: deepseekServer.URL,
			APIKey:  "test-key",
		},
	)

	ctx := context.Background()
	_, err := client.Generate(ctx, "test prompt")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no response choices available")
}

func TestInvalidJSON(t *testing.T) {
	ollamaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ollamaServer.Close()

	deepseekServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json"))
	}))
	defer deepseekServer.Close()

	client := NewClient(
		&Model{
			Type:    LocalOllama,
			Name:    ModelLlama3,
			BaseURL: ollamaServer.URL,
		},
		&Model{
			Type:    DeepSeekAPI,
			Name:    ModelDeepSeekR1,
			BaseURL: deepseekServer.URL,
			APIKey:  "test-key",
		},
	)

	ctx := context.Background()
	_, err := client.Generate(ctx, "test prompt")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid character")
}

func TestModelValidation(t *testing.T) {
	// Set up mock servers
	ollamaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"response": "Test response",
			"metrics": map[string]interface{}{
				"tokens": 1,
			},
			"done": true,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer ollamaServer.Close()

	deepseekServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"content": "Test response",
					},
				},
			},
			"usage": map[string]interface{}{
				"total_tokens": 1,
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer deepseekServer.Close()

	tests := []struct {
		name        string
		model       *Model
		expectError bool
	}{
		{
			name: "Valid Ollama Model",
			model: &Model{
				Type:    LocalOllama,
				Name:    ModelLlama3,
				BaseURL: ollamaServer.URL,
			},
			expectError: false,
		},
		{
			name: "Valid DeepSeek Model",
			model: &Model{
				Type:    DeepSeekAPI,
				Name:    ModelDeepSeekR1,
				BaseURL: deepseekServer.URL,
				APIKey:  "test-key",
			},
			expectError: false,
		},
		{
			name: "Missing BaseURL",
			model: &Model{
				Type: LocalOllama,
				Name: ModelLlama3,
			},
			expectError: true,
		},
		{
			name: "Missing API Key for DeepSeek",
			model: &Model{
				Type:    DeepSeekAPI,
				Name:    ModelDeepSeekR1,
				BaseURL: deepseekServer.URL,
			},
			expectError: true,
		},
		{
			name: "Invalid Model Type",
			model: &Model{
				Type:    ModelType(999),
				Name:    "invalid",
				BaseURL: ollamaServer.URL,
			},
			expectError: true,
		},
		{
			name: "Empty Model Name",
			model: &Model{
				Type:    LocalOllama,
				BaseURL: ollamaServer.URL,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.model, tt.model) // Use same model for primary and fallback
			_, err := client.Generate(context.Background(), "test")
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStreamContextCancellation(t *testing.T) {
	primaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer primaryServer.Close()

	fallbackServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer fallbackServer.Close()

	client := NewClient(
		&Model{
			Type:    LocalOllama,
			Name:    ModelLlama3,
			BaseURL: primaryServer.URL,
		},
		&Model{
			Type:    DeepSeekAPI,
			Name:    ModelDeepSeekR1,
			BaseURL: fallbackServer.URL,
			APIKey:  "test-key",
		},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	stream, err := client.Stream(ctx, "test prompt")
	require.NoError(t, err)

	var collectedResponses []string
	for resp := range stream {
		collectedResponses = append(collectedResponses, resp.Text)
	}

	assert.Empty(t, collectedResponses)
}

func TestStreamRateLimiting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		require.True(t, ok)

		response := `{"response":"Test","done":true,"metrics":{"tokens":1}}`
		_, err := fmt.Fprintln(w, response)
		require.NoError(t, err)
		flusher.Flush()
	}))
	defer server.Close()

	client := NewClient(
		&Model{
			Type:    LocalOllama,
			Name:    ModelLlama3,
			BaseURL: server.URL,
		},
		&Model{},
	)

	ctx := context.Background()
	start := time.Now()

	// Make multiple stream requests to test rate limiting
	for i := 0; i < 3; i++ {
		stream, err := client.Stream(ctx, "test prompt")
		require.NoError(t, err)
		for range stream {
			// Consume stream
		}
	}

	duration := time.Since(start)
	assert.True(t, duration >= 200*time.Millisecond, "Rate limiting should space out requests")
}

func TestStreamOllamaOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)

		// Verify all options are set
		assert.Equal(t, float64(4096), reqBody["context"])
		assert.Equal(t, "json", reqBody["format"])
		assert.Equal(t, "test-template", reqBody["template"])
		assert.Equal(t, "test-system", reqBody["system"])
		assert.Equal(t, map[string]interface{}{
			"temperature": 0.7,
			"top_p":      0.9,
		}, reqBody["options"])

		flusher, ok := w.(http.Flusher)
		require.True(t, ok)

		response := `{"response":"Test","done":true,"metrics":{"tokens":1,"total_duration":0.5}}`
		_, err = fmt.Fprintln(w, response)
		require.NoError(t, err)
		flusher.Flush()
	}))
	defer server.Close()

	client := NewClient(
		&Model{
			Type:     LocalOllama,
			Name:     ModelLlama3,
			BaseURL:  server.URL,
			Context:  4096,
			Format:   "json",
			Template: "test-template",
			System:   "test-system",
			Options: map[string]any{
				"temperature": 0.7,
				"top_p":      0.9,
			},
		},
		&Model{},
	)

	ctx := context.Background()
	stream, err := client.Stream(ctx, "test prompt")
	require.NoError(t, err)

	var collectedResponses []string
	for resp := range stream {
		collectedResponses = append(collectedResponses, resp.Text)
		assert.Equal(t, ModelLlama3, resp.ModelUsed)
		assert.Contains(t, resp.Metadata, "total_duration")
	}

	assert.Equal(t, []string{"Test"}, collectedResponses)
}

func TestStreamDeepSeekInvalidResponse(t *testing.T) {
	ollamaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ollamaServer.Close()

	deepseekServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		require.True(t, ok)

		// Send invalid responses
		responses := []string{
			"data: invalid json",
			"not a data prefix",
			"data: {\"choices\":[{\"delta\":{}}]}", // Empty content
			"data: [DONE]",
		}

		for _, resp := range responses {
			_, err := fmt.Fprintln(w, resp)
			require.NoError(t, err)
			flusher.Flush()
		}
	}))
	defer deepseekServer.Close()

	client := NewClient(
		&Model{
			Type:    LocalOllama,
			Name:    ModelLlama3,
			BaseURL: ollamaServer.URL,
		},
		&Model{
			Type:    DeepSeekAPI,
			Name:    ModelDeepSeekR1,
			BaseURL: deepseekServer.URL,
			APIKey:  "test-key",
		},
	)

	ctx := context.Background()
	stream, err := client.Stream(ctx, "test prompt")
	require.NoError(t, err)

	var collectedResponses []string
	for resp := range stream {
		collectedResponses = append(collectedResponses, resp.Text)
	}

	assert.Empty(t, collectedResponses)
}

func TestStreamOllamaReadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Close connection immediately to simulate read error
		panic("force connection close")
	}))
	defer server.Close()

	client := NewClient(
		&Model{
			Type:    LocalOllama,
			Name:    ModelLlama3,
			BaseURL: server.URL,
		},
		&Model{},
	)

	ctx := context.Background()
	stream, err := client.Stream(ctx, "test prompt")
	require.NoError(t, err)

	var collectedResponses []string
	for resp := range stream {
		collectedResponses = append(collectedResponses, resp.Text)
	}

	assert.Empty(t, collectedResponses)
}

func TestGenerateOllamaOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)

		// Verify all options are set
		assert.Equal(t, float64(4096), reqBody["context"])
		assert.Equal(t, "json", reqBody["format"])
		assert.Equal(t, "test-template", reqBody["template"])
		assert.Equal(t, "test-system", reqBody["system"])
		assert.Equal(t, map[string]interface{}{
			"temperature": 0.7,
			"top_p":      0.9,
		}, reqBody["options"])

		response := map[string]interface{}{
			"response": "Test response",
			"metrics": map[string]interface{}{
				"tokens":         42,
				"total_duration": 1.5,
			},
			"done": true,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(
		&Model{
			Type:     LocalOllama,
			Name:     ModelLlama3,
			BaseURL:  server.URL,
			Context:  4096,
			Format:   "json",
			Template: "test-template",
			System:   "test-system",
			Options: map[string]any{
				"temperature": 0.7,
				"top_p":      0.9,
			},
		},
		&Model{},
	)

	ctx := context.Background()
	resp, err := client.Generate(ctx, "test prompt")
	require.NoError(t, err)
	assert.Equal(t, "Test response", resp.Text)
	assert.Equal(t, ModelLlama3, resp.ModelUsed)
	assert.Equal(t, 42, resp.TokenCount)
}

func TestGenerateRequestErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"response": "Test response",
			"metrics": map[string]interface{}{
				"tokens": 1,
			},
			"done": true,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Test marshal error
	defaultClient := &DefaultClient{
		primaryModel: &Model{
			Type:    LocalOllama,
			Name:    ModelLlama3,
			BaseURL: server.URL,
			Options: map[string]any{
				"invalid": make(chan int), // Will cause marshal error
			},
		},
		httpClient:   &http.Client{},
		ollamaClient: &http.Client{},
		rateLimiter:  rate.NewLimiter(rate.Inf, 1),
	}

	ctx := context.Background()
	_, err := defaultClient.generateOllama(ctx, defaultClient.primaryModel, "test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "marshal request error")

	// Test create request error
	defaultClient = &DefaultClient{
		primaryModel: &Model{
			Type:    LocalOllama,
			Name:    ModelLlama3,
			BaseURL: string([]byte{0x7f}), // Invalid URL
		},
		httpClient:   &http.Client{},
		ollamaClient: &http.Client{},
		rateLimiter:  rate.NewLimiter(rate.Inf, 1),
	}

	_, err = defaultClient.generateOllama(ctx, defaultClient.primaryModel, "test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create request error")
}

func TestStreamInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		require.True(t, ok)

		_, err := fmt.Fprintln(w, "invalid json")
		require.NoError(t, err)
		flusher.Flush()
	}))
	defer server.Close()

	client := NewClient(
		&Model{
			Type:    LocalOllama,
			Name:    ModelLlama3,
			BaseURL: server.URL,
		},
		&Model{},
	)

	ctx := context.Background()
	stream, err := client.Stream(ctx, "test prompt")
	require.NoError(t, err)

	var collectedResponses []string
	for resp := range stream {
		collectedResponses = append(collectedResponses, resp.Text)
	}

	assert.Empty(t, collectedResponses)
}
