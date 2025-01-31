from fastapi import FastAPI, HTTPException, APIRouter
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from pydantic import BaseModel
from typing import Dict, Any, Optional, List, Literal, Union, Tuple
from datetime import datetime
import uvicorn
import asyncio
import aiohttp
import os
import psutil
import json
import numpy as np
import pandas as pd
from enum import Enum
from loguru import logger

from database.health import check_database_health, check_market_data_health
from ollama_client import OllamaClient
from deepseek_client import DeepseekClient

# Initialize clients
ollama_client = OllamaClient()
deepseek = DeepseekClient()
import pandas as pd
from datetime import datetime
import uvicorn
import asyncio
import aiohttp
import os
import psutil
from enum import Enum
from dotenv import load_dotenv
from loguru import logger
from deepseek_client import DeepseekClient
from ollama_client import OllamaClient
from database.health import (
    check_database_health,
    check_market_data_health,
    check_dydx_api_health,
    check_hyperliquid_api_health
)

class ServiceStatus(Enum):
    HEALTHY = "healthy"
    DEGRADED = "degraded"
    UNHEALTHY = "unhealthy"

# Configure logger
logger.add("ml_service.log", rotation="500 MB", level="INFO")

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

async def check_component_health(name: str, check_func, start_time: datetime) -> Tuple[Dict[str, Any], List[str], List[str]]:
    """Check health of a single component"""
    try:
        check_start = datetime.utcnow()
        status = await check_func()
        latency = (datetime.utcnow() - check_start).total_seconds() * 1000
        
        is_healthy = status.get("healthy", False)
        component_status = {
            "status": ServiceStatus.HEALTHY.value if is_healthy else ServiceStatus.DEGRADED.value,
            "last_check": datetime.utcnow().isoformat(),
            "details": status.get("details", {}),
            "latency_ms": latency
        }
        
        critical_errors = []
        degraded_services = []
        
        if not is_healthy:
            degraded_services.append(name)
            if error := status.get("error"):
                critical_errors.append(f"{name.title()} error: {error}")
                component_status["status"] = ServiceStatus.UNHEALTHY.value
                
        return component_status, critical_errors, degraded_services
    except Exception as e:
        logger.error(f"{name} health check failed: {e}")
        return {
            "status": ServiceStatus.UNHEALTHY.value,
            "error": str(e),
            "last_check": datetime.utcnow().isoformat(),
            "latency_ms": (datetime.utcnow() - start_time).total_seconds() * 1000
        }, [f"{name.title()} error: {str(e)}"], [name]

