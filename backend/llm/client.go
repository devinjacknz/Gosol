package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"gosol/backend/monitoring"

	"golang.org/x/time/rate"
)

// ModelType represents the type of LLM model
type ModelType int

const (
	LocalOllama ModelType = iota
	DeepSeekAPI
)

// Common model names
const (
	// Llama models
	ModelLlama3     = "llama2"
	ModelDeepSeekR1 = "deepseek-coder:1.5b"
	ModelPhi4       = "phi:latest"
	ModelGemma2     = "gemma:2b"
	ModelCodeLlama  = "codellama:7b"
	ModelMistral    = "mistral"
	ModelLlava      = "llava"
)

// Model represents an LLM model configuration
type Model struct {
	Type    ModelType
	Name    string
	BaseURL string
	APIKey  string
	// Additional Ollama specific options
	Context  int            `json:"context,omitempty"`  // Sets the size of the context window
	Format   string         `json:"format,omitempty"`   // Sets the format of the response (json)
	Template string         `json:"template,omitempty"` // Sets the prompt template
	System   string         `json:"system,omitempty"`   // Sets the system message
	Options  map[string]any `json:"options,omitempty"`  // Additional model parameters
}

// Response represents the LLM response
type Response struct {
	Text       string            `json:"text"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	ModelUsed  string            `json:"model_used"`
	TokenCount int               `json:"token_count"`
}

// Client defines the interface for LLM interactions
type Client interface {
	// Generate generates text from a prompt
	Generate(ctx context.Context, prompt string) (*Response, error)
	// Stream generates text from a prompt with streaming response
	Stream(ctx context.Context, prompt string) (<-chan *Response, error)
	// GetModel returns the current model configuration
	GetModel() *Model
	// SetModel sets the model configuration
	SetModel(model *Model) error
}

// DefaultClient implements the Client interface
type DefaultClient struct {
	primaryModel  *Model
	fallbackModel *Model
	httpClient    *http.Client
	ollamaClient  *http.Client
	retryCount    int
	retryDelay    time.Duration
	maxConcurrent int
	rateLimiter   *rate.Limiter
	mu            sync.RWMutex
}

// NewClient creates a new LLM client
func NewClient(primaryModel, fallbackModel *Model) Client {
	return &DefaultClient{
		primaryModel:  primaryModel,
		fallbackModel: fallbackModel,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		ollamaClient:  &http.Client{Timeout: 60 * time.Second},
		retryCount:    3,
		retryDelay:    time.Second,
		maxConcurrent: 5,
		rateLimiter:   rate.NewLimiter(rate.Limit(10), 1), // 10 requests per second
	}
}

// Generate implements text generation
func (c *DefaultClient) Generate(ctx context.Context, prompt string) (*Response, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	start := time.Now()
	// Try primary model first
	resp, err := c.generateWithModel(ctx, c.primaryModel, prompt)
	if err != nil {
		// Log primary model failure
		log.Printf("Primary model failed: %v, falling back to secondary model", err)
		monitoring.RecordLLMFallback()

		// Try fallback model
		resp, err = c.generateWithModel(ctx, c.fallbackModel, prompt)
		if err != nil {
			monitoring.RecordLLMRequest(c.fallbackModel.Name, "generate", time.Since(start), "error", 0)
			return nil, fmt.Errorf("both models failed: %w", err)
		}
	}

	monitoring.RecordLLMRequest(resp.ModelUsed, "generate", time.Since(start), "success", resp.TokenCount)
	return resp, nil
}

// Stream implements streaming text generation
func (c *DefaultClient) Stream(ctx context.Context, prompt string) (<-chan *Response, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	responseChan := make(chan *Response)
	start := time.Now()

	go func() {
		defer close(responseChan)
		totalTokens := 0

		// Try primary model first
		err := c.streamWithModel(ctx, c.primaryModel, prompt, responseChan)
		if err != nil {
			// Log primary model failure
			log.Printf("Primary model stream failed: %v, falling back to secondary model", err)
			monitoring.RecordLLMFallback()

			// Try fallback model
			if err := c.streamWithModel(ctx, c.fallbackModel, prompt, responseChan); err != nil {
				log.Printf("Both models failed for streaming: %v", err)
				monitoring.RecordLLMRequest(c.fallbackModel.Name, "stream", time.Since(start), "error", 0)
				return
			}
		}

		monitoring.RecordLLMRequest(c.primaryModel.Name, "stream", time.Since(start), "success", totalTokens)
	}()

	return responseChan, nil
}

func (c *DefaultClient) generateWithModel(ctx context.Context, model *Model, prompt string) (*Response, error) {
	switch model.Type {
	case LocalOllama:
		return c.generateOllama(ctx, model, prompt)
	case DeepSeekAPI:
		return c.generateDeepSeek(ctx, model, prompt)
	default:
		return nil, fmt.Errorf("unsupported model type: %v", model.Type)
	}
}

func (c *DefaultClient) streamWithModel(ctx context.Context, model *Model, prompt string, responseChan chan<- *Response) error {
	switch model.Type {
	case LocalOllama:
		return c.streamOllama(ctx, model, prompt, responseChan)
	case DeepSeekAPI:
		return c.streamDeepSeek(ctx, model, prompt, responseChan)
	default:
		return fmt.Errorf("unsupported model type: %v", model.Type)
	}
}

func (c *DefaultClient) generateOllama(ctx context.Context, model *Model, prompt string) (*Response, error) {
	reqBody := map[string]interface{}{
		"model":  model.Name,
		"prompt": prompt,
	}

	// Add optional parameters if set
	if model.Context > 0 {
		reqBody["context"] = model.Context
	}
	if model.Format != "" {
		reqBody["format"] = model.Format
	}
	if model.Template != "" {
		reqBody["template"] = model.Template
	}
	if model.System != "" {
		reqBody["system"] = model.System
	}
	if len(model.Options) > 0 {
		reqBody["options"] = model.Options
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", model.BaseURL+"/api/generate", bytes.NewReader(reqData))
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.ollamaClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var ollamaResp struct {
		Response string `json:"response"`
		Metrics  struct {
			TokenCount         int     `json:"tokens"`
			TotalDuration      float64 `json:"total_duration"`
			LoadDuration       float64 `json:"load_duration"`
			PromptEvalCount    int     `json:"prompt_eval_count"`
			PromptEvalDuration float64 `json:"prompt_eval_duration"`
			EvalCount          int     `json:"eval_count"`
			EvalDuration       float64 `json:"eval_duration"`
		} `json:"metrics"`
		Context []int `json:"context,omitempty"`
		Done    bool  `json:"done"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return &Response{
		Text:       ollamaResp.Response,
		ModelUsed:  model.Name,
		TokenCount: ollamaResp.Metrics.TokenCount,
		Metadata: map[string]string{
			"total_duration": fmt.Sprintf("%.2fs", ollamaResp.Metrics.TotalDuration),
			"load_duration":  fmt.Sprintf("%.2fs", ollamaResp.Metrics.LoadDuration),
			"eval_count":     fmt.Sprintf("%d", ollamaResp.Metrics.EvalCount),
			"eval_duration":  fmt.Sprintf("%.2fs", ollamaResp.Metrics.EvalDuration),
		},
	}, nil
}

