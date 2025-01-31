import asyncio
import logging
import os
from datetime import datetime
import pandas as pd
import pytest
from typing import Dict, Any, Optional, Callable
from unittest.mock import patch, MagicMock, AsyncMock
from trading_system import TradingSystem
from agent_system import AgentSystem, AgentConfig, TradeSignal
from trade_executor import TradeExecutor
from risk_management import RiskManager, RiskConfig
from market_data_service import MarketConfig

# Configure logging
logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger('service_test')

class MockRiskManager:
    def __init__(self, config: RiskConfig):
        self.config = config
        self.positions = {}
    
    def update_positions(self, market_data):
        logger.info(f"Updating positions with market data: {market_data}")
    
    def get_risk_report(self):
        return {
            'portfolio_summary': {
                'drawdown': 0.0,
                'daily_pnl': 0.0
            },
            'risk_metrics': {
                'var': 0.0,
                'sharpe': 0.0
            }
        }
    
    def get_position_summary(self):
        return pd.DataFrame()
    
    def check_position_risk(self, symbol, direction, size, price):
        return True
    
    def calculate_position_size(self, price: float, stop_loss: float, volatility: float) -> float:
        return 0.1  # Mock position size
    
    def _calculate_margin(self, size: float, price: float) -> float:
        return size * price * 0.1  # Mock margin calculation

class MockAgentSystem:
    def __init__(self):
        self.agents = {}
    
    def add_agent(self, config: AgentConfig):
        mock_agent = MagicMock()
        mock_agent.config = config
        mock_agent.analyze = MagicMock(return_value=TradeSignal(
            symbol=config.symbol,
            direction='buy',
            size=0.1,
            stop_loss=45000.0,
            take_profit=55000.0,
            confidence=0.8,
            agent_name=config.name,
            price=50000.0,
            timestamp=datetime.now(),
            metadata={}
        ))
        self.agents[config.name] = mock_agent
    
    def get_system_metrics(self):
        return {'total_agents': len(self.agents)}

class MockTradeExecutor:
    def __init__(self, agent_system):
        self.positions = {}
        self.agent_system = agent_system
    
    async def monitor_positions(self):
        logger.info("Monitoring positions...")
    
    async def process_signal(self, signal, price):
        logger.info(f"Processing signal for {signal.symbol} at price {price}")
    
    def get_performance_metrics(self):
        return {'total_trades': 0, 'win_rate': 0.0}

class MockConfig:
    @staticmethod
    def get_market_data_config():
        return MarketConfig(
            exchange='binance',
            symbols=['BTC/USDT', 'ETH/USDT'],
            timeframes=['1m', '5m', '15m', '1h', '4h', '1d'],
            api_key=os.getenv('TEST_API_KEY', 'mock_key'),
            api_secret=os.getenv('TEST_API_SECRET', 'mock_secret'),
            cache_size=1000,
            update_interval=1.0,
            db_path='market_data.db',
            perpetual_enabled=True,
            perpetual_symbols=['BTC/USDT', 'ETH/USDT'],
            funding_interval=8,
            max_leverage=20,
            default_leverage=1
        )
    
    @staticmethod
    def get_exchange_config():
        return {
            'name': 'test_exchange',
            'api_key': os.getenv('TEST_API_KEY', 'mock_key'),
            'api_secret': os.getenv('TEST_API_SECRET', 'mock_secret')
        }
    
    @staticmethod
    def get_risk_config():
        return {
            'max_leverage': 10.0,
            'max_position_value': 100000.0,
            'max_position_size': 10.0,  # Maximum position size in base currency
            'max_drawdown': 0.2,
            'daily_loss_limit': 5000.0,
            'position_limit': 5,
            'risk_per_trade': 0.02,
            'leverage_limit': 10.0,
            'correlation_limit': 0.7,
            'min_diversification': 3,
            'stop_loss_atr': 2.0,
            'take_profit_atr': 3.0,
            'min_maintenance_margin': 0.05,
            'funding_rate_interval': 8,
            'liquidation_threshold': 0.8,
            'margin_call_threshold': 0.9,
            'db_path': 'risk.db'
        }
    
    @staticmethod
    def get_log_config():
        return {
            'level': logging.DEBUG,
            'format': '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
        }

logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

MOCK_MARKET_DATA = {
    'BTC/USDT': {
        'price': 50000.0,
        'volume': 1000.0,
        'bids': [[49990.0, 1.0], [49980.0, 2.0]],
        'asks': [[50010.0, 1.0], [50020.0, 2.0]],
        'funding_rate': 0.0001,
        'volatility': 0.02,
        'atr': 100.0
    },
    'ETH/USDT': {
        'price': 3000.0,
        'volume': 5000.0,
        'bids': [[2990.0, 10.0], [2980.0, 20.0]],
        'asks': [[3010.0, 10.0], [3020.0, 20.0]],
        'funding_rate': 0.0002,
        'volatility': 0.03,
        'atr': 50.0
    }
}

class MockMarketDataService:
    def __init__(self, config: Optional[MarketConfig] = None):
        self.config = config if config else MockConfig.get_market_data_config()
        self.symbols = self.config.symbols
        self.perpetual_cache = {}
    
    def __getattr__(self, name: str) -> Any:
        try:
            return getattr(self.config, name)
        except AttributeError:
            raise AttributeError(f"'{self.__class__.__name__}' object has no attribute '{name}'")
    
    async def start(self) -> None:
        logger.info("[MockMarketData] Starting market data service")
    
    async def stop(self) -> None:
        logger.info("[MockMarketData] Stopping market data service")
    
    def get_latest_price(self, symbol: str) -> Optional[float]:
        logger.debug(f"[MockMarketData] Getting latest price for {symbol}")
        if symbol in MOCK_MARKET_DATA:
            return MOCK_MARKET_DATA[symbol]['price']
        return None
    
    def get_market_depth(self, symbol: str) -> Dict[str, Any]:
        logger.debug(f"[MockMarketData] Getting market depth for {symbol}")
        if symbol in MOCK_MARKET_DATA:
            return {
                'bids': MOCK_MARKET_DATA[symbol]['bids'],
                'asks': MOCK_MARKET_DATA[symbol]['asks']
            }
        return {'bids': [], 'asks': []}
    
    def get_funding_rate(self, symbol: str) -> Optional[float]:
        logger.debug(f"[MockMarketData] Getting funding rate for {symbol}")
        if symbol in MOCK_MARKET_DATA:
            return MOCK_MARKET_DATA[symbol]['funding_rate']
        return None
    
    def get_volatility(self, symbol: str) -> Optional[float]:
        logger.debug(f"[MockMarketData] Getting volatility for {symbol}")
        if symbol in MOCK_MARKET_DATA:
            return MOCK_MARKET_DATA[symbol]['volatility']
        return None
    
    def get_atr(self, symbol: str) -> Optional[float]:
        logger.debug(f"[MockMarketData] Getting ATR for {symbol}")
        if symbol in MOCK_MARKET_DATA:
            return MOCK_MARKET_DATA[symbol]['atr']
        return None
    
    def subscribe(self, event_type: str, callback: Callable) -> None:
        logger.info(f"[MockMarketData] Subscribed to {event_type} events")
    
    def get_ohlcv(self, symbol: str, timeframe: str = '1h') -> pd.DataFrame:
        logger.info(f"[MockMarketData] Getting OHLCV data for {symbol} on {timeframe}")
        if symbol not in MOCK_MARKET_DATA:
            return pd.DataFrame()
            
        data = pd.DataFrame({
            'timestamp': pd.date_range(start='2024-01-01', periods=100, freq='h'),
            'open': [MOCK_MARKET_DATA[symbol]['price']] * 100,
            'high': [MOCK_MARKET_DATA[symbol]['price'] * 1.02] * 100,
            'low': [MOCK_MARKET_DATA[symbol]['price'] * 0.98] * 100,
            'close': [MOCK_MARKET_DATA[symbol]['price']] * 100,
            'volume': [MOCK_MARKET_DATA[symbol]['volume']] * 100
        })
        return data
    
    def calculate_vwap(self, symbol: str) -> Optional[float]:
        logger.debug(f"[MockMarketData] Calculating VWAP for {symbol}")
        if symbol not in MOCK_MARKET_DATA:
            return None
        return 50250.0

