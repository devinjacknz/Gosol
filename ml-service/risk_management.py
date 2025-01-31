import logging
import pandas as pd
import numpy as np
from typing import Dict, List, Optional, Union
from datetime import datetime, timedelta
from dataclasses import dataclass
import json
import sqlite3
from pathlib import Path

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@dataclass
class RiskConfig:
    """风险配置"""
    max_position_size: float  # 最大仓位规模
    max_drawdown: float  # 最大回撤限制
    daily_loss_limit: float  # 日亏损限制
    position_limit: int  # 最大持仓数量
    risk_per_trade: float  # 单笔交易风险
    leverage_limit: float  # 杠杆限制
    correlation_limit: float  # 相关性限制
    min_diversification: int  # 最小分散化数量
    stop_loss_atr: float  # ATR倍数止损
    take_profit_atr: float  # ATR倍数止盈
    db_path: str  # 数据库路径
    max_leverage: float  # 最大允许杠杆
    max_position_value: float  # 单个仓位最大价值
    min_maintenance_margin: float  # 最小维持保证金率
    funding_rate_interval: int  # 资金费率收取间隔（小时）
    liquidation_threshold: float  # 强平阈值
    margin_call_threshold: float  # 追加保证金阈值

@dataclass
class Position:
    """持仓信息"""
    symbol: str
    direction: str
    size: float
    entry_price: float
    current_price: float
    stop_loss: float
    take_profit: float
    unrealized_pnl: float
    realized_pnl: float
    margin_used: float
    timestamp: datetime
    metadata: Dict

@dataclass
class PortfolioState:
    """组合状态"""
    total_equity: float
    used_margin: float
    free_margin: float
    margin_level: float
    total_pnl: float
    daily_pnl: float
    positions: Dict[str, Position]
    risk_metrics: Dict[str, float]
    drawdown: float
    exposure: Dict[str, float]

@dataclass
class ContractPosition(Position):
    """合约持仓"""
    leverage: float  # 杠杆倍数
    liquidation_price: float  # 强平价格
    margin_type: str  # 'isolated' 或 'cross'
    maintenance_margin: float  # 维持保证金
    funding_rate: float  # 资金费率
    next_funding_time: datetime  # 下次资金费时间

