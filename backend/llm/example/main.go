package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/leonzhao/gosol/backend/llm"
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
			"top_p":       0.9,
		},
	}

	// Create fallback model (DeepSeek API)
	fallbackModel := &llm.Model{
		Type:    llm.DeepSeekAPI,
		Name:    "deepseek-coder-33b-instruct",
		BaseURL: "https://api.deepseek.com",
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
	}

	// Create LLM client
	client := llm.NewClient(primaryModel, fallbackModel)

	// Test Generate with code generation prompt
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	prompt := `Write a Go function to implement a concurrent worker pool with the following requirements:
1. Accept number of workers and a job queue channel
2. Each worker processes jobs concurrently
3. Support graceful shutdown
4. Return results through a result channel`

	resp, err := client.Generate(ctx, prompt)
	if err != nil {
		log.Fatalf("Generate error: %v", err)
	}

	fmt.Printf("Generated response from %s:\n", resp.ModelUsed)
	fmt.Printf("Response:\n%s\n", resp.Text)
	fmt.Printf("\nMetrics:\n")
	for k, v := range resp.Metadata {
		fmt.Printf("- %s: %s\n", k, v)
	}

	// Test Stream with code review prompt
	fmt.Println("\nTesting streaming response...")
	prompt = `Review the following code and suggest improvements:

func process(data []int) []int {
    result := make([]int, len(data))
    for i := 0; i < len(data); i++ {
        result[i] = data[i] * 2
    }
    return result
}`

	streamCh, err := client.Stream(ctx, prompt)
	if err != nil {
		log.Fatalf("Stream error: %v", err)
	}

	fmt.Print("Streaming response: ")
	for resp := range streamCh {
		fmt.Print(resp.Text)
	}
	fmt.Println()
}
