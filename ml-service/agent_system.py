import logging
import pandas as pd
import numpy as np
from typing import Dict, List, Optional, Union, Any
from datetime import datetime, timedelta
from dataclasses import dataclass, field
import json
import sqlite3
from pathlib import Path
from abc import ABC, abstractmethod
from collections import deque
from decimal import Decimal

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def calculate_sma(data: pd.Series, period: int) -> pd.Series:
    """Calculate Simple Moving Average"""
    return data.rolling(window=period).mean()

def calculate_rsi(data: pd.Series, period: int = 14) -> pd.Series:
    """Calculate Relative Strength Index"""
    delta = pd.Series(data.diff(), dtype=float)
    gains = delta.clip(lower=0)
    losses = -delta.clip(upper=0)
    avg_gains = gains.rolling(window=period, min_periods=1).mean()
    avg_losses = losses.rolling(window=period, min_periods=1).mean()
    rs = avg_gains / avg_losses.replace(0, float('inf'))
    return pd.Series(100 - (100 / (1 + rs)), dtype=float)

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
    parameters: Dict[str, Any]
    confidence_threshold: float = 0.6
    risk_limit: float = 0.02
    max_positions: int = 1
    enable_ml: bool = False
    margin_type: str = 'isolated'
    max_leverage: int = 10
    funding_rate_threshold: float = 0.01
    liquidation_threshold: float = 0.8
    maintenance_margin: float = 0.05
    metadata: Dict[str, Any] = field(default_factory=dict)

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
    metadata: Dict[str, Any] = field(default_factory=dict)

    def __post_init__(self):
        if self.direction not in ('buy', 'sell'):
            raise ValueError("Direction must be 'buy' or 'sell'")
        if not isinstance(self.price, (int, float, Decimal)) or self.price <= 0:
            raise ValueError("Invalid price")
        if not isinstance(self.size, (int, float, Decimal)) or self.size <= 0:
            raise ValueError("Invalid size")
        if not isinstance(self.confidence, (int, float)) or not 0 <= self.confidence <= 1:
            raise ValueError("Confidence must be between 0 and 1")

