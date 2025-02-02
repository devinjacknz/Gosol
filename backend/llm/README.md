# LLM Client

A Go client library for interacting with LLM models, supporting both local Ollama and DeepSeek API.

## Features

- Support for multiple LLM models (Ollama and DeepSeek)
- Built-in support for popular models (Llama2, DeepSeek, Phi, Gemma, CodeLlama, etc.)
- Advanced model configuration options
- Automatic fallback to backup model
- Rate limiting and concurrent request handling
- Streaming response support
- Detailed performance metrics
- JSON response format support
- Context window management
- System prompts and templates

## Installation

```bash
go get -u gosol/backend/llm
```

## Usage

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "gosol/backend/llm"
)

func main() {
    // Create primary model (Ollama)
    primaryModel := &llm.Model{
        Type:    llm.LocalOllama,
        Name:    llm.ModelDeepSeekR1,
        BaseURL: "http://localhost:11434",
        System:  "You are a helpful coding assistant.",
        Format:  "json",
        Context: 4096,
        Options: map[string]any{
            "temperature": 0.7,
            "top_p": 0.9,
        },
    }

    // Create fallback model (DeepSeek API)
    fallbackModel := &llm.Model{
        Type:    llm.DeepSeekAPI,
        Name:    "deepseek-coder-33b-instruct",
        BaseURL: "https://api.deepseek.com",
        APIKey:  "your-api-key",
    }

    // Create LLM client
    client := llm.NewClient(primaryModel, fallbackModel)

    // Generate text
    ctx := context.Background()
    resp, err := client.Generate(ctx, "Write a Go function")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Response: %s\n", resp.Text)
    
    // Print metrics
    for k, v := range resp.Metadata {
        fmt.Printf("%s: %s\n", k, v)
    }
}
```

### Streaming Example

```go
ctx := context.Background()
streamCh, err := client.Stream(ctx, "Write a Python function")
if err != nil {
    log.Fatal(err)
}

for resp := range streamCh {
    fmt.Print(resp.Text)
}
```

## Configuration

### Model Configuration

- `Type`: Model type (LocalOllama or DeepSeekAPI)
- `Name`: Model name/version (see predefined constants)
- `BaseURL`: API base URL
- `APIKey`: API key (for DeepSeek API)
- `Context`: Context window size (default: 4096)
- `Format`: Response format (e.g., "json")
- `Template`: Custom prompt template
- `System`: System message for chat context
- `Options`: Additional model parameters
  - temperature: Controls randomness (0.0 - 1.0)
  - top_p: Controls diversity (0.0 - 1.0)
  - top_k: Controls vocabulary size
  - repeat_penalty: Prevents repetition
  - seed: Random seed for reproducibility

### Predefined Models

```go
const (
    ModelLlama3     = "llama2"
    ModelDeepSeekR1 = "deepseek-coder:1.5b"
    ModelPhi4       = "phi:latest"
    ModelGemma2     = "gemma:2b"
    ModelCodeLlama  = "codellama:7b"
    ModelMistral    = "mistral"
    ModelLlava      = "llava"
)
```

### Client Configuration

- Rate limiting: 10 requests per second by default
- HTTP timeouts: 30s for DeepSeek API, 60s for Ollama
- Retry count: 3 attempts
- Max concurrent requests: 5

## Performance Metrics

The client tracks detailed performance metrics:
- Request count and error rate
- Token count and processing speed
- Model load time and evaluation time
- Total processing duration
- Prompt evaluation metrics
- Context window usage

## Testing

Run tests:
```bash
go test -v ./...
```

## Dependencies

- golang.org/x/time/rate: Rate limiting
- github.com/stretchr/testify: Testing

## Best Practices

1. Model Selection:
   - Use Ollama for local development and testing
   - Use DeepSeek API for production workloads
   - Configure appropriate context window size
   - Set temperature based on task requirements

2. Error Handling:
   - Always check for errors
   - Use fallback models for reliability
   - Monitor error rates and latency

3. Performance:
   - Use streaming for long responses
   - Monitor token usage and costs
   - Adjust rate limits based on needs

4. Security:
   - Never hardcode API keys
   - Use environment variables
   - Validate and sanitize prompts 