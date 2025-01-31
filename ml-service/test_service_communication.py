import asyncio
import logging
from trading_system import TradingSystem

# Set up logging
logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

async def test_service_communication():
    # Create trading system instance which will initialize all services
    trading_system = TradingSystem()
    
    async def price_callback(symbol: str, data: dict):
        logger.debug(f"[ServiceTest] Price update received for {symbol}: {data}")
        # Verify data propagation to trading system
        current_price = trading_system.market_data_service.get_latest_price(symbol)
        logger.debug(f"[ServiceTest] Latest price in market service for {symbol}: {current_price}")
    
    async def orderbook_callback(symbol: str, data: dict):
        logger.debug(f"[ServiceTest] Orderbook update received for {symbol}: {data}")
        # Verify orderbook data
        depth = trading_system.market_data_service.get_market_depth(symbol)
        logger.debug(f"[ServiceTest] Market depth for {symbol}: {depth}")
    
    async def funding_callback(symbol: str, data: dict):
        logger.debug(f"[ServiceTest] Funding update received for {symbol}: {data}")
        # Verify perpetual data
        funding_rate = trading_system.market_data_service.get_funding_rate(symbol)
        logger.debug(f"[ServiceTest] Funding rate for {symbol}: {funding_rate}")
    
    # Set up logging
    logging.basicConfig(level=logging.DEBUG)
    
    # Subscribe to market data updates
    trading_system.market_data_service.subscribe('trades', price_callback)
    trading_system.market_data_service.subscribe('orderbook', orderbook_callback)
    trading_system.market_data_service.subscribe('funding', funding_callback)
    
    try:
        logger.info("[ServiceTest] Starting trading system...")
        await trading_system.start()
        
        # Monitor communication for 30 seconds
        logger.info("[ServiceTest] Monitoring inter-service communication...")
        await asyncio.sleep(30)
        
        # Get system status
        status = trading_system.get_system_status()
        logger.info(f"[ServiceTest] System status: {status}")
        
    except Exception as e:
        logger.error(f"[ServiceTest] Error during test: {str(e)}")
    finally:
        logger.info("[ServiceTest] Stopping trading system...")
        await trading_system.stop()
    
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