class BaseAgent(ABC):
    """Agent基类"""

    def __init__(self, config: AgentConfig):
        self.config = config
        self.positions: List[Dict[str, Any]] = []
        self.signals: deque[TradeSignal] = deque(maxlen=1000)
        self.performance_metrics: Dict[str, Union[int, float]] = {
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

    def update_performance(self, signal: TradeSignal, success: bool, return_pct: float) -> None:
        """更新性能指标"""
        if not isinstance(signal, TradeSignal):
            raise TypeError("signal must be a TradeSignal instance")
        if not isinstance(success, bool):
            raise TypeError("success must be a boolean")
        if not isinstance(return_pct, (int, float)):
            raise TypeError("return_pct must be a number")

        self.performance_metrics['total_signals'] += 1
        if success:
            self.performance_metrics['successful_signals'] += 1
        else:
            self.performance_metrics['failed_signals'] += 1

        self.performance_metrics['total_pnl'] += float(return_pct)

        total_signals = self.performance_metrics['total_signals']
        if total_signals > 0:
            self.performance_metrics['win_rate'] = (
                self.performance_metrics['successful_signals'] / total_signals
            )
            self.performance_metrics['avg_return'] = (
                self.performance_metrics['total_pnl'] / total_signals
            )

    def _calculate_position_size(self, price: float, stop_loss: float) -> float:
        """计算仓位大小"""
        if not isinstance(price, (int, float)) or price <= 0:
            raise ValueError("price must be a positive number")
        if not isinstance(stop_loss, (int, float)):
            raise ValueError("stop_loss must be a number")

        risk_amount = float(self.config.risk_limit * 100000)
        price_risk = abs(float(price) - float(stop_loss))
        if price_risk == 0:
            raise ValueError("price_risk cannot be zero")
        return risk_amount / price_risk

    def _calculate_stop_loss(self, data: pd.DataFrame, direction: str) -> float:
        """计算止损价格"""
        if not isinstance(data, pd.DataFrame):
            raise TypeError("data must be a pandas DataFrame")
        if direction not in ('buy', 'sell'):
            raise ValueError("direction must be 'buy' or 'sell'")

        if direction == 'buy':
            return float(data['low'].iloc[-10:].min())
        else:
            return float(data['high'].iloc[-10:].max())

    def _calculate_take_profit(self, entry_price: float, stop_loss: float) -> float:
        """计算止盈价格"""
        if not isinstance(entry_price, (int, float)) or entry_price <= 0:
            raise ValueError("entry_price must be a positive number")
        if not isinstance(stop_loss, (int, float)):
            raise ValueError("stop_loss must be a number")

        risk = abs(float(entry_price) - float(stop_loss))
        return entry_price + (risk * 2)

class TrendFollowingAgent(BaseAgent):
    """趋势跟踪Agent"""

    def analyze(self, data: pd.DataFrame) -> Optional[TradeSignal]:
        if not isinstance(data, pd.DataFrame):
            raise TypeError("data must be a pandas DataFrame")
        if len(data) < 50:
            return None

        required_columns = ['close', 'high', 'low']
        if not all(col in data.columns for col in required_columns):
            raise ValueError(f"DataFrame must contain columns: {required_columns}")

        try:
            data['sma20'] = calculate_sma(data['close'], 20)
            data['sma50'] = calculate_sma(data['close'], 50)
            data['rsi'] = calculate_rsi(data['close'], 14)
            data['atr'] = calculate_atr(data['high'], data['low'], data['close'], 14)

            current_price = float(data['close'].iloc[-1])
            sma20 = float(data['sma20'].iloc[-1])
            sma50 = float(data['sma50'].iloc[-1])
            rsi = float(data['rsi'].iloc[-1])
            atr = float(data['atr'].iloc[-1])

            if pd.isna(current_price) or pd.isna(sma20) or pd.isna(sma50) or pd.isna(rsi) or pd.isna(atr):
                logger.warning(f"NaN values detected in technical indicators for {self.config.symbol}")
                return None

            signal = None
            if sma20 > sma50 and rsi > 50:
                stop_loss = current_price - (atr * 2)
                take_profit = current_price + (atr * 4)
                confidence = min(1.0, (sma20 - sma50) / sma50 * 5)

                if confidence >= self.config.confidence_threshold:
                    try:
                        position_size = self._calculate_position_size(current_price, stop_loss)
                        signal = TradeSignal(
                            symbol=self.config.symbol,
                            direction='buy',
                            price=current_price,
                            stop_loss=stop_loss,
                            take_profit=take_profit,
                            size=position_size,
                            confidence=confidence,
                            agent_name=self.config.name,
                            timestamp=data.index[-1],
                            metadata={
                                'strategy': 'trend_following',
                                'sma20': float(sma20),
                                'sma50': float(sma50),
                                'rsi': float(rsi),
                                'atr': float(atr)
                            }
                        )
                    except (ValueError, TypeError) as e:
                        logger.error(f"Error creating buy signal: {e}")
                        return None

            elif sma20 < sma50 and rsi < 50:
                stop_loss = current_price + (atr * 2)
                take_profit = current_price - (atr * 4)
                confidence = min(1.0, (sma50 - sma20) / sma50 * 5)

                if confidence >= self.config.confidence_threshold:
                    try:
                        position_size = self._calculate_position_size(current_price, stop_loss)
                        signal = TradeSignal(
                            symbol=self.config.symbol,
                            direction='sell',
                            price=current_price,
                            stop_loss=stop_loss,
                            take_profit=take_profit,
                            size=position_size,
                            confidence=confidence,
                            agent_name=self.config.name,
                            timestamp=data.index[-1],
                            metadata={
                                'strategy': 'trend_following',
                                'sma20': float(sma20),
                                'sma50': float(sma50),
                                'rsi': float(rsi),
                                'atr': float(atr)
                            }
                        )
                    except (ValueError, TypeError) as e:
                        logger.error(f"Error creating sell signal: {e}")
                        return None

            if signal:
                self.signals.append(signal)

            return signal

        except Exception as e:
            logger.error(f"Error in TrendFollowingAgent analysis: {e}")
            return None

class MeanReversionAgent(BaseAgent):
    """均值回归Agent"""

    def analyze(self, data: pd.DataFrame) -> Optional[TradeSignal]:
        if not isinstance(data, pd.DataFrame):
            raise TypeError("data must be a pandas DataFrame")
        if len(data) < 50:
            return None

        required_columns = ['close']
        if not all(col in data.columns for col in required_columns):
            raise ValueError(f"DataFrame must contain columns: {required_columns}")

        try:
            data['sma20'] = calculate_sma(data['close'], 20)
            data['boll_upper'], data['boll_middle'], data['boll_lower'] = calculate_bbands(
                data['close'], period=20, num_std=2
            )
            data['rsi'] = calculate_rsi(data['close'], 14)

            current_price = float(data['close'].iloc[-1])
            sma20 = float(data['sma20'].iloc[-1])
            boll_upper = float(data['boll_upper'].iloc[-1])
            boll_lower = float(data['boll_lower'].iloc[-1])
            rsi = float(data['rsi'].iloc[-1])

            if pd.isna(current_price) or pd.isna(sma20) or pd.isna(boll_upper) or pd.isna(boll_lower) or pd.isna(rsi):
                logger.warning(f"NaN values detected in technical indicators for {self.config.symbol}")
                return None

            signal = None
            if current_price < boll_lower and rsi < 30:  # 超卖，买入信号
                stop_loss = current_price * 0.99  # 1%止损
                take_profit = sma20  # 均线作为目标
                confidence = min(1.0, (boll_lower - current_price) / current_price * 10)

                if confidence >= self.config.confidence_threshold:
                    try:
                        position_size = self._calculate_position_size(current_price, stop_loss)
                        signal = TradeSignal(
                            symbol=self.config.symbol,
                            direction='buy',
                            price=current_price,
                            stop_loss=stop_loss,
                            take_profit=take_profit,
                            size=position_size,
                            confidence=confidence,
                            agent_name=self.config.name,
                            timestamp=data.index[-1],
                            metadata={
                                'strategy': 'mean_reversion',
                                'sma20': float(sma20),
                                'boll_upper': float(boll_upper),
                                'boll_lower': float(boll_lower),
                                'rsi': float(rsi)
                            }
                        )
                    except (ValueError, TypeError) as e:
                        logger.error(f"Error creating buy signal: {e}")
                        return None

            elif current_price > boll_upper and rsi > 70:  # 超买，卖出信号
                stop_loss = current_price * 1.01  # 1%止损
                take_profit = sma20  # 均线作为目标
                confidence = min(1.0, (current_price - boll_upper) / current_price * 10)

                if confidence >= self.config.confidence_threshold:
                    try:
                        position_size = self._calculate_position_size(current_price, stop_loss)
                        signal = TradeSignal(
                            symbol=self.config.symbol,
                            direction='sell',
                            price=current_price,
                            stop_loss=stop_loss,
                            take_profit=take_profit,
                            size=position_size,
                            confidence=confidence,
                            agent_name=self.config.name,
                            timestamp=data.index[-1],
                            metadata={
                                'strategy': 'mean_reversion',
                                'sma20': float(sma20),
                                'boll_upper': float(boll_upper),
                                'boll_lower': float(boll_lower),
                                'rsi': float(rsi)
                            }
                        )
                    except (ValueError, TypeError) as e:
                        logger.error(f"Error creating sell signal: {e}")
                        return None

            if signal:
                self.signals.append(signal)

            return signal

        except Exception as e:
            logger.error(f"Error in MeanReversionAgent analysis: {e}")
            return None

class BreakoutAgent(BaseAgent):
    """突破交易Agent"""

    def analyze(self, data: pd.DataFrame) -> Optional[TradeSignal]:
        if not isinstance(data, pd.DataFrame):
            raise TypeError("data must be a pandas DataFrame")
        if len(data) < 50:
            return None

        required_columns = ['close', 'high', 'low', 'volume']
        if not all(col in data.columns for col in required_columns):
            raise ValueError(f"DataFrame must contain columns: {required_columns}")

        try:
            data['atr'] = calculate_atr(data['high'], data['low'], data['close'], 14)
            data['highest'] = data['high'].rolling(20).max()
            data['lowest'] = data['low'].rolling(20).min()
            data['volume_sma'] = calculate_sma(data['volume'], 20)

            current_price = float(data['close'].iloc[-1])
            current_volume = float(data['volume'].iloc[-1])
            atr = float(data['atr'].iloc[-1])
            highest = float(data['highest'].iloc[-2])  # 使用前一个周期的高点
            lowest = float(data['lowest'].iloc[-2])  # 使用前一个周期的低点
            volume_sma = float(data['volume_sma'].iloc[-1])

            # Validate data integrity
            if any(pd.isna(x) for x in [current_price, current_volume, atr, highest, lowest, volume_sma]):
                logger.warning(f"NaN values detected in technical indicators for {self.config.symbol}")
                return None

            if current_price <= 0 or current_volume <= 0 or atr <= 0:
                logger.error(f"Invalid negative or zero values detected for {self.config.symbol}")
                return None

            signal = None
            if (current_price > highest and
                current_volume > volume_sma * 1.5):  # 向上突破
                try:
                    stop_loss = current_price - (atr * 2)
                    take_profit = current_price + (atr * 4)
                    confidence = min(1.0, (current_price - highest) / highest * 10)

                    if confidence >= self.config.confidence_threshold:
                        position_size = self._calculate_position_size(current_price, stop_loss)
                        signal = TradeSignal(
                            symbol=self.config.symbol,
                            direction='buy',
                            price=current_price,
                            stop_loss=stop_loss,
                            take_profit=take_profit,
                            size=position_size,
                            confidence=confidence,
                            agent_name=self.config.name,
                            timestamp=data.index[-1],
                            metadata={
                                'strategy': 'breakout',
                                'atr': float(atr),
                                'highest': float(highest),
                                'volume_ratio': float(current_volume / volume_sma)
                            }
                        )
                except (ValueError, TypeError) as e:
                    logger.error(f"Error creating buy signal: {e}")
                    return None

            elif (current_price < lowest and
                  current_volume > volume_sma * 1.5):  # 向下突破
                try:
                    stop_loss = float(current_price + (atr * 2))
                    take_profit = float(current_price - (atr * 4))
                    confidence = float(min(1.0, (lowest - current_price) / lowest * 10))

                    if confidence >= self.config.confidence_threshold:
                        position_size = self._calculate_position_size(current_price, stop_loss)
                        signal = TradeSignal(
                            symbol=self.config.symbol,
                            direction='sell',
                            price=float(current_price),
                            stop_loss=stop_loss,
                            take_profit=take_profit,
                            size=float(position_size),
                            confidence=confidence,
                            agent_name=self.config.name,
                            timestamp=data.index[-1],
                            metadata={
                                'strategy': 'breakout',
                                'atr': float(atr),
                                'lowest': float(lowest),
                                'volume_ratio': float(current_volume / volume_sma)
                            }
                        )
                except (ValueError, TypeError) as e:
                    logger.error(f"Error creating sell signal: {e}")
                    return None
                except Exception as e:
                    logger.error(f"Unexpected error creating sell signal: {e}", exc_info=True)
                    return None

            if signal:
                try:
                    if not isinstance(signal, TradeSignal):
                        logger.error(f"Invalid signal type: {type(signal)}")
                        return None
                    self.signals.append(signal)
                    logger.info(f"Generated {signal.direction} signal for {signal.symbol} with confidence {signal.confidence:.2f}")
                except Exception as e:
                    logger.error(f"Error appending signal: {e}", exc_info=True)
                    return None

            return signal

        except Exception as e:
            logger.error(f"Error in BreakoutAgent analysis: {e}", exc_info=True)
            return None

class AgentSystem:
    """Agent系统"""

    def __init__(self):
        self.agents: Dict[str, BaseAgent] = {}
        self.agent_performance: Dict[str, Dict] = {}
        self.db_path = Path("agent_system.db").absolute()
        self._initialize_database()

        # Configure logging
        logging.basicConfig(
            level=logging.INFO,
            format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
            handlers=[
                logging.FileHandler("logs/agent_system.log"),
                logging.StreamHandler()
            ]
        )

    def _initialize_database(self):
        """初始化数据库"""
        Path("logs").mkdir(exist_ok=True)
        with sqlite3.connect(str(self.db_path)) as conn:
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
                    enable_ml BOOLEAN,
                    margin_type TEXT DEFAULT 'isolated',
                    max_leverage INTEGER DEFAULT 10,
                    funding_rate_threshold REAL DEFAULT 0.01,
                    liquidation_threshold REAL DEFAULT 0.8,
                    maintenance_margin REAL DEFAULT 0.05
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
                    margin_type TEXT,
                    leverage INTEGER,
                    funding_rate REAL,
                    liquidation_price REAL,
                    maintenance_margin REAL,
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

    def add_agent(self, config: AgentConfig) -> bool:
        """Add a new trading agent to the system"""
        if not isinstance(config, AgentConfig):
            logger.error("Invalid config type provided")
            return False

        if config.name in self.agents:
            logger.warning(f"Agent {config.name} already exists, updating configuration")

        try:
            agent_types = {
                'trend_following': TrendFollowingAgent,
                'mean_reversion': MeanReversionAgent,
                'breakout': BreakoutAgent
            }

            if config.strategy_type not in agent_types:
                logger.error(f"Unknown strategy type: {config.strategy_type}")
                return False

            agent = agent_types[config.strategy_type](config)

            with sqlite3.connect(self.db_path) as conn:
                try:
                    conn.execute("""
                        INSERT OR REPLACE INTO agent_config (
                            name, symbol, timeframe, strategy_type,
                            parameters, confidence_threshold,
                            risk_limit, max_positions, enable_ml,
                            margin_type, max_leverage, funding_rate_threshold,
                            liquidation_threshold, maintenance_margin
                        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                    """, (
                        config.name, config.symbol, config.timeframe,
                        config.strategy_type, json.dumps(config.parameters),
                        config.confidence_threshold, config.risk_limit,
                        config.max_positions, config.enable_ml,
                        config.margin_type, config.max_leverage,
                        config.funding_rate_threshold,
                        config.liquidation_threshold, config.maintenance_margin
                    ))
                    conn.commit()
                except sqlite3.Error as e:
                    logger.error(f"Database error while adding agent {config.name}: {e}")
                    return False

            self.agents[config.name] = agent
            logger.info(f"Successfully added/updated agent: {config.name}")
            return True

        except Exception as e:
            logger.error(f"Failed to add agent {config.name}: {e}", exc_info=True)
            return False

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
        try:
            if not isinstance(signal, TradeSignal):
                raise ValueError("Invalid signal type")

            with sqlite3.connect(self.db_path) as conn:
                conn.execute("""
                    INSERT INTO signals (
                        timestamp, symbol, direction, price,
                        stop_loss, take_profit, size,
                        confidence, agent_name, margin_type,
                        leverage, funding_rate, liquidation_price,
                        maintenance_margin, metadata
                    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                """, (
                    signal.timestamp.strftime('%Y-%m-%d %H:%M:%S'),
                    signal.symbol,
                    signal.direction,
                    float(signal.price),
                    float(signal.stop_loss),
                    float(signal.take_profit),
                    float(signal.size),
                    float(signal.confidence),
                    signal.agent_name,
                    signal.metadata.get('margin_type', 'isolated'),
                    int(signal.metadata.get('leverage', 1)),
                    float(signal.metadata.get('funding_rate', 0.0)),
                    float(signal.metadata.get('liquidation_price', 0.0)),
                    float(signal.metadata.get('maintenance_margin', 0.05)),
                    json.dumps(signal.metadata)
                ))
                conn.commit()
        except (sqlite3.Error, ValueError, TypeError, json.JSONDecodeError) as e:
            logger.error(f"Error saving signal: {e}")
            raise

    def update_agent_performance(self, agent_name: str, metrics: Dict):
        """更新Agent性能"""
        required_fields = {
            'total_signals': int,
            'successful_signals': int,
            'failed_signals': int,
            'total_pnl': float,
            'win_rate': float,
            'avg_return': float
        }

        try:
            # Validate metrics
            for field, field_type in required_fields.items():
                if field not in metrics:
                    raise ValueError(f"Missing required field: {field}")
                try:
                    metrics[field] = field_type(metrics[field])
                except (ValueError, TypeError):
                    raise ValueError(f"Invalid type for {field}")

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
                    datetime.now().strftime('%Y-%m-%d %H:%M:%S'),
                    agent_name,
                    metrics['total_signals'],
                    metrics['successful_signals'],
                    metrics['failed_signals'],
                    metrics['total_pnl'],
                    metrics['win_rate'],
                    metrics['avg_return']
                ))
                conn.commit()
        except Exception as e:
            logger.error(f"Error updating agent performance: {e}")
            raise

    def get_agent_performance(self, agent_name: str,
                            start_time: Optional[datetime] = None,
                            end_time: Optional[datetime] = None) -> pd.DataFrame:
        """获取Agent性能数据"""
        try:
            with sqlite3.connect(self.db_path) as conn:
                conn.row_factory = sqlite3.Row
                query = """
                    SELECT * FROM performance
                    WHERE agent_name = ?
                """
                params = [agent_name]

                if start_time:
                    query += " AND timestamp >= ?"
                    params.append(start_time.strftime('%Y-%m-%d %H:%M:%S'))
                if end_time:
                    query += " AND timestamp <= ?"
                    params.append(end_time.strftime('%Y-%m-%d %H:%M:%S'))

                query += " ORDER BY timestamp"

                df = pd.read_sql_query(query, conn, params=tuple(params), parse_dates=['timestamp'])
                if df.empty:
                    return pd.DataFrame(columns=[
                        'timestamp', 'agent_name', 'total_signals',
                        'successful_signals', 'failed_signals',
                        'total_pnl', 'win_rate', 'avg_return'
                    ])
                return df
        except sqlite3.Error as e:
            logger.error(f"Database error in get_agent_performance: {e}")
            raise

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
                params.append(start_time.strftime('%Y-%m-%d %H:%M:%S'))
            if end_time:
                query += " AND timestamp <= ?"
                params.append(end_time.strftime('%Y-%m-%d %H:%M:%S'))

            query += " ORDER BY timestamp"

            df = pd.read_sql_query(query, conn, params=tuple(params))
            df['timestamp'] = pd.to_datetime(df['timestamp'])
            return df

    def get_system_metrics(self) -> Dict[str, Union[int, float, Dict[str, float]]]:
        """Get system-wide performance metrics including perpetual trading metrics"""
        try:
            if not Path(str(self.db_path)).exists():
                logger.warning(f"Database file {self.db_path} does not exist")
                return self._get_default_metrics()

            metrics = {
                'total_agents': len(self.agents),
                'active_agents': 0,
                'total_signals': 0,
                'system_win_rate': 0.0,
                'system_avg_return': 0.0,
                'risk_metrics': {
                    'volatility': 0.0,
                    'sharpe_ratio': 0.0,
                    'max_drawdown': 0.0,
                    'win_loss_ratio': 0.0
                },
                'perpetual_metrics': {
                    'funding_rate_impact': 0.0,
                    'leverage_ratio': 0.0,
                    'position_concentration': 0.0,
                    'liquidation_risk': 0.0,
                    'margin_usage': 0.0,
                    'position_health': 0.0
                }
            }

            with sqlite3.connect(self.db_path) as conn:
                cursor = conn.execute("""
                    WITH signal_metrics AS (
                        SELECT
                            COUNT(*) as total_signals,
                            COUNT(DISTINCT agent_name) as active_agents,
                            AVG(CASE WHEN json_valid(metadata)
                                AND CAST(json_extract(metadata, '$.pnl') AS FLOAT) > 0
                                THEN 1 ELSE 0 END) as win_rate,
                            AVG(CASE WHEN json_valid(metadata)
                                THEN CAST(json_extract(metadata, '$.pnl') AS FLOAT)
                                ELSE 0 END) as avg_pnl,
                            AVG(CASE WHEN json_valid(metadata)
                                THEN CAST(json_extract(metadata, '$.funding_rate') AS FLOAT)
                                ELSE 0 END) as avg_funding_rate,
                            AVG(CASE WHEN json_valid(metadata)
                                THEN CAST(json_extract(metadata, '$.leverage') AS FLOAT)
                                ELSE 1 END) as avg_leverage,
                            AVG(CASE WHEN json_valid(metadata)
                                AND ABS(CAST(json_extract(metadata, '$.liquidation_price') AS FLOAT) - price) / price < 0.1
                                THEN 1 ELSE 0 END) as liquidation_risk,
                            AVG(CASE WHEN json_valid(metadata)
                                THEN CAST(json_extract(metadata, '$.maintenance_margin') AS FLOAT)
                                ELSE 0.05 END) as avg_margin
                        FROM signals
                        WHERE timestamp >= datetime('now', '-30 days')
                    )
                    SELECT * FROM signal_metrics
                """)
                row = cursor.fetchone()

                if row and row[0]:
                    metrics.update({
                        'total_signals': row[0],
                        'active_agents': row[1],
                        'system_win_rate': float(row[2] or 0.0),
                        'system_avg_return': float(row[3] or 0.0)
                    })

                    returns = pd.read_sql_query("""
                        SELECT
                            CAST(json_extract(metadata, '$.pnl') AS FLOAT) as pnl,
                            CAST(json_extract(metadata, '$.funding_rate') AS FLOAT) as funding_rate,
                            CAST(json_extract(metadata, '$.leverage') AS INTEGER) as leverage,
                            symbol,
                            price,
                            CAST(json_extract(metadata, '$.liquidation_price') AS FLOAT) as liquidation_price,
                            CAST(json_extract(metadata, '$.maintenance_margin') AS FLOAT) as maintenance_margin
                        FROM signals
                        WHERE timestamp >= datetime('now', '-7 days')
                        AND json_valid(metadata)
                        AND json_extract(metadata, '$.pnl') IS NOT NULL
                    """, conn)

                    if not returns.empty:
                        # Calculate risk metrics
                        returns['pnl'] = returns['pnl'].astype(float)
                        metrics['risk_metrics'].update({
                            'volatility': float(returns['pnl'].std()),
                            'sharpe_ratio': float(returns['pnl'].mean() / returns['pnl'].std() if returns['pnl'].std() > 0 else 0),
                            'max_drawdown': float(returns['pnl'].cumsum().diff().min()),
                            'win_loss_ratio': float(len(returns[returns['pnl'] > 0]) / len(returns[returns['pnl'] < 0]) if len(returns[returns['pnl'] < 0]) > 0 else 0)
                        })

                        # Calculate perpetual trading metrics
                        metrics['perpetual_metrics'].update({
                            'funding_rate_impact': float(returns['funding_rate'].mean()),
                            'leverage_ratio': float(returns['leverage'].mean()),
                            'position_concentration': float(returns.groupby('symbol').size().max() / len(returns)),
                            'liquidation_risk': float((abs(returns['liquidation_price'] - returns['price']) / returns['price'] < 0.1).mean()),
                            'margin_usage': float(returns['maintenance_margin'].mean()),
                            'position_health': float(1 - (returns['leverage'] > 10).mean())
                        })

            return metrics
        except Exception as e:
            logger.error(f"Error in get_system_metrics: {e}", exc_info=True)
            return self._get_default_metrics()

