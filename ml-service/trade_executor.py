import asyncio
import logging
import aiohttp
import time
import psutil
from typing import Dict, List, Optional, Any, Union, TypeVar, cast
from datetime import datetime
from dataclasses import dataclass
from ml_service.agent_system import TradeSignal, AgentSystem
from ml_service.config import Config
from ml_service.reporting_system import ReportingSystem, ExecutionReport, PerformanceReport
import numpy as np
import pandas as pd

logger = logging.getLogger(__name__)

@dataclass
class Position:
    """持仓信息"""
    symbol: str
    direction: str
    size: float
    entry_price: float
    stop_loss: Optional[float]
    take_profit: Optional[float]
    agent_name: str
    open_time: datetime
    metadata: Dict[str, Any]

@dataclass
class ContractPosition:
    """合约持仓信息"""
    symbol: str
    direction: str
    size: float
    entry_price: float
    leverage: float
    margin_type: str
    liquidation_price: float
    maintenance_margin: float
    funding_rate: float
    next_funding_time: datetime
    agent_name: str
    open_time: datetime
    metadata: Dict[str, Any]
    stop_loss: Optional[float] = None
    take_profit: Optional[float] = None

T = TypeVar('T', bound=Union[Position, ContractPosition])

def create_trade_from_dict(data: Dict[str, Any]) -> 'Trade':
    """Create a Trade object from dictionary data"""
    return Trade(
        symbol=data['symbol'],
        direction=data.get('direction', 'buy'),
        size=float(data['size']),
        entry_price=float(data['price']),
        exit_price=0.0,
        stop_loss=data.get('stop_loss'),
        take_profit=data.get('take_profit'),
        agent_name=data.get('agent_name', 'system'),
        open_time=datetime.now(),
        close_time=datetime.now(),
        pnl=0.0,
        metadata=data
    )
from ml_service.reporting_system import ReportingSystem, ExecutionReport, PerformanceReport
import numpy as np
import pandas as pd
from ml_service.config import Config
from ml_service.exchanges.hyperliquid_client import HyperliquidClient
from ml_service.exchanges.dydx_client import DydxClient

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

@dataclass
class Trade:
    """交易记录"""
    symbol: str
    direction: str
    size: float
    entry_price: float
    exit_price: float
    stop_loss: Optional[float]
    take_profit: Optional[float]
    agent_name: str
    open_time: datetime
    close_time: datetime
    pnl: float
    metadata: Dict[str, Any]

