import logging
import pandas as pd
import numpy as np
from typing import Dict, List, Optional, Union
from datetime import datetime, timedelta
from dataclasses import dataclass
from agent_system import AgentSystem, AgentConfig, TradeSignal
from reporting_system import ReportingSystem, ExecutionReport, PerformanceReport
from market_data_service import MarketDataService, MarketConfig
import json
from pathlib import Path
from risk_management import RiskConfig, Position
import sqlite3

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@dataclass
class BacktestConfig:
    """回测配置"""
    start_date: datetime
    end_date: datetime
    initial_capital: float
    symbols: List[str]
    timeframes: List[str]
    commission_rate: float = 0.001  # 0.1% 手续费
    slippage: float = 0.001  # 0.1% 滑点
    enable_fractional: bool = True  # 是否允许小数仓位
    data_source: str = "database"  # 数据来源：database或csv
    max_leverage: float = 10.0  # 最大杠杆倍数
    funding_rate_interval: int = 8  # 资金费率收取间隔（小时）
    min_maintenance_margin: float = 0.005  # 最小维持保证金率

@dataclass
class BacktestPosition:
    """回测持仓"""
    symbol: str
    direction: str
    size: float
    entry_price: float
    stop_loss: float
    take_profit: float
    agent_name: str
    open_time: datetime
    unrealized_pnl: float = 0.0
    metadata: Dict = None

@dataclass
class ContractBacktestPosition(BacktestPosition):
    """合约回测持仓"""
    leverage: float  # 杠杆倍数
    liquidation_price: float  # 强平价格
    margin_type: str  # 'isolated' 或 'cross'
    maintenance_margin: float  # 维持保证金
    funding_rate: float  # 资金费率
    next_funding_time: datetime  # 下次资金费时间
    notional_value: float  # 名义价值

@dataclass
class BacktestResult:
    """回测结果"""
    total_returns: float
    annual_returns: float
    sharpe_ratio: float
    max_drawdown: float
    win_rate: float
    profit_factor: float
    total_trades: int
    trades_history: List[Dict]
    equity_curve: pd.Series
    monthly_returns: pd.Series
    positions_history: List[Position]
    risk_metrics: Dict[str, float]

