import logging
import pandas as pd
import numpy as np
from typing import Dict, List, Optional, Union
from datetime import datetime, timedelta
from dataclasses import dataclass
import json
import sqlite3
from pathlib import Path
from abc import ABC, abstractmethod
from collections import deque

def calculate_sma(data: pd.Series, period: int) -> pd.Series:
    """Calculate Simple Moving Average"""
    return data.rolling(window=period).mean()

def calculate_rsi(data: pd.Series, period: int = 14) -> pd.Series:
    """Calculate Relative Strength Index"""
    delta = data.diff()
    gain = (delta.where(delta > 0, 0)).rolling(window=period).mean()
    loss = (-delta.where(delta < 0, 0)).rolling(window=period).mean()
    rs = gain / loss.replace(0, float('inf'))
    return 100 - (100 / (1 + rs))

def calculate_atr(high: pd.Series, low: pd.Series, close: pd.Series, period: int = 14) -> pd.Series:
    """Calculate Average True Range"""
    tr1 = high - low
    tr2 = abs(high - close.shift())
    tr3 = abs(low - close.shift())
    tr = pd.concat([tr1, tr2, tr3], axis=1).max(axis=1)
    return tr.rolling(window=period).mean()

def calculate_bbands(data: pd.Series, period: int = 20, num_std: float = 2) -> tuple[pd.Series, pd.Series, pd.Series]:
    """Calculate Bollinger Bands"""
    middle = calculate_sma(data, period)
    std = data.rolling(window=period).std()
    upper = middle + (std * num_std)
    lower = middle - (std * num_std)
    return upper, middle, lower

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@dataclass
class AgentConfig:
    """Agent配置"""
    name: str
    symbol: str
    timeframe: str
    strategy_type: str
    parameters: Dict
    confidence_threshold: float = 0.6
    risk_limit: float = 0.02  # 单次交易风险限制
    max_positions: int = 1  # 最大持仓数量
    enable_ml: bool = False  # 是否启用机器学习

@dataclass
class TradeSignal:
    """交易信号"""
    symbol: str
    direction: str  # 'buy' or 'sell'
    price: float
    stop_loss: float
    take_profit: float
    size: float
    confidence: float
    agent_name: str
    timestamp: datetime
    metadata: Dict

class BaseAgent(ABC):
    """Agent基类"""
    
    def __init__(self, config: AgentConfig):
        self.config = config
        self.positions: List[Dict] = []
        self.signals: deque = deque(maxlen=1000)
        self.performance_metrics = {
            'total_signals': 0,
            'successful_signals': 0,
            'failed_signals': 0,
            'total_pnl': 0.0,
            'win_rate': 0.0,
            'avg_return': 0.0
        }
    
    @abstractmethod
    def analyze(self, data: pd.DataFrame) -> Optional[TradeSignal]:
        """分析市场数据并生成交易信号"""
        pass
    
    def update_performance(self, signal: TradeSignal, success: bool, return_pct: float):
        """更新性能指标"""
        self.performance_metrics['total_signals'] += 1
        if success:
            self.performance_metrics['successful_signals'] += 1
        else:
            self.performance_metrics['failed_signals'] += 1
        
        self.performance_metrics['total_pnl'] += return_pct
        self.performance_metrics['win_rate'] = (
            self.performance_metrics['successful_signals'] /
            self.performance_metrics['total_signals']
        )
        self.performance_metrics['avg_return'] = (
            self.performance_metrics['total_pnl'] /
            self.performance_metrics['total_signals']
        )
    
    def _calculate_position_size(self, price: float, stop_loss: float) -> float:
        """计算仓位大小"""
        risk_amount = self.config.risk_limit * 100000  # 假设账户规模100,000
        price_risk = abs(price - stop_loss)
        return risk_amount / price_risk
    
    def _calculate_stop_loss(self, data: pd.DataFrame, direction: str) -> float:
        """计算止损价格"""
        if direction == 'buy':
            return data['low'].iloc[-10:].min()
        else:
            return data['high'].iloc[-10:].max()
    
    def _calculate_take_profit(self, entry_price: float, stop_loss: float) -> float:
        """计算止盈价格"""
        risk = abs(entry_price - stop_loss)
        return entry_price + (risk * 2)  # 风险收益比2:1

