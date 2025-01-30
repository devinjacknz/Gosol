import asyncio
import logging
from typing import Dict, List, Optional
from datetime import datetime
from dataclasses import dataclass
from agent_system import TradeSignal, AgentSystem
from reporting_system import ReportingSystem, ExecutionReport, PerformanceReport
import numpy as np
import pandas as pd
from config import Config

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@dataclass
class Position:
    """持仓信息"""
    symbol: str
    direction: str
    size: float
    entry_price: float
    stop_loss: float
    take_profit: float
    agent_name: str
    open_time: datetime
    metadata: Dict[str, any]

@dataclass
class Trade:
    """交易记录"""
    symbol: str
    direction: str
    size: float
    entry_price: float
    exit_price: float
    stop_loss: float
    take_profit: float
    agent_name: str
    open_time: datetime
    close_time: datetime
    pnl: float
    metadata: Dict[str, any]

class TradeExecutor:
    """交易执行系统"""
    
    def __init__(self, agent_system: AgentSystem):
        self.agent_system = agent_system
        self.positions: Dict[str, Position] = {}  # symbol -> Position
        self.trades: List[Trade] = []
        self.max_positions = 5
        self.max_risk_per_trade = 0.02  # 每笔交易最大风险2%
        
        # 风控参数
        self.max_drawdown = 0.1  # 最大回撤10%
        self.daily_loss_limit = 0.05  # 日亏损限制5%
        self.position_sizing_method = 'risk_based'  # 'risk_based' or 'equal_size'
        
        # 性能跟踪
        self.total_pnl = 0
        self.daily_pnl = 0
        self.max_equity = 0
        self.current_drawdown = 0
        
        # 报告系统
        self.reporting = ReportingSystem()
    
    async def process_signal(self, signal: TradeSignal, current_price: float) -> None:
        """处理交易信号"""
        try:
            # 检查是否是合约交易
            is_contract = signal.metadata.get('contract', False)
            if is_contract:
                await self._process_contract_signal(signal, current_price)
            else:
                await self._process_spot_signal(signal, current_price)
        except Exception as e:
            logger.error(f"Error processing signal: {e}")
            self._handle_error(signal, str(e))

    async def _process_contract_signal(self, signal: TradeSignal, current_price: float) -> None:
        """处理合约交易信号"""
        # 获取合约配置
        contract_config = Config.get_contract_config()
        
        # 验证交易对是否支持合约交易
        if signal.symbol not in contract_config['enabled_pairs']:
            raise ValueError(f"Contract trading not supported for {signal.symbol}")
        
        # 验证杠杆倍数
        leverage = signal.metadata.get('leverage', 1)
        if leverage not in contract_config['leverage_options']:
            raise ValueError(f"Invalid leverage: {leverage}")
        
        # 验证保证金模式
        margin_type = signal.metadata.get('margin_type', 'isolated')
        if margin_type not in contract_config['margin_types']:
            raise ValueError(f"Invalid margin type: {margin_type}")
        
        # 检查是否有足够保证金
        required_margin = self._calculate_required_margin(signal, current_price, leverage)
        if not self._check_margin_requirement(required_margin):
            raise ValueError("Insufficient margin")
        
        # 处理开仓或调整持仓
        position = self._get_position(signal.symbol)
        if position:
            await self._handle_existing_contract_position(signal, current_price, position)
        else:
            await self._open_new_contract_position(signal, current_price)

    async def _open_new_contract_position(self, signal: TradeSignal, current_price: float) -> None:
        """开新的合约仓位"""
        # 计算仓位大小
        size = self._calculate_contract_position_size(signal, current_price)
        
        # 创建合约持仓
        position = ContractPosition(
            symbol=signal.symbol,
            direction=signal.direction,
            size=size,
            entry_price=current_price,
            leverage=signal.metadata['leverage'],
            margin_type=signal.metadata['margin_type'],
            liquidation_price=self._calculate_liquidation_price(signal, current_price),
            maintenance_margin=self._calculate_maintenance_margin(size, current_price),
            funding_rate=self._get_funding_rate(signal.symbol),
            next_funding_time=self._get_next_funding_time(),
            agent_name=signal.agent_name,
            open_time=datetime.now(),
            metadata=signal.metadata
        )
        
        # 保存持仓信息
        await self._save_position(position)
        
        # 更新性能指标
        self._update_performance_metrics({
            'type': 'contract_open',
            'symbol': signal.symbol,
            'size': size,
            'price': current_price,
            'leverage': signal.metadata['leverage']
        })

    def _calculate_liquidation_price(self, signal: TradeSignal, current_price: float) -> float:
        """计算强平价格"""
        leverage = signal.metadata['leverage']
        margin_type = signal.metadata['margin_type']
        maintenance_margin = Config.get_risk_config()['min_maintenance_margin']
        
        if signal.direction == 'buy':
            return current_price * (1 - 1/leverage + maintenance_margin)
        else:
            return current_price * (1 + 1/leverage - maintenance_margin)

    def _calculate_maintenance_margin(self, size: float, price: float) -> float:
        """计算维持保证金"""
        min_rate = Config.get_risk_config()['min_maintenance_margin']
        return size * price * min_rate

    def _get_funding_rate(self, symbol: str) -> float:
        """获取当前资金费率"""
        # 这里应该从交易所API获取实际资金费率
        # 目前使用模拟值
        return 0.0001  # 0.01%/8h

    def _get_next_funding_time(self) -> datetime:
        """获取下次资金费时间"""
        now = datetime.now()
        interval = Config.get_risk_config()['funding_rate_interval']
        next_hour = (now.hour // interval + 1) * interval
        return now.replace(hour=next_hour, minute=0, second=0, microsecond=0)

    def _calculate_required_margin(self, signal: TradeSignal, price: float, leverage: float) -> float:
        """计算所需保证金"""
        return signal.size * price / leverage

    def _check_margin_requirement(self, required_margin: float) -> bool:
        """检查是否满足保证金要求"""
        available_margin = self._get_available_margin()
        return available_margin >= required_margin

    def _get_available_margin(self) -> float:
        """获取可用保证金"""
        # 这里应该从账户系统获取实际可用保证金
        # 目前返回模拟值
        return 100000.0

    async def _process_spot_signal(self, signal: TradeSignal, current_price: float) -> None:
        """处理现货交易信号"""
        try:
            # 检查风控限制
            if not self._check_risk_limits():
                logger.warning("Risk limits reached, cannot take new positions")
                return
            
            # 检查是否已有相同symbol的持仓
            if signal.symbol in self.positions:
                await self._handle_existing_position(signal, current_price)
            else:
                await self._open_new_position(signal, current_price)
                
        except Exception as e:
            logger.error(f"Error processing signal: {str(e)}")
    
    async def _open_new_position(self, signal: TradeSignal, current_price: float) -> None:
        """开新仓位"""
        # 检查是否达到最大持仓数
        if len(self.positions) >= self.max_positions:
            logger.warning("Maximum number of positions reached")
            return
        
        # 调整仓位大小
        adjusted_size = self._calculate_position_size(signal, current_price)
        
        # 创建新持仓
        position = Position(
            symbol=signal.symbol,
            direction=signal.direction,
            size=adjusted_size,
            entry_price=current_price,
            stop_loss=signal.stop_loss,
            take_profit=signal.take_profit,
            agent_name=signal.agent_name,
            open_time=datetime.now(),
            metadata=signal.metadata
        )
        
        # 保存持仓
        self.positions[signal.symbol] = position
        logger.info(f"Opened new position: {position}")
        
        # 创建并保存执行报告
        execution_report = ExecutionReport(
            timestamp=datetime.now(),
            symbol=signal.symbol,
            action='open',
            direction=signal.direction,
            price=current_price,
            size=adjusted_size,
            agent_name=signal.agent_name,
            confidence=signal.confidence,
            reason='SIGNAL',
            metadata={
                **signal.metadata,
                'stop_loss': signal.stop_loss,
                'take_profit': signal.take_profit
            }
        )
        await self.reporting.save_execution_report(execution_report)
    
    async def _handle_existing_position(self, signal: TradeSignal, 
                                      current_price: float) -> None:
        """处理已有持仓"""
        position = self.positions[signal.symbol]
        
        # 如果信号方向与持仓方向相反，则平仓
        if signal.direction != position.direction:
            await self._close_position(position, current_price, "SIGNAL")
        # 如果方向相同，可以考虑调整持仓大小或者更新止损止盈
        else:
            self._update_position_params(position, signal)
            
            # 创建修改报告
            execution_report = ExecutionReport(
                timestamp=datetime.now(),
                symbol=signal.symbol,
                action='modify',
                direction=position.direction,
                price=current_price,
                size=position.size,
                agent_name=signal.agent_name,
                confidence=signal.confidence,
                reason='UPDATE',
                metadata={
                    'old_stop_loss': position.stop_loss,
                    'new_stop_loss': signal.stop_loss,
                    'old_take_profit': position.take_profit,
                    'new_take_profit': signal.take_profit
                }
            )
            await self.reporting.save_execution_report(execution_report)
    
    async def _close_position(self, position: Position, current_price: float,
                            reason: str) -> None:
        """平仓"""
        # 计算盈亏
        pnl = self._calculate_pnl(position, current_price)
        
        # 创建交易记录
        trade = Trade(
            symbol=position.symbol,
            direction=position.direction,
            size=position.size,
            entry_price=position.entry_price,
            exit_price=current_price,
            stop_loss=position.stop_loss,
            take_profit=position.take_profit,
            agent_name=position.agent_name,
            open_time=position.open_time,
            close_time=datetime.now(),
            pnl=pnl,
            metadata={
                **position.metadata,
                'close_reason': reason
            }
        )
        
        # 更新交易历史
        self.trades.append(trade)
        
        # 更新性能指标
        self._update_performance_metrics(trade)
        
        # 从持仓中移除
        del self.positions[position.symbol]
        
        # 更新Agent表现
        self.agent_system.update_agent_performance(
            signal=None,  # 这里需要原始信号
            success=pnl > 0,
            return_pct=pnl / position.entry_price
        )
        
        # 创建平仓报告
        execution_report = ExecutionReport(
            timestamp=datetime.now(),
            symbol=position.symbol,
            action='close',
            direction=position.direction,
            price=current_price,
            size=position.size,
            agent_name=position.agent_name,
            confidence=1.0,
            reason=reason,
            metadata={
                'pnl': pnl,
                'hold_time': (datetime.now() - position.open_time).total_seconds() / 3600,
                'entry_price': position.entry_price
            }
        )
        await self.reporting.save_execution_report(execution_report)
        
        # 保存交易记录
        await self.reporting.save_trade(trade.__dict__)
        
        # 更新性能报告
        performance_report = PerformanceReport(
            timestamp=datetime.now(),
            total_pnl=self.total_pnl,
            daily_pnl=self.daily_pnl,
            total_trades=len(self.trades),
            winning_trades=sum(1 for t in self.trades if t.pnl > 0),
            losing_trades=sum(1 for t in self.trades if t.pnl < 0),
            win_rate=sum(1 for t in self.trades if t.pnl > 0) / len(self.trades),
            avg_profit=np.mean([t.pnl for t in self.trades if t.pnl > 0]),
            avg_loss=np.mean([t.pnl for t in self.trades if t.pnl < 0]),
            max_drawdown=self.current_drawdown,
            sharpe_ratio=self._calculate_sharpe_ratio(),
            agent_metrics=self.agent_system.get_system_metrics(),
            market_metrics={}  # TODO: 添加市场指标
        )
        await self.reporting.save_performance_report(performance_report)
        
        logger.info(f"Closed position: {trade}")
    
    def _calculate_position_size(self, signal: TradeSignal, 
                               current_price: float) -> float:
        """计算仓位大小"""
        if self.position_sizing_method == 'risk_based':
            # 基于风险的仓位计算
            risk_amount = self.max_risk_per_trade * self._get_account_balance()
            price_risk = abs(current_price - signal.stop_loss)
            return risk_amount / price_risk
        else:
            # 等分仓位
            return signal.size
    
    def _calculate_pnl(self, position: Position, current_price: float) -> float:
        """计算盈亏"""
        if position.direction == 'buy':
            return (current_price - position.entry_price) * position.size
        else:
            return (position.entry_price - current_price) * position.size
    
    def _update_position_params(self, position: Position, signal: TradeSignal) -> None:
        """更新持仓参数"""
        position.stop_loss = signal.stop_loss
        position.take_profit = signal.take_profit
        logger.info(f"Updated position parameters: {position}")
    
    def _check_risk_limits(self) -> bool:
        """检查风控限制"""
        # 检查回撤限制
        if self.current_drawdown > self.max_drawdown:
            return False
        
        # 检查日亏损限制
        if self.daily_pnl < -self.daily_loss_limit * self._get_account_balance():
            return False
        
        return True
    
    def _update_performance_metrics(self, trade: Trade) -> None:
        """更新性能指标"""
        self.total_pnl += trade.pnl
        self.daily_pnl += trade.pnl
        
        current_equity = self._get_account_balance()
        self.max_equity = max(self.max_equity, current_equity)
        self.current_drawdown = (self.max_equity - current_equity) / self.max_equity
    
    def _calculate_sharpe_ratio(self) -> float:
        """计算夏普比率"""
        if not self.trades:
            return 0.0
        
        returns = pd.Series([t.pnl / t.entry_price for t in self.trades])
        if returns.std() == 0:
            return 0.0
            
        return np.sqrt(252) * (returns.mean() - 0.02/252) / returns.std()
    
    def _get_account_balance(self) -> float:
        """获取账户余额"""
        # TODO: 实现获取实际账户余额的逻辑
        return 100000.0  # 示例固定值
    
    def get_performance_metrics(self) -> Dict[str, float]:
        """获取性能指标"""
        return {
            'total_pnl': self.total_pnl,
            'daily_pnl': self.daily_pnl,
            'current_drawdown': self.current_drawdown,
            'open_positions': len(self.positions),
            'total_trades': len(self.trades),
            'win_rate': self._calculate_win_rate(),
            'average_trade': self._calculate_average_trade()
        }
    
    def _calculate_win_rate(self) -> float:
        """计算胜率"""
        if not self.trades:
            return 0.0
        winning_trades = sum(1 for trade in self.trades if trade.pnl > 0)
        return winning_trades / len(self.trades)
    
    def _calculate_average_trade(self) -> float:
        """计算平均每笔交易盈亏"""
        if not self.trades:
            return 0.0
        return sum(trade.pnl for trade in self.trades) / len(self.trades)
    
    async def monitor_positions(self) -> None:
        """监控持仓"""
        while True:
            try:
                for symbol, position in list(self.positions.items()):
                    current_price = await self._get_current_price(symbol)
                    
                    # 检查止损止盈
                    if self._check_stop_loss(position, current_price):
                        await self._close_position(position, current_price, "STOP_LOSS")
                    elif self._check_take_profit(position, current_price):
                        await self._close_position(position, current_price, "TAKE_PROFIT")
                    
                await asyncio.sleep(1)  # 每秒检查一次
                
            except Exception as e:
                logger.error(f"Error monitoring positions: {str(e)}")
                await asyncio.sleep(5)  # 出错后等待5秒再试
    
    def _check_stop_loss(self, position: Position, current_price: float) -> bool:
        """检查是否触发止损"""
        if position.direction == 'buy':
            return current_price <= position.stop_loss
        else:
            return current_price >= position.stop_loss
    
    def _check_take_profit(self, position: Position, current_price: float) -> bool:
        """检查是否触发止盈"""
        if position.direction == 'buy':
            return current_price >= position.take_profit
        else:
            return current_price <= position.take_profit
    
    async def _get_current_price(self, symbol: str) -> float:
        """获取当前价格"""
        # TODO: 实现获取实时价格的逻辑
        return 0.0  # 示例返回值 