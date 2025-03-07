import logging
import pandas as pd
import numpy as np
from typing import Dict, List, Optional, Union, Callable, Any, Tuple, TypeVar
from datetime import datetime, timedelta
from dataclasses import dataclass, field
import ccxt.async_support as ccxt
import asyncio
import json
import sqlite3
from pathlib import Path
import aiohttp
from aiohttp import ClientWebSocketResponse, WSMsgType, ClientSession
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
    update_interval: float = 1.0
    db_path: str = "market_data.db"
    perpetual_enabled: bool = False
    perpetual_symbols: List[str] = field(default_factory=list)
    funding_interval: int = 8  # hours
    max_leverage: int = 20
    default_leverage: int = 1

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
        self.perpetual_cache: Dict[str, Dict] = {}
        
        # WebSocket连接
        self.ws_connections: Dict[str, ClientWebSocketResponse] = {}
        self.ws_subscriptions: Dict[str, List[str]] = {}
        self.session: Optional[ClientSession] = None
        
        # 事件订阅者
        self.subscribers: Dict[str, List[Callable]] = {
            'ohlcv': [],
            'orderbook': [],
            'trades': [],
            'ticker': [],
            'funding': [],
            'perpetual': []
        }
        
        # 永续合约数据缓存
        self.perpetual_data: Dict[str, Dict[str, Any]] = {}
        
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
            
            # Perpetual trading data
            conn.execute("""
                CREATE TABLE IF NOT EXISTS perpetual_data (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    timestamp DATETIME,
                    symbol TEXT,
                    funding_rate REAL,
                    mark_price REAL,
                    index_price REAL,
                    open_interest REAL,
                    next_funding_time DATETIME,
                    exchange TEXT
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
        self.session = ClientSession()
        
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
        
        # 关闭HTTP会话
        if self.session:
            await self.session.close()
        
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
        if not hasattr(self.exchange, 'urls') or 'ws' not in self.exchange.urls:
            logger.warning(f"Exchange {self.config.exchange} does not support WebSocket")
            return
        
        for symbol in self.config.symbols:
            try:
                # 创建WebSocket连接
                ws = await self.session.ws_connect(self.exchange.urls['ws'])
                self.ws_connections[symbol] = ws
                
                # 订阅数据
                channels = ['ticker', 'orderbook', 'trades']
                if self.config.perpetual_enabled and symbol in self.config.perpetual_symbols:
                    channels.append('funding')
                
                for channel in channels:
                    await ws.send_json({
                        'type': 'subscribe',
                        'channel': channel,
                        'symbol': symbol
                    })
                
                self.ws_subscriptions[symbol] = channels
                
                # 启动消息处理
                asyncio.create_task(self._handle_ws_messages(symbol, ws))
                
                # 启动消息处理
                asyncio.create_task(self._handle_ws_messages(symbol, ws))
                
            except Exception as e:
                logger.error(f"Error starting WebSocket for {symbol}: {str(e)}")
    
    async def _handle_ws_messages(self, symbol: str, ws: ClientWebSocketResponse):
        """处理WebSocket消息"""
        try:
            async for msg in ws:
                if not self.is_running:
                    break
                
                if msg.type == WSMsgType.TEXT:
                    try:
                        data = json.loads(msg.data)
                        if 'type' in data:
                            if data['type'] == 'ticker':
                                await self._handle_ticker(symbol, data)
                            elif data['type'] == 'orderbook':
                                await self._handle_orderbook(symbol, data)
                            elif data['type'] == 'trade':
                                await self._handle_trade(symbol, data)
                            elif data['type'] == 'funding':
                                await self._handle_funding(symbol, data)
                    except json.JSONDecodeError as e:
                        logger.error(f"Failed to parse WebSocket message: {e}")
                elif msg.type == WSMsgType.ERROR:
                    logger.error(f"WebSocket error for {symbol}: {ws.exception()}")
                elif msg.type == WSMsgType.CLOSED:
                    logger.warning(f"WebSocket connection closed for {symbol}")
                    break
                
        except Exception as e:
            logger.error(f"Error handling WebSocket messages for {symbol}: {str(e)}")
        finally:
            if not ws.closed:
                await ws.close()
            # 尝试重新连接
            await asyncio.sleep(5)
            await self._start_websocket()
    
    async def _update_market_data(self):
        """更新市场数据"""
        while self.is_running:
            try:
                logger.debug("[MarketDataService] Starting market data update cycle")
                for symbol in self.config.symbols:
                    logger.debug(f"[MarketDataService] Updating data for {symbol}")
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
                            
                # 更新永续合约数据
                await self.update_perpetual_data()
                
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
        
    async def _handle_funding(self, symbol: str, data: Dict):
        """处理资金费率数据"""
        try:
            funding_data = {
                'timestamp': pd.Timestamp.now(),
                'funding_rate': float(data['funding_rate']),
                'mark_price': float(data['mark_price']),
                'index_price': float(data['index_price']),
                'next_funding_time': pd.Timestamp(data['next_funding_time']),
                'exchange': data.get('exchange', 'unknown')
            }
            
            if 'open_interest' in data:
                funding_data['open_interest'] = float(data['open_interest'])
            
            # 更新缓存
            self.perpetual_cache[symbol] = funding_data
            
            # 保存到数据库
            with sqlite3.connect(self.db_path) as conn:
                conn.execute("""
                    INSERT INTO perpetual_data (
                        timestamp, symbol, funding_rate, mark_price,
                        index_price, open_interest, next_funding_time, exchange
                    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                """, (
                    funding_data['timestamp'],
                    symbol,
                    funding_data['funding_rate'],
                    funding_data['mark_price'],
                    funding_data['index_price'],
                    funding_data.get('open_interest', 0),
                    funding_data['next_funding_time'],
                    funding_data['exchange']
                ))
                conn.commit()
            
            # 通知订阅者
            await self._notify_subscribers('funding', symbol, funding_data)
            
        except (ValueError, KeyError) as e:
            logger.error(f"Error processing funding data for {symbol}: {str(e)}")
        except Exception as e:
            logger.error(f"Unexpected error processing funding data for {symbol}: {str(e)}")
            raise
    
    async def _notify_subscribers(self, event_type: str, symbol: str, data: Dict):
        """通知订阅者"""
        for callback in self.subscribers[event_type]:
            try:
                await callback(symbol, data)
            except Exception as e:
                logger.error(f"Error notifying subscriber: {str(e)}")
    
    def subscribe(self, event_type: str, callback: Callable):
        """订阅事件"""
        if event_type in self.subscribers:
            self.subscribers[event_type].append(callback)
    
    def unsubscribe(self, event_type: str, callback: Callable):
        """取消订阅"""
        if event_type in self.subscribers:
            self.subscribers[event_type].remove(callback)
    
    def get_ohlcv(self, symbol: str, timeframe: str,
                  start_time: Optional[datetime] = None,
                  end_time: Optional[datetime] = None) -> pd.DataFrame:
        """获取K线数据"""
        logger.info(f"Fetching OHLCV data for {symbol} on {timeframe}")
        if start_time is None:
            # 返回缓存数据
            data = list(self.ohlcv_cache[symbol][timeframe])
            df = pd.DataFrame(data)
            logger.info(f"Retrieved {len(df)} OHLCV records from cache for {symbol}")
            return df
        
        # 从数据库查询
        with sqlite3.connect(self.db_path) as conn:
            query = """
                SELECT * FROM ohlcv
                WHERE symbol = ? AND timeframe = ?
                AND timestamp BETWEEN ? AND ?
                ORDER BY timestamp
            """
            params = [
                symbol,
                timeframe,
                start_time.isoformat() if start_time else None,
                end_time.isoformat() if end_time else None
            ]
            df = pd.read_sql_query(query, conn, params=params)
            logger.info(f"Retrieved {len(df)} OHLCV records from database for {symbol}")
            return df
    
    def get_orderbook(self, symbol: str) -> Dict:
        """获取订单簿"""
        return self.orderbook_cache[symbol]
    
    def get_trades(self, symbol: str,
                  start_time: Optional[datetime] = None,
                  end_time: Optional[datetime] = None) -> pd.DataFrame:
        """获取交易数据"""
        logger.info(f"Fetching trade data for {symbol}")
        if start_time is None:
            # 返回缓存数据
            data = list(self.trades_cache[symbol])
            df = pd.DataFrame(data)
            logger.info(f"Retrieved {len(df)} trade records from cache for {symbol}")
            return df
        
        # 从数据库查询
        with sqlite3.connect(self.db_path) as conn:
            query = """
                SELECT * FROM trades
                WHERE symbol = ?
                AND timestamp BETWEEN ? AND ?
                ORDER BY timestamp
            """
            params = [
                symbol,
                start_time.isoformat() if start_time else None,
                end_time.isoformat() if end_time else None
            ]
            df = pd.read_sql_query(query, conn, params=params)
            logger.info(f"Retrieved {len(df)} trade records from database for {symbol}")
            return df
    
    def get_latest_price(self, symbol: str) -> Optional[float]:
        """获取最新价格"""
        if symbol in self.trades_cache and self.trades_cache[symbol]:
            price = self.trades_cache[symbol][-1]['price']
            logger.debug(f"[MarketDataService] Latest price for {symbol}: {price}")
            return price
        logger.debug(f"[MarketDataService] No price data available for {symbol}")
        return None
    
    def get_market_depth(self, symbol: str, depth: int = 10) -> Dict:
        """获取市场深度"""
        orderbook = self.orderbook_cache[symbol]
        return {
            'timestamp': orderbook['timestamp'],
            'bids': orderbook['bids'][:depth],
            'asks': orderbook['asks'][:depth]
        }
    
    def calculate_vwap(self, symbol: str, window: int = 100) -> Optional[float]:
        """计算成交量加权平均价格"""
        trades = list(self.trades_cache[symbol])[-window:]
        if not trades:
            return None
        
        volume_price = sum(trade['price'] * trade['amount'] for trade in trades)
        volume = sum(trade['amount'] for trade in trades)
        return volume_price / volume if volume > 0 else None
        
    def get_volatility(self, symbol: str, window: int = 20) -> Optional[float]:
        """计算价格波动率"""
        if symbol not in self.trades_cache:
            logger.warning(f"No trade data available for {symbol}")
            return None
            
        prices = [trade['price'] for trade in list(self.trades_cache[symbol])[-window:]]
        if not prices:
            return None
            
        returns = pd.Series(prices).pct_change().dropna()
        return float(returns.std() * (252 ** 0.5))  # Annualized volatility
        
    def get_atr(self, symbol: str, timeframe: str = '1h', period: int = 14) -> Optional[float]:
        """计算平均真实范围"""
        try:
            df = self.get_ohlcv(symbol, timeframe)
            if df.empty:
                logger.warning(f"No OHLCV data available for {symbol}")
                return None
                
            df['high_low'] = df['high'] - df['low']
            df['high_close'] = abs(df['high'] - df['close'].shift())
            df['low_close'] = abs(df['low'] - df['close'].shift())
            df['tr'] = df[['high_low', 'high_close', 'low_close']].max(axis=1)
            atr = df['tr'].rolling(period).mean().iloc[-1]
            logger.info(f"Calculated ATR for {symbol}: {atr}")
            return float(atr)
        except Exception as e:
            logger.error(f"Error calculating ATR for {symbol}: {str(e)}")
            return None
        
    async def fetch_perpetual_data(self, symbol: str) -> Optional[Dict[str, Any]]:
        """获取永续合约数据"""
        logger.info(f"Fetching perpetual data for {symbol}")
        if not self.config.perpetual_enabled or not self.config.perpetual_symbols:
            logger.warning("Perpetual trading not enabled or no symbols configured")
            return None
            
        if symbol not in self.config.perpetual_symbols:
            logger.warning(f"Symbol {symbol} not in perpetual symbols list")
            return None
            
        try:
            from exchanges.dydx_client import DydxClient
            from exchanges.hyperliquid_client import HyperliquidClient
            
            dydx = DydxClient(self.config.api_key, self.config.api_secret)
            hyperliquid = HyperliquidClient(self.config.api_key, self.config.api_secret)
            
            # Try dYdX first
            try:
                data = await dydx.get_funding_rate(symbol)
                if data:
                    data['exchange'] = 'dydx'
                    data['open_interest'] = await dydx.get_open_interest(symbol)
                    logger.info(f"Successfully fetched dYdX data for {symbol}")
                    return data
            except Exception as e:
                logger.warning(f"Failed to fetch dYdX data for {symbol}: {str(e)}")
            
            # Try Hyperliquid as fallback
            try:
                data = await hyperliquid.get_funding_rate(symbol)
                if data:
                    data['exchange'] = 'hyperliquid'
                    data['open_interest'] = await hyperliquid.get_open_interest(symbol)
                    logger.info(f"Successfully fetched Hyperliquid data for {symbol}")
                    return data
            except Exception as e:
                logger.error(f"Failed to fetch Hyperliquid data for {symbol}: {str(e)}")
                return None
                
        except Exception as e:
            logger.error(f"Error fetching perpetual data for {symbol}: {str(e)}")
            return None
            
    def save_perpetual_data(self, symbol: str, data: Dict) -> None:
        """保存永续合约数据"""
        if not data:
            return
            
        self.perpetual_cache[symbol] = data
        
        with sqlite3.connect(self.db_path) as conn:
            conn.execute("""
                INSERT INTO perpetual_data (
                    timestamp, symbol, funding_rate, mark_price,
                    index_price, open_interest, next_funding_time, exchange
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
            """, (
                datetime.now(),
                symbol,
                data['funding_rate'],
                data['mark_price'],
                data['index_price'],
                data['open_interest'],
                data['next_funding_time'],
                data['exchange']
            ))
            conn.commit()
            
    async def update_perpetual_data(self) -> None:
        """更新永续合约数据"""
        if not self.config.perpetual_enabled:
            logger.debug("Perpetual trading not enabled")
            return
            
        if not self.config.perpetual_symbols:
            logger.debug("No perpetual symbols configured")
            return
            
        for symbol in self.config.perpetual_symbols:
            try:
                data = await self.fetch_perpetual_data(symbol)
                if data:
                    self.save_perpetual_data(symbol, data)
                    await self._notify_subscribers('perpetual', symbol, data)
                else:
                    logger.warning(f"No perpetual data available for {symbol}")
            except Exception as e:
                logger.error(f"Failed to update perpetual data for {symbol}: {str(e)}")
                continue
                
    def get_funding_rate(self, symbol: str) -> Optional[float]:
        """获取当前资金费率"""
        if symbol in self.perpetual_cache:
            return self.perpetual_cache[symbol]['funding_rate']
        return None
        
    def get_mark_price(self, symbol: str) -> Optional[float]:
        """获取标记价格"""
        if symbol in self.perpetual_cache:
            return self.perpetual_cache[symbol]['mark_price']
        return None                           