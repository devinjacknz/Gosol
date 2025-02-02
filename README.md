# GoSol Trading System

A high-performance trading system with LLM-powered analysis capabilities.

## Features

- **LLM Integration**
  - Local Ollama support (Llama3, DeepSeek-R1, Phi-4, Gemma2)
  - DeepSeek API fallback
  - Streaming responses
  - Performance monitoring

- **Trading Features**
  - Market data analysis
  - Position management
  - Risk controls
  - Performance analytics

## Quick Start

1. Prerequisites:
```bash
# Install Ollama
curl https://ollama.ai/install.sh | sh

# Pull required models
ollama pull deepseek-coder:1.5b
ollama pull llama2
ollama pull phi:latest
ollama pull gemma:2b
```

2. Environment Setup:
```bash
# Copy example env file
cp .env.example .env

# Edit .env file with your configuration
vim .env
```

3. Build and Run:
```bash
# Build
go build -o gosol ./cmd/gosol

# Run
./gosol
```

## Project Structure

```
.
├── backend/
│   ├── llm/           # LLM integration
│   ├── monitoring/    # Metrics and monitoring
│   └── trading/       # Trading logic
├── cmd/
│   └── gosol/         # Main application
├── config/            # Configuration files
├── docs/             # Documentation
└── scripts/          # Utility scripts
```

## Configuration

See [Configuration Guide](docs/configuration.md) for detailed settings.

## Development

1. Install dependencies:
```bash
go mod tidy
```

2. Run tests:
```bash
go test ./...
```

3. Run linter:
```bash
golangci-lint run
```

## Documentation

- [API Documentation](docs/api.md)
- [LLM Integration Guide](docs/llm.md)
- [Trading System Guide](docs/trading.md)
- [Monitoring Guide](docs/monitoring.md)

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details