func (c *DefaultClient) generateDeepSeek(ctx context.Context, model *Model, prompt string) (*Response, error) {
	reqBody := map[string]interface{}{
		"model":    model.Name,
		"messages": []map[string]string{{"role": "user", "content": prompt}},
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", model.BaseURL+"/v1/chat/completions", bytes.NewReader(reqData))
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+model.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var deepseekResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&deepseekResp); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	if len(deepseekResp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices available")
	}

	return &Response{
		Text:       deepseekResp.Choices[0].Message.Content,
		ModelUsed:  model.Name,
		TokenCount: deepseekResp.Usage.TotalTokens,
	}, nil
}

func (c *DefaultClient) streamOllama(ctx context.Context, model *Model, prompt string, responseChan chan<- *Response) error {
	reqBody := map[string]interface{}{
		"model":  model.Name,
		"prompt": prompt,
		"stream": true,
	}

	// Add optional parameters if set
	if model.Context > 0 {
		reqBody["context"] = model.Context
	}
	if model.Format != "" {
		reqBody["format"] = model.Format
	}
	if model.Template != "" {
		reqBody["template"] = model.Template
	}
	if model.System != "" {
		reqBody["system"] = model.System
	}
	if len(model.Options) > 0 {
		reqBody["options"] = model.Options
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", model.BaseURL+"/api/generate", bytes.NewReader(reqData))
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.ollamaClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read response error: %w", err)
		}

		var streamResp struct {
			Response string `json:"response"`
			Done     bool   `json:"done"`
			Context  []int  `json:"context,omitempty"`
			Metrics  struct {
				TokenCount    int     `json:"tokens"`
				TotalDuration float64 `json:"total_duration"`
			} `json:"metrics,omitempty"`
		}

		if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
			continue
		}

		responseChan <- &Response{
			Text:       streamResp.Response,
			ModelUsed:  model.Name,
			TokenCount: streamResp.Metrics.TokenCount,
			Metadata: map[string]string{
				"total_duration": fmt.Sprintf("%.2fs", streamResp.Metrics.TotalDuration),
			},
		}

		if streamResp.Done {
			break
		}
	}

	return nil
}

func (c *DefaultClient) streamDeepSeek(ctx context.Context, model *Model, prompt string, responseChan chan<- *Response) error {
	reqBody := map[string]interface{}{
		"model":    model.Name,
		"messages": []map[string]string{{"role": "user", "content": prompt}},
		"stream":   true,
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", model.BaseURL+"/v1/chat/completions", bytes.NewReader(reqData))
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+model.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read response error: %w", err)
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		line = strings.TrimPrefix(line, "data: ")

		if line == "[DONE]" {
			break
		}

		var streamResp struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}

		if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
			continue
		}

		if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta.Content != "" {
			responseChan <- &Response{
				Text:      streamResp.Choices[0].Delta.Content,
				ModelUsed: model.Name,
			}
		}
	}

	return nil
}

// GetModel returns the current model configuration
func (c *DefaultClient) GetModel() *Model {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.primaryModel
}

// SetModel sets the model configuration
func (c *DefaultClient) SetModel(model *Model) error {
	if model == nil {
		return fmt.Errorf("model cannot be nil")
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.primaryModel = model
	return nil
}
