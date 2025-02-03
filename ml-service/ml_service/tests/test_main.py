import pytest
from fastapi.testclient import TestClient
from unittest.mock import AsyncMock, patch
import pandas as pd
import numpy as np
from datetime import datetime

from ml_service.main import app, market_data, ollama_client, deepseek

client = TestClient(app)

@pytest.fixture
def mock_market_data():
    return {
        'price': [100.0, 101.0, 102.0],
        'volume': [1000.0, 1100.0, 1200.0],
        'timestamp': pd.date_range(start='2024-01-01', periods=3, freq='h')
    }

@pytest.mark.asyncio
async def test_health_check():
    with patch('database.health.check_database_health', AsyncMock(return_value=True)), \
         patch('database.health.check_market_data_health', AsyncMock(return_value=True)), \
         patch('ollama_client.OllamaClient.check_health', AsyncMock(return_value=True)), \
         patch('deepseek_client.DeepseekClient.check_health', AsyncMock(return_value=True)):
        
        response = client.get("/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"
        assert all(v == "healthy" for v in data["components"].values())

@pytest.mark.asyncio
async def test_metrics():
    response = client.get("/metrics")
    assert response.status_code == 200
    data = response.json()
    assert "system" in data
    assert "analyzer" in data

@pytest.mark.asyncio
async def test_market_data():
    with patch('market_data.MarketData.fetch_market_data', AsyncMock(return_value=pd.DataFrame({
        'price': [100.0, 101.0, 102.0],
        'volume': [1000.0, 1100.0, 1200.0],
        'timestamp': pd.date_range(start='2024-01-01', periods=3, freq='h')
    }))):
        response = client.get("/api/v1/market/BTC-USD")
        assert response.status_code == 200
        data = response.json()
        assert "prices" in data
        assert "volumes" in data
        assert len(data["prices"]) == 3

@pytest.mark.asyncio
async def test_prediction():
    test_data = {
        "token_address": "0x123",
        "price_history": [100.0, 101.0, 102.0],
        "volume_history": [1000.0, 1100.0, 1200.0],
        "timestamp": int(datetime.now().timestamp()),
        "market_cap": 1000000.0,
        "holders": 1000
    }
    
    with patch('ollama_client.OllamaClient.analyze_market_sentiment', AsyncMock(return_value={
        "prediction": 1,
        "confidence": 0.85,
        "reasoning": "Test reasoning"
    })):
        response = client.post("/api/v1/predict", json=test_data)
        assert response.status_code == 200
        data = response.json()
        assert "prediction" in data
        assert "confidence" in data
        assert data["confidence"] > 0

@pytest.mark.asyncio
async def test_model_fallback():
    test_data = {
        "token_address": "0x123",
        "price_history": [100.0, 101.0, 102.0]
    }
    
    # Test Ollama failure with DeepSeek fallback
    with patch('ollama_client.OllamaClient.analyze_market_sentiment', AsyncMock(side_effect=Exception("Ollama error"))), \
         patch('deepseek_client.DeepseekClient.analyze_market_sentiment', AsyncMock(return_value={
             "prediction": 1,
             "confidence": 0.75,
             "reasoning": "Fallback reasoning"
         })):
        response = client.post("/api/v1/predict", json=test_data)
        assert response.status_code == 200
        data = response.json()
        assert data["model"] == "deepseek"
        assert data["confidence"] > 0

@pytest.mark.asyncio
async def test_error_handling():
    # Test invalid input
    response = client.post("/api/v1/predict", json={})
    assert response.status_code == 422
    
    # Test both models failing
    test_data = {"token_address": "0x123", "price_history": [100.0, 101.0, 102.0]}
    with patch('ollama_client.OllamaClient.analyze_market_sentiment', AsyncMock(side_effect=Exception("Ollama error"))), \
         patch('deepseek_client.DeepseekClient.analyze_market_sentiment', AsyncMock(side_effect=Exception("DeepSeek error"))):
        response = client.post("/api/v1/predict", json=test_data)
        assert response.status_code == 503
        data = response.json()
        assert "error" in data
