import os
from pathlib import Path
from typing import Dict, Any

class Config:
    """系统配置"""
    
    # 基础配置
    BASE_DIR = Path(__file__).parent
    DB_DIR = BASE_DIR / "database"
    LOG_DIR = BASE_DIR / "logs"
    
    # 创建必要的目录
    DB_DIR.mkdir(exist_ok=True)
    LOG_DIR.mkdir(exist_ok=True)
    
    # 交易所配置
    EXCHANGE_CONFIG = {
        "name": "binance",
        "api_key": os.getenv("EXCHANGE_API_KEY", ""),
        "api_secret": os.getenv("EXCHANGE_API_SECRET", ""),
        "testnet": os.getenv("USE_TESTNET", "true").lower() == "true"
    }
    
    # 市场数据配置
    MARKET_DATA_CONFIG = {
        "symbols": ["BTC/USDT", "ETH/USDT"],
        "timeframes": ["1m", "5m", "15m", "30m", "1h", "4h", "1d"],
        "cache_size": 1000,
        "update_interval": 1.0,
        "db_path": str(DB_DIR / "market_data.db")
    }
    
    # 风险管理配置
    RISK_CONFIG = {
        "max_position_size": 0.1,  # 最大仓位为账户价值的10%
        "max_drawdown": 0.2,       # 最大回撤20%
        "daily_loss_limit": 0.05,  # 日亏损限制5%
        "position_limit": 10,      # 最大同时持仓10个
        "risk_per_trade": 0.01,    # 每笔交易风险1%
        "leverage_limit": 3.0,     # 最大杠杆3倍
        "correlation_limit": 0.7,   # 相关性限制0.7
        "min_diversification": 3,   # 最小分散化数量3个
        "stop_loss_atr": 2.0,      # 2倍ATR止损
        "take_profit_atr": 3.0,    # 3倍ATR止盈
        "db_path": str(DB_DIR / "risk_management.db"),
        "max_leverage": 3.0,       # 最大允许杠杆
        "max_position_value": 100000.0,  # 单个仓位最大价值
        "min_maintenance_margin": 0.005,  # 最小维持保证金率
        "funding_rate_interval": 8,       # 资金费率收取间隔（小时）
        "liquidation_threshold": 0.05,    # 强平阈值
        "margin_call_threshold": 0.1      # 追加保证金阈值
    }
    
    # 合约交易配置
    CONTRACT_CONFIG = {
        "enabled_pairs": ["BTC/USDT", "ETH/USDT"],  # 支持合约交易的交易对
        "leverage_options": [1, 2, 3],               # 可选杠杆倍数
        "margin_types": ["isolated", "cross"],       # 保证金模式
        "min_contract_size": {                       # 最小合约数量
            "BTC/USDT": 0.001,
            "ETH/USDT": 0.01
        },
        "price_precision": {                         # 价格精度
            "BTC/USDT": 2,
            "ETH/USDT": 2
        },
        "size_precision": {                          # 数量精度
            "BTC/USDT": 3,
            "ETH/USDT": 3
        },
        "funding_rate_limit": 0.0075,               # 最大资金费率限制
        "max_slippage": 0.001                       # 最大滑点限制
    }
    
    # 数据库配置
    DB_CONFIG = {
        "market_data": str(DB_DIR / "market_data.db"),
        "risk_management": str(DB_DIR / "risk_management.db"),
        "agent_system": str(DB_DIR / "agent_system.db")
    }
    
    # 日志配置
    LOG_CONFIG = {
        "level": "INFO",
        "format": "%(asctime)s - %(name)s - %(levelname)s - %(message)s",
        "file": str(LOG_DIR / "trading_system.log")
    }
    
    @classmethod
    def get_exchange_config(cls) -> Dict[str, Any]:
        """获取交易所配置"""
        return cls.EXCHANGE_CONFIG.copy()
    
    @classmethod
    def get_market_data_config(cls) -> Dict[str, Any]:
        """获取市场数据配置"""
        return cls.MARKET_DATA_CONFIG.copy()
    
    @classmethod
    def get_risk_config(cls) -> Dict[str, Any]:
        """获取风险管理配置"""
        return cls.RISK_CONFIG.copy()
    
    @classmethod
    def get_contract_config(cls) -> Dict[str, Any]:
        """获取合约交易配置"""
        return cls.CONTRACT_CONFIG.copy()
    
    @classmethod
    def get_db_config(cls) -> Dict[str, str]:
        """获取数据库配置"""
        return cls.DB_CONFIG.copy()
    
    @classmethod
    def get_log_config(cls) -> Dict[str, Any]:
        """获取日志配置"""
        return cls.LOG_CONFIG.copy() 