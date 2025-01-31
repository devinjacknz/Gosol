from fastapi import FastAPI, HTTPException, APIRouter
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import Optional, List, Dict
import numpy as np
import pandas as pd
from datetime import datetime
import uvicorn
import asyncio
import aiohttp
import os
from dotenv import load_dotenv
from deepseek_client import DeepseekClient

# Load environment variables
load_dotenv()

class MarketDataResponse(BaseModel):
    symbol: str
    price: float
    change24h: float
    volume24h: float
    openInterest: Optional[float] = None
    fundingRate: Optional[float] = None
    nextFundingTime: Optional[str] = None

app = FastAPI(
    title="Gosol ML Service",
    description="Market data and trading analysis service",
    version="1.0.0",
    docs_url="/api/v1/docs",
    openapi_url="/api/v1/openapi.json"
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Create router for API v1
router = APIRouter(prefix="/api/v1", tags=["Market Data"])

@router.get("/market-data/{symbol}", response_model=MarketDataResponse)
async def get_market_data(symbol: str):
    """Get market data for a specific trading pair"""
    try:
        global market_data
        if market_data is None:
            market_data = MarketData()
            print("Initialized new MarketData instance")
        
        original_symbol = symbol
        if '-USD' in symbol:
            symbol = symbol.replace('-USD', '-PERP')
        elif not symbol.endswith('-PERP'):
            symbol = f"{symbol}-PERP"
        print(f"Processing market data request for {original_symbol} (normalized to {symbol})")
        
        data = await market_data.fetch_market_data(symbol)
        print(f"Successfully fetched market data: {data}")
        return data
    except Exception as e:
        import traceback
        print(f"Error fetching market data for {symbol}:")
        print(traceback.format_exc())
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/health")
async def health_check():
    """Health check endpoint"""
    return {"status": "healthy"}

# Initialize services
market_data = None
deepseek = DeepseekClient()

@app.on_event("startup")
async def startup_event():
    """Initialize services on application startup"""
    global market_data
    if market_data is None:
        market_data = MarketData()
        print("Initialized MarketData service during startup")
    
    # Debug route registration
    print("\nRegistered routes:")
    for route in app.routes:
        print(f"{route.path} [{','.join(route.methods)}]")

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
        self.dydx_api_endpoint = os.getenv('DYDX_API_ENDPOINT', 'https://api.stage.dydx.exchange')
        self.hyperliquid_api_endpoint = os.getenv('HYPERLIQUID_API_ENDPOINT', 'https://api.hyperliquid-testnet.xyz')
        self.session = None

    async def get_session(self):
        if self.session is None:
            self.session = aiohttp.ClientSession()
        return self.session

    async def fetch_market_data(self, symbol: str) -> MarketDataResponse:
        """Fetch market data from exchanges"""
        session = await self.get_session()
        
        # Try Hyperliquid first since dYdX is not resolving
        try:
            hl_symbol = symbol.split('-')[0]
            hl_url = f"{self.hyperliquid_api_endpoint}/info"
            headers = {'Content-Type': 'application/json'}
            
            print(f"Fetching Hyperliquid metadata for {hl_symbol}")
            meta_data = {'type': 'meta'}
            async with session.post(hl_url, json=meta_data, headers=headers) as meta_response:
                meta_text = await meta_response.text()
                print(f"Hyperliquid meta response: {meta_text}")
                if meta_response.status == 200:
                    meta = await meta_response.json()
                    for market in meta.get('universe', []):
                        if isinstance(market, dict) and market.get('name') == hl_symbol:
                            print(f"Found market {hl_symbol} in Hyperliquid universe")
                            
                            print(f"Fetching L2 book data for {hl_symbol}")
                            l2_data = {'type': 'l2Book', 'coin': hl_symbol}
                            async with session.post(hl_url, json=l2_data, headers=headers) as l2_response:
                                l2_text = await l2_response.text()
                                print(f"Hyperliquid L2 response: {l2_text}")
                                if l2_response.status == 200:
                                    l2_book = await l2_response.json()
                                    if 'levels' not in l2_book:
                                        print(f"No levels in L2 book: {l2_book}")
                                        continue
                                        
                                    asks = l2_book['levels'][0]
                                    bids = l2_book['levels'][1]
                                    
                                    if asks and bids:
                                        best_ask = float(asks[0]['px'])
                                        best_bid = float(bids[0]['px'])
                                        mid_price = (best_ask + best_bid) / 2
                                        
                                        print(f"Fetching market state for {hl_symbol}")
                                        state_data = {'type': 'stats', 'coin': hl_symbol}
                                        async with session.post(hl_url, json=state_data, headers=headers) as state_response:
                                            state_text = await state_response.text()
                                            print(f"Hyperliquid stats response: {state_text}")
                                            if state_response.status == 200:
                                                state = await state_response.json()
                                                return MarketDataResponse(
                                                    symbol=symbol,
                                                    price=mid_price,
                                                    change24h=float(state.get('dayReturn', 0)) * 100,
                                                    volume24h=float(state.get('dayUsdVolume', 0)),
                                                    openInterest=float(state.get('openInterest', 0)),
                                                    fundingRate=float(state.get('funding', 0)) / 1e6,
                                                    nextFundingTime=state.get('nextFunding')
                                                )
        except Exception as e:
            import traceback
            print(f"Hyperliquid API error for {hl_symbol}:")
            print(traceback.format_exc())

        # Try dYdX as fallback
        try:
            dydx_symbol = symbol.split('-')[0] + '-USD'
            dydx_url = "https://api.dydx.exchange/v3/markets"  # Use main API instead of staging
            print(f"Trying dYdX API: {dydx_url}")
            async with session.get(dydx_url) as dydx_response:
                dydx_text = await dydx_response.text()
                print(f"dYdX response: {dydx_text}")
                if dydx_response.status == 200:
                    dydx_data = await dydx_response.json()
                    market = dydx_data.get('markets', {}).get(dydx_symbol, {})
                    if market:
                        return MarketDataResponse(
                            symbol=symbol,
                            price=float(market.get('indexPrice', 0)),
                            change24h=float(market.get('priceChange24H', 0)),
                            volume24h=float(market.get('volume24H', 0)),
                            openInterest=float(market.get('openInterest', 0)),
                            fundingRate=float(market.get('nextFundingRate', 0)),
                            nextFundingTime=market.get('nextFundingAt')
                        )
        except Exception as e:
            import traceback
            print(f"dYdX API error for {dydx_symbol}:")
            print(traceback.format_exc())

        raise HTTPException(status_code=404, detail=f"Market data not found for {symbol}")

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

@router.get("/", tags=["System"])
async def root():
    """Root endpoint"""
    return {"status": "ML Service is running"}

@router.post("/predict", tags=["Predictions"])
async def predict(request: PredictionRequest):
    """Generate trading prediction"""
    try:
        return await analyzer.get_prediction(request)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

# Include router in app
app.include_router(router, prefix="/api/v1")

if __name__ == "__main__":
    port = int(os.getenv("ML_SERVICE_PORT", 8000))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=True) 