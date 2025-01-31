import aiohttp
from datetime import datetime
from loguru import logger
from typing import Dict, Any

async def check_database_health() -> Dict[str, Any]:
    """Check if database connections are healthy and return detailed status"""
    status = {
        "healthy": True,
        "databases": {},
        "error": None
    }
    
    try:
        import sqlite3
        import os
        
        required_dbs = ['market_data.db', 'trading_data.db', 'agent_system.db']
        for db_name in required_dbs:
            db_status = {
                "exists": os.path.exists(db_name),
                "connection": False,
                "error": None
            }
            
            if not db_status["exists"]:
                logger.warning(f"Database {db_name} does not exist")
                status["healthy"] = False
            else:
                try:
                    conn = sqlite3.connect(db_name)
                    cursor = conn.cursor()
                    cursor.execute("SELECT 1")
                    cursor.fetchone()
                    conn.close()
                    db_status["connection"] = True
                except sqlite3.Error as e:
                    db_status["error"] = str(e)
                    db_status["connection"] = False
                    status["healthy"] = False
                    logger.error(f"Failed to connect to {db_name}: {e}")
            
            status["databases"][db_name] = db_status
            
        return status
    except Exception as e:
        error_msg = f"Database health check failed: {str(e)}"
        logger.error(error_msg)
        status["healthy"] = False
        status["error"] = error_msg
        return status

async def check_dydx_api_health() -> Dict[str, Any]:
    """Check dYdX API health status"""
    status = {
        "healthy": False,
        "latency_ms": 0,
        "error": None
    }
    try:
        start_time = datetime.utcnow()
        async with aiohttp.ClientSession() as session:
            async with session.get("https://api.dydx.exchange/v3/markets") as response:
                status["latency_ms"] = (datetime.utcnow() - start_time).total_seconds() * 1000
                status["healthy"] = response.status == 200
                if not status["healthy"]:
                    status["error"] = f"API returned status {response.status}"
    except Exception as e:
        status["error"] = str(e)
        logger.error(f"dYdX API health check failed: {e}")
    return status

async def check_hyperliquid_api_health() -> Dict[str, Any]:
    """Check Hyperliquid API health status"""
    status = {
        "healthy": False,
        "latency_ms": 0,
        "error": None
    }
    try:
        start_time = datetime.utcnow()
        async with aiohttp.ClientSession() as session:
            async with session.post(
                "https://api.hyperliquid-testnet.xyz/info",
                json={"type": "meta"}
            ) as response:
                status["latency_ms"] = (datetime.utcnow() - start_time).total_seconds() * 1000
                status["healthy"] = response.status == 200
                if not status["healthy"]:
                    status["error"] = f"API returned status {response.status}"
    except Exception as e:
        status["error"] = str(e)
        logger.error(f"Hyperliquid API health check failed: {e}")
    return status

async def check_market_data_health() -> Dict[str, Any]:
    """Check if market data service is healthy"""
    status = {
        "healthy": True,
        "services": {},
        "error": None
    }
    
    try:
        dydx_status = await check_dydx_api_health()
        hyperliquid_status = await check_hyperliquid_api_health()
        
        status["services"]["dydx"] = dydx_status
        status["services"]["hyperliquid"] = hyperliquid_status
        
        # System is healthy if at least one exchange API is working
        status["healthy"] = dydx_status["healthy"] or hyperliquid_status["healthy"]
        
        if not status["healthy"]:
            status["error"] = "All exchange APIs are unavailable"
            
        return status
    except Exception as e:
        status["healthy"] = False
        status["error"] = str(e)
        logger.error(f"Market data health check failed: {e}")
        return status