class TrendFollowingAgent(BaseAgent):
    """趋势跟踪Agent"""
    
    def analyze(self, data: pd.DataFrame) -> Optional[TradeSignal]:
        if len(data) < 50:
            return None
        
        # 计算技术指标
        data['sma20'] = calculate_sma(data['close'], 20)
        data['sma50'] = calculate_sma(data['close'], 50)
        data['rsi'] = calculate_rsi(data['close'], 14)
        data['atr'] = calculate_atr(data['high'], data['low'], data['close'], 14)
        
        current_price = data['close'].iloc[-1]
        sma20 = data['sma20'].iloc[-1]
        sma50 = data['sma50'].iloc[-1]
        rsi = data['rsi'].iloc[-1]
        atr = data['atr'].iloc[-1]
        
        # 生成信号
        signal = None
        if sma20 > sma50 and rsi > 50:  # 买入信号
            stop_loss = current_price - (atr * 2)
            take_profit = current_price + (atr * 4)
            confidence = min(1.0, (sma20 - sma50) / sma50 * 5)
            
            if confidence >= self.config.confidence_threshold:
                signal = TradeSignal(
                    symbol=self.config.symbol,
                    direction='buy',
                    price=current_price,
                    stop_loss=stop_loss,
                    take_profit=take_profit,
                    size=self._calculate_position_size(current_price, stop_loss),
                    confidence=confidence,
                    agent_name=self.config.name,
                    timestamp=data.index[-1],
                    metadata={
                        'strategy': 'trend_following',
                        'sma20': sma20,
                        'sma50': sma50,
                        'rsi': rsi,
                        'atr': atr
                    }
                )
        
        elif sma20 < sma50 and rsi < 50:  # 卖出信号
            stop_loss = current_price + (atr * 2)
            take_profit = current_price - (atr * 4)
            confidence = min(1.0, (sma50 - sma20) / sma50 * 5)
            
            if confidence >= self.config.confidence_threshold:
                signal = TradeSignal(
                    symbol=self.config.symbol,
                    direction='sell',
                    price=current_price,
                    stop_loss=stop_loss,
                    take_profit=take_profit,
                    size=self._calculate_position_size(current_price, stop_loss),
                    confidence=confidence,
                    agent_name=self.config.name,
                    timestamp=data.index[-1],
                    metadata={
                        'strategy': 'trend_following',
                        'sma20': sma20,
                        'sma50': sma50,
                        'rsi': rsi,
                        'atr': atr
                    }
                )
        
        if signal:
            self.signals.append(signal)
        
        return signal

class MeanReversionAgent(BaseAgent):
    """均值回归Agent"""
    
    def analyze(self, data: pd.DataFrame) -> Optional[TradeSignal]:
        if len(data) < 50:
            return None
        
        # 计算技术指标
        data['sma20'] = calculate_sma(data['close'], 20)
        data['boll_upper'], data['boll_middle'], data['boll_lower'] = calculate_bbands(
            data['close'], period=20, num_std=2
        )
        data['rsi'] = calculate_rsi(data['close'], 14)
        
        current_price = data['close'].iloc[-1]
        sma20 = data['sma20'].iloc[-1]
        boll_upper = data['boll_upper'].iloc[-1]
        boll_lower = data['boll_lower'].iloc[-1]
        rsi = data['rsi'].iloc[-1]
        
        # 生成信号
        signal = None
        if current_price < boll_lower and rsi < 30:  # 超卖，买入信号
            stop_loss = current_price * 0.99  # 1%止损
            take_profit = sma20  # 均线作为目标
            confidence = min(1.0, (boll_lower - current_price) / current_price * 10)
            
            if confidence >= self.config.confidence_threshold:
                signal = TradeSignal(
                    symbol=self.config.symbol,
                    direction='buy',
                    price=current_price,
                    stop_loss=stop_loss,
                    take_profit=take_profit,
                    size=self._calculate_position_size(current_price, stop_loss),
                    confidence=confidence,
                    agent_name=self.config.name,
                    timestamp=data.index[-1],
                    metadata={
                        'strategy': 'mean_reversion',
                        'sma20': sma20,
                        'boll_upper': boll_upper,
                        'boll_lower': boll_lower,
                        'rsi': rsi
                    }
                )
        
        elif current_price > boll_upper and rsi > 70:  # 超买，卖出信号
            stop_loss = current_price * 1.01  # 1%止损
            take_profit = sma20  # 均线作为目标
            confidence = min(1.0, (current_price - boll_upper) / current_price * 10)
            
            if confidence >= self.config.confidence_threshold:
                signal = TradeSignal(
                    symbol=self.config.symbol,
                    direction='sell',
                    price=current_price,
                    stop_loss=stop_loss,
                    take_profit=take_profit,
                    size=self._calculate_position_size(current_price, stop_loss),
                    confidence=confidence,
                    agent_name=self.config.name,
                    timestamp=data.index[-1],
                    metadata={
                        'strategy': 'mean_reversion',
                        'sma20': sma20,
                        'boll_upper': boll_upper,
                        'boll_lower': boll_lower,
                        'rsi': rsi
                    }
                )
        
        if signal:
            self.signals.append(signal)
        
        return signal

