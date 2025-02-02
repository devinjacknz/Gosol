package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLLMClient(t *testing.T) {
	// Mock servers
	ollamaSrv := mockOllamaServer()
	defer ollamaSrv.Close()

	deepseekSrv := mockDeepSeekServer()
	defer deepseekSrv.Close()

	// Test cases
	tests := []struct {
		name          string
		primaryModel  *Model
		fallbackModel *Model
		prompt        string
		wantErr       bool
		wantContains  string
	}{
		{
			name: "Ollama Success",
			primaryModel: &Model{
				Type:    LocalOllama,
				Name:    "deepseek-coder:1.5b",
				BaseURL: ollamaSrv.URL,
			},
			prompt:       "Write a Go function",
			wantErr:      false,
			wantContains: "func",
		},
		{
			name: "DeepSeek Success",
			primaryModel: &Model{
				Type:    DeepSeekAPI,
				Name:    "deepseek-coder-33b-instruct",
				BaseURL: deepseekSrv.URL,
				APIKey:  "test-key",
			},
			prompt:       "Write a Python function",
			wantErr:      false,
			wantContains: "def",
		},
		{
			name: "Fallback Success",
			primaryModel: &Model{
				Type:    LocalOllama,
				Name:    "invalid",
				BaseURL: "http://invalid",
			},
			fallbackModel: &Model{
				Type:    DeepSeekAPI,
				Name:    "deepseek-coder-33b-instruct",
				BaseURL: deepseekSrv.URL,
				APIKey:  "test-key",
			},
			prompt:       "Write code",
			wantErr:      false,
			wantContains: "def",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create client
			client := NewClient(tt.primaryModel, tt.fallbackModel)

			// Test Generate
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resp, err := client.Generate(ctx, tt.prompt)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Contains(t, resp.Text, tt.wantContains)
			assert.NotEmpty(t, resp.ModelUsed)

			// Test Stream
			streamCh, err := client.Stream(ctx, tt.prompt)
			require.NoError(t, err)

			var fullResponse string
			for resp := range streamCh {
				fullResponse += resp.Text
			}

			assert.Contains(t, fullResponse, tt.wantContains)
		})
	}
}

func mockOllamaServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		response := `{"response": "func add(a, b int) int {\n    return a + b\n}", "done": true, "metrics":{"tokens":20}}`
		if r.Header.Get("Content-Type") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(response))
		}
	}))
}

func mockDeepSeekServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		response := `{
			"choices": [
				{
					"message": {
						"content": "def add(a: int, b: int) -> int:\n    return a + b"
					}
				}
			],
			"usage": {
				"total_tokens": 25
			}
		}`

		if r.Header.Get("Content-Type") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(response))
		}
	}))
}
