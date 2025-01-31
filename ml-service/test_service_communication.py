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
    
    risk_config = RiskConfig(
        max_position_size=1.0,
        max_drawdown=0.1,
        daily_loss_limit=1000.0,
        position_limit=5,
        risk_per_trade=0.02,
        leverage_limit=3.0,
        correlation_limit=0.7,
        min_diversification=2,
        stop_loss_atr=2.0,
        take_profit_atr=3.0,
        db_path=":memory:",
        max_leverage=5.0,
        max_position_value=10000.0
    )
    risk_manager = RiskManager(risk_config)
    
    trading_system = TradingSystem(
        market_data=market_service,
        agent_system=agent_system,
        risk_manager=risk_manager
    )
    
    async def price_callback(symbol: str, data: dict):
        logger.debug(f"[ServiceTest] Price update received for {symbol}: {data}")
    
    async def orderbook_callback(symbol: str, data: dict):
        logger.debug(f"[ServiceTest] Orderbook update received for {symbol}: {data}")
    
    async def funding_callback(symbol: str, data: dict):
        logger.debug(f"[ServiceTest] Funding update received for {symbol}: {data}")
    
    market_service.subscribe('trades', price_callback)
    market_service.subscribe('orderbook', orderbook_callback)
    market_service.subscribe('funding', funding_callback)
    
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