class BreakoutAgent(BaseAgent):
    """突破交易Agent"""
    
    def analyze(self, data: pd.DataFrame) -> Optional[TradeSignal]:
        if len(data) < 50:
            return None
        
        # 计算技术指标
        data['atr'] = calculate_atr(data['high'], data['low'], data['close'], 14)
        data['highest'] = data['high'].rolling(20).max()
        data['lowest'] = data['low'].rolling(20).min()
        data['volume_sma'] = calculate_sma(data['volume'], 20)
        
        current_price = data['close'].iloc[-1]
        current_volume = data['volume'].iloc[-1]
        atr = data['atr'].iloc[-1]
        highest = data['highest'].iloc[-2]  # 使用前一个周期的高点
        lowest = data['lowest'].iloc[-2]  # 使用前一个周期的低点
        volume_sma = data['volume_sma'].iloc[-1]
        
        # 生成信号
        signal = None
        if (current_price > highest and 
            current_volume > volume_sma * 1.5):  # 向上突破
            stop_loss = current_price - (atr * 2)
            take_profit = current_price + (atr * 4)
            confidence = min(1.0, (current_price - highest) / highest * 10)
            
            if confidence >= self.config.confidence_threshold:
                signal = TradeSignal(
                    symbol=self.config.symbol,
                    direction='buy',
                    price=current_price,
                    stop_loss=stop_loss,
                    take_profit=take_profit,
                    size=self._calculate_position_size(current_price, stop_loss),
                    confidence=confidence,
                    agent_name=self.config.name,
                    timestamp=data.index[-1],
                    metadata={
                        'strategy': 'breakout',
                        'atr': atr,
                        'highest': highest,
                        'volume_ratio': current_volume / volume_sma
                    }
                )
        
        elif (current_price < lowest and 
              current_volume > volume_sma * 1.5):  # 向下突破
            stop_loss = current_price + (atr * 2)
            take_profit = current_price - (atr * 4)
            confidence = min(1.0, (lowest - current_price) / lowest * 10)
            
            if confidence >= self.config.confidence_threshold:
                signal = TradeSignal(
                    symbol=self.config.symbol,
                    direction='sell',
                    price=current_price,
                    stop_loss=stop_loss,
                    take_profit=take_profit,
                    size=self._calculate_position_size(current_price, stop_loss),
                    confidence=confidence,
                    agent_name=self.config.name,
                    timestamp=data.index[-1],
                    metadata={
                        'strategy': 'breakout',
                        'atr': atr,
                        'lowest': lowest,
                        'volume_ratio': current_volume / volume_sma
                    }
                )
        
        if signal:
            self.signals.append(signal)
        
        return signal

