# Configuration Guide

## Environment Variables

### Environment Settings
- `ENVIRONMENT`: Runtime environment (development/staging/production)
- `LOG_LEVEL`: Logging level (debug/info/warn/error)

### LLM Configuration
- `OLLAMA_BASE_URL`: Ollama API base URL
- `DEEPSEEK_API_KEY`: DeepSeek API key
- `DEEPSEEK_BASE_URL`: DeepSeek API base URL

### Model Configuration
- `DEFAULT_MODEL`: Primary model name
- `FALLBACK_MODEL`: Fallback model name
- `CONTEXT_SIZE`: Model context window size
- `RESPONSE_FORMAT`: Response format (text/json)

### Performance Settings
- `MAX_CONCURRENT_REQUESTS`: Maximum concurrent requests
- `REQUESTS_PER_SECOND`: Rate limit for requests
- `HTTP_TIMEOUT_SECONDS`: HTTP client timeout
- `STREAM_TIMEOUT_SECONDS`: Streaming response timeout

### Monitoring Configuration
- `ENABLE_METRICS`: Enable Prometheus metrics
- `METRICS_PORT`: Metrics server port
- `METRICS_PATH`: Metrics endpoint path

### Security Settings
- `ENABLE_RATE_LIMIT`: Enable rate limiting
- `ENABLE_REQUEST_VALIDATION`: Enable request validation
- `MAX_REQUEST_SIZE_MB`: Maximum request size

## Model Configuration

### Ollama Models

1. DeepSeek Coder
```yaml
name: deepseek-coder:1.5b
context: 4096
format: json
system: "You are a helpful coding assistant."
options:
  temperature: 0.7
  top_p: 0.9
  top_k: 40
  repeat_penalty: 1.1
```

2. Llama2
```yaml
name: llama2
context: 4096
format: json
system: "You are a helpful assistant."
options:
  temperature: 0.8
  top_p: 0.9
```

3. Phi-4
```yaml
name: phi:latest
context: 2048
format: json
options:
  temperature: 0.7
```

4. Gemma2
```yaml
name: gemma:2b
context: 8192
format: json
options:
  temperature: 0.7
```

### DeepSeek API

```yaml
name: deepseek-coder-33b-instruct
format: json
options:
  temperature: 0.7
  top_p: 0.9
  max_tokens: 2048
```

## Performance Tuning

### Rate Limiting
- Default: 10 requests per second
- Adjust based on:
  - Model response time
  - System resources
  - API limits

### Concurrent Requests
- Default: 5 concurrent requests
- Consider:
  - Memory usage
  - CPU utilization
  - Network capacity

### Timeouts
- HTTP timeout: 30 seconds
- Stream timeout: 60 seconds
- Adjust for:
  - Model complexity
  - Response length
  - Network conditions

## Monitoring Setup

### Metrics
- Request count and latency
- Token usage
- Error rates
- Model performance
- Resource utilization

### Prometheus Configuration
```yaml
scrape_configs:
  - job_name: 'llm-metrics'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:2112']
```

### Grafana Dashboard
- Trading metrics
- LLM performance
- System resources
- Error tracking

## Security Best Practices

1. API Keys
   - Use environment variables
   - Rotate regularly
   - Restrict permissions

2. Rate Limiting
   - Enable by default
   - Configure per client
   - Monitor abuse

3. Request Validation
   - Validate input size
   - Check content type
   - Sanitize prompts

4. Error Handling
   - Don't expose internals
   - Log securely
   - Fail gracefully 