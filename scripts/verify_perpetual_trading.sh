#!/bin/bash
set -e

# Configuration
ML_SERVICE_URL="http://localhost:8000"
BACKEND_URL="http://localhost:8080"
SYMBOLS=("BTC-USD" "ETH-USD" "SOL-USD" "AVAX-USD")

echo "Starting perpetual trading verification..."

# Test market data endpoints with funding rates
test_market_data() {
    local symbol=$1
    echo "Testing market data for $symbol..."
    response=$(curl -s -f "$ML_SERVICE_URL/api/v1/market-data/$symbol")
    
    if echo "$response" | grep -q "fundingRate"; then
        echo "✓ Funding rate present for $symbol"
    else
        echo "✗ Missing funding rate for $symbol"
        exit 1
    fi
    
    if echo "$response" | grep -q "nextFundingTime"; then
        echo "✓ Next funding time present for $symbol"
    else
        echo "✗ Missing next funding time for $symbol"
        exit 1
    fi
}

# Test WebSocket market data stream
test_websocket_stream() {
    local symbol=$1
    echo "Testing WebSocket stream for $symbol..."
    timeout 30s websocat "ws://${BACKEND_URL#http://}/ws/$symbol" | while read -r line; do
        if echo "$line" | grep -q "fundingRate"; then
            echo "✓ Received funding rate update for $symbol"
            break
        fi
    done
}

# Run verification
echo "Starting verification process..."

# 1. Verify environment
if ! curl -s "$ML_SERVICE_URL/api/v1/health" > /dev/null; then
    echo "ML service not running"
    exit 1
fi

# 2. Test market data for all symbols
for symbol in "${SYMBOLS[@]}"; do
    test_market_data "$symbol"
    test_websocket_stream "$symbol"
done

# 3. Test prediction endpoint with perpetual data
echo "Testing prediction endpoint with perpetual data..."
curl -s -X POST \
     -H "Content-Type: application/json" \
     -d '{
         "token_address": "BTC-USD",
         "price_history": [50000,51000,52000,51500,51200],
         "volume_history": [1000,1200,1100,900,950],
         "timestamp": '$(date +%s)',
         "market_cap": 1000000000000,
         "holders": 1000000
     }' \
     "$ML_SERVICE_URL/api/v1/predict"

echo "Perpetual trading verification completed"
