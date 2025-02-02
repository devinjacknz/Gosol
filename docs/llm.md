# LLM Integration Guide

## Overview

The LLM integration provides a unified interface for interacting with both local Ollama models and the DeepSeek API, featuring:
- Multiple model support
- Automatic fallback
- Streaming responses
- Performance monitoring

## Supported Models

### Ollama Models
- Llama 3.3 (`llama2`)
- DeepSeek Coder 1.5b (`deepseek-coder:1.5b`)
- Phi-4 (`phi:latest`)
- Gemma 2 (`gemma:2b`)
- CodeLlama (`codellama:7b`)
- Mistral (`mistral`)
- Llava (`llava`)

### DeepSeek API
- DeepSeek Coder 33B (`deepseek-coder-33b-instruct`)

## Installation

1. Install Ollama:
```bash
curl https://ollama.ai/install.sh | sh
```

2. Pull required models:
```bash
ollama pull deepseek-coder:1.5b
ollama pull llama2
ollama pull phi:latest
ollama pull gemma:2b
```

## Usage Examples

### Basic Text Generation

```go
package main

import (
    "context"
    "fmt"
    "log"
    "gosol/backend/llm"
)

func main() {
    // Create client with models
    client := llm.NewClient(
        &llm.Model{
            Type:    llm.LocalOllama,
            Name:    llm.ModelDeepSeekR1,
            BaseURL: "http://localhost:11434",
            System:  "You are a helpful coding assistant.",
        },
        &llm.Model{
            Type:    llm.DeepSeekAPI,
            Name:    "deepseek-coder-33b-instruct",
            BaseURL: "https://api.deepseek.com",
            APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
        },
    )

    // Generate text
    resp, err := client.Generate(context.Background(), "Write a Go function")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Response: %s\n", resp.Text)
}
```

### Streaming Response

```go
streamCh, err := client.Stream(ctx, "Write a Python function")
if err != nil {
    log.Fatal(err)
}

for resp := range streamCh {
    fmt.Print(resp.Text)
}
```

### Advanced Configuration

```go
model := &llm.Model{
    Type:    llm.LocalOllama,
    Name:    llm.ModelDeepSeekR1,
    BaseURL: "http://localhost:11434",
    System:  "You are a helpful coding assistant.",
    Format:  "json",
    Context: 4096,
    Options: map[string]any{
        "temperature": 0.7,
        "top_p": 0.9,
        "top_k": 40,
        "repeat_penalty": 1.1,
    },
}
```

## Performance Optimization

### Context Window
- Default: 4096 tokens
- Adjust based on:
  - Model capabilities
  - Input/output length
  - Memory constraints

### Temperature Settings
- Lower (0.1-0.4): More focused, deterministic responses
- Medium (0.5-0.7): Balanced creativity and precision
- Higher (0.8-1.0): More creative, diverse responses

### Response Format
- Text: Raw text output
- JSON: Structured response format
- Template: Custom response templates

## Error Handling

### Common Errors
1. Rate Limit Exceeded
```go
if err != nil && strings.Contains(err.Error(), "rate limit") {
    time.Sleep(time.Second)
    // Retry request
}
```

2. Context Timeout
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

3. Model Not Found
```go
if err != nil && strings.Contains(err.Error(), "model not found") {
    // Fall back to default model
}
```

## Best Practices

1. Model Selection
   - Use local models for development
   - Use API for production
   - Match model to task complexity

2. Performance
   - Enable rate limiting
   - Use appropriate timeouts
   - Monitor resource usage

3. Error Handling
   - Implement fallback logic
   - Log errors with context
   - Graceful degradation

4. Security
   - Validate inputs
   - Sanitize outputs
   - Secure API keys

## Integration Examples

### Code Generation

```go
func GenerateCode(ctx context.Context, client llm.Client, prompt string) (string, error) {
    model := &llm.Model{
        Type:    llm.LocalOllama,
        Name:    llm.ModelDeepSeekR1,
        System:  "You are an expert Go programmer.",
        Format:  "json",
        Options: map[string]any{
            "temperature": 0.3, // Lower for code generation
        },
    }
    
    resp, err := client.Generate(ctx, prompt)
    if err != nil {
        return "", fmt.Errorf("generate code: %w", err)
    }
    
    return resp.Text, nil
}
```

### Code Review

```go
func ReviewCode(ctx context.Context, client llm.Client, code string) (string, error) {
    prompt := fmt.Sprintf("Review this code and suggest improvements:\n\n%s", code)
    
    model := &llm.Model{
        Type:    llm.LocalOllama,
        Name:    llm.ModelDeepSeekR1,
        System:  "You are an expert code reviewer.",
        Options: map[string]any{
            "temperature": 0.7, // Higher for creative suggestions
        },
    }
    
    resp, err := client.Generate(ctx, prompt)
    if err != nil {
        return "", fmt.Errorf("review code: %w", err)
    }
    
    return resp.Text, nil
}
```

### Trading Analysis

```go
func AnalyzeMarket(ctx context.Context, client llm.Client, data string) (string, error) {
    prompt := fmt.Sprintf("Analyze this market data and provide insights:\n\n%s", data)
    
    model := &llm.Model{
        Type:    llm.LocalOllama,
        Name:    llm.ModelLlama3,
        System:  "You are an expert market analyst.",
        Format:  "json",
        Options: map[string]any{
            "temperature": 0.5,
            "top_k": 50,
        },
    }
    
    resp, err := client.Generate(ctx, prompt)
    if err != nil {
        return "", fmt.Errorf("analyze market: %w", err)
    }
    
    return resp.Text, nil
}
``` 