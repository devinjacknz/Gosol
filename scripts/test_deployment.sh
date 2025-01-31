#!/bin/bash
set -e

echo "Starting test deployment..."

# Build and start services
docker-compose -f docker-compose.test.yml build
docker-compose -f docker-compose.test.yml up -d

# Wait for services to be healthy
echo "Waiting for services to be healthy..."
sleep 30

# Test ML service health
echo "Testing ML service health..."
curl -f http://localhost:8000/api/v1/health

# Test market data endpoints
echo "Testing market data endpoints..."
curl -f http://localhost:8000/api/v1/market-data/BTC-USD

# Test WebSocket connection
echo "Testing WebSocket connection..."
wscat -c ws://localhost:8080/ws

# Run stress test
echo "Running stress test..."
for i in {1..100}; do
    curl -s -f http://localhost:8000/api/v1/market-data/BTC-USD &
    curl -s -f http://localhost:8000/api/v1/market-data/ETH-USD &
    if [ $((i % 10)) -eq 0 ]; then
        echo "Completed $i requests"
    fi
done

wait

echo "Deployment test completed"