class RiskManager:
    """风险管理系统"""
    
    def __init__(self, config: RiskConfig):
        self.config = config
        self.db_path = config.db_path
        self.positions: Dict[str, Union[Position, ContractPosition]] = {}
        self.max_leverage = config.max_leverage  # 最大允许杠杆
        self.min_maintenance_margin = 0.005  # 最小维持保证金率
        self.max_position_value = config.max_position_value  # 单个仓位最大价值
        self.portfolio_history: List[PortfolioState] = []
        self.daily_stats: Dict[str, float] = {
            'high_equity': 0.0,
            'low_equity': float('inf'),
            'total_trades': 0,
            'winning_trades': 0,
            'losing_trades': 0,
            'total_pnl': 0.0
        }
        
        # 初始化数据库
        self._initialize_database()
    
    def _initialize_database(self):
        """初始化数据库"""
        with sqlite3.connect(self.db_path) as conn:
            # 持仓记录表
            conn.execute("""
                CREATE TABLE IF NOT EXISTS positions (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    timestamp DATETIME,
                    symbol TEXT,
                    direction TEXT,
                    size REAL,
                    entry_price REAL,
                    current_price REAL,
                    stop_loss REAL,
                    take_profit REAL,
                    unrealized_pnl REAL,
                    realized_pnl REAL,
                    margin_used REAL,
                    metadata TEXT
                )
            """)
            
            # 组合状态表
            conn.execute("""
                CREATE TABLE IF NOT EXISTS portfolio_states (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    timestamp DATETIME,
                    total_equity REAL,
                    used_margin REAL,
                    free_margin REAL,
                    margin_level REAL,
                    total_pnl REAL,
                    daily_pnl REAL,
                    drawdown REAL,
                    risk_metrics TEXT,
                    exposure TEXT
                )
            """)
            
            # 风险事件表
            conn.execute("""
                CREATE TABLE IF NOT EXISTS risk_events (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    timestamp DATETIME,
                    event_type TEXT,
                    severity TEXT,
                    description TEXT,
                    metadata TEXT
                )
            """)
            
            conn.commit()
    
    def check_position_risk(self, symbol: str, direction: str,
                          size: float, price: float) -> bool:
        """检查持仓风险"""
        # 检查持仓数量限制
        if len(self.positions) >= self.config.position_limit:
            logger.warning("Position limit reached")
            return False
        
        # 计算所需保证金
        margin_required = self._calculate_margin(size, price)
        portfolio_state = self._get_portfolio_state()
        
        # 检查可用保证金
        if margin_required > portfolio_state.free_margin:
            logger.warning("Insufficient margin")
            return False
        
        # 检查风险敞口
        total_exposure = sum(abs(p.size * p.current_price)
                           for p in self.positions.values())
        new_exposure = total_exposure + (size * price)
        if new_exposure > portfolio_state.total_equity * self.config.leverage_limit:
            logger.warning("Leverage limit exceeded")
            return False
        
        # 检查相关性
        if not self._check_correlation(symbol):
            logger.warning("Correlation limit exceeded")
            return False
        
        return True
    
    def calculate_position_size(self, price: float, stop_loss: float,
                              volatility: float) -> float:
        """计算仓位大小"""
        portfolio_state = self._get_portfolio_state()
        risk_amount = portfolio_state.total_equity * self.config.risk_per_trade
        
        # 基于波动率调整风险
        volatility_factor = 1.0
        if volatility > 0.02:  # 高波动率时降低仓位
            volatility_factor = 0.02 / volatility
        
        # 计算仓位大小
        price_risk = abs(price - stop_loss)
        position_size = (risk_amount * volatility_factor) / price_risk
        
        # 应用最大仓位限制
        max_size = self.config.max_position_size * portfolio_state.total_equity / price
        position_size = min(position_size, max_size)
        
        return position_size
    
    def update_positions(self, market_data: Dict[str, Dict]):
        """更新持仓状态"""
        for symbol, position in self.positions.items():
            if symbol in market_data:
                current_price = market_data[symbol]['price']
                position.current_price = current_price
                
                # 更新未实现盈亏
                if position.direction == 'buy':
                    position.unrealized_pnl = (
                        (current_price - position.entry_price) * position.size
                    )
                else:
                    position.unrealized_pnl = (
                        (position.entry_price - current_price) * position.size
                    )
                
                # 检查止损止盈
                self._check_stop_loss_take_profit(position)
        
        # 更新组合状态
        portfolio_state = self._get_portfolio_state()
        self.portfolio_history.append(portfolio_state)
        self._save_portfolio_state(portfolio_state)
    
    def _check_stop_loss_take_profit(self, position: Position) -> bool:
        """检查止损止盈"""
        if position.direction == 'buy':
            if position.current_price <= position.stop_loss:
                self._close_position(position, 'stop_loss')
                return True
            elif position.current_price >= position.take_profit:
                self._close_position(position, 'take_profit')
                return True
        else:
            if position.current_price >= position.stop_loss:
                self._close_position(position, 'stop_loss')
                return True
            elif position.current_price <= position.take_profit:
                self._close_position(position, 'take_profit')
                return True
        
        return False
    
    def _close_position(self, position: Position, reason: str):
        """平仓"""
        # 计算已实现盈亏
        if position.direction == 'buy':
            realized_pnl = (
                (position.current_price - position.entry_price) * position.size
            )
        else:
            realized_pnl = (
                (position.entry_price - position.current_price) * position.size
            )
        
        position.realized_pnl = realized_pnl
        
        # 更新每日统计
        self.daily_stats['total_trades'] += 1
        if realized_pnl > 0:
            self.daily_stats['winning_trades'] += 1
        else:
            self.daily_stats['losing_trades'] += 1
        self.daily_stats['total_pnl'] += realized_pnl
        
        # 保存平仓记录
        with sqlite3.connect(self.db_path) as conn:
            conn.execute("""
                UPDATE positions SET
                    current_price = ?,
                    unrealized_pnl = 0,
                    realized_pnl = ?,
                    metadata = ?
                WHERE symbol = ?
            """, (
                position.current_price,
                realized_pnl,
                json.dumps({**position.metadata, 'close_reason': reason}),
                position.symbol
            ))
            conn.commit()
        
        # 从持仓中移除
        del self.positions[position.symbol]
        
        # 记录风险事件
        self._log_risk_event(
            'position_closed',
            'info',
            f"Position closed: {position.symbol} ({reason})",
            {
                'symbol': position.symbol,
                'direction': position.direction,
                'realized_pnl': realized_pnl,
                'reason': reason
            }
        )
    
    def _get_portfolio_state(self) -> PortfolioState:
        """获取组合状态"""
        total_equity = 100000.0  # 初始资金
        used_margin = sum(p.margin_used for p in self.positions.values())
        unrealized_pnl = sum(p.unrealized_pnl for p in self.positions.values())
        
        # 计算总权益
        total_equity += unrealized_pnl + self.daily_stats['total_pnl']
        
        # 计算可用保证金
        free_margin = total_equity - used_margin
        margin_level = total_equity / used_margin if used_margin > 0 else float('inf')
        
        # 计算回撤
        self.daily_stats['high_equity'] = max(
            self.daily_stats['high_equity'],
            total_equity
        )
        drawdown = (
            (self.daily_stats['high_equity'] - total_equity) /
            self.daily_stats['high_equity']
        )
        
        # 计算风险指标
        risk_metrics = self._calculate_risk_metrics()
        
        # 计算风险敞口
        exposure = {}
        for symbol, position in self.positions.items():
            exposure[symbol] = position.size * position.current_price
        
        return PortfolioState(
            total_equity=total_equity,
            used_margin=used_margin,
            free_margin=free_margin,
            margin_level=margin_level,
            total_pnl=self.daily_stats['total_pnl'],
            daily_pnl=unrealized_pnl + self.daily_stats['total_pnl'],
            positions=self.positions.copy(),
            risk_metrics=risk_metrics,
            drawdown=drawdown,
            exposure=exposure
        )
    
    def _calculate_risk_metrics(self) -> Dict[str, float]:
        """计算风险指标"""
        if not self.portfolio_history:
            return {
                'sharpe_ratio': 0.0,
                'sortino_ratio': 0.0,
                'max_drawdown': 0.0,
                'var_95': 0.0,
                'expected_shortfall': 0.0
            }
        
        # 计算收益率序列
        returns = []
        for i in range(1, len(self.portfolio_history)):
            prev = self.portfolio_history[i-1]
            curr = self.portfolio_history[i]
            returns.append(
                (curr.total_equity - prev.total_equity) /
                prev.total_equity
            )
        
        returns = np.array(returns)
        
        # 计算风险指标
        if len(returns) > 1:
            avg_return = returns.mean()
            std_dev = returns.std()
            downside_returns = returns[returns < 0]
            downside_std = downside_returns.std() if len(downside_returns) > 0 else 0
            
            # 夏普比率
            risk_free_rate = 0.02 / 252  # 假设年化2%的无风险利率
            sharpe_ratio = (
                (avg_return - risk_free_rate) / std_dev * np.sqrt(252)
                if std_dev > 0 else 0
            )
            
            # 索提诺比率
            sortino_ratio = (
                (avg_return - risk_free_rate) / downside_std * np.sqrt(252)
                if downside_std > 0 else 0
            )
            
            # 最大回撤
            max_drawdown = max(p.drawdown for p in self.portfolio_history)
            
            # VaR和ES
            var_95 = np.percentile(returns, 5)
            expected_shortfall = returns[returns <= var_95].mean()
            
            return {
                'sharpe_ratio': sharpe_ratio,
                'sortino_ratio': sortino_ratio,
                'max_drawdown': max_drawdown,
                'var_95': var_95,
                'expected_shortfall': expected_shortfall
            }
        
        return {
            'sharpe_ratio': 0.0,
            'sortino_ratio': 0.0,
            'max_drawdown': 0.0,
            'var_95': 0.0,
            'expected_shortfall': 0.0
        }
    
    def _calculate_margin(self, size: float, price: float) -> float:
        """计算所需保证金"""
        # 假设5%的初始保证金要求
        return size * price * 0.05
    
    def _check_correlation(self, new_symbol: str) -> bool:
        """检查相关性限制"""
        if not self.positions:
            return True
        
        try:
            # 获取所有相关的交易对
            symbols = [new_symbol] + list(self.positions.keys())
            
            # 从数据库获取历史价格数据
            with sqlite3.connect(self.db_path) as conn:
                prices_data = {}
                for symbol in symbols:
                    query = """
                        SELECT timestamp, close
                        FROM ohlcv
                        WHERE symbol = ?
                        AND timestamp >= datetime('now', '-30 day')
                        ORDER BY timestamp
                    """
                    df = pd.read_sql_query(query, conn, params=(symbol,))
                    if not df.empty:
                        df.set_index('timestamp', inplace=True)
                        prices_data[symbol] = df['close']
                
                if not prices_data:
                    logger.warning("No price data available for correlation check")
                    return True
                
                # 创建价格数据框
                prices_df = pd.DataFrame(prices_data)
                
                # 计算收益率
                returns_df = prices_df.pct_change().dropna()
                
                # 计算相关性矩阵
                corr_matrix = returns_df.corr()
                
                # 检查新交易对与现有持仓的相关性
                for symbol in self.positions.keys():
                    if symbol in corr_matrix.columns:
                        correlation = abs(corr_matrix.loc[new_symbol, symbol])
                        if correlation > self.config.correlation_limit:
                            logger.warning(
                                f"High correlation ({correlation:.2f}) between "
                                f"{new_symbol} and {symbol}"
                            )
                            return False
                
                # 检查投资组合分散度
                if len(self.positions) >= self.config.min_diversification:
                    portfolio_corr = corr_matrix.abs().mean().mean()
                    if portfolio_corr > self.config.correlation_limit:
                        logger.warning(
                            f"Portfolio correlation ({portfolio_corr:.2f}) "
                            f"exceeds limit ({self.config.correlation_limit})"
                        )
                        return False
                
                return True
                
        except Exception as e:
            logger.error(f"Error in correlation check: {str(e)}")
            return True  # 如果出错，允许交易继续
    
    def _save_portfolio_state(self, state: PortfolioState):
        """保存组合状态"""
        with sqlite3.connect(self.db_path) as conn:
            conn.execute("""
                INSERT INTO portfolio_states (
                    timestamp, total_equity, used_margin,
                    free_margin, margin_level, total_pnl,
                    daily_pnl, drawdown, risk_metrics,
                    exposure
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            """, (
                datetime.now(),
                state.total_equity,
                state.used_margin,
                state.free_margin,
                state.margin_level,
                state.total_pnl,
                state.daily_pnl,
                state.drawdown,
                json.dumps(state.risk_metrics),
                json.dumps(state.exposure)
            ))
            conn.commit()
    
    def _log_risk_event(self, event_type: str, severity: str,
                       description: str, metadata: Dict):
        """记录风险事件"""
        with sqlite3.connect(self.db_path) as conn:
            conn.execute("""
                INSERT INTO risk_events (
                    timestamp, event_type, severity,
                    description, metadata
                ) VALUES (?, ?, ?, ?, ?)
            """, (
                datetime.now(),
                event_type,
                severity,
                description,
                json.dumps(metadata)
            ))
            conn.commit()
    
    def get_risk_report(self) -> Dict:
        """生成风险报告"""
        portfolio_state = self._get_portfolio_state()
        
        return {
            'portfolio_summary': {
                'total_equity': portfolio_state.total_equity,
                'used_margin': portfolio_state.used_margin,
                'free_margin': portfolio_state.free_margin,
                'margin_level': portfolio_state.margin_level,
                'total_pnl': portfolio_state.total_pnl,
                'daily_pnl': portfolio_state.daily_pnl,
                'drawdown': portfolio_state.drawdown
            },
            'risk_metrics': portfolio_state.risk_metrics,
            'positions': {
                symbol: {
                    'size': pos.size,
                    'entry_price': pos.entry_price,
                    'current_price': pos.current_price,
                    'unrealized_pnl': pos.unrealized_pnl,
                    'margin_used': pos.margin_used
                }
                for symbol, pos in portfolio_state.positions.items()
            },
            'exposure': portfolio_state.exposure,
            'daily_stats': self.daily_stats
        }
    
    def get_position_summary(self) -> pd.DataFrame:
        """获取持仓摘要"""
        if not self.positions:
            return pd.DataFrame()
        
        data = []
        for symbol, position in self.positions.items():
            data.append({
                'symbol': symbol,
                'direction': position.direction,
                'size': position.size,
                'entry_price': position.entry_price,
                'current_price': position.current_price,
                'unrealized_pnl': position.unrealized_pnl,
                'margin_used': position.margin_used,
                'stop_loss': position.stop_loss,
                'take_profit': position.take_profit
            })
        
        return pd.DataFrame(data)
    
    def get_risk_events(self, start_time: Optional[datetime] = None,
                       end_time: Optional[datetime] = None,
                       severity: Optional[str] = None) -> pd.DataFrame:
        """获取风险事件"""
        with sqlite3.connect(self.db_path) as conn:
            query = "SELECT * FROM risk_events WHERE 1=1"
            params = []
            
            if start_time:
                query += " AND timestamp >= ?"
                params.append(start_time)
            if end_time:
                query += " AND timestamp <= ?"
                params.append(end_time)
            if severity:
                query += " AND severity = ?"
                params.append(severity)
            
            query += " ORDER BY timestamp DESC"
            
            return pd.read_sql_query(query, conn, params=params)
    
    def _check_leverage(self, leverage: float) -> bool:
        """检查杠杆是否在允许范围内"""
        return leverage <= self.max_leverage

    def _calculate_liquidation_price(self, position: ContractPosition) -> float:
        """计算预估强平价格"""
        if position.margin_type == 'isolated':
            if position.direction == 'long':
                return position.entry_price * (1 - 1/position.leverage + self.min_maintenance_margin)
            else:
                return position.entry_price * (1 + 1/position.leverage - self.min_maintenance_margin)
        else:
            # cross模式下需要考虑账户总权益
            total_equity = self._calculate_total_equity()
            return self._calculate_cross_liquidation_price(position, total_equity)

    def _calculate_cross_liquidation_price(self, position: ContractPosition, total_equity: float) -> float:
        """计算全仓模式下的强平价格"""
        total_margin = sum(p.margin_used for p in self.positions.values())
        available_margin = total_equity - total_margin + position.margin_used
        
        if position.direction == 'long':
            return position.entry_price * (1 - available_margin/(position.size * position.entry_price))
        else:
            return position.entry_price * (1 + available_margin/(position.size * position.entry_price))

    def _check_margin_requirement(self, position: ContractPosition) -> bool:
        """检查保证金要求"""
        required_margin = position.size * position.entry_price / position.leverage
        maintenance_margin = required_margin * self.min_maintenance_margin
        
        if position.margin_type == 'isolated':
            return position.margin_used >= required_margin
        else:
            total_equity = self._calculate_total_equity()
            return total_equity >= sum(p.margin_used for p in self.positions.values()) + required_margin

    def can_open_position(self, order: Dict) -> bool:
        """检查是否可以开仓"""
        if 'leverage' in order:  # 合约订单
            if not self._check_leverage(order['leverage']):
                logger.warning(f"Leverage {order['leverage']} exceeds maximum allowed {self.max_leverage}")
                return False
                
            position_value = order['size'] * order['price']
            if position_value > self.max_position_value:
                logger.warning(f"Position value {position_value} exceeds maximum allowed {self.max_position_value}")
                return False
            
            # 检查保证金要求
            required_margin = position_value / order['leverage']
            if required_margin > self._calculate_free_margin():
                logger.warning("Insufficient margin for the order")
                return False
        
        return super().can_open_position(order)

    def _update_contract_positions(self):
        """更新合约持仓状态"""
        for symbol, position in list(self.positions.items()):
            if isinstance(position, ContractPosition):
                # 更新强平价格
                position.liquidation_price = self._calculate_liquidation_price(position)
                
                # 检查是否接近强平价格
                price_distance = abs(position.current_price - position.liquidation_price) / position.current_price
                if price_distance < 0.05:  # 如果价格接近强平价格的5%以内
                    self._log_risk_event(
                        'LIQUIDATION_RISK',
                        'HIGH',
                        f'Position {symbol} is near liquidation price',
                        {
                            'current_price': position.current_price,
                            'liquidation_price': position.liquidation_price,
                            'distance_percentage': price_distance * 100
                        }
                    )
                
                # 更新资金费用
                if datetime.now() >= position.next_funding_time:
                    funding_payment = position.size * position.current_price * position.funding_rate
                    position.realized_pnl -= funding_payment
                    position.next_funding_time += timedelta(hours=8)  # 更新下次资金费时间
                    
                    self._log_risk_event(
                        'FUNDING_PAYMENT',
                        'LOW',
                        f'Funding payment for {symbol}',
                        {
                            'amount': funding_payment,
                            'funding_rate': position.funding_rate,
                            'next_funding_time': position.next_funding_time.isoformat()
                        }
                    )  