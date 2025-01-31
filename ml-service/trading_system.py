import asyncio
import logging
import pandas as pd
from typing import Dict, List, Optional, Any, Union
from datetime import datetime
from ml_service.config import Config
from ml_service.market_data_service import MarketDataService, MarketConfig
from ml_service.risk_management import RiskManager, RiskConfig, Position, ContractPosition
from ml_service.ml_agent import DeepSeekAgent
from ml_service.agent_system import TradeSignal, AgentSystem, AgentConfig
from ml_service.trade_executor import TradeExecutor

# 配置日志
logging.basicConfig(**Config.get_log_config())
logger = logging.getLogger(__name__)

class TradingSystem:
    """交易系统主类"""
    
    def __init__(self):
        # 加载配置
        market_config = Config.get_market_data_config()
        exchange_config = Config.get_exchange_config()
        risk_config_dict = Config.get_risk_config()
        
        # 初始化市场数据服务
        self.market_data_service = MarketDataService(MarketConfig(
            exchange=exchange_config['name'],
            symbols=market_config.symbols,
            timeframes=market_config.timeframes,
            api_key=exchange_config['api_key'],
            api_secret=exchange_config['api_secret'],
            cache_size=market_config.cache_size,
            update_interval=market_config.update_interval,
            db_path=market_config.db_path
        ))
        
        # 初始化Agent系统
        self.agent_system = AgentSystem()
        
        # 初始化交易执行器
        self.trade_executor = TradeExecutor(self.agent_system)
        
        # 系统状态
        self.is_running = False
        self.last_update = None
        
        # 初始化风险管理器
        self.risk_manager = RiskManager(RiskConfig(**risk_config_dict))
    
    def initialize_agents(self) -> None:
        """初始化分析Agent"""
        # 趋势跟踪Agent
        self.agent_system.add_agent(AgentConfig(
            name="trend_follower",
            symbol="BTC/USDT",
            timeframe="1h",
            strategy_type="trend_following",
            parameters={
                'stop_loss_atr': 2.0,
                'take_profit_atr': 3.0
            },
            confidence_threshold=0.7,
            risk_limit=0.02,
            max_positions=1,
            enable_ml=False
        ))
        
        # 均值回归Agent
        self.agent_system.add_agent(AgentConfig(
            name="mean_reversal",
            symbol="BTC/USDT",
            timeframe="5m",
            strategy_type="mean_reversion",
            parameters={
                'stop_loss_atr': 1.0,
                'take_profit_atr': 2.0
            },
            confidence_threshold=0.75,
            risk_limit=0.01,
            max_positions=1,
            enable_ml=False
        ))
        
        # 突破交易Agent
        self.agent_system.add_agent(AgentConfig(
            name="breakout_trader",
            symbol="ETH/USDT",
            timeframe="15m",
            strategy_type="breakout",
            parameters={
                'stop_loss_atr': 1.5,
                'take_profit_atr': 3.0
            },
            confidence_threshold=0.8,
            risk_limit=0.015,
            max_positions=1,
            enable_ml=False
        ))
        
        # DeepSeek ML Agent
        self.agent_system.add_agent(AgentConfig(
            name="deepseek_ml",
            symbol="BTC/USDT",
            timeframe="1h",
            strategy_type="ml",
            parameters={
                'stop_loss_atr': 2.0,
                'take_profit_atr': 3.0,
                'model_path': 'models/deepseek-coder-1.5b-base'
            },
            confidence_threshold=0.8,
            risk_limit=0.01,
            max_positions=1,
            enable_ml=True
        ))
    
    async def start(self) -> None:
        """启动交易系统"""
        logger.info("Starting trading system...")
        self.is_running = True
        
        # 启动市场数据服务
        await self.market_data_service.start()
        
        # 初始化Agent
        self.initialize_agents()
        
        # 启动所有异步任务
        await asyncio.gather(
            self._analysis_loop(),
            self.trade_executor.monitor_positions()
        )
    
    async def stop(self) -> None:
        """停止交易系统"""
        logger.info("Stopping trading system...")
        self.is_running = False
        
        # 关闭所有持仓
        for symbol in self.market_data_service.config.symbols:
            current_price = self.market_data_service.get_latest_price(symbol)
            if current_price and symbol in self.trade_executor.positions:
                await self.trade_executor._close_position(
                    self.trade_executor.positions[symbol],
                    current_price,
                    "SYSTEM_SHUTDOWN"
                )
        
        # 停止市场数据服务
        await self.market_data_service.stop()
    
    async def _analysis_loop(self) -> None:
        """市场分析循环"""
        logger.debug("[TradingSystem] Starting market analysis loop")
        while self.is_running:
            try:
                for symbol in self.market_data_service.config.symbols:
                    logger.debug(f"[TradingSystem] Starting analysis cycle for {symbol}")
                    
                    # 获取每个Agent需要的时间周期数据
                    for agent in self.agent_system.agents.values():
                        logger.debug(f"[TradingSystem] Running agent {agent.config.name} for {symbol}")
                        timeframe = agent.config.timeframe
                        try:
                            data = self.market_data_service.get_ohlcv(symbol, timeframe)
                            if data.empty:
                                logger.warning(f"No OHLCV data available for {symbol} on {timeframe}")
                                continue
                                
                            logger.info(f"Retrieved OHLCV data for {symbol} ({len(data)} rows)")
                            data['symbol'] = symbol
                            
                            # 运行Agent分析
                            logger.info(f"Running {agent.config.name} analysis for {symbol}")
                            signal = agent.analyze(data)
                            
                            if not signal:
                                logger.info(f"No signal generated for {symbol}")
                                continue
                                
                            if signal.confidence < agent.config.confidence_threshold:
                                logger.info(f"Signal confidence {signal.confidence} below threshold for {symbol}")
                                continue
                                
                            logger.info(f"Got valid signal for {symbol} with confidence {signal.confidence}")
                            
                            # 获取市场数据
                            current_price = self.market_data_service.get_latest_price(symbol)
                            if current_price is None:
                                logger.error(f"Failed to get latest price for {symbol}")
                                continue
                                
                            orderbook = self.market_data_service.get_market_depth(symbol)
                            if not orderbook or not orderbook.get('bids'):
                                logger.error(f"Failed to get valid orderbook for {symbol}")
                                continue
                                
                            logger.info(f"Market data received - Price: {current_price}, Orderbook depth: {len(orderbook['bids'])}")
                            
                            # 获取风险指标
                            volatility = self.market_data_service.get_volatility(symbol)
                            atr = self.market_data_service.get_atr(symbol)
                            logger.info(f"Risk metrics - Volatility: {volatility}, ATR: {atr}")
                            
                            # 处理交易信号
                            try:
                                adjusted_price = self._optimize_execution_price(
                                    signal.direction,
                                    current_price,
                                    orderbook
                                )
                                logger.info(f"Optimized price for {symbol}: {adjusted_price}")
                                
                                await self.trade_executor.process_signal(
                                    signal,
                                    adjusted_price
                                )
                                logger.info(f"Successfully processed trade signal for {symbol}")
                            except Exception as e:
                                logger.error(f"Error processing trade for {symbol}: {str(e)}")
                                continue
                                
                        except Exception as e:
                            logger.error(f"Error in analysis cycle for {symbol}: {str(e)}")
                            continue
                            
                self.last_update = datetime.now()
                await asyncio.sleep(1)  # 每秒分析一次
                
            except Exception as e:
                logger.error(f"Error in analysis loop: {str(e)}")
                await asyncio.sleep(5)
    
    def _optimize_execution_price(self, direction: str, 
                                current_price: float, 
                                orderbook: Dict) -> float:
        """优化执行价格"""
        if not orderbook or not orderbook['bids'] or not orderbook['asks']:
            return current_price
            
        if direction == 'buy':
            # 买入时，检查卖单深度，找到合适的价格点
            for price, volume in orderbook['asks']:
                if volume >= 0.1:  # 假设最小交易量为0.1
                    return min(price * 1.001, current_price * 1.002)  # 加上0.1%的滑点
        else:
            # 卖出时，检查买单深度，找到合适的价格点
            for price, volume in orderbook['bids']:
                if volume >= 0.1:
                    return max(price * 0.999, current_price * 0.998)  # 减去0.1%的滑点
        
        return current_price
    
    async def process_signal(self, signal: TradeSignal):
        """处理交易信号"""
        symbol = signal.symbol
        direction = signal.direction
        confidence = signal.confidence
        
        # 获取市场数据
        current_price = self.market_data_service.get_latest_price(symbol)
        if current_price is None:
            logger.error(f"Cannot process signal: no current price for {symbol}")
            return
            
        volatility = self.market_data_service.get_volatility(symbol)
        if volatility is None:
            logger.error(f"Cannot process signal: no volatility data for {symbol}")
            return
            
        atr = self.market_data_service.get_atr(symbol)
        if atr is None:
            logger.error(f"Cannot process signal: no ATR data for {symbol}")
            return
            
        logger.info(f"Processing signal for {symbol} - Price: {current_price}, Vol: {volatility}, ATR: {atr}")
        
        # 计算止损止盈价格
        stop_loss_mult = float(self.risk_manager.config.stop_loss_atr)
        take_profit_mult = float(self.risk_manager.config.take_profit_atr)
        
        if direction == 'buy':
            stop_loss = current_price - (atr * stop_loss_mult)
            take_profit = current_price + (atr * take_profit_mult)
        else:
            stop_loss = current_price + (atr * stop_loss_mult)
            take_profit = current_price - (atr * take_profit_mult)
            
        logger.info(f"Calculated prices - Stop: {stop_loss}, Take: {take_profit}")
        
        # 计算建议仓位大小
        try:
            position_size = self.risk_manager.calculate_position_size(
                float(current_price), float(stop_loss), float(volatility)
            )
            logger.info(f"Calculated position size: {position_size}")
        except Exception as e:
            logger.error(f"Failed to calculate position size: {str(e)}")
            return
            
        # 检查风险限制
        try:
            if not self.risk_manager.check_position_risk(
                symbol, direction, float(position_size), float(current_price)
            ):
                logger.warning(f"Risk check failed for {symbol}")
                return
            logger.info(f"Risk check passed for {symbol}")
        except Exception as e:
            logger.error(f"Failed risk check: {str(e)}")
            return
        
        # 创建交易信号
        try:
            signal = TradeSignal(
                symbol=symbol,
                direction=direction,
                size=position_size,
                stop_loss=stop_loss,
                take_profit=take_profit,
                confidence=confidence,
                agent_name="SYSTEM",
                price=current_price,
                timestamp=datetime.now(),
                metadata={
                    'confidence': confidence,
                    'volatility': volatility,
                    'atr': atr,
                    'margin_used': self.risk_manager._calculate_margin(position_size, current_price)
                }
            )
            await self.trade_executor.process_signal(signal, current_price)
            logger.info(f"New position opened: {symbol}")
        except Exception as e:
            logger.error(f"Failed to execute order: {e}")
    
    async def update(self):
        """更新系统状态"""
        # 更新市场数据
        market_data = {
            symbol: {
                'price': self.market_data_service.get_latest_price(symbol)
            }
            for symbol in self.market_data_service.config.symbols
        }
        
        # 更新风险管理系统
        self.risk_manager.update_positions(market_data)
        
        # 检查风险指标
        risk_report = self.risk_manager.get_risk_report()
        if risk_report['portfolio_summary']['drawdown'] > self.risk_manager.config.max_drawdown:
            logger.warning("Maximum drawdown exceeded")
            await self._handle_risk_event('max_drawdown_exceeded')
        
        if risk_report['portfolio_summary']['daily_pnl'] < -self.risk_manager.config.daily_loss_limit:
            logger.warning("Daily loss limit exceeded")
            await self._handle_risk_event('daily_loss_limit_exceeded')
    
    async def _handle_risk_event(self, event_type: str):
        """处理风险事件"""
        if event_type in ['max_drawdown_exceeded', 'daily_loss_limit_exceeded']:
            # 关闭所有仓位
            for symbol, position in list(self.risk_manager.positions.items()):
                try:
                    current_price = self.market_data_service.get_latest_price(symbol)
                    if current_price:
                        if isinstance(position, (Position, ContractPosition)):
                            try:
                                await self.trade_executor._close_position(position, current_price, "RISK_EVENT")
                            except TypeError:
                                logger.error(f"Type mismatch closing position for {symbol}")
                            except Exception as e:
                                logger.error(f"Error closing position for {symbol}: {e}")
                except Exception as e:
                    logger.error(f"Failed to handle position {symbol}: {e}")
            
            # 暂停交易
            self.is_running = False
            logger.info("Trading disabled due to risk event")
    
    def get_system_status(self) -> Dict:
        """获取系统状态"""
        status = {
            'is_running': self.is_running,
            'last_update': self.last_update,
            'active_agents': len(self.agent_system.agents),
            'open_positions': len(self.trade_executor.positions),
            'performance_metrics': self.trade_executor.get_performance_metrics(),
            'agent_metrics': self.agent_system.get_system_metrics(),
            'market_data': {
                symbol: {
                    'last_price': self.market_data_service.get_latest_price(symbol),
                    'spread': self.market_data_service.get_market_depth(symbol)['asks'][0][0] - self.market_data_service.get_market_depth(symbol)['bids'][0][0],
                    'vwap': self.market_data_service.calculate_vwap(symbol)
                }
                for symbol in self.market_data_service.config.symbols
            }
        }
        
        # 添加风险管理状态
        risk_report = self.risk_manager.get_risk_report()
        status['risk_management'] = {
            'portfolio_summary': risk_report['portfolio_summary'],
            'risk_metrics': risk_report['risk_metrics'],
            'positions': self.risk_manager.get_position_summary().to_dict('records'),
            'trading_enabled': self.is_running
        }
        
        return status

async def main():
    """主函数"""
    # 创建交易系统实例
    system = TradingSystem()
    
    try:
        # 启动系统
        await system.start()
        
        # 保持运行直到收到停止信号
        while True:
            await asyncio.sleep(1)
            
    except KeyboardInterrupt:
        logger.info("Received shutdown signal")
        await system.stop()
    
    except Exception as e:
        logger.error(f"System error: {str(e)}")
        await system.stop()

if __name__ == "__main__":
    # 运行主程序
    asyncio.run(main())                                        