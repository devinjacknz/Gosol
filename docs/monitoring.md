# Monitoring Guide

## Overview

The monitoring system provides comprehensive metrics for:
- LLM performance and usage
- System health and resources
- Error tracking and alerts

## Metrics

### LLM Metrics

1. Request Metrics
```go
// Request count by model and operation
llm_request_total{model="deepseek-coder",operation="generate"}

// Request duration
llm_request_duration_seconds{model="deepseek-coder",operation="generate"}

// Token count
llm_tokens_total{model="deepseek-coder"}

// Fallback count
llm_fallback_total
```

2. Performance Metrics
```go
// Model load time
llm_model_load_duration_seconds{model="deepseek-coder"}

// Prompt evaluation time
llm_prompt_eval_duration_seconds{model="deepseek-coder"}

// Total processing time
llm_total_duration_seconds{model="deepseek-coder"}
```

3. Error Metrics
```go
// Error count by type
llm_error_total{model="deepseek-coder",type="timeout"}

// Rate limit exceeded
llm_rate_limit_exceeded_total{model="deepseek-coder"}
```

## Prometheus Configuration

### Scrape Config
```yaml
scrape_configs:
  - job_name: 'llm-metrics'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:2112']
    metrics_path: '/metrics'
```

### Recording Rules
```yaml
groups:
  - name: llm_rules
    rules:
      - record: llm:request_rate_5m
        expr: rate(llm_request_total[5m])
      
      - record: llm:error_rate_5m
        expr: rate(llm_error_total[5m])
      
      - record: llm:token_rate_5m
        expr: rate(llm_tokens_total[5m])
```

## Grafana Dashboards

### LLM Performance Dashboard
```json
{
  "title": "LLM Performance",
  "panels": [
    {
      "title": "Request Rate",
      "type": "graph",
      "targets": [
        {
          "expr": "rate(llm_request_total[5m])",
          "legendFormat": "{{model}}"
        }
      ]
    },
    {
      "title": "Response Time",
      "type": "heatmap",
      "targets": [
        {
          "expr": "rate(llm_request_duration_seconds_bucket[5m])",
          "legendFormat": "{{le}}"
        }
      ]
    },
    {
      "title": "Token Usage",
      "type": "graph",
      "targets": [
        {
          "expr": "rate(llm_tokens_total[5m])",
          "legendFormat": "{{model}}"
        }
      ]
    }
  ]
}
```

### Error Tracking Dashboard
```json
{
  "title": "LLM Errors",
  "panels": [
    {
      "title": "Error Rate",
      "type": "graph",
      "targets": [
        {
          "expr": "rate(llm_error_total[5m])",
          "legendFormat": "{{type}}"
        }
      ]
    },
    {
      "title": "Fallback Rate",
      "type": "graph",
      "targets": [
        {
          "expr": "rate(llm_fallback_total[5m])"
        }
      ]
    }
  ]
}
```

## Alerting

### Alert Rules
```yaml
groups:
  - name: llm_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(llm_error_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High LLM error rate"
          
      - alert: HighLatency
        expr: histogram_quantile(0.95, rate(llm_request_duration_seconds_bucket[5m])) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High LLM latency"
```

## Best Practices

1. Metric Collection
   - Use appropriate labels
   - Keep cardinality under control
   - Monitor resource usage

2. Dashboard Design
   - Group related metrics
   - Use appropriate visualizations
   - Add helpful descriptions

3. Alerting
   - Set meaningful thresholds
   - Avoid alert fatigue
   - Include actionable information

4. Performance
   - Monitor scrape duration
   - Use recording rules
   - Optimize queries 