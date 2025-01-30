import logging
import pandas as pd
import numpy as np
from typing import Dict, List, Optional, Union
from datetime import datetime, timedelta
from dataclasses import dataclass
import ccxt.async_support as ccxt
import asyncio
import json
import sqlite3
from pathlib import Path
import websockets
from collections import deque

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@dataclass
class MarketConfig:
    """市场配置"""
    exchange: str
    symbols: List[str]
    timeframes: List[str]
    api_key: Optional[str] = None
    api_secret: Optional[str] = None
    cache_size: int = 1000
    update_interval: float = 1.0  # 秒
    db_path: str = "market_data.db"

class MarketDataService:
    """市场数据服务"""
    
    def __init__(self, config: MarketConfig):
        self.config = config
        self.db_path = Path(config.db_path)
        
        # 初始化交易所
        self.exchange = getattr(ccxt, config.exchange.lower())({
            'apiKey': config.api_key,
            'secret': config.api_secret,
            'enableRateLimit': True,
            'options': {'defaultType': 'spot'}
        })
        
        # 数据缓存
        self.ohlcv_cache: Dict[str, Dict[str, deque]] = {}
        self.orderbook_cache: Dict[str, Dict] = {}
        self.trades_cache: Dict[str, deque] = {}
        
        # WebSocket连接
        self.ws_connections: Dict[str, websockets.WebSocketClientProtocol] = {}
        self.ws_subscriptions: Dict[str, List[str]] = {}
        
        # 事件订阅者
        self.subscribers: Dict[str, List[callable]] = {
            'ohlcv': [],
            'orderbook': [],
            'trades': [],
            'ticker': []
        }
        
        # 初始化数据库
        self._initialize_database()
        
        # 服务状态
        self.is_running = False
    
    def _initialize_database(self):
        """初始化数据库"""
        with sqlite3.connect(self.db_path) as conn:
            # OHLCV数据表
            conn.execute("""
                CREATE TABLE IF NOT EXISTS ohlcv (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    timestamp DATETIME,
                    symbol TEXT,
                    timeframe TEXT,
                    open REAL,
                    high REAL,
                    low REAL,
                    close REAL,
                    volume REAL
                )
            """)
            
            # 订单簿数据表
            conn.execute("""
                CREATE TABLE IF NOT EXISTS orderbook (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    timestamp DATETIME,
                    symbol TEXT,
                    bids TEXT,
                    asks TEXT,
                    bids_volume REAL,
                    asks_volume REAL
                )
            """)
            
            # 交易数据表
            conn.execute("""
                CREATE TABLE IF NOT EXISTS trades (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    timestamp DATETIME,
                    symbol TEXT,
                    price REAL,
                    amount REAL,
                    side TEXT,
                    trade_id TEXT
                )
            """)
            
            conn.commit()
    
    async def start(self):
        """启动服务"""
        logger.info("Starting market data service...")
        self.is_running = True
        
        # 初始化缓存
        for symbol in self.config.symbols:
            self.ohlcv_cache[symbol] = {}
            for timeframe in self.config.timeframes:
                self.ohlcv_cache[symbol][timeframe] = deque(maxlen=self.config.cache_size)
            
            self.orderbook_cache[symbol] = {
                'timestamp': None,
                'bids': [],
                'asks': []
            }
            self.trades_cache[symbol] = deque(maxlen=self.config.cache_size)
        
        # 加载历史数据
        await self._load_historical_data()
        
        # 启动WebSocket连接
        await self._start_websocket()
        
        # 启动数据更新任务
        asyncio.create_task(self._update_market_data())
    
    async def stop(self):
        """停止服务"""
        logger.info("Stopping market data service...")
        self.is_running = False
        
        # 关闭WebSocket连接
        for ws in self.ws_connections.values():
            await ws.close()
        
        # 关闭交易所连接
        await self.exchange.close()
    
    async def _load_historical_data(self):
        """加载历史数据"""
        for symbol in self.config.symbols:
            for timeframe in self.config.timeframes:
                try:
                    # 获取最近的1000条K线数据
                    ohlcv = await self.exchange.fetch_ohlcv(
                        symbol, timeframe, limit=1000
                    )
                    
                    # 转换为pandas DataFrame
                    df = pd.DataFrame(
                        ohlcv,
                        columns=['timestamp', 'open', 'high', 'low', 'close', 'volume']
                    )
                    df['timestamp'] = pd.to_datetime(df['timestamp'], unit='ms')
                    
                    # 保存到缓存
                    for _, row in df.iterrows():
                        self.ohlcv_cache[symbol][timeframe].append(row.to_dict())
                    
                    # 保存到数据库
                    with sqlite3.connect(self.db_path) as conn:
                        for _, row in df.iterrows():
                            conn.execute("""
                                INSERT INTO ohlcv (
                                    timestamp, symbol, timeframe,
                                    open, high, low, close, volume
                                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                            """, (
                                row['timestamp'], symbol, timeframe,
                                row['open'], row['high'], row['low'],
                                row['close'], row['volume']
                            ))
                        conn.commit()
                    
                except Exception as e:
                    logger.error(f"Error loading historical data for {symbol} {timeframe}: {str(e)}")
    
    async def _start_websocket(self):
        """启动WebSocket连接"""
        if not hasattr(self.exchange, 'ws'):
            logger.warning(f"Exchange {self.config.exchange} does not support WebSocket")
            return
        
        for symbol in self.config.symbols:
            try:
                # 创建WebSocket连接
                ws = await self.exchange.ws.connect()
                self.ws_connections[symbol] = ws
                
                # 订阅数据
                channels = ['ticker', 'orderbook', 'trades']
                subscriptions = []
                for channel in channels:
                    sub = await ws.subscribe(channel, symbol)
                    subscriptions.append(sub)
                
                self.ws_subscriptions[symbol] = subscriptions
                
                # 启动消息处理
                asyncio.create_task(self._handle_ws_messages(symbol, ws))
                
            except Exception as e:
                logger.error(f"Error starting WebSocket for {symbol}: {str(e)}")
    
    async def _handle_ws_messages(self, symbol: str, ws: websockets.WebSocketClientProtocol):
        """处理WebSocket消息"""
        try:
            async for message in ws:
                if not self.is_running:
                    break
                
                data = json.loads(message)
                
                # 处理不同类型的消息
                if 'type' in data:
                    if data['type'] == 'ticker':
                        await self._handle_ticker(symbol, data)
                    elif data['type'] == 'orderbook':
                        await self._handle_orderbook(symbol, data)
                    elif data['type'] == 'trade':
                        await self._handle_trade(symbol, data)
                
        except Exception as e:
            logger.error(f"Error handling WebSocket messages for {symbol}: {str(e)}")
            
            # 尝试重新连接
            await asyncio.sleep(5)
            await self._start_websocket()
    
    async def _update_market_data(self):
        """更新市场数据"""
        while self.is_running:
            try:
                for symbol in self.config.symbols:
                    # 更新K线数据
                    for timeframe in self.config.timeframes:
                        ohlcv = await self.exchange.fetch_ohlcv(
                            symbol, timeframe, limit=1
                        )
                        if ohlcv:
                            latest = ohlcv[0]
                            data = {
                                'timestamp': pd.to_datetime(latest[0], unit='ms'),
                                'open': latest[1],
                                'high': latest[2],
                                'low': latest[3],
                                'close': latest[4],
                                'volume': latest[5]
                            }
                            self.ohlcv_cache[symbol][timeframe].append(data)
                            
                            # 通知订阅者
                            await self._notify_subscribers('ohlcv', symbol, data)
                    
                    # 更新订单簿
                    if not self.ws_connections.get(symbol):
                        orderbook = await self.exchange.fetch_order_book(symbol)
                        self.orderbook_cache[symbol] = {
                            'timestamp': pd.Timestamp.now(),
                            'bids': orderbook['bids'],
                            'asks': orderbook['asks']
                        }
                        
                        # 通知订阅者
                        await self._notify_subscribers('orderbook', symbol, self.orderbook_cache[symbol])
                
                await asyncio.sleep(self.config.update_interval)
                
            except Exception as e:
                logger.error(f"Error updating market data: {str(e)}")
                await asyncio.sleep(self.config.update_interval)
    
    async def _handle_ticker(self, symbol: str, data: Dict):
        """处理Ticker数据"""
        # 通知订阅者
        await self._notify_subscribers('ticker', symbol, data)
    
    async def _handle_orderbook(self, symbol: str, data: Dict):
        """处理订单簿数据"""
        self.orderbook_cache[symbol] = {
            'timestamp': pd.Timestamp.now(),
            'bids': data['bids'],
            'asks': data['asks']
        }
        
        # 保存到数据库
        with sqlite3.connect(self.db_path) as conn:
            conn.execute("""
                INSERT INTO orderbook (
                    timestamp, symbol, bids, asks,
                    bids_volume, asks_volume
                ) VALUES (?, ?, ?, ?, ?, ?)
            """, (
                self.orderbook_cache[symbol]['timestamp'],
                symbol,
                json.dumps(data['bids']),
                json.dumps(data['asks']),
                sum(bid[1] for bid in data['bids']),
                sum(ask[1] for ask in data['asks'])
            ))
            conn.commit()
        
        # 通知订阅者
        await self._notify_subscribers('orderbook', symbol, self.orderbook_cache[symbol])
    
    async def _handle_trade(self, symbol: str, data: Dict):
        """处理交易数据"""
        trade = {
            'timestamp': pd.Timestamp.now(),
            'price': data['price'],
            'amount': data['amount'],
            'side': data['side'],
            'trade_id': data.get('id')
        }
        
        self.trades_cache[symbol].append(trade)
        
        # 保存到数据库
        with sqlite3.connect(self.db_path) as conn:
            conn.execute("""
                INSERT INTO trades (
                    timestamp, symbol, price,
                    amount, side, trade_id
                ) VALUES (?, ?, ?, ?, ?, ?)
            """, (
                trade['timestamp'],
                symbol,
                trade['price'],
                trade['amount'],
                trade['side'],
                trade['trade_id']
            ))
            conn.commit()
        
        # 通知订阅者
        await self._notify_subscribers('trades', symbol, trade)
    
    async def _notify_subscribers(self, event_type: str, symbol: str, data: Dict):
        """通知订阅者"""
        for callback in self.subscribers[event_type]:
            try:
                await callback(symbol, data)
            except Exception as e:
                logger.error(f"Error notifying subscriber: {str(e)}")
    
    def subscribe(self, event_type: str, callback: callable):
        """订阅事件"""
        if event_type in self.subscribers:
            self.subscribers[event_type].append(callback)
    
    def unsubscribe(self, event_type: str, callback: callable):
        """取消订阅"""
        if event_type in self.subscribers:
            self.subscribers[event_type].remove(callback)
    
    def get_ohlcv(self, symbol: str, timeframe: str,
                  start_time: Optional[datetime] = None,
                  end_time: Optional[datetime] = None) -> pd.DataFrame:
        """获取K线数据"""
        if start_time is None:
            # 返回缓存数据
            data = list(self.ohlcv_cache[symbol][timeframe])
            return pd.DataFrame(data)
        
        # 从数据库查询
        with sqlite3.connect(self.db_path) as conn:
            query = """
                SELECT * FROM ohlcv
                WHERE symbol = ? AND timeframe = ?
                AND timestamp BETWEEN ? AND ?
                ORDER BY timestamp
            """
            return pd.read_sql_query(
                query, conn,
                params=(symbol, timeframe, start_time, end_time)
            )
    
    def get_orderbook(self, symbol: str) -> Dict:
        """获取订单簿"""
        return self.orderbook_cache[symbol]
    
    def get_trades(self, symbol: str,
                  start_time: Optional[datetime] = None,
                  end_time: Optional[datetime] = None) -> pd.DataFrame:
        """获取交易数据"""
        if start_time is None:
            # 返回缓存数据
            data = list(self.trades_cache[symbol])
            return pd.DataFrame(data)
        
        # 从数据库查询
        with sqlite3.connect(self.db_path) as conn:
            query = """
                SELECT * FROM trades
                WHERE symbol = ?
                AND timestamp BETWEEN ? AND ?
                ORDER BY timestamp
            """
            return pd.read_sql_query(
                query, conn,
                params=(symbol, start_time, end_time)
            )
    
    def get_latest_price(self, symbol: str) -> Optional[float]:
        """获取最新价格"""
        if symbol in self.trades_cache and self.trades_cache[symbol]:
            return self.trades_cache[symbol][-1]['price']
        return None
    
    def get_market_depth(self, symbol: str, depth: int = 10) -> Dict:
        """获取市场深度"""
        orderbook = self.orderbook_cache[symbol]
        return {
            'timestamp': orderbook['timestamp'],
            'bids': orderbook['bids'][:depth],
            'asks': orderbook['asks'][:depth]
        }
    
    def calculate_vwap(self, symbol: str, window: int = 100) -> float:
        """计算成交量加权平均价格"""
        trades = list(self.trades_cache[symbol])[-window:]
        if not trades:
            return None
        
        volume_price = sum(trade['price'] * trade['amount'] for trade in trades)
        volume = sum(trade['amount'] for trade in trades)
        return volume_price / volume if volume > 0 else None 