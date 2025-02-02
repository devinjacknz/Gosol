#!/bin/bash

# Exit on any error
set -e

ENV=$1
MAX_RETRIES=30
RETRY_INTERVAL=10

if [ -z "$ENV" ]; then
    echo "Usage: $0 <environment>"
    echo "Example: $0 staging"
    exit 1
fi

case "$ENV" in
    staging)
        BASE_URL="https://staging-api.gosol.com"
        ;;
    production)
        BASE_URL="https://api.gosol.com"
        ;;
    *)
        echo "Invalid environment: $ENV"
        echo "Supported environments: staging, production"
        exit 1
        ;;
esac

echo "Verifying deployment in $ENV environment..."
echo "Base URL: $BASE_URL"

# Function to check endpoint health
check_endpoint() {
    local endpoint=$1
    local expected_status=$2
    local description=$3
    
    echo "Checking $description..."
    
    for ((i=1; i<=MAX_RETRIES; i++)); do
        response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL$endpoint")
        
        if [ "$response" -eq "$expected_status" ]; then
            echo "✅ $description is healthy"
            return 0
        else
            if [ $i -eq $MAX_RETRIES ]; then
                echo "❌ $description check failed after $MAX_RETRIES attempts"
                return 1
            fi
            echo "Attempt $i/$MAX_RETRIES: $description returned $response, expected $expected_status"
            sleep $RETRY_INTERVAL
        fi
    done
}

# Check various endpoints
check_endpoint "/api/health" 200 "Health check endpoint"
check_endpoint "/api/metrics" 200 "Metrics endpoint"

# Additional checks can be added here

echo "✅ All deployment verification checks passed for $ENV environment"
exit 0
