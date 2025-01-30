import json
import logging
import numpy as np
import pandas as pd
from dataclasses import dataclass
from typing import Dict, List, Optional, Tuple, Union
from datetime import datetime, timedelta

from technical_analysis import TechnicalAnalysis
from market_analyzer import MarketAnalyzer

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@dataclass
class StrategyConfig:
    name: str
    description: str
    symbols: List[str]
    timeframe: str
    indicators: Dict[str, Dict]
    entry_conditions: List[Dict]
    exit_conditions: List[Dict]
    position_size: float
    stop_loss: float
    take_profit: float
    max_positions: int
    risk_per_trade: float

@dataclass
class Position:
    symbol: str
    side: str
    entry_price: float
    amount: float
    stop_loss: float
    take_profit: float
    entry_time: datetime
    pnl: float = 0.0
    status: str = "open"

class StrategyEngine:
    def __init__(self, config: StrategyConfig):
        self.config = config
        self.positions: Dict[str, Position] = {}
        self.ta = TechnicalAnalysis()
        self.analyzer = MarketAnalyzer()
        self.historical_data: Dict[str, pd.DataFrame] = {}
        self.indicators: Dict[str, Dict] = {}
        
    def initialize(self):
        """Initialize strategy with historical data and indicators"""
        for symbol in self.config.symbols:
            # Load historical data
            self.historical_data[symbol] = self._load_historical_data(symbol)
            
            # Calculate indicators
            self.indicators[symbol] = {}
            for ind_name, ind_config in self.config.indicators.items():
                self.indicators[symbol][ind_name] = self._calculate_indicator(
                    symbol, ind_name, ind_config
                )
    
    def update(self, market_data: Dict):
        """Update strategy with new market data"""
        symbol = market_data["symbol"]
        price = market_data["price"]
        timestamp = market_data["timestamp"]
        
        # Update historical data
        self._update_historical_data(symbol, price, timestamp)
        
        # Update indicators
        self._update_indicators(symbol)
        
        # Check positions
        self._check_positions(symbol, price)
        
        # Generate signals
        return self._generate_signals(symbol, price)
    
    def _load_historical_data(self, symbol: str) -> pd.DataFrame:
        """Load historical data for a symbol"""
        try:
            # TODO: Implement data loading from database
            data = pd.DataFrame()  # Placeholder
            logger.info(f"Loaded historical data for {symbol}")
            return data
        except Exception as e:
            logger.error(f"Failed to load historical data for {symbol}: {e}")
            raise
    
    def _update_historical_data(self, symbol: str, price: float, timestamp: datetime):
        """Update historical data with new price"""
        new_data = pd.DataFrame({
            "timestamp": [timestamp],
            "price": [price]
        })
        self.historical_data[symbol] = pd.concat([
            self.historical_data[symbol],
            new_data
        ]).tail(1000)  # Keep last 1000 points
    
    def _calculate_indicator(self, symbol: str, name: str, config: Dict) -> pd.Series:
        """Calculate technical indicator"""
        data = self.historical_data[symbol]
        return getattr(self.ta, name)(data, **config)
    
    def _update_indicators(self, symbol: str):
        """Update all indicators for a symbol"""
        for ind_name, ind_config in self.config.indicators.items():
            self.indicators[symbol][ind_name] = self._calculate_indicator(
                symbol, ind_name, ind_config
            )
    
    def _check_positions(self, symbol: str, current_price: float):
        """Check and update existing positions"""
        if symbol in self.positions:
            position = self.positions[symbol]
            
            # Calculate PnL
            if position.side == "long":
                position.pnl = (current_price - position.entry_price) * position.amount
            else:
                position.pnl = (position.entry_price - current_price) * position.amount
            
            # Check stop loss
            if position.side == "long" and current_price <= position.stop_loss:
                self._close_position(symbol, current_price, "stop_loss")
            elif position.side == "short" and current_price >= position.stop_loss:
                self._close_position(symbol, current_price, "stop_loss")
            
            # Check take profit
            if position.side == "long" and current_price >= position.take_profit:
                self._close_position(symbol, current_price, "take_profit")
            elif position.side == "short" and current_price <= position.take_profit:
                self._close_position(symbol, current_price, "take_profit")
    
    def _close_position(self, symbol: str, price: float, reason: str):
        """Close a position"""
        position = self.positions[symbol]
        position.status = "closed"
        
        # Calculate final PnL
        if position.side == "long":
            pnl = (price - position.entry_price) * position.amount
        else:
            pnl = (position.entry_price - price) * position.amount
        
        logger.info(f"Closed position for {symbol}: {reason}, PnL: {pnl}")
        del self.positions[symbol]
    
    def _generate_signals(self, symbol: str, current_price: float) -> Optional[Dict]:
        """Generate trading signals based on strategy conditions"""
        # Skip if already in position
        if symbol in self.positions:
            return None
        
        # Check entry conditions
        if self._check_conditions(symbol, self.config.entry_conditions):
            # Calculate position size
            account_size = 10000  # TODO: Get from config
            risk_amount = account_size * self.config.risk_per_trade
            position_size = self._calculate_position_size(
                current_price, risk_amount, self.config.stop_loss
            )
            
            # Generate entry signal
            signal = {
                "type": "entry",
                "symbol": symbol,
                "side": "long",  # TODO: Support short positions
                "price": current_price,
                "amount": position_size,
                "stop_loss": current_price * (1 - self.config.stop_loss),
                "take_profit": current_price * (1 + self.config.take_profit)
            }
            
            logger.info(f"Generated entry signal for {symbol}: {signal}")
            return signal
        
        return None
    
    def _check_conditions(self, symbol: str, conditions: List[Dict]) -> bool:
        """Check if conditions are met"""
        for condition in conditions:
            indicator = self.indicators[symbol][condition["indicator"]]
            current_value = indicator.iloc[-1]
            
            if condition["type"] == "above":
                if current_value <= condition["value"]:
                    return False
            elif condition["type"] == "below":
                if current_value >= condition["value"]:
                    return False
            elif condition["type"] == "cross_above":
                if not (indicator.iloc[-2] <= condition["value"] and 
                       current_value > condition["value"]):
                    return False
            elif condition["type"] == "cross_below":
                if not (indicator.iloc[-2] >= condition["value"] and 
                       current_value < condition["value"]):
                    return False
        
        return True
    
    def _calculate_position_size(self, price: float, risk_amount: float,
                               stop_loss_pct: float) -> float:
        """Calculate position size based on risk parameters"""
        stop_loss_points = price * stop_loss_pct
        position_size = risk_amount / stop_loss_points
        return position_size
    
    def get_strategy_state(self) -> Dict:
        """Get current strategy state"""
        return {
            "name": self.config.name,
            "positions": {
                symbol: {
                    "side": pos.side,
                    "entry_price": pos.entry_price,
                    "amount": pos.amount,
                    "pnl": pos.pnl,
                    "entry_time": pos.entry_time.isoformat()
                }
                for symbol, pos in self.positions.items()
            },
            "indicators": {
                symbol: {
                    name: values.iloc[-1]
                    for name, values in symbol_indicators.items()
                }
                for symbol, symbol_indicators in self.indicators.items()
            }
        } 