class BacktestSystem:
    """回测系统"""
    
    def __init__(self, config: BacktestConfig):
        self.config = config
        self.agent_system = AgentSystem()
        self.reporting = ReportingSystem("backtest_results.db")
        
        # 回测状态
        self.current_time: datetime = config.start_date
        self.positions: Dict[str, Position] = {}
        self.trades_history: List[Dict] = []
        self.equity_curve: List[float] = [config.initial_capital]
        self.equity_timestamps: List[datetime] = [config.start_date]
        
        # 性能指标
        self.total_pnl = 0.0
        self.current_drawdown = 0.0
        self.max_drawdown = 0.0
        self.peak_equity = config.initial_capital
        
        # 加载历史数据
        self.historical_data = self._load_historical_data()
        
        self.funding_rates: Dict[str, float] = {}  # 各交易对的资金费率
        self.next_funding_time = config.start_date + timedelta(hours=config.funding_rate_interval)
    
    def _load_historical_data(self) -> Dict[str, Dict[str, pd.DataFrame]]:
        """加载历史数据"""
        data = {}
        
        if self.config.data_source == "database":
            with sqlite3.connect("database/market_data.db") as conn:
                for symbol in self.config.symbols:
                    data[symbol] = {}
                    for timeframe in self.config.timeframes:
                        query = """
                            SELECT * FROM ohlcv
                            WHERE symbol = ? AND timeframe = ?
                            AND timestamp BETWEEN ? AND ?
                            ORDER BY timestamp
                        """
                        df = pd.read_sql_query(
                            query, conn,
                            params=(
                                symbol, timeframe,
                                self.config.start_date,
                                self.config.end_date
                            )
                        )
                        if not df.empty:
                            df.set_index('timestamp', inplace=True)
                            data[symbol][timeframe] = df
        
        else:  # csv数据源
            data_dir = Path("data")
            for symbol in self.config.symbols:
                data[symbol] = {}
                for timeframe in self.config.timeframes:
                    file_path = data_dir / f"{symbol}_{timeframe}.csv"
                    if file_path.exists():
                        df = pd.read_csv(file_path)
                        df['timestamp'] = pd.to_datetime(df['timestamp'])
                        df.set_index('timestamp', inplace=True)
                        mask = (
                            (df.index >= self.config.start_date) &
                            (df.index <= self.config.end_date)
                        )
                        data[symbol][timeframe] = df[mask]
        
        return data
    
    def run(self) -> BacktestResult:
        """运行回测"""
        logger.info("Starting backtest...")
        
        # 初始化Agent
        self._initialize_agents()
        
        # 回测主循环
        while self.current_time <= self.config.end_date:
            try:
                # 更新市场数据
                current_data = self._get_current_data()
                if not current_data:
                    self.current_time += timedelta(minutes=1)
                    continue
                
                # 更新持仓
                self._update_positions(current_data)
                
                # 生成交易信号
                for symbol in self.config.symbols:
                    for agent in self.agent_system.agents.values():
                        if agent.config.symbol == symbol:
                            timeframe_data = self._get_timeframe_data(
                                symbol,
                                agent.config.timeframe,
                                self.current_time
                            )
                            
                            if timeframe_data is not None:
                                signal = agent.analyze(timeframe_data)
                                if signal:
                                    self._process_signal(signal, current_data[symbol])
                
                # 记录权益
                current_equity = self._calculate_total_equity(current_data)
                self.equity_curve.append(current_equity)
                self.equity_timestamps.append(self.current_time)
                
                # 更新最大回撤
                self.peak_equity = max(self.peak_equity, current_equity)
                self.current_drawdown = (self.peak_equity - current_equity) / self.peak_equity
                self.max_drawdown = max(self.max_drawdown, self.current_drawdown)
                
                # 更新时间
                self.current_time += timedelta(minutes=1)
                
            except Exception as e:
                logger.error(f"Error in backtest loop: {str(e)}")
                break
        
        # 平掉所有持仓
        self._close_all_positions()
        
        # 计算回测结果
        return self._calculate_results()
    
    def _initialize_agents(self):
        """初始化交易代理"""
        # 这里添加需要测试的交易代理
        pass
    
    def _get_current_data(self) -> Optional[Dict[str, Dict]]:
        """获取当前时间点的市场数据"""
        data = {}
        for symbol in self.config.symbols:
            # 获取1分钟数据
            if '1m' not in self.historical_data[symbol]:
                continue
            
            df = self.historical_data[symbol]['1m']
            mask = (df.index <= self.current_time)
            if not mask.any():
                continue
            
            current = df[mask].iloc[-1]
            data[symbol] = {
                'price': current['close'],
                'volume': current['volume'],
                'high': current['high'],
                'low': current['low']
            }
        
        return data if data else None
    
    def _get_timeframe_data(self, symbol: str, timeframe: str,
                           current_time: datetime) -> Optional[pd.DataFrame]:
        """获取指定时间周期的历史数据"""
        if timeframe not in self.historical_data[symbol]:
            return None
        
        df = self.historical_data[symbol][timeframe]
        mask = (df.index <= current_time)
        if not mask.any():
            return None
        
        return df[mask]
    
    def _process_signal(self, signal: Dict, market_data: Dict):
        """处理交易信号"""
        symbol = signal['symbol']
        direction = signal['direction']
        size = signal['size']
        
        # 检查是否是合约交易
        is_contract = signal.get('contract', False)
        leverage = signal.get('leverage', 1.0)
        margin_type = signal.get('margin_type', 'isolated')
        
        if is_contract:
            # 检查杠杆是否超过限制
            if leverage > self.config.max_leverage:
                logger.warning(f"Leverage {leverage} exceeds maximum allowed {self.config.max_leverage}")
                return
            
            # 计算所需保证金
            notional_value = size * market_data['price']
            required_margin = notional_value / leverage
            
            # 检查是否有足够保证金
            if required_margin > self._calculate_free_margin():
                logger.warning("Insufficient margin for the contract order")
                return
            
            # 计算强平价格
            liquidation_price = self._calculate_liquidation_price(
                direction, market_data['price'], leverage, margin_type
            )
            
            # 创建合约持仓
            position = ContractBacktestPosition(
                symbol=symbol,
                direction=direction,
                size=size,
                entry_price=market_data['price'],
                stop_loss=signal['stop_loss'],
                take_profit=signal['take_profit'],
                agent_name=signal.get('agent_name', 'unknown'),
                open_time=self.current_time,
                unrealized_pnl=0.0,
                metadata=signal.get('metadata', {}),
                leverage=leverage,
                liquidation_price=liquidation_price,
                margin_type=margin_type,
                maintenance_margin=required_margin * self.config.min_maintenance_margin,
                funding_rate=self.funding_rates.get(symbol, 0.0),
                next_funding_time=self.next_funding_time,
                notional_value=notional_value
            )
        else:
            # 现货交易逻辑
            position = BacktestPosition(
                symbol=symbol,
                direction=direction,
                size=size,
                entry_price=market_data['price'],
                stop_loss=signal['stop_loss'],
                take_profit=signal['take_profit'],
                agent_name=signal.get('agent_name', 'unknown'),
                open_time=self.current_time,
                unrealized_pnl=0.0,
                metadata=signal.get('metadata', {})
            )
        
        # 记录交易
        trade = {
            'timestamp': self.current_time,
            'symbol': symbol,
            'direction': direction,
            'size': size,
            'price': market_data['price'],
            'commission': required_margin * self.config.commission_rate if is_contract else size * market_data['price'] * self.config.commission_rate,
            'type': 'contract' if is_contract else 'spot',
            'leverage': leverage if is_contract else 1.0,
            'notional_value': notional_value if is_contract else size * market_data['price']
        }
        self.trades_history.append(trade)
        
        # 添加持仓
        self.positions[symbol] = position
    
    def _update_positions(self, current_data: Dict[str, Dict]):
        """更新持仓状态"""
        # 检查是否需要更新资金费率
        if self.current_time >= self.next_funding_time:
            self._update_funding_rates()
            self.next_funding_time += timedelta(hours=self.config.funding_rate_interval)
        
        for symbol, position in list(self.positions.items()):
            if symbol not in current_data:
                continue
            
            current_price = current_data[symbol]['price']
            position.current_price = current_price
            
            if isinstance(position, ContractBacktestPosition):
                # 更新合约持仓
                self._update_contract_position(position, current_price)
            else:
                # 更新现货持仓
                if position.direction == 'buy':
                    position.unrealized_pnl = (current_price - position.entry_price) * position.size
                else:
                    position.unrealized_pnl = (position.entry_price - current_price) * position.size
            
            # 检查止损止盈
            if self._check_stop_loss_take_profit(position):
                self._close_position(position, current_price, "SL/TP")
    
    def _update_contract_position(self, position: ContractBacktestPosition, current_price: float):
        """更新合约持仓状态"""
        # 更新未实现盈亏
        if position.direction == 'buy':
            position.unrealized_pnl = (current_price - position.entry_price) * position.size * position.leverage
        else:
            position.unrealized_pnl = (position.entry_price - current_price) * position.size * position.leverage
        
        # 检查是否触及强平价格
        if position.direction == 'buy' and current_price <= position.liquidation_price:
            self._close_position(position, current_price, "LIQUIDATION")
        elif position.direction == 'sell' and current_price >= position.liquidation_price:
            self._close_position(position, current_price, "LIQUIDATION")
        
        # 更新资金费用
        if self.current_time >= position.next_funding_time:
            funding_payment = position.notional_value * position.funding_rate
            position.realized_pnl -= funding_payment
            position.next_funding_time += timedelta(hours=self.config.funding_rate_interval)
    
    def _update_funding_rates(self):
        """更新资金费率"""
        # 这里可以实现更复杂的资金费率计算逻辑
        # 目前使用简单的随机生成
        for symbol in self.config.symbols:
            self.funding_rates[symbol] = np.random.normal(0.0001, 0.0002)  # 均值0.01%，标准差0.02%
    
    def _calculate_liquidation_price(self, direction: str, entry_price: float,
                                   leverage: float, margin_type: str) -> float:
        """计算强平价格"""
        maintenance_margin = self.config.min_maintenance_margin
        
        if direction == 'buy':
            return entry_price * (1 - 1/leverage + maintenance_margin)
        else:
            return entry_price * (1 + 1/leverage - maintenance_margin)
    
    def _close_position(self, position: Position, current_price: float, reason: str):
        """平仓"""
        commission = position.size * current_price * self.config.commission_rate
        
        if position.direction == 'buy':
            pnl = (current_price - position.entry_price) * position.size
        else:
            pnl = (position.entry_price - current_price) * position.size
        
        # 记录交易
        trade = {
            'timestamp': self.current_time,
            'symbol': position.symbol,
            'direction': 'sell' if position.direction == 'buy' else 'buy',
            'size': position.size,
            'price': current_price,
            'commission': commission,
            'pnl': pnl,
            'reason': reason,
            'type': 'close'
        }
        self.trades_history.append(trade)
        
        # 更新总盈亏
        self.total_pnl += pnl - commission
        
        # 移除持仓
        del self.positions[position.symbol]
    
    def _check_stop_loss_take_profit(self, position: Position) -> bool:
        """检查止损止盈"""
        if position.direction == 'buy':
            if position.current_price <= position.stop_loss:
                return True
            elif position.current_price >= position.take_profit:
                return True
        else:
            if position.current_price >= position.stop_loss:
                return True
            elif position.current_price <= position.take_profit:
                return True
        
        return False
    
    def _calculate_total_equity(self, current_data: Dict[str, Dict]) -> float:
        """计算总权益"""
        total = self.equity_curve[-1]
        for symbol, position in self.positions.items():
            if symbol in current_data:
                if position.direction == 'buy':
                    pnl = (
                        (current_data[symbol]['price'] - position.entry_price) *
                        position.size
                    )
                else:
                    pnl = (
                        (position.entry_price - current_data[symbol]['price']) *
                        position.size
                    )
                total += pnl
        return total
    
    def _calculate_free_margin(self) -> float:
        """计算可用保证金"""
        used_margin = sum(p.margin_used for p in self.positions.values())
        return self.equity_curve[-1] - used_margin
    
    def _close_all_positions(self):
        """平掉所有持仓"""
        for symbol, position in list(self.positions.items()):
            current_price = self._get_current_price(symbol)
            if current_price is not None:
                self._close_position(position, current_price, 'BACKTEST_END')
    
    def _get_current_price(self, symbol: str) -> Optional[float]:
        """获取当前价格"""
        if symbol not in self.historical_data:
            return None
        
        # 使用最小时间周期的数据
        min_timeframe = min(self.config.timeframes)
        if min_timeframe not in self.historical_data[symbol]:
            return None
        
        df = self.historical_data[symbol][min_timeframe]
        if self.current_time not in df.index:
            return None
        
        return df.loc[self.current_time, 'close']
    
    def _calculate_results(self) -> BacktestResult:
        """计算回测结果"""
        equity_curve = pd.Series(self.equity_curve, index=pd.date_range(
            self.config.start_date,
            self.config.end_date,
            freq='1min'
        )[:len(self.equity_curve)])
        
        # 计算收益率
        returns = equity_curve.pct_change().dropna()
        total_returns = (equity_curve[-1] - self.config.initial_capital) / self.config.initial_capital
        
        # 计算年化收益率
        days = (self.config.end_date - self.config.start_date).days
        annual_returns = (1 + total_returns) ** (365 / days) - 1
        
        # 计算夏普比率
        risk_free_rate = 0.02  # 假设无风险利率2%
        excess_returns = returns - risk_free_rate/252
        sharpe_ratio = np.sqrt(252) * excess_returns.mean() / returns.std()
        
        # 计算最大回撤
        cummax = equity_curve.cummax()
        drawdown = (cummax - equity_curve) / cummax
        max_drawdown = drawdown.max()
        
        # 计算交易统计
        trades_df = pd.DataFrame(self.trades_history)
        if not trades_df.empty:
            winning_trades = trades_df[trades_df['pnl'] > 0]
            win_rate = len(winning_trades) / len(trades_df)
            
            total_profit = winning_trades['pnl'].sum()
            total_loss = abs(trades_df[trades_df['pnl'] < 0]['pnl'].sum())
            profit_factor = total_profit / total_loss if total_loss != 0 else float('inf')
        else:
            win_rate = 0
            profit_factor = 0
        
        # 计算月度收益率
        monthly_returns = equity_curve.resample('M').last().pct_change()
        
        # 计算风险指标
        risk_metrics = {
            'volatility': returns.std() * np.sqrt(252),
            'sortino_ratio': self._calculate_sortino_ratio(returns),
            'calmar_ratio': annual_returns / max_drawdown if max_drawdown != 0 else float('inf'),
            'avg_drawdown': drawdown.mean(),
            'avg_drawdown_days': self._calculate_avg_drawdown_days(equity_curve)
        }
        
        return BacktestResult(
            total_returns=total_returns,
            annual_returns=annual_returns,
            sharpe_ratio=sharpe_ratio,
            max_drawdown=max_drawdown,
            win_rate=win_rate,
            profit_factor=profit_factor,
            total_trades=len(trades_df),
            trades_history=self.trades_history,
            equity_curve=equity_curve,
            monthly_returns=monthly_returns,
            positions_history=list(self.positions.values()),
            risk_metrics=risk_metrics
        )
    
    def _calculate_sortino_ratio(self, returns: pd.Series) -> float:
        """计算索提诺比率"""
        risk_free_rate = 0.02  # 假设无风险利率2%
        excess_returns = returns - risk_free_rate/252
        downside_returns = returns[returns < 0]
        downside_std = downside_returns.std() * np.sqrt(252)
        
        return excess_returns.mean() * 252 / downside_std if downside_std != 0 else 0
    
    def _calculate_avg_drawdown_days(self, equity_curve: pd.Series) -> float:
        """计算平均回撤持续天数"""
        cummax = equity_curve.cummax()
        drawdown = (cummax - equity_curve) / cummax
        
        # 找到回撤开始和结束点
        drawdown_periods = []
        in_drawdown = False
        start_idx = None
        
        for i, value in enumerate(drawdown):
            if not in_drawdown and value > 0:
                in_drawdown = True
                start_idx = i
            elif in_drawdown and value == 0:
                in_drawdown = False
                if start_idx is not None:
                    drawdown_periods.append(i - start_idx)
        
        if drawdown_periods:
            return np.mean(drawdown_periods) / (24 * 60)  # 转换为天数
        return 0 