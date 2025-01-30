import logging
import pandas as pd
import numpy as np
from typing import Dict, List, Optional
from datetime import datetime, timedelta
from dataclasses import dataclass
import json
import sqlite3
from pathlib import Path

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@dataclass
class ExecutionReport:
    """执行报告"""
    timestamp: datetime
    symbol: str
    action: str  # 'open', 'close', 'modify'
    direction: str  # 'buy', 'sell'
    price: float
    size: float
    agent_name: str
    confidence: float
    reason: str
    metadata: Dict

@dataclass
class PerformanceReport:
    """绩效报告"""
    timestamp: datetime
    total_pnl: float
    daily_pnl: float
    total_trades: int
    winning_trades: int
    losing_trades: int
    win_rate: float
    avg_profit: float
    avg_loss: float
    max_drawdown: float
    sharpe_ratio: float
    agent_metrics: Dict
    market_metrics: Dict

class ReportingSystem:
    """报告系统"""
    
    def __init__(self, db_path: str = "trading_data.db"):
        self.db_path = db_path
        self.reports_path = Path("reports")
        self.reports_path.mkdir(exist_ok=True)
        
        # 初始化数据库
        self._initialize_database()
    
    def _initialize_database(self):
        """初始化数据库表"""
        with sqlite3.connect(self.db_path) as conn:
            # 执行报告表
            conn.execute("""
                CREATE TABLE IF NOT EXISTS execution_reports (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    timestamp DATETIME,
                    symbol TEXT,
                    action TEXT,
                    direction TEXT,
                    price REAL,
                    size REAL,
                    agent_name TEXT,
                    confidence REAL,
                    reason TEXT,
                    metadata TEXT
                )
            """)
            
            # 绩效报告表
            conn.execute("""
                CREATE TABLE IF NOT EXISTS performance_reports (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    timestamp DATETIME,
                    total_pnl REAL,
                    daily_pnl REAL,
                    total_trades INTEGER,
                    winning_trades INTEGER,
                    losing_trades INTEGER,
                    win_rate REAL,
                    avg_profit REAL,
                    avg_loss REAL,
                    max_drawdown REAL,
                    sharpe_ratio REAL,
                    agent_metrics TEXT,
                    market_metrics TEXT
                )
            """)
            
            # 市场数据表
            conn.execute("""
                CREATE TABLE IF NOT EXISTS market_data (
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
            
            # 交易记录表
            conn.execute("""
                CREATE TABLE IF NOT EXISTS trades (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    symbol TEXT,
                    direction TEXT,
                    open_time DATETIME,
                    close_time DATETIME,
                    entry_price REAL,
                    exit_price REAL,
                    size REAL,
                    pnl REAL,
                    agent_name TEXT,
                    metadata TEXT
                )
            """)
            
            conn.commit()
    
    async def save_execution_report(self, report: ExecutionReport):
        """保存执行报告"""
        try:
            with sqlite3.connect(self.db_path) as conn:
                conn.execute("""
                    INSERT INTO execution_reports (
                        timestamp, symbol, action, direction, price,
                        size, agent_name, confidence, reason, metadata
                    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                """, (
                    report.timestamp, report.symbol, report.action,
                    report.direction, report.price, report.size,
                    report.agent_name, report.confidence, report.reason,
                    json.dumps(report.metadata)
                ))
                conn.commit()
            
            logger.info(f"Saved execution report for {report.symbol}")
            
        except Exception as e:
            logger.error(f"Error saving execution report: {str(e)}")
    
    async def save_performance_report(self, report: PerformanceReport):
        """保存绩效报告"""
        try:
            with sqlite3.connect(self.db_path) as conn:
                conn.execute("""
                    INSERT INTO performance_reports (
                        timestamp, total_pnl, daily_pnl, total_trades,
                        winning_trades, losing_trades, win_rate,
                        avg_profit, avg_loss, max_drawdown,
                        sharpe_ratio, agent_metrics, market_metrics
                    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                """, (
                    report.timestamp, report.total_pnl, report.daily_pnl,
                    report.total_trades, report.winning_trades,
                    report.losing_trades, report.win_rate,
                    report.avg_profit, report.avg_loss,
                    report.max_drawdown, report.sharpe_ratio,
                    json.dumps(report.agent_metrics),
                    json.dumps(report.market_metrics)
                ))
                conn.commit()
            
            logger.info("Saved performance report")
            
        except Exception as e:
            logger.error(f"Error saving performance report: {str(e)}")
    
    async def save_market_data(self, symbol: str, timeframe: str, data: pd.DataFrame):
        """保存市场数据"""
        try:
            with sqlite3.connect(self.db_path) as conn:
                for idx, row in data.iterrows():
                    conn.execute("""
                        INSERT INTO market_data (
                            timestamp, symbol, timeframe,
                            open, high, low, close, volume
                        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                    """, (
                        idx, symbol, timeframe,
                        row['open'], row['high'], row['low'],
                        row['close'], row['volume']
                    ))
                conn.commit()
            
            logger.info(f"Saved market data for {symbol} {timeframe}")
            
        except Exception as e:
            logger.error(f"Error saving market data: {str(e)}")
    
    async def save_trade(self, trade: Dict):
        """保存交易记录"""
        try:
            with sqlite3.connect(self.db_path) as conn:
                conn.execute("""
                    INSERT INTO trades (
                        symbol, direction, open_time, close_time,
                        entry_price, exit_price, size, pnl,
                        agent_name, metadata
                    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                """, (
                    trade['symbol'], trade['direction'],
                    trade['open_time'], trade['close_time'],
                    trade['entry_price'], trade['exit_price'],
                    trade['size'], trade['pnl'],
                    trade['agent_name'], json.dumps(trade['metadata'])
                ))
                conn.commit()
            
            logger.info(f"Saved trade record for {trade['symbol']}")
            
        except Exception as e:
            logger.error(f"Error saving trade: {str(e)}")
    
    def generate_daily_report(self, date: datetime) -> Dict:
        """生成每日报告"""
        try:
            with sqlite3.connect(self.db_path) as conn:
                # 获取当日交易
                trades_df = pd.read_sql_query("""
                    SELECT * FROM trades
                    WHERE date(close_time) = date(?)
                """, conn, params=(date,))
                
                # 获取绩效报告
                perf_df = pd.read_sql_query("""
                    SELECT * FROM performance_reports
                    WHERE date(timestamp) = date(?)
                """, conn, params=(date,))
                
                # 生成报告
                report = {
                    'date': date.strftime('%Y-%m-%d'),
                    'trading_summary': {
                        'total_trades': len(trades_df),
                        'winning_trades': len(trades_df[trades_df['pnl'] > 0]),
                        'total_pnl': trades_df['pnl'].sum(),
                        'max_profit': trades_df['pnl'].max(),
                        'max_loss': trades_df['pnl'].min(),
                        'avg_trade_pnl': trades_df['pnl'].mean()
                    },
                    'performance_metrics': {
                        'sharpe_ratio': perf_df['sharpe_ratio'].iloc[-1],
                        'max_drawdown': perf_df['max_drawdown'].iloc[-1],
                        'win_rate': perf_df['win_rate'].iloc[-1]
                    },
                    'trades': trades_df.to_dict('records')
                }
                
                # 保存报告
                report_path = self.reports_path / f"daily_report_{date.strftime('%Y%m%d')}.json"
                with open(report_path, 'w') as f:
                    json.dump(report, f, indent=2)
                
                return report
                
        except Exception as e:
            logger.error(f"Error generating daily report: {str(e)}")
            return {}
    
    def get_historical_performance(self, 
                                 start_date: datetime,
                                 end_date: datetime) -> pd.DataFrame:
        """获取历史绩效数据"""
        try:
            with sqlite3.connect(self.db_path) as conn:
                query = """
                    SELECT * FROM performance_reports
                    WHERE timestamp BETWEEN ? AND ?
                    ORDER BY timestamp
                """
                return pd.read_sql_query(query, conn, 
                                       params=(start_date, end_date))
                
        except Exception as e:
            logger.error(f"Error getting historical performance: {str(e)}")
            return pd.DataFrame()
    
    def get_agent_performance(self, agent_name: str) -> Dict:
        """获取Agent绩效数据"""
        try:
            with sqlite3.connect(self.db_path) as conn:
                trades_df = pd.read_sql_query("""
                    SELECT * FROM trades
                    WHERE agent_name = ?
                """, conn, params=(agent_name,))
                
                return {
                    'total_trades': len(trades_df),
                    'winning_trades': len(trades_df[trades_df['pnl'] > 0]),
                    'total_pnl': trades_df['pnl'].sum(),
                    'avg_profit': trades_df[trades_df['pnl'] > 0]['pnl'].mean(),
                    'avg_loss': trades_df[trades_df['pnl'] < 0]['pnl'].mean(),
                    'win_rate': len(trades_df[trades_df['pnl'] > 0]) / len(trades_df),
                    'profit_factor': abs(trades_df[trades_df['pnl'] > 0]['pnl'].sum() / 
                                      trades_df[trades_df['pnl'] < 0]['pnl'].sum())
                }
                
        except Exception as e:
            logger.error(f"Error getting agent performance: {str(e)}")
            return {} 