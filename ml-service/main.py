from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import numpy as np
import pandas as pd
from datetime import datetime
import uvicorn
import asyncio
import aiohttp
from typing import List, Dict
import os
from dotenv import load_dotenv
from deepseek_client import DeepseekClient

# Load environment variables
load_dotenv()

app = FastAPI()
deepseek = DeepseekClient()

class PredictionRequest(BaseModel):
    token_address: str
    price_history: List[float]
    volume_history: List[float]
    timestamp: int
    market_cap: float
    holders: int

class PredictionResponse(BaseModel):
    prediction: float
    confidence: float
    recommendation: str
    analysis: Dict[str, float]
    deepseek_analysis: Dict

class MarketData:
    def __init__(self):
        self.price_cache = {}
        self.volume_cache = {}

    async def fetch_market_data(self, token_address: str) -> Dict:
        """Fetch market data from DEX"""
        async with aiohttp.ClientSession() as session:
            # Replace with actual DEX API endpoint
            url = f"https://api.dex.example/v1/token/{token_address}"
            async with session.get(url) as response:
                if response.status == 200:
                    return await response.json()
                raise HTTPException(status_code=response.status, detail="Failed to fetch market data")

class TradingAnalyzer:
    def __init__(self):
        self.market_data = MarketData()
        # Initialize your ML model here
        # self.model = load_model()

    def preprocess_data(self, price_history: List[float], volume_history: List[float]) -> np.ndarray:
        """Preprocess data for model input"""
        df = pd.DataFrame({
            'price': price_history,
            'volume': volume_history
        })
        
        # Add technical indicators
        df['price_change'] = df['price'].pct_change()
        df['volume_change'] = df['volume'].pct_change()
        df['volatility'] = df['price'].rolling(window=10).std()
        
        # Fill NaN values
        df = df.fillna(method='bfill')
        
        return df.values

    def analyze_market(self, data: np.ndarray) -> Dict[str, float]:
        """Analyze market conditions"""
        price_data = data[:, 0]
        volume_data = data[:, 1]
        
        analysis = {
            'trend': np.mean(np.diff(price_data[-10:])),
            'volatility': np.std(price_data[-20:]),
            'volume_trend': np.mean(np.diff(volume_data[-10:])),
            'momentum': np.sum(np.diff(price_data[-5:]))
        }
        
        return analysis

    async def get_prediction(self, request: PredictionRequest) -> PredictionResponse:
        """Generate trading prediction"""
        # Preprocess data
        data = self.preprocess_data(request.price_history, request.volume_history)
        
        # Analyze market conditions
        analysis = self.analyze_market(data)
        
        # Prepare data for Deepseek analysis
        token_data = {
            'current_price': request.price_history[-1],
            'price_change_24h': ((request.price_history[-1] / request.price_history[-24]) - 1) * 100 if len(request.price_history) >= 24 else 0,
            'price_change_7d': ((request.price_history[-1] / request.price_history[-168]) - 1) * 100 if len(request.price_history) >= 168 else 0,
            'volume_24h': sum(request.volume_history[-24:]) if len(request.volume_history) >= 24 else request.volume_history[-1],
            'volume_change': ((request.volume_history[-1] / request.volume_history[-24]) - 1) * 100 if len(request.volume_history) >= 24 else 0,
            'market_cap': request.market_cap,
            'holders': request.holders
        }
        
        # Get Deepseek analysis
        deepseek_analysis = await deepseek.analyze_market_sentiment(token_data)
        
        # Combine traditional analysis with Deepseek insights
        prediction = request.price_history[-1] * (1 + np.random.normal(0, 0.01))
        confidence = max(0, min(1, deepseek_analysis['confidence']))
        
        return PredictionResponse(
            prediction=float(prediction),
            confidence=confidence,
            recommendation=deepseek_analysis['recommendation'],
            analysis=analysis,
            deepseek_analysis=deepseek_analysis
        )

analyzer = TradingAnalyzer()

@app.get("/")
async def root():
    return {"status": "ML Service is running"}

@app.post("/predict")
async def predict(request: PredictionRequest):
    try:
        return await analyzer.get_prediction(request)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/health")
async def health_check():
    return {"status": "healthy"}

if __name__ == "__main__":
    port = int(os.getenv("ML_SERVICE_PORT", 8000))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=True) 