import logging
import pandas as pd
import numpy as np
from typing import Dict, List, Optional
from datetime import datetime, timedelta
from dataclasses import dataclass
import json
import sqlite3
from pathlib import Path
import psutil
import time
import asyncio
from collections import deque

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@dataclass
class SystemMetrics:
    """系统指标"""
    cpu_usage: float
    memory_usage: float
    disk_usage: float
    network_io: Dict[str, float]
    process_time: float

@dataclass
class TradingMetrics:
    """交易指标"""
    execution_latency: float
    signal_processing_time: float
    order_success_rate: float
    slippage: float
    fill_ratio: float

@dataclass
class AgentMetrics:
    """Agent指标"""
    signal_count: int
    signal_quality: float
    response_time: float
    cpu_usage: float
    memory_usage: float

class PerformanceMonitor:
    """性能监控系统"""
    
    def __init__(self, db_path: str = "performance_metrics.db"):
        self.db_path = db_path
        self.metrics_cache_size = 1000
        self.system_metrics: deque = deque(maxlen=self.metrics_cache_size)
        self.trading_metrics: deque = deque(maxlen=self.metrics_cache_size)
        self.agent_metrics: Dict[str, deque] = {}
        
        # 初始化数据库
        self._initialize_database()
        
        # 启动监控
        self.is_running = False
        self.monitor_interval = 1  # 1秒
    
    def _initialize_database(self):
        """初始化数据库"""
        with sqlite3.connect(self.db_path) as conn:
            # 系统指标表
            conn.execute("""
                CREATE TABLE IF NOT EXISTS system_metrics (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    timestamp DATETIME,
                    cpu_usage REAL,
                    memory_usage REAL,
                    disk_usage REAL,
                    network_io TEXT,
                    process_time REAL
                )
            """)
            
            # 交易指标表
            conn.execute("""
                CREATE TABLE IF NOT EXISTS trading_metrics (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    timestamp DATETIME,
                    execution_latency REAL,
                    signal_processing_time REAL,
                    order_success_rate REAL,
                    slippage REAL,
                    fill_ratio REAL
                )
            """)
            
            # Agent指标表
            conn.execute("""
                CREATE TABLE IF NOT EXISTS agent_metrics (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    timestamp DATETIME,
                    agent_name TEXT,
                    signal_count INTEGER,
                    signal_quality REAL,
                    response_time REAL,
                    cpu_usage REAL,
                    memory_usage REAL
                )
            """)
            
            conn.commit()
    
    async def start_monitoring(self):
        """启动监控"""
        self.is_running = True
        while self.is_running:
            try:
                # 收集系统指标
                system_metrics = self._collect_system_metrics()
                self.system_metrics.append(system_metrics)
                await self._save_system_metrics(system_metrics)
                
                # 等待下一个监控周期
                await asyncio.sleep(self.monitor_interval)
                
            except Exception as e:
                logger.error(f"Error in monitoring: {str(e)}")
                await asyncio.sleep(self.monitor_interval)
    
    def stop_monitoring(self):
        """停止监控"""
        self.is_running = False
    
    def _collect_system_metrics(self) -> SystemMetrics:
        """收集系统指标"""
        # CPU使用率
        cpu_usage = psutil.cpu_percent(interval=None)
        
        # 内存使用率
        memory = psutil.virtual_memory()
        memory_usage = memory.percent
        
        # 磁盘使用率
        disk = psutil.disk_usage('/')
        disk_usage = disk.percent
        
        # 网络IO
        network = psutil.net_io_counters()
        network_io = {
            'bytes_sent': network.bytes_sent,
            'bytes_recv': network.bytes_recv
        }
        
        # 进程时间
        process = psutil.Process()
        process_time = sum(process.cpu_times()[:2])
        
        return SystemMetrics(
            cpu_usage=cpu_usage,
            memory_usage=memory_usage,
            disk_usage=disk_usage,
            network_io=network_io,
            process_time=process_time
        )
    
    async def record_trading_metrics(self, metrics: TradingMetrics):
        """记录交易指标"""
        self.trading_metrics.append(metrics)
        await self._save_trading_metrics(metrics)
    
    async def record_agent_metrics(self, agent_name: str, metrics: AgentMetrics):
        """记录Agent指标"""
        if agent_name not in self.agent_metrics:
            self.agent_metrics[agent_name] = deque(maxlen=self.metrics_cache_size)
        
        self.agent_metrics[agent_name].append(metrics)
        await self._save_agent_metrics(agent_name, metrics)
    
    async def _save_system_metrics(self, metrics: SystemMetrics):
        """保存系统指标"""
        try:
            with sqlite3.connect(self.db_path) as conn:
                conn.execute("""
                    INSERT INTO system_metrics (
                        timestamp, cpu_usage, memory_usage,
                        disk_usage, network_io, process_time
                    ) VALUES (?, ?, ?, ?, ?, ?)
                """, (
                    datetime.now(),
                    metrics.cpu_usage,
                    metrics.memory_usage,
                    metrics.disk_usage,
                    json.dumps(metrics.network_io),
                    metrics.process_time
                ))
                conn.commit()
        except Exception as e:
            logger.error(f"Error saving system metrics: {str(e)}")
    
    async def _save_trading_metrics(self, metrics: TradingMetrics):
        """保存交易指标"""
        try:
            with sqlite3.connect(self.db_path) as conn:
                conn.execute("""
                    INSERT INTO trading_metrics (
                        timestamp, execution_latency,
                        signal_processing_time, order_success_rate,
                        slippage, fill_ratio
                    ) VALUES (?, ?, ?, ?, ?, ?)
                """, (
                    datetime.now(),
                    metrics.execution_latency,
                    metrics.signal_processing_time,
                    metrics.order_success_rate,
                    metrics.slippage,
                    metrics.fill_ratio
                ))
                conn.commit()
        except Exception as e:
            logger.error(f"Error saving trading metrics: {str(e)}")
    
    async def _save_agent_metrics(self, agent_name: str, metrics: AgentMetrics):
        """保存Agent指标"""
        try:
            with sqlite3.connect(self.db_path) as conn:
                conn.execute("""
                    INSERT INTO agent_metrics (
                        timestamp, agent_name, signal_count,
                        signal_quality, response_time,
                        cpu_usage, memory_usage
                    ) VALUES (?, ?, ?, ?, ?, ?, ?)
                """, (
                    datetime.now(),
                    agent_name,
                    metrics.signal_count,
                    metrics.signal_quality,
                    metrics.response_time,
                    metrics.cpu_usage,
                    metrics.memory_usage
                ))
                conn.commit()
        except Exception as e:
            logger.error(f"Error saving agent metrics: {str(e)}")
    
    def get_system_metrics(self, 
                          start_time: datetime,
                          end_time: datetime) -> pd.DataFrame:
        """获取系统指标"""
        try:
            with sqlite3.connect(self.db_path) as conn:
                query = """
                    SELECT * FROM system_metrics
                    WHERE timestamp BETWEEN ? AND ?
                    ORDER BY timestamp
                """
                return pd.read_sql_query(query, conn,
                                       params=(start_time, end_time))
        except Exception as e:
            logger.error(f"Error getting system metrics: {str(e)}")
            return pd.DataFrame()
    
    def get_trading_metrics(self,
                           start_time: datetime,
                           end_time: datetime) -> pd.DataFrame:
        """获取交易指标"""
        try:
            with sqlite3.connect(self.db_path) as conn:
                query = """
                    SELECT * FROM trading_metrics
                    WHERE timestamp BETWEEN ? AND ?
                    ORDER BY timestamp
                """
                return pd.read_sql_query(query, conn,
                                       params=(start_time, end_time))
        except Exception as e:
            logger.error(f"Error getting trading metrics: {str(e)}")
            return pd.DataFrame()
    
    def get_agent_metrics(self,
                         agent_name: str,
                         start_time: datetime,
                         end_time: datetime) -> pd.DataFrame:
        """获取Agent指标"""
        try:
            with sqlite3.connect(self.db_path) as conn:
                query = """
                    SELECT * FROM agent_metrics
                    WHERE agent_name = ?
                    AND timestamp BETWEEN ? AND ?
                    ORDER BY timestamp
                """
                return pd.read_sql_query(query, conn,
                                       params=(agent_name, start_time, end_time))
        except Exception as e:
            logger.error(f"Error getting agent metrics: {str(e)}")
            return pd.DataFrame()
    
    def analyze_performance(self,
                          start_time: datetime,
                          end_time: datetime) -> Dict:
        """分析性能"""
        system_metrics = self.get_system_metrics(start_time, end_time)
        trading_metrics = self.get_trading_metrics(start_time, end_time)
        
        if system_metrics.empty or trading_metrics.empty:
            return {}
        
        analysis = {
            'system_performance': {
                'avg_cpu_usage': system_metrics['cpu_usage'].mean(),
                'max_cpu_usage': system_metrics['cpu_usage'].max(),
                'avg_memory_usage': system_metrics['memory_usage'].mean(),
                'max_memory_usage': system_metrics['memory_usage'].max(),
                'avg_disk_usage': system_metrics['disk_usage'].mean(),
                'process_time_increase': (
                    system_metrics['process_time'].iloc[-1] -
                    system_metrics['process_time'].iloc[0]
                )
            },
            'trading_performance': {
                'avg_latency': trading_metrics['execution_latency'].mean(),
                'max_latency': trading_metrics['execution_latency'].max(),
                'avg_processing_time': trading_metrics['signal_processing_time'].mean(),
                'order_success_rate': trading_metrics['order_success_rate'].mean(),
                'avg_slippage': trading_metrics['slippage'].mean(),
                'avg_fill_ratio': trading_metrics['fill_ratio'].mean()
            }
        }
        
        # 添加性能警告
        warnings = []
        if analysis['system_performance']['max_cpu_usage'] > 80:
            warnings.append("High CPU usage detected")
        if analysis['system_performance']['max_memory_usage'] > 80:
            warnings.append("High memory usage detected")
        if analysis['trading_performance']['avg_latency'] > 1.0:
            warnings.append("High execution latency")
        if analysis['trading_performance']['order_success_rate'] < 0.95:
            warnings.append("Low order success rate")
        
        analysis['warnings'] = warnings
        
        return analysis
    
    def optimize_performance(self, analysis: Dict) -> List[str]:
        """优化建议"""
        recommendations = []
        
        # 系统性能优化建议
        if analysis['system_performance']['avg_cpu_usage'] > 70:
            recommendations.append(
                "Consider reducing the number of concurrent operations"
            )
        if analysis['system_performance']['avg_memory_usage'] > 70:
            recommendations.append(
                "Consider implementing memory cleanup or reducing cache sizes"
            )
        
        # 交易性能优化建议
        if analysis['trading_performance']['avg_latency'] > 0.5:
            recommendations.append(
                "Consider optimizing network connections or reducing data processing"
            )
        if analysis['trading_performance']['order_success_rate'] < 0.98:
            recommendations.append(
                "Review order execution strategy and error handling"
            )
        if analysis['trading_performance']['avg_slippage'] > 0.001:
            recommendations.append(
                "Consider implementing better price improvement strategies"
            )
        
        return recommendations 