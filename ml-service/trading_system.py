import asyncio
import logging
from typing import Dict, List
from datetime import datetime
from agent_system import AgentSystem, AgentConfig
from trade_executor import TradeExecutor
from market_data_service import MarketDataService, MarketConfig
from risk_management import RiskManager, RiskConfig, Position
from config import Config
import pandas as pd
from ml_agent import DeepSeekAgent

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
            symbols=market_config['symbols'],
            timeframes=market_config['timeframes'],
            api_key=exchange_config['api_key'],
            api_secret=exchange_config['api_secret'],
            cache_size=market_config['cache_size'],
            update_interval=market_config['update_interval'],
            db_path=market_config['db_path']
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
        while self.is_running:
            try:
                for symbol in self.market_data_service.config.symbols:
                    # 获取每个Agent需要的时间周期数据
                    for agent in self.agent_system.agents.values():
                        timeframe = agent.config.timeframe
                        data = self.market_data_service.get_ohlcv(symbol, timeframe)
                        
                        if data.empty:
                            continue
                        
                        # 添加symbol信息到数据中
                        data['symbol'] = symbol
                        
                        # 运行Agent分析
                        signal = await agent.analyze(data)
                        
                        if signal and signal.confidence >= agent.config.confidence_threshold:
                            # 获取最新价格和订单簿数据
                            current_price = self.market_data_service.get_latest_price(symbol)
                            orderbook = self.market_data_service.get_latest_orderbook(symbol)
                            
                            if current_price:
                                # 使用订单簿数据优化执行价格
                                adjusted_price = self._optimize_execution_price(
                                    signal.direction,
                                    current_price,
                                    orderbook
                                )
                                
                                # 执行交易
                                await self.trade_executor.process_signal(
                                    signal,
                                    adjusted_price
                                )
                
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
    
    def process_signal(self, signal: Dict):
        """处理交易信号"""
        symbol = signal['symbol']
        direction = signal['direction']
        confidence = signal.get('confidence', 0.5)
        
        # 获取市场数据
        current_price = self.market_data_service.get_latest_price(symbol)
        volatility = self.market_data_service.get_volatility(symbol)
        atr = self.market_data_service.get_atr(symbol)
        
        # 计算止损止盈价格
        if direction == 'buy':
            stop_loss = current_price - atr * self.risk_manager.config.stop_loss_atr
            take_profit = current_price + atr * self.risk_manager.config.take_profit_atr
        else:
            stop_loss = current_price + atr * self.risk_manager.config.stop_loss_atr
            take_profit = current_price - atr * self.risk_manager.config.take_profit_atr
        
        # 计算建议仓位大小
        position_size = self.risk_manager.calculate_position_size(
            current_price, stop_loss, volatility
        )
        
        # 检查风险限制
        if not self.risk_manager.check_position_risk(
            symbol, direction, position_size, current_price
        ):
            logger.warning(f"Risk check failed for {symbol}")
            return
        
        # 创建新仓位
        position = Position(
            symbol=symbol,
            direction=direction,
            size=position_size,
            entry_price=current_price,
            current_price=current_price,
            stop_loss=stop_loss,
            take_profit=take_profit,
            unrealized_pnl=0.0,
            realized_pnl=0.0,
            margin_used=self.risk_manager._calculate_margin(position_size, current_price),
            timestamp=datetime.now(),
            metadata={
                'confidence': confidence,
                'volatility': volatility,
                'atr': atr
            }
        )
        
        # 执行交易
        try:
            order = self.execute_order(position)
            if order['status'] == 'filled':
                self.risk_manager.positions[symbol] = position
                logger.info(f"New position opened: {symbol}")
        except Exception as e:
            logger.error(f"Failed to execute order: {e}")
    
    def update(self):
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
            self._handle_risk_event('max_drawdown_exceeded')
        
        if risk_report['portfolio_summary']['daily_pnl'] < -self.risk_manager.config.daily_loss_limit:
            logger.warning("Daily loss limit exceeded")
            self._handle_risk_event('daily_loss_limit_exceeded')
    
    def _handle_risk_event(self, event_type: str):
        """处理风险事件"""
        if event_type in ['max_drawdown_exceeded', 'daily_loss_limit_exceeded']:
            # 关闭所有仓位
            for symbol, position in list(self.risk_manager.positions.items()):
                try:
                    self.close_position(symbol)
                except Exception as e:
                    logger.error(f"Failed to close position {symbol}: {e}")
            
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
                    'spread': self.market_data_service.calculate_spread(symbol),
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