class TradeExecutor:
    """交易执行系统"""

    def __init__(self, agent_system: AgentSystem):
        self.agent_system = agent_system
        self.positions: Dict[str, Union[Position, ContractPosition]] = {}  # symbol -> Position
        self.trades: List[Trade] = []
        self.max_positions = 5
        self.max_risk_per_trade = 0.02  # 每笔交易最大风险2%
        self.hyperliquid_client = HyperliquidClient()
        self.dydx_client = DydxClient()

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
            await self.reporting.save_error_report({
                'timestamp': datetime.now(),
                'error': str(e),
                'signal': signal.__dict__,
                'type': 'SIGNAL_PROCESSING_ERROR'
            })

    async def _process_contract_signal(self, signal: TradeSignal, current_price: float) -> None:
        """处理合约交易信号"""
        try:
            if not isinstance(signal, TradeSignal):
                raise TypeError("Invalid signal type")
            if not isinstance(current_price, (int, float)) or current_price <= 0:
                raise ValueError(f"Invalid current price: {current_price}")
                
            contract_config = Config.get_contract_config()
            if not isinstance(contract_config, dict):
                raise RuntimeError("Invalid contract configuration")

            if not signal.symbol or signal.symbol not in contract_config.get('enabled_pairs', []):
                raise ValueError(f"Contract trading not supported for {signal.symbol}")

            leverage = signal.metadata.get('leverage', 1)
            if not isinstance(leverage, (int, float)) or leverage <= 0:
                raise ValueError(f"Invalid leverage value: {leverage}")
            if leverage not in contract_config.get('leverage_options', []):
                raise ValueError(f"Unsupported leverage option: {leverage}")

            margin_type = signal.metadata.get('margin_type', 'isolated')
            if not isinstance(margin_type, str) or margin_type not in contract_config.get('margin_types', []):
                raise ValueError(f"Invalid margin type: {margin_type}")

            required_margin = self._calculate_required_margin(signal, current_price, leverage)
            if not isinstance(required_margin, (int, float)) or required_margin <= 0:
                raise ValueError(f"Invalid required margin: {required_margin}")
            if not self._check_margin_requirement(required_margin):
                raise ValueError(f"Insufficient margin. Required: {required_margin}")

            position = self.positions.get(signal.symbol)
            if position:
                if isinstance(position, ContractPosition):
                    await self._update_contract_position(position, signal, current_price)
                else:
                    logger.warning(f"Invalid position type for contract trading: {type(position)}")
                    await self._close_position(position, current_price, "INVALID_POSITION_TYPE")
            else:
                await self._open_new_contract_position(signal, current_price)
                
            logger.info(f"Successfully processed contract signal for {signal.symbol}")
        except (TypeError, ValueError) as e:
            logger.error(f"Validation error processing contract signal: {str(e)}")
            raise
        except Exception as e:
            logger.error(f"Unexpected error processing contract signal: {str(e)}", exc_info=True)
            raise

    async def _open_new_contract_position(self, signal: TradeSignal, current_price: float) -> None:
        """开新的合约仓位"""
        try:
            if not isinstance(signal, TradeSignal):
                raise TypeError("Invalid signal type")
            if not isinstance(current_price, (int, float)) or current_price <= 0:
                raise ValueError(f"Invalid current price: {current_price}")

            # 计算仓位大小并验证
            size = self._calculate_position_size(signal, current_price)
            if not isinstance(size, (int, float)) or size <= 0:
                raise ValueError(f"Invalid position size calculated: {size}")

            # 验证必要的元数据
            leverage = signal.metadata.get('leverage')
            margin_type = signal.metadata.get('margin_type')
            if not isinstance(leverage, (int, float)) or leverage <= 0:
                raise ValueError(f"Invalid leverage: {leverage}")
            if not isinstance(margin_type, str) or margin_type not in ['isolated', 'cross']:
                raise ValueError(f"Invalid margin type: {margin_type}")

            # 计算并验证关键参数
            liquidation_price = self._calculate_liquidation_price(signal, current_price)
            maintenance_margin = self._calculate_maintenance_margin(size, current_price)
            funding_rate = self._get_funding_rate(signal.symbol)
            next_funding_time = self._get_next_funding_time()

            if not isinstance(liquidation_price, (int, float)) or liquidation_price <= 0:
                raise ValueError(f"Invalid liquidation price: {liquidation_price}")
            if not isinstance(maintenance_margin, (int, float)) or maintenance_margin <= 0:
                raise ValueError(f"Invalid maintenance margin: {maintenance_margin}")

            # 创建合约持仓
            position = ContractPosition(
                symbol=signal.symbol,
                direction=signal.direction,
                size=size,
                entry_price=current_price,
                leverage=leverage,
                margin_type=margin_type,
                liquidation_price=liquidation_price,
                maintenance_margin=maintenance_margin,
                funding_rate=funding_rate,
                next_funding_time=next_funding_time,
                agent_name=signal.agent_name,
                open_time=datetime.now(),
                metadata=signal.metadata
            )

            # 保存持仓信息
            self.positions[position.symbol] = position
            logger.info(f"Successfully opened new contract position: {position}")

            # 更新性能指标
            trade = {
                'type': 'contract_open',
                'symbol': signal.symbol,
                'size': size,
                'price': current_price,
                'leverage': leverage,
                'margin_type': margin_type,
                'liquidation_price': liquidation_price,
                'maintenance_margin': maintenance_margin,
                'funding_rate': funding_rate
            }
            self._update_performance_metrics(trade)
            
        except (TypeError, ValueError) as e:
            logger.error(f"Validation error opening contract position: {str(e)}")
            raise
        except Exception as e:
            logger.error(f"Unexpected error opening contract position: {str(e)}", exc_info=True)
            raise

    async def _update_contract_position(self, position: ContractPosition, signal: TradeSignal, current_price: float) -> None:
        try:
            if not isinstance(position, ContractPosition):
                raise TypeError(f"Invalid position type: {type(position)}")
            if not isinstance(signal, TradeSignal):
                raise TypeError(f"Invalid signal type: {type(signal)}")
            if not isinstance(current_price, (int, float)) or current_price <= 0:
                raise ValueError(f"Invalid current price: {current_price}")

            if signal.direction != position.direction:
                await self._close_position(position, current_price, "SIGNAL_DIRECTION_CHANGE")
                await self._open_new_contract_position(signal, current_price)
            else:
                position.stop_loss = signal.stop_loss
                position.take_profit = signal.take_profit
                position.funding_rate = await self._get_funding_rate(signal.symbol)
                position.next_funding_time = self._get_next_funding_time()
                position.metadata.update(signal.metadata or {})
                
                # Update liquidation price if leverage or margin type changed
                if signal.metadata.get('leverage') != position.leverage or signal.metadata.get('margin_type') != position.margin_type:
                    position.liquidation_price = self._calculate_liquidation_price(signal, current_price)
                    position.maintenance_margin = self._calculate_maintenance_margin(position.size, current_price)
                
                logger.info(f"Updated contract position: {position}")
        except Exception as e:
            logger.error(f"Error updating contract position: {e}", exc_info=True)
            raise

    def _calculate_liquidation_price(self, signal: TradeSignal, current_price: float) -> float:
        """计算强平价格"""
        try:
            if not isinstance(signal, TradeSignal):
                raise TypeError("Invalid signal type")
            if not isinstance(current_price, (int, float)) or current_price <= 0:
                raise ValueError(f"Invalid current price: {current_price}")
                
            leverage = signal.metadata.get('leverage')
            if not isinstance(leverage, (int, float)) or leverage <= 0:
                raise ValueError(f"Invalid leverage: {leverage}")
                
            margin_type = signal.metadata.get('margin_type')
            if not isinstance(margin_type, str) or margin_type not in ['isolated', 'cross']:
                raise ValueError(f"Invalid margin type: {margin_type}")
                
            risk_config = Config.get_risk_config()
            if not isinstance(risk_config, dict):
                raise RuntimeError("Invalid risk configuration")
                
            maintenance_margin = risk_config.get('min_maintenance_margin')
            if not isinstance(maintenance_margin, (int, float)) or maintenance_margin <= 0:
                raise ValueError(f"Invalid maintenance margin: {maintenance_margin}")

            if signal.direction == 'buy':
                liquidation_price = current_price * (1 - 1/leverage + maintenance_margin)
            else:
                liquidation_price = current_price * (1 + 1/leverage - maintenance_margin)
                
            if liquidation_price <= 0:
                raise ValueError(f"Invalid liquidation price calculated: {liquidation_price}")
                
            return float(liquidation_price)
        except Exception as e:
            logger.error(f"Error calculating liquidation price: {e}", exc_info=True)
            raise

    def _calculate_maintenance_margin(self, size: float, price: float) -> float:
        """计算维持保证金"""
        try:
            if not isinstance(size, (int, float)) or size <= 0:
                raise ValueError(f"Invalid size: {size}")
            if not isinstance(price, (int, float)) or price <= 0:
                raise ValueError(f"Invalid price: {price}")
                
            risk_config = Config.get_risk_config()
            min_rate = risk_config.get('min_maintenance_margin')
            if not isinstance(min_rate, (int, float)) or min_rate <= 0:
                logger.warning("Invalid maintenance margin rate, using default 0.05")
                min_rate = 0.05
                
            return float(size * price * min_rate)
        except Exception as e:
            logger.error(f"Error calculating maintenance margin: {e}")
            return float(size * price * 0.05)

    async def _get_funding_rate(self, symbol: str) -> float:
        """获取当前资金费率"""
        try:
            if not isinstance(symbol, str) or not symbol:
                raise ValueError("Invalid symbol")
            
            funding_rate = None
            
            try:
                funding_rate = await self.hyperliquid_client.get_funding_rate(symbol)
            except Exception as e:
                logger.warning(f"Failed to get funding rate from Hyperliquid: {e}")
            
            if not funding_rate:
                try:
                    funding_rate = await self.dydx_client.get_funding_rate(symbol)
                except Exception as e:
                    logger.warning(f"Failed to get funding rate from dYdX: {e}")
            
            if funding_rate and isinstance(funding_rate, (int, float)):
                return float(funding_rate)
                
            return 0.0001  # Default fallback value (0.01%/8h)
        except Exception as e:
            logger.error(f"Error getting funding rate: {e}")
            return 0.0001

    def _get_next_funding_time(self) -> datetime:
        """获取下次资金费时间"""
        try:
            now = datetime.now()
            risk_config = Config.get_risk_config()
            
            if not isinstance(risk_config, dict):
                logger.error("Invalid risk configuration")
                interval = 8  # Default 8-hour interval
            else:
                interval = risk_config.get('funding_rate_interval', 8)
                if not isinstance(interval, int) or interval <= 0:
                    logger.warning(f"Invalid funding rate interval: {interval}, using default 8 hours")
                    interval = 8
            
            next_hour = (now.hour // interval + 1) * interval
            if next_hour >= 24:
                next_hour = next_hour % 24
                
            return now.replace(hour=next_hour, minute=0, second=0, microsecond=0)
        except Exception as e:
            logger.error(f"Error calculating next funding time: {e}")
            return now.replace(hour=((now.hour + 8) % 24), minute=0, second=0, microsecond=0)

    def _calculate_required_margin(self, signal: TradeSignal, price: float, leverage: float) -> float:
        """计算所需保证金"""
        try:
            if not isinstance(signal, TradeSignal):
                raise TypeError(f"Invalid signal type: {type(signal)}")
            if not isinstance(signal.size, (int, float)) or signal.size <= 0:
                raise ValueError(f"Invalid position size: {signal.size}")
            if not isinstance(price, (int, float)) or price <= 0:
                raise ValueError(f"Invalid price: {price}")
            if not isinstance(leverage, (int, float)) or leverage <= 0:
                raise ValueError(f"Invalid leverage: {leverage}")
                
            required_margin = float(signal.size * price / leverage)
            if required_margin <= 0:
                raise ValueError(f"Invalid required margin calculation result: {required_margin}")
                
            return required_margin
        except Exception as e:
            logger.error(f"Error calculating required margin: {e}", exc_info=True)
            raise

    def _check_margin_requirement(self, required_margin: float) -> bool:
        """检查是否满足保证金要求"""
        try:
            if not isinstance(required_margin, (int, float)):
                raise ValueError(f"Invalid required margin type: {type(required_margin)}")
            if required_margin < 0:
                raise ValueError(f"Required margin cannot be negative: {required_margin}")
                
            available_margin = self._get_available_margin()
            if not isinstance(available_margin, (int, float)):
                raise ValueError(f"Invalid available margin type: {type(available_margin)}")
            if available_margin < 0:
                raise ValueError(f"Available margin cannot be negative: {available_margin}")
                
            return float(available_margin) >= float(required_margin)
        except Exception as e:
            logger.error(f"Error checking margin requirement: {e}", exc_info=True)
            return False

    def _get_available_margin(self) -> float:
        """获取可用保证金"""
        try:
            # TODO: 实现从账户系统获取实际可用保证金的逻辑
            account_balance = self._get_account_balance()
            if not isinstance(account_balance, (int, float)):
                raise ValueError(f"Invalid account balance type: {type(account_balance)}")
            if account_balance < 0:
                raise ValueError(f"Account balance cannot be negative: {account_balance}")
                
            # 计算已用保证金
            used_margin = 0.0
            for position in self.positions.values():
                if isinstance(position, ContractPosition):
                    try:
                        position_value = position.size * position.entry_price
                        position_margin = position_value / position.leverage
                        used_margin += float(position_margin)
                    except Exception as e:
                        logger.error(f"Error calculating position margin: {e}")
                        continue
                        
            available_margin = float(account_balance - used_margin)
            if available_margin < 0:
                logger.warning(f"Available margin is negative: {available_margin}")
                return 0.0
                
            return available_margin
        except Exception as e:
            logger.error(f"Error getting available margin: {e}", exc_info=True)
            return 0.0

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

    async def _close_position(self, position: Union[Position, ContractPosition], current_price: float,
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
        try:
            if not isinstance(signal, TradeSignal):
                raise TypeError(f"Invalid signal type: {type(signal)}")
            if not isinstance(current_price, (int, float)) or current_price <= 0:
                raise ValueError(f"Invalid current price: {current_price}")
            if not isinstance(self.max_risk_per_trade, (int, float)) or not (0 < self.max_risk_per_trade < 1):
                raise ValueError(f"Invalid max risk per trade: {self.max_risk_per_trade}")

            if not signal.stop_loss:
                if not isinstance(signal.size, (int, float)) or signal.size <= 0:
                    raise ValueError(f"Invalid signal size: {signal.size}")
                logger.warning("No stop loss provided, using default position size")
                return float(signal.size)

            if not isinstance(signal.stop_loss, (int, float)) or signal.stop_loss <= 0:
                raise ValueError(f"Invalid stop loss: {signal.stop_loss}")

            if self.position_sizing_method == 'risk_based':
                try:
                    account_balance = self._get_account_balance()
                    if account_balance <= 0:
                        raise ValueError(f"Invalid account balance: {account_balance}")

                    risk_amount = float(self.max_risk_per_trade * account_balance)
                    price_risk = abs(current_price - signal.stop_loss)
                    
                    if price_risk <= 0:
                        raise ValueError(f"Invalid price risk: {price_risk}")
                        
                    position_size = risk_amount / price_risk
                    max_position_size = account_balance * 0.5
                    
                    if position_size <= 0:
                        raise ValueError(f"Invalid position size calculation: {position_size}")
                    if max_position_size <= 0:
                        raise ValueError(f"Invalid max position size: {max_position_size}")
                        
                    final_size = min(position_size, max_position_size)
                    
                    if isinstance(signal, ContractPosition):
                        leverage = signal.metadata.get('leverage', 1)
                        if not isinstance(leverage, (int, float)) or leverage <= 0:
                            raise ValueError(f"Invalid leverage: {leverage}")
                        final_size = final_size * leverage
                        
                    return float(final_size)
                except Exception as e:
                    logger.error(f"Error in risk-based position sizing: {e}")
                    if not isinstance(signal.size, (int, float)) or signal.size <= 0:
                        raise ValueError(f"Invalid fallback signal size: {signal.size}")
                    return float(signal.size)
            else:
                if not isinstance(signal.size, (int, float)) or signal.size <= 0:
                    raise ValueError(f"Invalid signal size: {signal.size}")
                return float(signal.size)
        except Exception as e:
            logger.error(f"Critical error calculating position size: {e}", exc_info=True)
            return 0.0

    def _calculate_pnl(self, position: Union[Position, ContractPosition], current_price: float) -> float:
        """计算盈亏，包含资金费率影响"""
        try:
            if not isinstance(position, (Position, ContractPosition)):
                raise TypeError(f"Invalid position type: {type(position)}")
            if not isinstance(current_price, (int, float)) or current_price <= 0:
                raise ValueError(f"Invalid current price: {current_price}")
            if not isinstance(position.size, (int, float)) or position.size <= 0:
                raise ValueError(f"Invalid position size: {position.size}")
            if not isinstance(position.entry_price, (int, float)) or position.entry_price <= 0:
                raise ValueError(f"Invalid entry price: {position.entry_price}")
            if not isinstance(position.direction, str) or position.direction not in ['buy', 'sell']:
                raise ValueError(f"Invalid position direction: {position.direction}")

            # 计算基础盈亏
            multiplier = 1.0
            funding_pnl = 0.0
            
            if isinstance(position, ContractPosition):
                try:
                    # 验证并获取杠杆率
                    leverage = position.leverage
                    if not isinstance(leverage, (int, float)) or leverage <= 0:
                        logger.warning(f"Invalid leverage value: {leverage}, using default 1.0")
                    else:
                        multiplier = float(leverage)
                        if multiplier > 100:  # 最大杠杆率检查
                            logger.warning(f"Unusually high leverage: {multiplier}x")
                    
                    # 验证并计算资金费率影响
                    if hasattr(position, 'funding_rate') and hasattr(position, 'open_time'):
                        try:
                            funding_rate = float(position.funding_rate)
                            time_held = (datetime.now() - position.open_time).total_seconds() / 3600.0
                            
                            if time_held > 0:
                                # 计算资金费率间隔（每8小时一次）
                                funding_intervals = time_held / 8.0
                                notional_value = float(position.size * current_price)
                                
                                # 验证资金费率合理性
                                if abs(funding_rate) > 0.01:  # 1%最大资金费率
                                    logger.warning(
                                        f"High funding rate detected:\n"
                                        f"Symbol: {position.symbol}\n"
                                        f"Rate: {funding_rate:.4%}\n"
                                        f"Time Held: {time_held:.1f}h"
                                    )
                                
                                # 计算资金费用
                                funding_multiplier = 1 if position.direction == 'buy' else -1
                                funding_pnl = -(funding_rate * funding_intervals * notional_value * funding_multiplier)
                                
                                # 验证资金费用合理性
                                if not isinstance(funding_pnl, (int, float)):
                                    logger.error("Invalid funding PnL calculation result")
                                    funding_pnl = 0.0
                                elif abs(funding_pnl) > notional_value * 0.1:
                                    logger.warning(
                                        f"Large funding PnL detected:\n"
                                        f"Symbol: {position.symbol}\n"
                                        f"Funding PnL: {funding_pnl}\n"
                                        f"Notional Value: {notional_value}\n"
                                        f"Percentage: {(funding_pnl/notional_value*100):.2f}%"
                                    )
                                    
                                # 记录资金费用统计
                                logger.info(
                                    f"Funding fee calculation:\n"
                                    f"Symbol: {position.symbol}\n"
                                    f"Rate: {funding_rate:.4%}\n"
                                    f"Intervals: {funding_intervals:.2f}\n"
                                    f"PnL: {funding_pnl:.2f}"
                                )
                        except (TypeError, ValueError) as e:
                            logger.error(f"Error in funding calculation: {e}")
                            funding_pnl = 0.0
                    else:
                        logger.warning(
                            f"Missing funding data for {position.symbol}:\n"
                            f"Has Rate: {hasattr(position, 'funding_rate')}\n"
                            f"Has Time: {hasattr(position, 'open_time')}"
                        )
                except Exception as e:
                    logger.error(f"Error calculating contract PnL: {e}", exc_info=True)
                    multiplier = 1.0
                    funding_pnl = 0.0

            # 计算价格变动盈亏
            try:
                # 验证价格差异的合理性
                price_diff = current_price - position.entry_price if position.direction == 'buy' \
                            else position.entry_price - current_price
                
                if abs(price_diff) > position.entry_price * 0.5:  # 价格变动超过50%
                    logger.warning(
                        f"Large price movement detected:\n"
                        f"Symbol: {position.symbol}\n"
                        f"Entry: {position.entry_price}\n"
                        f"Current: {current_price}\n"
                        f"Change: {(price_diff/position.entry_price*100):.2f}%"
                    )
                
                # 计算基础盈亏
                price_pnl = float(price_diff * position.size * multiplier)
                
                # 验证基础盈亏
                if not isinstance(price_pnl, (int, float)):
                    raise ValueError(f"Invalid price PnL calculation: {price_pnl}")
                
                notional_value = float(position.size * current_price * multiplier)
                if abs(price_pnl) > notional_value:
                    logger.warning(
                        f"Price PnL exceeds position value:\n"
                        f"Symbol: {position.symbol}\n"
                        f"PnL: {price_pnl}\n"
                        f"Notional Value: {notional_value}\n"
                        f"Percentage: {(price_pnl/notional_value*100):.2f}%"
                    )
                
                # 合并所有盈亏组件
                total_pnl = price_pnl + funding_pnl
                
                # 验证最终盈亏
                if not isinstance(total_pnl, (int, float)):
                    raise ValueError(f"Invalid total PnL: {total_pnl}")
                
                # 记录详细的盈亏分析
                logger.info(
                    f"PnL calculation details:\n"
                    f"Symbol: {position.symbol}\n"
                    f"Direction: {position.direction}\n"
                    f"Size: {position.size}\n"
                    f"Entry Price: {position.entry_price}\n"
                    f"Current Price: {current_price}\n"
                    f"Price PnL: {price_pnl}\n"
                    f"Funding PnL: {funding_pnl}\n"
                    f"Total PnL: {total_pnl}\n"
                    f"Return: {(total_pnl/notional_value*100):.2f}%"
                )
                
                return float(total_pnl)
            except Exception as e:
                logger.error(
                    f"Error calculating final PnL:\n"
                    f"Symbol: {position.symbol}\n"
                    f"Error: {e}",
                    exc_info=True
                )
                return 0.0
        except Exception as e:
            logger.error(f"Critical error calculating PnL: {e}", exc_info=True)
            return 0.0

    def _update_position_params(self, position: Union[Position, ContractPosition], signal: TradeSignal) -> None:
        """更新持仓参数，包含风险控制和资金费率更新"""
        try:
            if not isinstance(position, (Position, ContractPosition)):
                raise TypeError(f"Invalid position type: {type(position)}")
            if not isinstance(signal, TradeSignal):
                raise TypeError(f"Invalid signal type: {type(signal)}")

            # 验证并更新止损止盈
            if signal.stop_loss is not None:
                if not isinstance(signal.stop_loss, (int, float)):
                    raise ValueError(f"Invalid stop loss: {signal.stop_loss}")
                if signal.stop_loss > 0:
                    position.stop_loss = signal.stop_loss

            if signal.take_profit is not None:
                if not isinstance(signal.take_profit, (int, float)):
                    raise ValueError(f"Invalid take profit: {signal.take_profit}")
                if signal.take_profit > 0:
                    position.take_profit = signal.take_profit

            # 合约持仓特殊处理
            if isinstance(position, ContractPosition):
                # 更新资金费率相关参数
                new_funding_rate = self._get_funding_rate(signal.symbol)
                if isinstance(new_funding_rate, (int, float)):
                    position.funding_rate = float(new_funding_rate)
                position.next_funding_time = self._get_next_funding_time()

                # 验证并更新杠杆率
                new_leverage = signal.metadata.get('leverage')
                if new_leverage is not None:
                    if not isinstance(new_leverage, (int, float)) or new_leverage <= 0:
                        raise ValueError(f"Invalid leverage: {new_leverage}")
                    position.leverage = float(new_leverage)

                # 验证并更新保证金类型
                new_margin_type = signal.metadata.get('margin_type')
                if new_margin_type is not None:
                    if not isinstance(new_margin_type, str) or new_margin_type not in ['isolated', 'cross']:
                        raise ValueError(f"Invalid margin type: {new_margin_type}")
                    position.margin_type = new_margin_type

                # 重新计算清算价格和维持保证金
                try:
                    position.liquidation_price = self._calculate_liquidation_price(signal, position.entry_price)
                    position.maintenance_margin = self._calculate_maintenance_margin(position.size, position.entry_price)
                except Exception as e:
                    logger.error(f"Error updating liquidation parameters: {e}")

            # 更新元数据，保留重要字段
            try:
                position.metadata.update({
                    k: v for k, v in signal.metadata.items()
                    if v is not None and k not in [
                        'created_at', 'updated_at', 'timestamp',
                        'funding_rate', 'next_funding_time'
                    ]
                })
            except Exception as e:
                logger.error(f"Error updating metadata: {e}")

            # 记录重要参数变更
            old_stop_loss = position.stop_loss
            old_take_profit = position.take_profit
            
            # 记录参数更新结果
            logger.info(
                f"Position parameters updated for {position.symbol}:\n"
                f"Stop Loss: {old_stop_loss} -> {position.stop_loss}\n"
                f"Take Profit: {old_take_profit} -> {position.take_profit}"
            )

        except Exception as e:
            logger.error(f"Critical error updating position parameters: {e}", exc_info=True)
            raise  # 重要参数更新失败应该抛出异常以便上层处理

    def _check_risk_limits(self) -> bool:
        """检查风控限制，包括回撤、日亏损和持仓限制"""
        try:
            # 验证风控参数
            if not isinstance(self.max_drawdown, (int, float)) or self.max_drawdown <= 0 or self.max_drawdown >= 1:
                logger.error(f"Invalid max drawdown setting: {self.max_drawdown}")
                return False
            if not isinstance(self.daily_loss_limit, (int, float)) or self.daily_loss_limit <= 0 or self.daily_loss_limit >= 1:
                logger.error(f"Invalid daily loss limit setting: {self.daily_loss_limit}")
                return False
            if not isinstance(self.max_positions, int) or self.max_positions <= 0:
                logger.error(f"Invalid max positions setting: {self.max_positions}")
                return False

            # 检查回撤限制
            if not isinstance(self.current_drawdown, (int, float)):
                logger.error(f"Invalid drawdown value: {self.current_drawdown}")
                return False
            if self.current_drawdown > self.max_drawdown:
                logger.warning(f"Max drawdown limit reached: {self.current_drawdown:.2%}")
                return False

            # 获取并验证账户余额
            try:
                account_balance = self._get_account_balance()
                if account_balance <= 0:
                    logger.error(f"Invalid account balance: {account_balance}")
                    return False
            except Exception as e:
                logger.error(f"Error getting account balance: {e}")
                return False

            # 检查日亏损限制
            if not isinstance(self.daily_pnl, (int, float)):
                logger.error(f"Invalid daily PnL value: {self.daily_pnl}")
                return False
            daily_loss_threshold = -self.daily_loss_limit * account_balance
            if self.daily_pnl < daily_loss_threshold:
                logger.warning(f"Daily loss limit reached: {self.daily_pnl:.2f} (threshold: {daily_loss_threshold:.2f})")
                return False

            # 检查持仓数量限制
            if len(self.positions) >= self.max_positions:
                logger.warning(f"Maximum number of positions reached: {len(self.positions)}")
                return False

            # 计算并验证总持仓价值
            total_position_value = 0.0
            for symbol, position in self.positions.items():
                try:
                    if not isinstance(position, (Position, ContractPosition)):
                        logger.error(f"Invalid position type for {symbol}: {type(position)}")
                        continue
                    if not all(isinstance(x, (int, float)) for x in [position.size, position.entry_price]):
                        logger.error(f"Invalid numeric values in position {symbol}")
                        continue
                    
                    position_value = float(position.size * position.entry_price)
                    if isinstance(position, ContractPosition):
                        if not isinstance(position.leverage, (int, float)) or position.leverage <= 0:
                            logger.error(f"Invalid leverage for {symbol}: {position.leverage}")
                            continue
                        position_value *= float(position.leverage)
                    
                    if position_value < 0:
                        logger.error(f"Negative position value for {symbol}: {position_value}")
                        continue
                        
                    total_position_value += position_value
                except Exception as e:
                    logger.error(f"Error calculating position value for {symbol}: {e}")
                    continue

            # 验证总持仓价值限制
            max_position_value = account_balance * 3  # 最大杠杆率3倍
            if total_position_value >= max_position_value:
                logger.warning(f"Maximum position value reached: {total_position_value:.2f} (limit: {max_position_value:.2f})")
                return False

            return True

        except Exception as e:
            logger.error(f"Critical error checking risk limits: {e}", exc_info=True)
            return False

    def _update_performance_metrics(self, trade_data: Union[Trade, Dict[str, Any]]) -> None:
        if isinstance(trade_data, dict):
            trade = create_trade_from_dict(trade_data)
        else:
            trade = trade_data

        self.total_pnl += trade.pnl if hasattr(trade, 'pnl') else 0.0
        self.daily_pnl += trade.pnl if hasattr(trade, 'pnl') else 0.0

        current_equity = self._get_account_balance()
        self.max_equity = max(self.max_equity, current_equity)
        self.current_drawdown = (self.max_equity - current_equity) / self.max_equity

    def _calculate_sharpe_ratio(self) -> float:
        """计算夏普比率，带有完整的错误处理和数据验证"""
        try:
            if not self.trades:
                return 0.0

            # 计算收益率时进行数据验证
            returns = []
            for trade in self.trades:
                try:
                    if not isinstance(trade.pnl, (int, float)) or not isinstance(trade.entry_price, (int, float)):
                        logger.warning(f"Invalid trade data: pnl={trade.pnl}, entry_price={trade.entry_price}")
                        continue
                        
                    if trade.entry_price <= 0:
                        logger.warning(f"Invalid entry price: {trade.entry_price}")
                        continue
                        
                    returns.append(trade.pnl / trade.entry_price)
                except ZeroDivisionError:
                    logger.warning(f"Zero entry price in trade: {trade}")
                    continue
                except Exception as e:
                    logger.error(f"Error processing trade for Sharpe ratio: {e}")
                    continue

            if not returns:
                return 0.0

            returns_array = np.array(returns, dtype=float)
            std = np.std(returns_array) if len(returns_array) > 0 else 0.0
            
            if std == 0:
                return 0.0

            annualization_factor = np.sqrt(252)
            risk_free_rate = 0.02  # 2% annual risk-free rate
            daily_rf_rate = float(risk_free_rate / 252)
            
            mean_return = np.mean(returns_array) if len(returns_array) > 0 else 0.0
            
            excess_return = float(mean_return - daily_rf_rate)
            sharpe = float(annualization_factor * excess_return / std)
            
            # 限制异常值
            return float(np.clip(sharpe, -10, 10))
            
        except Exception as e:
            logger.error(f"Error calculating Sharpe ratio: {e}", exc_info=True)
            return 0.0

    def _get_account_balance(self) -> float:
        """获取账户余额"""
        try:
            # TODO: 实现从账户系统获取实际余额的逻辑
            # 目前返回模拟值，后续需要对接实际账户系统
            balance = 100000.0
            
            if not isinstance(balance, (int, float)):
                raise ValueError(f"Invalid balance type: {type(balance)}")
            if balance < 0:
                logger.warning(f"Negative balance detected: {balance}")
                return 0.0
                
            return float(balance)
        except Exception as e:
            logger.error(f"Error getting account balance: {e}", exc_info=True)
            return 0.0

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
        try:
            if not isinstance(self.trades, list):
                logger.error(f"Invalid trades type: {type(self.trades)}")
                return 0.0
            if not self.trades:
                return 0.0
                
            winning_trades = 0
            total_valid_trades = 0
            
            for trade in self.trades:
                try:
                    if not hasattr(trade, 'pnl') or not isinstance(trade.pnl, (int, float)):
                        logger.warning(f"Invalid PnL value in trade: {trade}")
                        continue
                    total_valid_trades += 1
                    if trade.pnl > 0:
                        winning_trades += 1
                except Exception as e:
                    logger.warning(f"Error processing trade for win rate: {e}")
                    continue
                    
            if total_valid_trades == 0:
                return 0.0
                
            win_rate = float(winning_trades) / total_valid_trades
            return float(np.clip(win_rate, 0.0, 1.0))
        except Exception as e:
            logger.error(f"Error calculating win rate: {e}", exc_info=True)
            return 0.0

    def _calculate_average_trade(self) -> float:
        """计算平均每笔交易盈亏"""
        try:
            if not isinstance(self.trades, list):
                logger.error(f"Invalid trades type: {type(self.trades)}")
                return 0.0
            if not self.trades:
                return 0.0
                
            total_pnl = 0.0
            valid_trades = 0
            
            for trade in self.trades:
                try:
                    if not hasattr(trade, 'pnl') or not isinstance(trade.pnl, (int, float)):
                        logger.warning(f"Invalid PnL value in trade: {trade}")
                        continue
                    total_pnl += float(trade.pnl)
                    valid_trades += 1
                except Exception as e:
                    logger.warning(f"Error processing trade for average calculation: {e}")
                    continue
                    
            if valid_trades == 0:
                return 0.0
                
            return float(total_pnl / valid_trades)
        except Exception as e:
            logger.error(f"Error calculating average trade: {e}", exc_info=True)
            return 0.0

    async def monitor_positions(self) -> None:
        """监控持仓，包含永续合约特定的监控逻辑"""
        monitoring_interval = 1.0  # 基础监控间隔（秒）
        error_backoff = 1.0  # 错误回退时间（秒）
        max_backoff = 30.0  # 最大回退时间（秒）
        consecutive_errors = 0
        last_health_check = time.time()
        health_check_interval = 60.0  # 健康检查间隔（秒）
        
        while True:
            try:
                # 系统健康检查
                current_time = time.time()
                if current_time - last_health_check >= health_check_interval:
                    try:
                        # 检查风险限制
                        if not self._check_risk_limits():
                            logger.critical("Risk limits exceeded, suspending position monitoring")
                            await asyncio.sleep(health_check_interval)
                            continue
                            
                        # 检查系统资源
                        memory_info = psutil.Process().memory_info()
                        memory_usage = memory_info.rss / 1024 / 1024  # MB
                        if memory_usage > 1024:  # 1GB内存警告
                            logger.warning(f"High memory usage: {memory_usage:.1f}MB")
                            
                        # 重置错误计数和回退时间
                        if consecutive_errors > 0:
                            logger.info(f"System recovered after {consecutive_errors} consecutive errors")
                            consecutive_errors = 0
                            error_backoff = 1.0
                            
                        last_health_check = current_time
                    except Exception as health_error:
                        logger.error(f"Health check failed: {health_error}", exc_info=True)
                
                positions_to_monitor = list(self.positions.items())
                if not positions_to_monitor:
                    await asyncio.sleep(monitoring_interval)
                    continue
                    
                for symbol, position in positions_to_monitor:
                    try:
                        # 验证position对象完整性
                        if not isinstance(position, (Position, ContractPosition)):
                            logger.error(f"Invalid position type for {symbol}: {type(position)}")
                            continue
                            
                        required_attrs = ['symbol', 'direction', 'size', 'entry_price']
                        missing_attrs = [attr for attr in required_attrs if not hasattr(position, attr)]
                        if missing_attrs:
                            logger.error(f"Position missing required attributes: {missing_attrs}")
                            continue

                        current_price = await self._get_current_price(symbol)
                        if current_price <= 0:
                            logger.warning(f"Invalid price received for {symbol}: {current_price}")
                            continue

                        # 计算当前盈亏
                        pnl = self._calculate_pnl(position, current_price)
                        
                        # 合约持仓特殊处理
                        if isinstance(position, ContractPosition):
                            try:
                                # 更新资金费率
                                new_funding_rate = await self._get_funding_rate(symbol)
                                if isinstance(new_funding_rate, (int, float)):
                                    old_rate = position.funding_rate
                                    position.funding_rate = float(new_funding_rate)
                                    if abs(new_funding_rate) > 0.01:  # 1%警告阈值
                                        logger.warning(f"High funding rate for {symbol}: {new_funding_rate}")
                                    if old_rate != new_funding_rate:
                                        logger.info(f"Funding rate changed for {symbol}: {old_rate} -> {new_funding_rate}")

                                # 更新清算价格
                                new_liquidation_price = self._calculate_liquidation_price(position, current_price)
                                if new_liquidation_price > 0:
                                    position.liquidation_price = new_liquidation_price
                                    
                                    # 检查清算风险
                                    price_to_liquidation = abs(current_price - position.liquidation_price)
                                    liquidation_threshold = current_price * 0.1  # 10%警告阈值
                                    if price_to_liquidation < liquidation_threshold:
                                        logger.warning(
                                            f"Position {symbol} near liquidation: "
                                            f"Current={current_price}, Liquidation={position.liquidation_price}"
                                        )
                                        
                                # 检查维持保证金
                                maintenance_margin = self._calculate_maintenance_margin(position.size, current_price)
                                if maintenance_margin > 0:
                                    position.maintenance_margin = maintenance_margin
                                    
                                # 检查清算条件
                                try:
                                    if not isinstance(position.liquidation_price, (int, float)):
                                        raise ValueError(f"Invalid liquidation price: {position.liquidation_price}")
                                        
                                    price_to_liquidation = abs(current_price - position.liquidation_price)
                                    warning_threshold = current_price * 0.05  # 5%警告阈值
                                    danger_threshold = current_price * 0.02   # 2%危险阈值
                                    
                                    if price_to_liquidation <= danger_threshold:
                                        logger.critical(
                                            f"CRITICAL: Position {symbol} extremely close to liquidation:\n"
                                            f"Current Price: {current_price}\n"
                                            f"Liquidation Price: {position.liquidation_price}\n"
                                            f"Distance to Liquidation: {price_to_liquidation:.2f} ({(price_to_liquidation/current_price*100):.2f}%)"
                                        )
                                    elif price_to_liquidation <= warning_threshold:
                                        logger.warning(
                                            f"WARNING: Position {symbol} approaching liquidation:\n"
                                            f"Current Price: {current_price}\n"
                                            f"Liquidation Price: {position.liquidation_price}\n"
                                            f"Distance to Liquidation: {price_to_liquidation:.2f} ({(price_to_liquidation/current_price*100):.2f}%)"
                                        )
                                    
                                    if (position.direction == 'buy' and current_price <= position.liquidation_price) or \
                                       (position.direction == 'sell' and current_price >= position.liquidation_price):
                                        logger.critical(f"LIQUIDATION EVENT: {position.direction.capitalize()} position {symbol} at {current_price}")
                                        await self._close_position(position, current_price, "LIQUIDATION")
                                        continue
                                        
                                except Exception as e:
                                    logger.error(f"Error checking liquidation conditions for {symbol}: {e}", exc_info=True)
                                    
                            except Exception as e:
                                logger.error(f"Error updating contract position parameters for {symbol}: {e}", exc_info=True)

                        # 检查止损止盈
                        if self._check_stop_loss(position, current_price):
                            await self._close_position(position, current_price, "STOP_LOSS")
                            logger.info(f"Stop loss triggered for {symbol} at {current_price}")
                            continue

                        if self._check_take_profit(position, current_price):
                            await self._close_position(position, current_price, "TAKE_PROFIT")
                            logger.info(f"Take profit triggered for {symbol} at {current_price}")
                            continue

                    except ValueError as e:
                        logger.error(f"Value error monitoring position for {symbol}: {e}")
                        consecutive_errors += 1
                    except aiohttp.ClientError as e:
                        logger.error(
                            f"Network error monitoring position for {symbol} (attempt {consecutive_errors + 1}):\n"
                            f"Error: {e}\n"
                            f"Next retry in {error_backoff}s"
                        )
                        consecutive_errors += 1
                        await asyncio.sleep(error_backoff)
                    except asyncio.TimeoutError as e:
                        logger.error(
                            f"Timeout monitoring position for {symbol} (attempt {consecutive_errors + 1}):\n"
                            f"Error: {e}\n"
                            f"Next retry in {error_backoff}s"
                        )
                        consecutive_errors += 1
                        await asyncio.sleep(error_backoff)
                    except Exception as e:
                        logger.error(
                            f"Unexpected error monitoring position for {symbol} (attempt {consecutive_errors + 1}):\n"
                            f"Error: {e}\n"
                            f"Next retry in {error_backoff}s",
                            exc_info=True
                        )
                        consecutive_errors += 1
                        await asyncio.sleep(error_backoff)
                    finally:
                        # 更新持仓状态
                        if isinstance(position, ContractPosition):
                            try:
                                # 验证并更新下次资金费用时间
                                next_funding = self._get_next_funding_time()
                                if not isinstance(next_funding, datetime):
                                    raise ValueError(f"Invalid next funding time: {next_funding}")
                                position.next_funding_time = next_funding
                                
                                # 验证必要字段
                                required_fields = {
                                    'direction': position.direction,
                                    'size': position.size,
                                    'entry_price': position.entry_price,
                                    'funding_rate': position.funding_rate,
                                    'liquidation_price': position.liquidation_price,
                                    'maintenance_margin': position.maintenance_margin,
                                    'leverage': position.leverage
                                }
                                
                                for field, value in required_fields.items():
                                    if value is None:
                                        raise ValueError(f"Missing required field: {field}")
                                    if field in ['size', 'entry_price', 'liquidation_price', 'maintenance_margin', 'leverage']:
                                        if not isinstance(value, (int, float)) or value <= 0:
                                            raise ValueError(f"Invalid {field}: {value}")
                                    if field == 'direction' and value not in ['buy', 'sell']:
                                        raise ValueError(f"Invalid direction: {value}")
                                    if field == 'funding_rate' and not isinstance(value, (int, float)):
                                        raise ValueError(f"Invalid funding rate: {value}")
                                
                                # 构建状态对象
                                status = {
                                    'symbol': symbol,
                                    'direction': position.direction,
                                    'size': float(position.size),
                                    'entry_price': float(position.entry_price),
                                    'current_price': float(current_price),
                                    'pnl': float(pnl),
                                    'funding_rate': float(position.funding_rate),
                                    'next_funding': position.next_funding_time.isoformat(),
                                    'liquidation_price': float(position.liquidation_price),
                                    'maintenance_margin': float(position.maintenance_margin),
                                    'leverage': float(position.leverage),
                                    'margin_ratio': float(position.maintenance_margin / (position.size * current_price)),
                                    'unrealized_pnl': float(pnl),
                                    'timestamp': int(time.time())
                                }
                                
                                # 记录详细状态
                                logger.info(
                                    f"Contract position status update:\n"
                                    f"Symbol: {symbol}\n"
                                    f"Direction: {status['direction']}\n"
                                    f"Size: {status['size']}\n"
                                    f"Entry Price: {status['entry_price']}\n"
                                    f"Current Price: {status['current_price']}\n"
                                    f"PnL: {status['pnl']}\n"
                                    f"Funding Rate: {status['funding_rate']:.4%}\n"
                                    f"Next Funding: {status['next_funding']}\n"
                                    f"Liquidation Price: {status['liquidation_price']}\n"
                                    f"Margin Ratio: {status['margin_ratio']:.2%}\n"
                                    f"Leverage: {status['leverage']}x"
                                )
                                
                                # 发送状态更新到报告系统
                                try:
                                    # 使用带重试的状态保存
                                    retry_count = 0
                                    max_retries = 3
                                    retry_delay = 1.0
                                    
                                    while retry_count < max_retries:
                                        try:
                                            await asyncio.wait_for(
                                                self.reporting.save_position_status(status),
                                                timeout=5.0
                                            )
                                            if retry_count > 0:
                                                logger.info(f"Successfully saved position status after {retry_count + 1} attempts")
                                            break
                                        except asyncio.TimeoutError:
                                            retry_count += 1
                                            if retry_count < max_retries:
                                                logger.warning(
                                                    f"Timeout saving position status for {symbol} "
                                                    f"(attempt {retry_count}/{max_retries})"
                                                )
                                                await asyncio.sleep(retry_delay * (2 ** retry_count))
                                            else:
                                                logger.error(f"Failed to save position status after {max_retries} attempts")
                                        except Exception as save_error:
                                            retry_count += 1
                                            if retry_count < max_retries:
                                                logger.warning(
                                                    f"Error saving position status for {symbol} "
                                                    f"(attempt {retry_count}/{max_retries}): {save_error}"
                                                )
                                                await asyncio.sleep(retry_delay * (2 ** retry_count))
                                            else:
                                                logger.error(
                                                    f"Failed to save position status after {max_retries} attempts: {save_error}",
                                                    exc_info=True
                                                )
                                                
                                except Exception as e:
                                    logger.error(f"Critical error in status save retry loop: {e}", exc_info=True)
                                
                            except ValueError as ve:
                                logger.error(f"Validation error updating position status: {ve}")
                                consecutive_errors += 1
                            except Exception as status_error:
                                logger.error(f"Unexpected error updating position status: {status_error}", exc_info=True)
                                consecutive_errors += 1
                                
                        # 动态调整休眠时间
                        sleep_time = 0.1 * (1 + (consecutive_errors * 0.5))  # 错误越多，休眠越长
                        await asyncio.sleep(min(sleep_time, 1.0))  # 最长1秒

                # 检查系统资源使用情况
                try:
                    process = psutil.Process()
                    memory_info = process.memory_info()
                    cpu_percent = process.cpu_percent()
                    
                    # 记录系统资源使用情况
                    if memory_info.rss > 1024 * 1024 * 1024:  # 1GB内存警告
                        logger.warning(
                            f"High memory usage detected:\n"
                            f"Memory: {memory_info.rss / (1024*1024):.2f}MB\n"
                            f"CPU: {cpu_percent}%"
                        )
                    
                    # 根据系统负载动态调整监控间隔
                    if cpu_percent > 80:  # CPU使用率超过80%
                        monitoring_interval = max(monitoring_interval, 2.0)
                        logger.warning(f"High CPU usage ({cpu_percent}%), increased monitoring interval to {monitoring_interval}s")
                    elif memory_info.rss > 1024 * 1024 * 512:  # 内存使用超过512MB
                        monitoring_interval = max(monitoring_interval, 1.5)
                        logger.warning(f"High memory usage ({memory_info.rss / (1024*1024):.2f}MB), adjusted monitoring interval")
                    
                except Exception as e:
                    logger.error(f"Error monitoring system resources: {e}")
                
                # 根据错误状态调整监控间隔
                if consecutive_errors > 0:
                    monitoring_interval = min(error_backoff * (2 ** consecutive_errors), max_backoff)
                    logger.warning(
                        f"Adjusted monitoring interval to {monitoring_interval}s "
                        f"after {consecutive_errors} errors"
                    )
                else:
                    base_interval = 1.0
                    if consecutive_errors > 0:  # 错误恢复
                        logger.info(f"System recovered after {consecutive_errors} consecutive errors")
                        consecutive_errors = 0
                        error_backoff = 1.0
                    monitoring_interval = base_interval
                
                # 确保监控间隔在合理范围内
                monitoring_interval = max(min(monitoring_interval, max_backoff), 0.5)

                # 内存清理和资源回收
                try:
                    if consecutive_errors == 0:  # 只在系统稳定时执行清理
                        process = psutil.Process()
                        memory_before = process.memory_info().rss
                        
                        # 强制垃圾回收
                        import gc
                        gc.collect()
                        
                        # 清理过期的缓存数据
                        for symbol in list(self.positions.keys()):
                            if symbol not in self.positions:
                                continue
                            position = self.positions[symbol]
                            if (datetime.now() - position.open_time).days > 7:  # 清理7天以上的数据
                                logger.info(f"Cleaning up old position data for {symbol}")
                                del self.positions[symbol]
                        
                        memory_after = process.memory_info().rss
                        memory_freed = memory_before - memory_after
                        if memory_freed > 0:
                            logger.info(f"Memory cleanup freed {memory_freed / (1024*1024):.2f}MB")
                except Exception as e:
                    logger.error(f"Error during memory cleanup: {e}")
                
                await asyncio.sleep(monitoring_interval)
            except Exception as e:
                consecutive_errors += 1
                error_backoff = min(error_backoff * 2, max_backoff)
                logger.error(
                    f"Critical error in monitor loop (attempt {consecutive_errors}):\n"
                    f"Error: {e}\n"
                    f"Next retry in {error_backoff}s",
                    exc_info=True
                )
                await asyncio.sleep(error_backoff)

    def _check_stop_loss(self, position: Union[Position, ContractPosition], current_price: float) -> bool:
        """检查是否触发止损，包含完整的数据验证"""
        try:
            if not isinstance(position, (Position, ContractPosition)):
                logger.error(f"Invalid position type: {type(position)}")
                return False
                
            if not isinstance(current_price, (int, float)) or current_price <= 0:
                logger.error(f"Invalid current price: {current_price}")
                return False
                
            if not hasattr(position, 'stop_loss') or position.stop_loss is None:
                return False
                
            if not isinstance(position.stop_loss, (int, float)) or position.stop_loss <= 0:
                logger.error(f"Invalid stop loss value: {position.stop_loss}")
                return False
                
            if not hasattr(position, 'direction') or position.direction not in ['buy', 'sell']:
                logger.error(f"Invalid position direction: {getattr(position, 'direction', None)}")
                return False
                
            # 计算止损距离并记录
            price_to_stop = abs(current_price - position.stop_loss)
            distance_percent = (price_to_stop / current_price) * 100
            
            if distance_percent < 0.5:  # 接近止损点0.5%时发出警告
                logger.warning(
                    f"Position {position.symbol} approaching stop loss:\n"
                    f"Current Price: {current_price}\n"
                    f"Stop Loss: {position.stop_loss}\n"
                    f"Distance: {distance_percent:.2f}%"
                )
                
            return (position.direction == 'buy' and current_price <= position.stop_loss) or \
                   (position.direction == 'sell' and current_price >= position.stop_loss)
                   
        except Exception as e:
            logger.error(f"Error checking stop loss: {e}", exc_info=True)
            return False

    def _check_take_profit(self, position: Union[Position, ContractPosition], current_price: float) -> bool:
        """检查是否触发止盈，包含完整的数据验证"""
        try:
            if not isinstance(position, (Position, ContractPosition)):
                logger.error(f"Invalid position type: {type(position)}")
                return False
                
            if not isinstance(current_price, (int, float)) or current_price <= 0:
                logger.error(f"Invalid current price: {current_price}")
                return False
                
            if not hasattr(position, 'take_profit') or position.take_profit is None:
                return False
                
            if not isinstance(position.take_profit, (int, float)) or position.take_profit <= 0:
                logger.error(f"Invalid take profit value: {position.take_profit}")
                return False
                
            if not hasattr(position, 'direction') or position.direction not in ['buy', 'sell']:
                logger.error(f"Invalid position direction: {getattr(position, 'direction', None)}")
                return False
                
            # 计算止盈距离并记录
            price_to_target = abs(current_price - position.take_profit)
            distance_percent = (price_to_target / current_price) * 100
            
            if distance_percent < 1.0:  # 接近止盈点1%时发出警告
                logger.warning(
                    f"Position {position.symbol} approaching take profit:\n"
                    f"Current Price: {current_price}\n"
                    f"Take Profit: {position.take_profit}\n"
                    f"Distance: {distance_percent:.2f}%"
                )
                
            return (position.direction == 'buy' and current_price >= position.take_profit) or \
                   (position.direction == 'sell' and current_price <= position.take_profit)
                   
        except Exception as e:
            logger.error(f"Error checking take profit: {e}", exc_info=True)
            return False

    async def _get_current_price(self, symbol: str) -> float:
        """获取当前价格，按优先级尝试从不同来源获取，包含资金费率和流动性检查"""
        if not isinstance(symbol, str) or not symbol.strip():
            raise ValueError("Invalid or empty symbol")

        sources = [
            (self.hyperliquid_client.get_price, "Hyperliquid"),
            (self.dydx_client.get_price, "dYdX"),
            (self._get_market_data_price, "Market Data Service")
        ]
        
        last_error = None
        best_price = None
        best_source = None
        prices = []
        source_metadata = {}
        
        for get_price, source_name in sources:
            try:
                start_time = time.time()
                price = await get_price(symbol)
                response_time = time.time() - start_time
                
                if price is not None and isinstance(price, (int, float)):
                    if price <= 0:
                        logger.warning(f"Non-positive price ({price}) from {source_name}")
                        continue
                        
                    if price > 1_000_000_000:  # 10亿上限检查
                        logger.warning(f"Unusually high price ({price}) from {source_name}")
                        continue
                        
                    prices.append((price, source_name))
                    source_metadata[source_name] = {
                        'response_time': response_time,
                        'timestamp': time.time()
                    }
                    
                    # 更新最佳价格（考虑流动性、资金费率和响应时间）
                    if best_price is None or await self._is_better_price_source(
                        symbol, price, source_name, best_price, best_source,
                        response_time=response_time
                    ):
                        best_price = price
                        best_source = source_name
                        
            except aiohttp.ClientError as e:
                last_error = f"Network error from {source_name}: {e}"
                logger.warning(f"{last_error} (latency: {time.time() - start_time:.3f}s)")
            except ValueError as e:
                last_error = f"Invalid price from {source_name}: {e}"
                logger.warning(last_error)
            except Exception as e:
                last_error = f"Unexpected error from {source_name}: {e}"
                logger.warning(f"{last_error}", exc_info=True)
        
        if best_price is not None:
            if len(prices) > 1:
                # 使用加权平均价格
                total_weight = 0
                weighted_sum = 0
                
                for price, source in prices:
                    # 基于响应时间的权重计算
                    response_time = source_metadata[source]['response_time']
                    weight = 1.0 / (1.0 + response_time)  # 响应越快权重越大
                    
                    weighted_sum += price * weight
                    total_weight += weight
                
                avg_price = weighted_sum / total_weight
                max_deviation = 0.05  # 5%最大偏差
                
                # 检查异常价格
                for price, source in prices:
                    deviation = abs(price - avg_price) / avg_price
                    if deviation > max_deviation:
                        logger.warning(
                            f"Large price deviation from {source}:\n"
                            f"Price: {price}\n"
                            f"Average: {avg_price}\n"
                            f"Deviation: {deviation:.2%}\n"
                            f"Response Time: {source_metadata[source]['response_time']:.3f}s"
                        )
            
            logger.info(
                f"Selected price for {symbol}:\n"
                f"Price: {best_price}\n"
                f"Source: {best_source}\n"
                f"Response Time: {source_metadata[best_source]['response_time']:.3f}s"
            )
            return float(best_price)
            
        error_msg = f"Failed to fetch valid price for {symbol} from any source. Last error: {last_error}"
        logger.error(error_msg)
        raise ValueError(error_msg)
        
    async def _is_better_price_source(
        self, symbol: str, new_price: float, new_source: str,
        current_price: float, current_source: str, response_time: float = None
    ) -> bool:
        """评估价格源的质量，考虑流动性、资金费率和响应时间"""
        try:
            if not isinstance(symbol, str) or not symbol.strip():
                logger.error("Invalid symbol")
                return False
                
            if not isinstance(new_price, (int, float)) or new_price <= 0:
                logger.error(f"Invalid new price: {new_price}")
                return False
                
            if not isinstance(current_price, (int, float)) or current_price <= 0:
                logger.error(f"Invalid current price: {current_price}")
                return False
                
            # 检查价格偏差
            price_deviation = abs(new_price - current_price) / current_price
            if price_deviation > 0.05:  # 5%最大偏差
                logger.warning(
                    f"Large price deviation between sources:\n"
                    f"New ({new_source}): {new_price}\n"
                    f"Current ({current_source}): {current_price}\n"
                    f"Deviation: {price_deviation:.2%}"
                )
                return False
                
            # 获取资金费率和流动性数据
            funding_rate = None
            liquidity = None
            
            if new_source == "Hyperliquid":
                try:
                    funding_rate = await self.hyperliquid_client.get_funding_rate(symbol)
                    liquidity = await self.hyperliquid_client.get_liquidity(symbol)
                except Exception as e:
                    logger.error(f"Error fetching Hyperliquid data: {e}")
                    return False
            elif new_source == "dYdX":
                try:
                    funding_rate = await self.dydx_client.get_funding_rate(symbol)
                    liquidity = await self.dydx_client.get_liquidity(symbol)
                except Exception as e:
                    logger.error(f"Error fetching dYdX data: {e}")
                    return False
            else:
                return False  # 优先使用DEX价格源
                
            # 验证资金费率和流动性
            if not isinstance(funding_rate, (int, float)):
                logger.error(f"Invalid funding rate from {new_source}: {funding_rate}")
                return False
                
            if not isinstance(liquidity, (int, float)) or liquidity < 0:
                logger.error(f"Invalid liquidity from {new_source}: {liquidity}")
                return False
                
            # 评分系统 (0-100)
            score = 0
            
            # 资金费率评分 (0-40分)
            funding_score = 40 * (1 - min(abs(funding_rate), 0.01) / 0.01)
            score += funding_score
            
            # 流动性评分 (0-40分)
            liquidity_score = 40 * min(liquidity / 1_000_000, 1.0)
            score += liquidity_score
            
            # 响应时间评分 (0-20分)
            if response_time is not None:
                latency_score = 20 * (1 - min(response_time, 1.0))
                score += latency_score
            
            logger.info(
                f"Price source quality metrics for {new_source}:\n"
                f"Symbol: {symbol}\n"
                f"Funding Rate: {funding_rate:.4%}\n"
                f"Liquidity: {liquidity:,.0f}\n"
                f"Response Time: {response_time:.3f}s\n"
                f"Total Score: {score:.1f}/100"
            )
            
            return score >= 60  # 及格分为60分
            
        except Exception as e:
            logger.error(f"Error evaluating price source quality: {e}", exc_info=True)
            return False

    async def _get_market_data_price(self, symbol: str) -> Optional[float]:
        """从市场数据服务获取价格，包含永续合约相关数据验证"""
        if not isinstance(symbol, str) or not symbol.strip():
            logger.error("Invalid or empty symbol provided")
            return None

        timeout = aiohttp.ClientTimeout(total=5, connect=3)
        max_retries = 3
        retry_delay = 1.0
        backoff_factor = 2.0
        
        # 记录系统资源状态
        try:
            process = psutil.Process()
            memory_info = process.memory_info()
            cpu_percent = process.cpu_percent()
            
            if memory_info.rss > 1024 * 1024 * 512:  # 超过512MB
                logger.warning(
                    f"High memory usage in price retrieval:\n"
                    f"Memory: {memory_info.rss / (1024*1024):.2f}MB\n"
                    f"CPU: {cpu_percent}%\n"
                    f"Symbol: {symbol}"
                )
                
            if cpu_percent > 80:
                logger.warning(f"High CPU usage ({cpu_percent}%) during price retrieval")
        except Exception as e:
            logger.error(f"Error monitoring system resources: {e}")
        
        # 记录请求开始时间
        request_start = time.time()
        last_error = None
        request_id = f"{symbol}-{int(time.time())}"
        
        for attempt in range(max_retries):
            session = None
            try:
                session = aiohttp.ClientSession(timeout=timeout)
                url = f"{Config.get_market_data_url()}/api/v1/market-data/{symbol}"
                request_id = f"{symbol}-{datetime.now().timestamp()}"
                
                headers = {
                    'Content-Type': 'application/json',
                    'X-Request-ID': request_id,
                    'X-Client-Version': '1.0.0',
                    'X-Retry-Count': str(attempt)
                }
                
                async with session.get(url, headers=headers) as response:
                    response_time = time.time() - request_start
                    
                    if response.status == 200:
                        try:
                            data = await response.json()
                        except aiohttp.ContentTypeError as e:
                            logger.error(f"Invalid JSON response: {e}")
                            last_error = "Invalid JSON response"
                            continue
                            
                        if not isinstance(data, dict):
                            logger.error(f"Invalid response format: {data}")
                            last_error = "Invalid response format"
                            continue
                            
                        required_fields = ['price', 'timestamp', 'volume']
                        missing_fields = [f for f in required_fields if f not in data]
                        if missing_fields:
                            logger.error(f"Missing required fields: {missing_fields}")
                            last_error = f"Missing fields: {missing_fields}"
                            continue
                            
                        try:
                            validation_results = {
                                'price': float(data['price']),
                                'timestamp': int(data['timestamp']),
                                'volume': float(data['volume']),
                                'request_id': request_id,
                                'symbol': symbol
                            }
                            
                            # 价格验证
                            if validation_results['price'] <= 0:
                                raise ValueError(f"Non-positive price: {validation_results['price']}")
                                
                            # 成交量验证
                            if validation_results['volume'] < 0:
                                raise ValueError(f"Negative volume: {validation_results['volume']}")
                                
                            # 数据时效性验证
                            age = time.time() - validation_results['timestamp'] / 1000
                            if age > 60:
                                logger.warning(
                                    f"Stale market data:\n"
                                    f"Symbol: {validation_results['symbol']}\n"
                                    f"Age: {age:.1f}s\n"
                                    f"Request ID: {validation_results['request_id']}\n"
                                    f"Price: {validation_results['price']}\n"
                                    f"Volume: {validation_results['volume']}"
                                )
                                
                            # 价格异常检测
                            if 'last_price' in data:
                                last_price = float(data['last_price'])
                                price_change = abs(validation_results['price'] - last_price) / last_price
                                if price_change > 0.1:  # 10%价格变化
                                    logger.warning(
                                        f"Large price change detected:\n"
                                        f"Symbol: {validation_results['symbol']}\n"
                                        f"Current: {validation_results['price']}\n"
                                        f"Previous: {last_price}\n"
                                        f"Change: {price_change:.2%}\n"
                                        f"Request ID: {validation_results['request_id']}"
                                    )
                                
                            # 资金费率和标记价格验证
                            validation_metrics = {}
                            
                            if 'funding_rate' in data:
                                try:
                                    funding_rate = float(data['funding_rate'])
                                    validation_metrics['funding_rate'] = funding_rate
                                    
                                    if abs(funding_rate) > 0.01:  # 1%阈值
                                        logger.warning(
                                            f"High funding rate detected:\n"
                                            f"Symbol: {validation_results['symbol']}\n"
                                            f"Rate: {funding_rate:.4%}\n"
                                            f"Threshold: 1.00%\n"
                                            f"Request ID: {validation_results['request_id']}"
                                        )
                                    elif abs(funding_rate) > 0.005:  # 0.5%警告阈值
                                        logger.info(
                                            f"Elevated funding rate:\n"
                                            f"Symbol: {validation_results['symbol']}\n"
                                            f"Rate: {funding_rate:.4%}"
                                        )
                                except (TypeError, ValueError) as e:
                                    logger.error(f"Invalid funding rate format: {e}")
                                    
                            if 'mark_price' in data:
                                try:
                                    mark_price = float(data['mark_price'])
                                    validation_metrics['mark_price'] = mark_price
                                    price_diff = abs(mark_price - validation_results['price']) / validation_results['price']
                                    
                                    if price_diff > 0.01:  # 1%严重偏差
                                        logger.warning(
                                            f"Critical mark price deviation:\n"
                                            f"Symbol: {validation_results['symbol']}\n"
                                            f"Spot Price: {validation_results['price']}\n"
                                            f"Mark Price: {mark_price}\n"
                                            f"Deviation: {price_diff:.2%}\n"
                                            f"Request ID: {validation_results['request_id']}"
                                        )
                                    elif price_diff > 0.001:  # 0.1%轻微偏差
                                        logger.info(
                                            f"Minor mark price deviation:\n"
                                            f"Symbol: {validation_results['symbol']}\n"
                                            f"Deviation: {price_diff:.2%}"
                                        )
                                except (TypeError, ValueError) as e:
                                    logger.error(f"Invalid mark price format: {e}")
                                    
                            if 'open_interest' in data:
                                try:
                                    open_interest = float(data['open_interest'])
                                    validation_metrics['open_interest'] = open_interest
                                    
                                    if open_interest < 0:
                                        logger.error(
                                            f"Invalid open interest:\n"
                                            f"Symbol: {validation_results['symbol']}\n"
                                            f"Value: {open_interest}\n"
                                            f"Request ID: {validation_results['request_id']}"
                                        )
                                    elif open_interest == 0:
                                        logger.warning(
                                            f"Zero open interest:\n"
                                            f"Symbol: {validation_results['symbol']}"
                                        )
                                except (TypeError, ValueError) as e:
                                    logger.error(f"Invalid open interest format: {e}")
                                    
                            # 记录验证指标
                            if validation_metrics:
                                logger.info(
                                    f"Market metrics validated:\n"
                                    f"Symbol: {validation_results['symbol']}\n"
                                    f"Metrics: {validation_metrics}\n"
                                    f"Request ID: {validation_results['request_id']}"
                                )
                                    
                            logger.info(
                                f"Market data retrieved:\n"
                                f"Symbol: {symbol}\n"
                                f"Price: {validation_results['price']}\n"
                                f"Volume: {validation_results['volume']}\n"
                                f"Response Time: {response_time:.3f}s\n"
                                f"Request ID: {request_id}"
                            )
                            
                            return validation_results['price']
                            
                        except (TypeError, ValueError) as e:
                            logger.error(f"Data validation error: {e}")
                            last_error = f"Validation error: {e}"
                            continue
                            
                    elif response.status == 404:
                        logger.warning(f"No market data for symbol: {symbol}")
                        return None
                        
                    elif response.status == 429:
                        retry_after = float(response.headers.get('Retry-After', 
                            retry_delay * (backoff_factor ** attempt)))
                        
                        # 记录速率限制信息
                        rate_limit_info = {
                            'symbol': symbol,
                            'retry_after': retry_after,
                            'attempt': attempt + 1,
                            'total_attempts': max_retries,
                            'request_id': request_id,
                            'remaining_calls': response.headers.get('X-RateLimit-Remaining', 'unknown'),
                            'reset_time': response.headers.get('X-RateLimit-Reset', 'unknown')
                        }
                        
                        logger.warning(
                            f"Rate limit exceeded:\n"
                            f"Symbol: {rate_limit_info['symbol']}\n"
                            f"Retry After: {rate_limit_info['retry_after']}s\n"
                            f"Attempt: {rate_limit_info['attempt']}/{rate_limit_info['total_attempts']}\n"
                            f"Remaining API Calls: {rate_limit_info['remaining_calls']}\n"
                            f"Rate Limit Reset: {rate_limit_info['reset_time']}\n"
                            f"Request ID: {rate_limit_info['request_id']}"
                        )
                        
                        # 动态调整重试延迟
                        adjusted_delay = max(
                            retry_after,
                            retry_delay * (backoff_factor ** attempt)
                        )
                        
                        logger.info(f"Adjusting retry delay to {adjusted_delay:.1f}s")
                        await asyncio.sleep(adjusted_delay)
                        continue
                        
                    else:
                        response_text = await response.text()
                        logger.error(
                            f"Market data service error:\n"
                            f"Symbol: {symbol}\n"
                            f"Status: {response.status}\n"
                            f"Response: {response_text}\n"
                            f"Request ID: {request_id}"
                        )
                        last_error = f"Service error: {response.status}"
                        if attempt < max_retries - 1:
                            await asyncio.sleep(retry_delay * (backoff_factor ** attempt))
                            continue
                        return None
                        
            except (aiohttp.ClientError, asyncio.TimeoutError) as e:
                error_type = "Timeout" if isinstance(e, asyncio.TimeoutError) else "Network"
                logger.error(
                    f"{error_type} error:\n"
                    f"Symbol: {symbol}\n"
                    f"Error: {e}\n"
                    f"Attempt: {attempt + 1}/{max_retries}\n"
                    f"Next retry in: {retry_delay * (backoff_factor ** attempt):.1f}s"
                )
                last_error = f"{error_type} error: {e}"
                
                # 检查网络连接和服务健康状态
                try:
                    import socket
                    import concurrent.futures
                    
                    with concurrent.futures.ThreadPoolExecutor() as executor:
                        # DNS解析检查
                        dns_future = executor.submit(
                            socket.gethostbyname,
                            Config.get_market_data_url().split("://")[1].split("/")[0]
                        )
                        try:
                            dns_future.result(timeout=3)
                        except Exception as dns_error:
                            logger.critical(f"DNS resolution failed: {dns_error}")
                            return None
                            
                        # 基础网络连接检查
                        conn_future = executor.submit(
                            socket.create_connection,
                            ("8.8.8.8", 53),
                            3
                        )
                        try:
                            conn_future.result(timeout=3)
                        except Exception as conn_error:
                            logger.critical(f"Network connectivity issue: {conn_error}")
                            return None
                            
                    # 服务健康状态检查
                    health_url = f"{Config.get_market_data_url()}/health"
                    try:
                        async with aiohttp.ClientSession(timeout=aiohttp.ClientTimeout(total=3)) as health_session:
                            async with health_session.get(health_url) as health_response:
                                if health_response.status != 200:
                                    logger.critical(f"Service health check failed: {health_response.status}")
                                    return None
                    except Exception as health_error:
                        logger.critical(f"Service health check error: {health_error}")
                        return None
                        
                except Exception as check_error:
                    logger.critical(f"Connection check error: {check_error}")
                    return None
                    
                if attempt < max_retries - 1:
                    retry_time = retry_delay * (backoff_factor ** attempt)
                    logger.info(f"Network checks passed. Retrying in {retry_time:.1f}s...")
                    await asyncio.sleep(retry_time)
                    continue
                return None
                
            except Exception as e:
                error_context = {
                    'symbol': symbol,
                    'attempt': attempt + 1,
                    'total_attempts': max_retries,
                    'elapsed_time': time.time() - request_start,
                    'request_id': request_id,
                    'error_type': type(e).__name__,
                    'error_msg': str(e)
                }
                
                logger.error(
                    f"Unexpected error in market data retrieval:\n"
                    f"Symbol: {error_context['symbol']}\n"
                    f"Error Type: {error_context['error_type']}\n"
                    f"Error: {error_context['error_msg']}\n"
                    f"Attempt: {error_context['attempt']}/{error_context['total_attempts']}\n"
                    f"Elapsed Time: {error_context['elapsed_time']:.3f}s\n"
                    f"Request ID: {error_context['request_id']}",
                    exc_info=True,
                    extra=error_context
                )
                
                last_error = f"{error_context['error_type']}: {error_context['error_msg']}"
                
                # 检查是否需要进行资源清理
                try:
                    import gc
                    if error_context['elapsed_time'] > 10:  # 如果单次请求耗时超过10秒
                        gc.collect()
                        logger.info("Performed garbage collection after long-running request")
                except Exception as gc_error:
                    logger.error(f"Failed to perform garbage collection: {gc_error}")
                
                if attempt < max_retries - 1:
                    retry_time = retry_delay * (backoff_factor ** attempt)
                    logger.info(
                        f"Retrying request:\n"
                        f"Symbol: {symbol}\n"
                        f"Next attempt in: {retry_time:.1f}s\n"
                        f"Request ID: {request_id}"
                    )
                    await asyncio.sleep(retry_time)
                    continue
                return None
                
            finally:
                if session:
                    try:
                        await session.close()
                    except Exception as close_error:
                        logger.error(f"Error closing session: {close_error}")
                    
        total_time = time.time() - request_start
        logger.error(
            f"Failed to fetch market data after all retries:\n"
            f"Symbol: {symbol}\n"
            f"Total Attempts: {max_retries}\n"
            f"Total Time: {total_time:.3f}s\n"
            f"Last Error: {last_error}\n"
            f"Request ID: {request_id}"
        )
        return None
