#!/bin/bash
set -e

# Configuration
SYMBOLS=("BTC-USD" "ETH-USD" "SOL-USD" "AVAX-USD")
CONCURRENT_REQUESTS=50
TOTAL_REQUESTS=1000
ML_SERVICE_URL="http://localhost:8000"
BACKEND_URL="http://localhost:8080"

echo "Starting stress test for trading platform..."

# Test market data endpoints
test_market_data() {
    echo "Testing market data endpoints with $CONCURRENT_REQUESTS concurrent requests..."
    for ((i=1; i<=$TOTAL_REQUESTS; i++)); do
        if [ $((i % CONCURRENT_REQUESTS)) -eq 0 ]; then
            echo "Completed $i requests"
            wait
        fi
        
        symbol=${SYMBOLS[$((RANDOM % ${#SYMBOLS[@]}))]}
        curl -s -f "$ML_SERVICE_URL/api/v1/market-data/$symbol" > /dev/null &
    done
    wait
    echo "Market data endpoint test completed"
}

# Test prediction endpoints
test_predictions() {
    echo "Testing prediction endpoints..."
    for symbol in "${SYMBOLS[@]}"; do
        echo "Testing predictions for $symbol"
        curl -s -X POST \
             -H "Content-Type: application/json" \
             -d "{
                 \"token_address\": \"$symbol\",
                 \"price_history\": [50000,51000,52000,51500,51200],
                 \"volume_history\": [1000,1200,1100,900,950],
                 \"timestamp\": $(date +%s),
                 \"market_cap\": 1000000000000,
                 \"holders\": 1000000
             }" \
             "$ML_SERVICE_URL/api/v1/predict" > /dev/null
    done
    echo "Prediction endpoint test completed"
}

# Test WebSocket connections
test_websocket() {
    echo "Testing WebSocket connections..."
    for symbol in "${SYMBOLS[@]}"; do
        timeout 10s websocat "ws://${BACKEND_URL#http://}/ws/$symbol" || true
    done
    echo "WebSocket test completed"
}

# Run tests
echo "Starting stress tests..."
test_market_data
test_predictions
test_websocket

echo "All stress tests completed successfully"