class AgentSystem:
    """Agent系统"""
    
    def __init__(self):
        self.agents: Dict[str, BaseAgent] = {}
        self.agent_performance: Dict[str, Dict] = {}
        
        # 初始化数据库
        self.db_path = "agent_system.db"
        self._initialize_database()
    
    def _initialize_database(self):
        """初始化数据库"""
        with sqlite3.connect(self.db_path) as conn:
            # Agent配置表
            conn.execute("""
                CREATE TABLE IF NOT EXISTS agent_config (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    name TEXT UNIQUE,
                    symbol TEXT,
                    timeframe TEXT,
                    strategy_type TEXT,
                    parameters TEXT,
                    confidence_threshold REAL,
                    risk_limit REAL,
                    max_positions INTEGER,
                    enable_ml BOOLEAN
                )
            """)
            
            # 信号记录表
            conn.execute("""
                CREATE TABLE IF NOT EXISTS signals (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    timestamp DATETIME,
                    symbol TEXT,
                    direction TEXT,
                    price REAL,
                    stop_loss REAL,
                    take_profit REAL,
                    size REAL,
                    confidence REAL,
                    agent_name TEXT,
                    metadata TEXT
                )
            """)
            
            # 性能记录表
            conn.execute("""
                CREATE TABLE IF NOT EXISTS performance (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    timestamp DATETIME,
                    agent_name TEXT,
                    total_signals INTEGER,
                    successful_signals INTEGER,
                    failed_signals INTEGER,
                    total_pnl REAL,
                    win_rate REAL,
                    avg_return REAL
                )
            """)
            
            conn.commit()
    
    def add_agent(self, config: AgentConfig):
        """添加Agent"""
        # 创建Agent实例
        if config.strategy_type == 'trend_following':
            agent = TrendFollowingAgent(config)
        elif config.strategy_type == 'mean_reversion':
            agent = MeanReversionAgent(config)
        elif config.strategy_type == 'breakout':
            agent = BreakoutAgent(config)
        else:
            raise ValueError(f"Unknown strategy type: {config.strategy_type}")
        
        # 保存Agent配置
        with sqlite3.connect(self.db_path) as conn:
            conn.execute("""
                INSERT INTO agent_config (
                    name, symbol, timeframe, strategy_type,
                    parameters, confidence_threshold,
                    risk_limit, max_positions, enable_ml
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
            """, (
                config.name, config.symbol, config.timeframe,
                config.strategy_type, json.dumps(config.parameters),
                config.confidence_threshold, config.risk_limit,
                config.max_positions, config.enable_ml
            ))
            conn.commit()
        
        self.agents[config.name] = agent
        logger.info(f"Added agent: {config.name}")
    
    def remove_agent(self, agent_name: str):
        """移除Agent"""
        if agent_name in self.agents:
            del self.agents[agent_name]
            
            # 从数据库删除Agent配置
            with sqlite3.connect(self.db_path) as conn:
                conn.execute("""
                    DELETE FROM agent_config WHERE name = ?
                """, (agent_name,))
                conn.commit()
            
            logger.info(f"Removed agent: {agent_name}")
    
    def get_agent(self, agent_name: str) -> Optional[BaseAgent]:
        """获取Agent"""
        return self.agents.get(agent_name)
    
    def get_all_agents(self) -> List[str]:
        """获取所有Agent名称"""
        return list(self.agents.keys())
    
    def save_signal(self, signal: TradeSignal):
        """保存交易信号"""
        with sqlite3.connect(self.db_path) as conn:
            conn.execute("""
                INSERT INTO signals (
                    timestamp, symbol, direction, price,
                    stop_loss, take_profit, size,
                    confidence, agent_name, metadata
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            """, (
                signal.timestamp, signal.symbol, signal.direction,
                signal.price, signal.stop_loss, signal.take_profit,
                signal.size, signal.confidence, signal.agent_name,
                json.dumps(signal.metadata)
            ))
            conn.commit()
    
    def update_agent_performance(self, agent_name: str, metrics: Dict):
        """更新Agent性能"""
        self.agent_performance[agent_name] = metrics
        
        # 保存到数据库
        with sqlite3.connect(self.db_path) as conn:
            conn.execute("""
                INSERT INTO performance (
                    timestamp, agent_name, total_signals,
                    successful_signals, failed_signals,
                    total_pnl, win_rate, avg_return
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
            """, (
                datetime.now(), agent_name,
                metrics['total_signals'], metrics['successful_signals'],
                metrics['failed_signals'], metrics['total_pnl'],
                metrics['win_rate'], metrics['avg_return']
            ))
            conn.commit()
    
    def get_agent_performance(self, agent_name: str,
                            start_time: Optional[datetime] = None,
                            end_time: Optional[datetime] = None) -> pd.DataFrame:
        """获取Agent性能数据"""
        with sqlite3.connect(self.db_path) as conn:
            query = """
                SELECT * FROM performance
                WHERE agent_name = ?
            """
            params = [agent_name]
            
            if start_time:
                query += " AND timestamp >= ?"
                params.append(start_time)
            if end_time:
                query += " AND timestamp <= ?"
                params.append(end_time)
            
            query += " ORDER BY timestamp"
            
            return pd.read_sql_query(query, conn, params=params)
    
    def get_agent_signals(self, agent_name: str,
                         start_time: Optional[datetime] = None,
                         end_time: Optional[datetime] = None) -> pd.DataFrame:
        """获取Agent信号数据"""
        with sqlite3.connect(self.db_path) as conn:
            query = """
                SELECT * FROM signals
                WHERE agent_name = ?
            """
            params = [agent_name]
            
            if start_time:
                query += " AND timestamp >= ?"
                params.append(start_time)
            if end_time:
                query += " AND timestamp <= ?"
                params.append(end_time)
            
            query += " ORDER BY timestamp"
            
            return pd.read_sql_query(query, conn, params=params)
    
    def get_system_metrics(self) -> Dict:
        """获取系统指标"""
        metrics = {
            'total_agents': len(self.agents),
            'active_agents': sum(1 for agent in self.agents.values()
                               if agent.performance_metrics['total_signals'] > 0),
            'total_signals': sum(agent.performance_metrics['total_signals']
                               for agent in self.agents.values()),
            'system_win_rate': 0.0,
            'system_avg_return': 0.0
        }
        
        # 计算系统级别的胜率和平均收益
        total_successful = sum(agent.performance_metrics['successful_signals']
                             for agent in self.agents.values())
        total_signals = metrics['total_signals']
        
        if total_signals > 0:
            metrics['system_win_rate'] = total_successful / total_signals
            metrics['system_avg_return'] = sum(
                agent.performance_metrics['total_pnl']
                for agent in self.agents.values()
            ) / total_signals
        
        return metrics  