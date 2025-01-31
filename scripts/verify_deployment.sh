#!/bin/bash
set -e

function test_environment() {
    local env=$1
    echo "Starting $env environment verification..."
    
    # Build and start services
    docker-compose -f docker-compose.$env.yml build
    docker-compose -f docker-compose.$env.yml up -d
    
    # Wait for services to be healthy
    echo "Waiting for services to be healthy..."
    sleep 30
    
    # Test ML service health
    echo "Testing ML service health..."
    curl -f http://localhost:8000/api/v1/health
    
    # Test market data endpoints with multiple symbols
    echo "Testing market data endpoints..."
    symbols=("BTC-USD" "ETH-USD" "SOL-USD" "AVAX-USD")
    for symbol in "${symbols[@]}"; do
        echo "Testing $symbol..."
        curl -f "http://localhost:8000/api/v1/market-data/$symbol"
    done
    
    # Run parallel load test
    echo "Running load test..."
    for i in {1..100}; do
        symbol=${symbols[$((RANDOM % ${#symbols[@]}))]}
        curl -s -f "http://localhost:8000/api/v1/market-data/$symbol" &
        if [ $((i % 20)) -eq 0 ]; then
            echo "Completed $i requests"
            wait
        fi
    done
    
    wait
    
    # Test prediction endpoint
    echo "Testing prediction endpoint..."
    curl -X POST -H "Content-Type: application/json" -d '{
        "token_address": "BTC-USD",
        "price_history": [50000,51000,52000,51500,51200],
        "volume_history": [1000,1200,1100,900,950],
        "timestamp": '$(date +%s)',
        "market_cap": 1000000000000,
        "holders": 1000000
    }' http://localhost:8000/api/v1/predict
    
    # Test WebSocket connection
    echo "Testing WebSocket connection..."
    timeout 10s websocat "ws://localhost:8080/ws" || true
    
    echo "$env environment verification completed"
}

# Make script executable
chmod +x "$0"

# Test environments
if [ "$1" = "test" ]; then
    test_environment "test"
elif [ "$1" = "staging" ]; then
    test_environment "staging"
else
    echo "Usage: $0 [test|staging]"
    exit 1
fi