@pytest.mark.asyncio
async def test_service_communication():
    with patch('trading_system.Config', MockConfig), \
         patch('trading_system.MarketDataService', MockMarketDataService), \
         patch('trading_system.AgentSystem', MockAgentSystem), \
         patch('trading_system.TradeExecutor', MockTradeExecutor), \
         patch('trading_system.RiskManager', MockRiskManager):
        trading_system = TradingSystem()
        
        try:
            logger.info("[ServiceTest] Starting trading system...")
            await trading_system.start()
            
            # Verify market data service configuration
            assert trading_system.market_data_service.config.perpetual_enabled, "Perpetual trading should be enabled"
            assert len(trading_system.market_data_service.config.perpetual_symbols) > 0, "Perpetual symbols should be configured"
            
            # Verify market data service functionality
            for symbol in MOCK_MARKET_DATA:
                # Test price data
                price = trading_system.market_data_service.get_latest_price(symbol)
                logger.info(f"[ServiceTest] Price for {symbol}: {price}")
                assert price == MOCK_MARKET_DATA[symbol]['price'], f"Price mismatch for {symbol}"
                
                # Test market depth
                depth = trading_system.market_data_service.get_market_depth(symbol)
                logger.info(f"[ServiceTest] Market depth for {symbol}: {depth}")
                assert depth['bids'] == MOCK_MARKET_DATA[symbol]['bids'], f"Bids mismatch for {symbol}"
                assert depth['asks'] == MOCK_MARKET_DATA[symbol]['asks'], f"Asks mismatch for {symbol}"
                
                # Test perpetual data
                funding = trading_system.market_data_service.get_funding_rate(symbol)
                logger.info(f"[ServiceTest] Funding rate for {symbol}: {funding}")
                assert funding == MOCK_MARKET_DATA[symbol]['funding_rate'], f"Funding rate mismatch for {symbol}"
                
                # Test technical indicators
                volatility = trading_system.market_data_service.get_volatility(symbol)
                logger.info(f"[ServiceTest] Volatility for {symbol}: {volatility}")
                assert volatility == MOCK_MARKET_DATA[symbol]['volatility'], f"Volatility mismatch for {symbol}"
                
                atr = trading_system.market_data_service.get_atr(symbol)
                logger.info(f"[ServiceTest] ATR for {symbol}: {atr}")
                assert atr == MOCK_MARKET_DATA[symbol]['atr'], f"ATR mismatch for {symbol}"
            
            # Verify system components
            logger.info("[ServiceTest] Verifying system components...")
            
            # Verify agent system
            assert len(trading_system.agent_system.agents) > 0, "Agent system should have initialized agents"
            agent_metrics = trading_system.agent_system.get_system_metrics()
            logger.info(f"[ServiceTest] Agent system metrics: {agent_metrics}")
            
            # Verify risk manager
            risk_report = trading_system.risk_manager.get_risk_report()
            logger.info(f"[ServiceTest] Risk report: {risk_report}")
            assert 'portfolio_summary' in risk_report, "Risk report should contain portfolio summary"
            assert 'risk_metrics' in risk_report, "Risk report should contain risk metrics"
            
            # Wait for analysis loop to process data
            logger.info("[ServiceTest] Waiting for analysis loop to process data...")
            await asyncio.sleep(2)
            
            logger.info("[ServiceTest] All verifications passed successfully")
            
        except Exception as e:
            logger.error(f"[ServiceTest] Error during test: {str(e)}")
            raise
        finally:
            logger.info("[ServiceTest] Stopping trading system...")
            await trading_system.stop()

if __name__ == "__main__":
    asyncio.run(test_service_communication())
