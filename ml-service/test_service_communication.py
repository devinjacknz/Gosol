import asyncio
import logging
from market_data_service import MarketDataService, MarketConfig
from trading_system import TradingSystem
from agent_system import AgentSystem, AgentConfig
from risk_management import RiskManager, RiskConfig

logging.basicConfig(level=logging.DEBUG)
logger = logging.getLogger(__name__)

async def test_service_communication():
    market_config = MarketConfig(
        exchange="binance",
        symbols=["BTC/USDT", "ETH/USDT"],
        timeframes=["1m", "5m", "15m"],
        perpetual_enabled=True,
        perpetual_symbols=["BTC/USDT", "ETH/USDT"]
    )
    
    market_service = MarketDataService(market_config)
    agent_system = AgentSystem()
    risk_manager = RiskManager(RiskConfig())
    
    trading_system = TradingSystem(
        market_service,
        agent_system,
        risk_manager
    )
    
    async def price_callback(symbol: str, data: dict):
        logger.info(f"Price update received for {symbol}: {data}")
    
    async def orderbook_callback(symbol: str, data: dict):
        logger.info(f"Orderbook update received for {symbol}: {data}")
    
    market_service.subscribe('trades', price_callback)
    market_service.subscribe('orderbook', orderbook_callback)
    
    try:
        await market_service.start()
        await trading_system.start()
        
        logger.info("Services started, monitoring communication...")
        await asyncio.sleep(30)  # Monitor for 30 seconds
        
    except Exception as e:
        logger.error(f"Error during test: {str(e)}")
    finally:
        await trading_system.stop()
        await market_service.stop()

if __name__ == "__main__":
    asyncio.run(test_service_communication())