@router.get("/health", response_model=Dict[str, Any])
async def health_check() -> JSONResponse:
    """Check the health of all ML service components with detailed status"""
    try:
        start_time = datetime.utcnow()
        process = psutil.Process()
        components_status = {}
        critical_errors = []
        degraded_services = []
        
        # Get system metrics
        system_metrics = {
            "cpu_percent": process.cpu_percent(interval=0.1),
            "memory_mb": process.memory_info().rss / (1024 * 1024),
            "memory_percent": process.memory_percent(),
            "threads": process.num_threads(),
            "open_files": len(process.open_files()),
            "connections": len(process.connections())
        }
        
        # Get analyzer metrics
        try:
            analyzer_metrics = analyzer.get_performance_metrics()
        except Exception as e:
            logger.error(f"Failed to get analyzer metrics: {e}")
            analyzer_metrics = {
                "error": str(e),
                "total_inferences": 0,
                "error_rate": 1.0,
                "uptime_seconds": 0
            }
            critical_errors.append(f"Analyzer metrics unavailable: {e}")
            status = ServiceStatus.DEGRADED

        # Define health check functions
        async def check_ollama_health():
            try:
                healthy = await ollama_client.check_health()
                details = await ollama_client.get_model_info() if healthy else {}
                return {"healthy": healthy, "details": details}
            except Exception as e:
                logger.error(f"Ollama health check failed: {e}")
                return {"healthy": False, "error": str(e)}
            
        async def check_deepseek_health():
            try:
                healthy = await deepseek.check_health()
                return {"healthy": healthy}
            except Exception as e:
                logger.error(f"Deepseek health check failed: {e}")
                return {"healthy": False, "error": str(e)}
            
        # Check all components concurrently
        components = {
            "database": check_database_health,
            "market_data": check_market_data_health,
            "ollama": check_ollama_health,
            "deepseek": check_deepseek_health
        }

        # Track service status
        critical_services = ["database", "market_data"]
        ml_services = ["ollama", "deepseek"]
        critical_errors = []
        degraded_services = []
        status = ServiceStatus.HEALTHY
        
        # Check all components concurrently
        tasks = []
        for name, check_func in components.items():
            tasks.append(check_component_health(name, check_func, start_time))
        
        results = await asyncio.gather(*tasks, return_exceptions=True)
        
        # Process results
        for name, (component_status, errors, degraded) in zip(components.keys(), results):
            if isinstance(component_status, Exception):
                logger.error(f"Health check failed for {name}: {component_status}")
                components_status[name] = {
                    "status": ServiceStatus.UNHEALTHY.value,
                    "error": str(component_status),
                    "last_check": datetime.utcnow().isoformat()
                }
                critical_errors.append(f"{name} health check failed: {component_status}")
                degraded_services.append(name)
                if name in critical_services:
                    status = ServiceStatus.UNHEALTHY
                elif status != ServiceStatus.UNHEALTHY:
                    status = ServiceStatus.DEGRADED
                continue
                
            components_status[name] = component_status
            critical_errors.extend(errors)
            degraded_services.extend(degraded)
            
            if name in critical_services and component_status["status"] == ServiceStatus.UNHEALTHY.value:
                status = ServiceStatus.UNHEALTHY
            elif name in ml_services and component_status["status"] == ServiceStatus.UNHEALTHY.value:
                if status != ServiceStatus.UNHEALTHY:
                    status = ServiceStatus.DEGRADED
                    
        # Determine final status based on component health
        if any(error for error in critical_errors if "database" in error.lower() or "market_data" in error.lower()):
            status = ServiceStatus.UNHEALTHY
        elif len(critical_errors) > 0:
            status = ServiceStatus.DEGRADED
            
        # Set status code
        status_code = {
            ServiceStatus.HEALTHY: 200,
            ServiceStatus.DEGRADED: 207,
            ServiceStatus.UNHEALTHY: 503
        }[status]
        
        response_time = (datetime.utcnow() - start_time).total_seconds() * 1000
        
        # Build response with detailed component status
        response = {
            "status": status.value,
            "timestamp": datetime.utcnow().isoformat(),
            "components": components_status,
            "analyzer_metrics": analyzer_metrics,
            "system_metrics": system_metrics,
            "response_time_ms": response_time
        }
        
        if critical_errors:
            response["critical_errors"] = critical_errors
        if degraded_services:
            response["degraded_services"] = degraded_services
            
        return JSONResponse(
            status_code=status_code,
            content=response,
            headers={"Cache-Control": "no-cache"}
        )
    except Exception as e:
        logger.error(f"Health check failed: {e}")
        return JSONResponse(
            status_code=503,
            content={
                "status": ServiceStatus.UNHEALTHY.value,
                "error": str(e),
                "timestamp": datetime.utcnow().isoformat(),
                "components": {
                    "database": {"status": ServiceStatus.UNHEALTHY.value},
                    "market_data": {"status": ServiceStatus.UNHEALTHY.value},
                    "ollama": {"status": ServiceStatus.UNHEALTHY.value},
                    "deepseek": {"status": ServiceStatus.UNHEALTHY.value}
                }
            },
            headers={"Cache-Control": "no-cache"}
        )
    except Exception as e:
        logger.error(f"Health check failed: {e}")
        return JSONResponse(
            status_code=503,
            content={
                "status": "unhealthy",
                "timestamp": datetime.utcnow().isoformat(),
                "error": str(e),
                "components": {k: "unhealthy" for k in ["database", "market_data", "ollama", "deepseek"]}
            }
        )

# Initialize services
market_data = None
deepseek = None
ollama_client = None

@app.on_event("startup")
async def startup_event():
    """Initialize services on application startup"""
    global market_data, deepseek, ollama_client
    
    # Initialize market data service
    if market_data is None:
        market_data = MarketData()
        await market_data.initialize()
        print("Initialized MarketData service during startup")
    
    # Initialize Deepseek client
    if deepseek is None:
        deepseek = DeepseekClient()
        print("Initialized Deepseek client during startup")
    
    # Initialize Ollama client
    if ollama_client is None:
        ollama_client = OllamaClient()
        print("Initialized Ollama client during startup")
        
        # Pull model if not already available
        try:
            model_info = await ollama_client.get_model_info()
            if model_info.get("status") != "loaded":
                print("Model not found, attempting to pull...")
                # Model pull is handled by the client
        except Exception as e:
            print(f"Error checking model status: {e}")
    
    # Debug route registration
    print("\nRegistered routes:")
    for route in app.routes:
        print(f"{route.path} [{','.join(route.methods)}]")
async def startup_event():
    """Initialize services on application startup"""
    global market_data, deepseek, ollama_client
    
    # Initialize market data service
    if market_data is None:
        market_data = MarketData()
        await market_data.initialize()
        logger.info("Initialized MarketData service during startup")
    
    # Initialize Deepseek client
    if deepseek is None:
        deepseek = DeepseekClient()
        logger.info("Initialized Deepseek client during startup")
    
    # Initialize Ollama client
    if ollama_client is None:
        ollama_client = OllamaClient()
        logger.info("Initialized Ollama client during startup")
        
        # Pull model if not already available
        try:
            model_info = await ollama_client.get_model_info()
            if model_info.get("status") != "loaded":
                logger.info("Model not found, attempting to pull...")
        except Exception as e:
            logger.error(f"Error checking model status: {e}")
    
    # Debug route registration
    logger.info("\nRegistered routes:")
    for route in app.routes:
        logger.info(f"{route.path} [{','.join(route.methods)}]")

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
        self.initialized = False
        
    async def initialize(self):
        if not self.initialized:
            self.session = aiohttp.ClientSession()
            self.initialized = True
            return True
        return False

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
        self.last_inference_time = None
        self.total_inferences = 0
        self.inference_times = []
        self.error_count = 0
        self.start_time = datetime.utcnow()

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
        start_time = datetime.utcnow()
        try:
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
            
            response = PredictionResponse(
                prediction=float(prediction),
                confidence=confidence,
                recommendation=deepseek_analysis['recommendation'],
                analysis=analysis,
                deepseek_analysis=deepseek_analysis
            )
            
            # Update metrics
            self.total_inferences += 1
            self.last_inference_time = datetime.utcnow()
            inference_time = (self.last_inference_time - start_time).total_seconds()
            self.inference_times.append(inference_time)
            if len(self.inference_times) > 1000:
                self.inference_times.pop(0)
            
            return response
            
        except Exception as e:
            self.error_count += 1
            logger.error(f"Prediction error: {str(e)}")
            raise

    def get_performance_metrics(self) -> Dict[str, float]:
        """Get analyzer performance metrics"""
        current_time = datetime.utcnow()
        uptime = (current_time - self.start_time).total_seconds()
        
        metrics = {
            "total_inferences": self.total_inferences,
            "error_rate": self.error_count / max(1, self.total_inferences),
            "uptime_seconds": uptime,
            "inferences_per_second": self.total_inferences / max(1, uptime),
            "average_latency": sum(self.inference_times) / max(1, len(self.inference_times)),
            "last_inference_age": (current_time - (self.last_inference_time or current_time)).total_seconds()
        }
        
        return metrics

analyzer = TradingAnalyzer()

@router.get("/", tags=["System"])
async def root():
    """Root endpoint"""
    return {"status": "ML Service is running"}

@router.get("/metrics", tags=["System"])
async def get_metrics():
    """Get detailed system and analyzer metrics with component status"""
    try:
        # Get analyzer metrics with error handling
        try:
            analyzer_metrics = analyzer.get_performance_metrics()
        except Exception as e:
            logger.error(f"Failed to get analyzer metrics: {str(e)}")
            analyzer_metrics = {
                "error": str(e),
                "total_inferences": 0,
                "error_rate": 1.0,
                "uptime_seconds": 0
            }

        # Get system metrics
        process = psutil.Process()
        memory_info = process.memory_info()
        
        # Get component status
        components_status = {}
        
        # Check Ollama
        try:
            ollama_healthy = await ollama_client.check_health()
            components_status["ollama"] = {
                "status": "healthy" if ollama_healthy else "unhealthy",
                "model": "deepseek-r1:1.5b",
                "last_check": datetime.utcnow().isoformat()
            }
        except Exception as e:
            components_status["ollama"] = {
                "status": "error",
                "error": str(e),
                "last_check": datetime.utcnow().isoformat()
            }
            
        # Check database
        try:
            db_healthy = await check_database_health()
            components_status["database"] = {
                "status": "healthy" if db_healthy else "unhealthy",
                "last_check": datetime.utcnow().isoformat()
            }
        except Exception as e:
            components_status["database"] = {
                "status": "error",
                "error": str(e),
                "last_check": datetime.utcnow().isoformat()
            }

        # Check market data
        try:
            market_data_healthy = await check_market_data_health()
            components_status["market_data"] = {
                "status": "healthy" if market_data_healthy else "unhealthy",
                "last_check": datetime.utcnow().isoformat()
            }
        except Exception as e:
            components_status["market_data"] = {
                "status": "error",
                "error": str(e),
                "last_check": datetime.utcnow().isoformat()
            }

        return {
            "analyzer": analyzer_metrics,
            "system": {
                "cpu_percent": process.cpu_percent(),
                "memory_mb": memory_info.rss / 1024 / 1024,
                "memory_percent": process.memory_percent(),
                "threads": process.num_threads(),
                "open_files": len(process.open_files()),
                "connections": len(process.connections())
            },
            "components": components_status,
            "timestamp": datetime.utcnow().isoformat()
        }
    except Exception as e:
        logger.error(f"Error getting metrics: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail={
                "error": str(e),
                "timestamp": datetime.utcnow().isoformat()
            }
        )

@router.post("/predict", tags=["Predictions"])
async def predict(request: PredictionRequest):
    """Generate trading prediction with fallback mechanisms"""
    start_time = datetime.utcnow()
    
    try:
        # Try primary prediction path with analyzer
        try:
            prediction = await analyzer.get_prediction(request)
            logger.info("Primary prediction successful")
            return prediction
        except Exception as primary_error:
            logger.warning(f"Primary prediction failed: {str(primary_error)}, attempting fallback")
            
            # Attempt fallback to Ollama model
            try:
                token_data = {
                    "current_price": request.price_history[-1] if request.price_history else None,
                    "price_change_24h": ((request.price_history[-1] / request.price_history[-24]) - 1) * 100 if len(request.price_history) >= 24 else None,
                    "price_change_7d": ((request.price_history[-1] / request.price_history[-168]) - 1) * 100 if len(request.price_history) >= 168 else None,
                    "volume_24h": sum(request.volume_history[-24:]) if len(request.volume_history) >= 24 else None,
                    "volume_change": ((sum(request.volume_history[-24:]) / sum(request.volume_history[-48:-24])) - 1) * 100 if len(request.volume_history) >= 48 else None,
                    "market_cap": request.market_cap,
                    "holders": request.holders
                }
                
                analysis = await ollama_client.analyze_market_sentiment(token_data)
                logger.info("Fallback prediction successful using Ollama model")
                
                return PredictionResponse(
                    prediction=request.price_history[-1] * (1 + float(analysis.get('risk_level', 5.0)) / 100),
                    confidence=float(analysis.get('confidence', 0.5)),
                    recommendation=analysis.get('recommendation', 'HOLD'),
                    analysis={
                        'risk_level': float(analysis.get('risk_level', 5.0)),
                        'sentiment': analysis.get('sentiment', 'neutral'),
                        'manipulation_risk': analysis.get('risk_analysis', {}).get('manipulation_risk', 'medium'),
                        'liquidity_risk': analysis.get('risk_analysis', {}).get('liquidity_risk', 'medium')
                    },
                    deepseek_analysis=None
                )
            except Exception as fallback_error:
                logger.error(f"Fallback prediction failed: {str(fallback_error)}")
                raise HTTPException(
                    status_code=503,
                    detail={
                        "error": "All prediction services failed",
                        "primary_error": str(primary_error),
                        "fallback_error": str(fallback_error),
                        "timestamp": datetime.utcnow().isoformat()
                    }
                )
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Unexpected error in prediction endpoint: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail={
                "error": "Internal server error",
                "message": str(e),
                "timestamp": datetime.utcnow().isoformat()
            }
        )
    finally:
        # Update analyzer metrics
        analyzer.total_inferences += 1
        analyzer.last_inference_time = datetime.utcnow()
        inference_time = (analyzer.last_inference_time - start_time).total_seconds()
        analyzer.inference_times.append(inference_time)
        if len(analyzer.inference_times) > 1000:
            analyzer.inference_times.pop(0)

# Include router in app
# Router already has /api/v1 prefix, so we include it at root
app.include_router(router)

if __name__ == "__main__":
    port = int(os.getenv("ML_SERVICE_PORT", 8000))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=True)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      