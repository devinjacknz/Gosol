import pandas as pd
import numpy as np
from typing import Dict, List, Optional
from dataclasses import dataclass
from datetime import datetime
import logging
from strategy_engine import StrategyEngine, StrategyConfig, Signal

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@dataclass
class Trade:
    """Represents a trade in the backtest"""
    id: str
    symbol: str
    side: str
    entry_time: datetime
    entry_price: float
    exit_time: Optional[datetime]
    exit_price: Optional[float]
    amount: float
    pnl: float = 0
    fee: float = 0

@dataclass
class BacktestResult:
    """Represents the results of a backtest"""
    total_pnl: float
    win_rate: float
    sharpe_ratio: float
    max_drawdown: float
    trades: List[Trade]
    equity_curve: pd.Series
    drawdown_curve: pd.Series
    metrics: Dict

class BacktestEngine:
    """Backtesting engine for trading strategies"""
    
    def __init__(self, 
                 strategy_config: StrategyConfig,
                 initial_capital: float = 100000,
                 fee_rate: float = 0.001):
        self.strategy = StrategyEngine(strategy_config)
        self.initial_capital = initial_capital
        self.fee_rate = fee_rate
        self.capital = initial_capital
        self.trades: List[Trade] = []
        self.equity_curve: List[float] = []
        self.positions: Dict[str, float] = {}  # symbol -> amount
    
    def run(self, data: pd.DataFrame) -> BacktestResult:
        """Run backtest on historical data"""
        logger.info(f"Starting backtest with initial capital: ${self.initial_capital}")
        
        # Reset state
        self.capital = self.initial_capital
        self.trades = []
        self.equity_curve = [self.initial_capital]
        self.positions = {}
        
        # Ensure data is sorted by time
        data = data.sort_index()
        
        # Run strategy on each bar
        for timestamp, bar in data.iterrows():
            # Update equity curve
            self.update_equity_curve(bar)
            
            # Generate trading signals
            signal = self.strategy.generate_signal(data.loc[:timestamp])
            if signal:
                self.execute_signal(signal, bar, timestamp)
        
        # Close any remaining positions
        self.close_all_positions(data.iloc[-1], data.index[-1])
        
        # Calculate performance metrics
        return self.calculate_results()
    
    def execute_signal(self, signal: Signal, bar: pd.Series, timestamp: datetime):
        """Execute a trading signal"""
        position = self.positions.get(signal.symbol, 0)
        
        # Calculate position size based on risk management
        amount = self.calculate_position_size(signal, bar)
        if amount == 0:
            return
            
        # Calculate fees
        fee = abs(amount * signal.price * self.fee_rate)
        
        if signal.side == 'BUY':
            # Open long position
            if position <= 0:
                # Close existing short position if any
                if position < 0:
                    self.close_position(signal.symbol, bar, timestamp)
                
                # Open new long position
                self.positions[signal.symbol] = amount
                self.capital -= (amount * signal.price + fee)
                
                self.trades.append(Trade(
                    id=f"trade_{len(self.trades)}",
                    symbol=signal.symbol,
                    side='BUY',
                    entry_time=timestamp,
                    entry_price=signal.price,
                    exit_time=None,
                    exit_price=None,
                    amount=amount,
                    fee=fee
                ))
                
        elif signal.side == 'SELL':
            # Open short position
            if position >= 0:
                # Close existing long position if any
                if position > 0:
                    self.close_position(signal.symbol, bar, timestamp)
                
                # Open new short position
                self.positions[signal.symbol] = -amount
                self.capital -= fee
                
                self.trades.append(Trade(
                    id=f"trade_{len(self.trades)}",
                    symbol=signal.symbol,
                    side='SELL',
                    entry_time=timestamp,
                    entry_price=signal.price,
                    exit_time=None,
                    exit_price=None,
                    amount=amount,
                    fee=fee
                ))
    
    def close_position(self, symbol: str, bar: pd.Series, timestamp: datetime):
        """Close an open position"""
        position = self.positions.get(symbol, 0)
        if position == 0:
            return
            
        # Find the corresponding trade
        for trade in reversed(self.trades):
            if trade.symbol == symbol and trade.exit_time is None:
                # Calculate PnL
                exit_price = bar['close']
                fee = abs(position * exit_price * self.fee_rate)
                
                if trade.side == 'BUY':
                    pnl = position * (exit_price - trade.entry_price) - trade.fee - fee
                else:
                    pnl = position * (trade.entry_price - exit_price) - trade.fee - fee
                
                # Update trade
                trade.exit_time = timestamp
                trade.exit_price = exit_price
                trade.pnl = pnl
                trade.fee += fee
                
                # Update capital
                self.capital += pnl
                if trade.side == 'BUY':
                    self.capital += position * exit_price
                
                break
        
        # Clear position
        self.positions[symbol] = 0
    
    def close_all_positions(self, bar: pd.Series, timestamp: datetime):
        """Close all open positions"""
        for symbol in list(self.positions.keys()):
            self.close_position(symbol, bar, timestamp)
    
    def calculate_position_size(self, signal: Signal, bar: pd.Series) -> float:
        """Calculate position size based on risk management rules"""
        risk_config = self.strategy.config.risk_management
        position_size = risk_config.get('positionSize', 1.0)  # Percentage of capital
        
        # Calculate maximum position size based on available capital
        max_amount = (self.capital * position_size) / signal.price
        
        # Apply additional risk management rules
        if 'maxPositionSize' in risk_config:
            max_amount = min(max_amount, risk_config['maxPositionSize'])
        
        return max_amount
    
    def update_equity_curve(self, bar: pd.Series):
        """Update equity curve with current portfolio value"""
        portfolio_value = self.capital
        
        # Add unrealized PnL from open positions
        for symbol, amount in self.positions.items():
            if amount != 0:
                portfolio_value += amount * bar['close']
        
        self.equity_curve.append(portfolio_value)
    
    def calculate_results(self) -> BacktestResult:
        """Calculate backtest results and performance metrics"""
        equity_curve = pd.Series(self.equity_curve)
        returns = equity_curve.pct_change().dropna()
        
        # Calculate drawdown
        drawdown = (equity_curve - equity_curve.cummax()) / equity_curve.cummax() * 100
        
        # Calculate basic metrics
        total_pnl = equity_curve.iloc[-1] - self.initial_capital
        winning_trades = len([t for t in self.trades if t.pnl > 0])
        total_trades = len(self.trades)
        win_rate = winning_trades / total_trades if total_trades > 0 else 0
        
        # Calculate Sharpe Ratio (assuming risk-free rate = 0)
        sharpe_ratio = np.sqrt(252) * (returns.mean() / returns.std()) if len(returns) > 0 else 0
        
        # Calculate additional metrics
        metrics = {
            'total_trades': total_trades,
            'winning_trades': winning_trades,
            'losing_trades': total_trades - winning_trades,
            'avg_trade': total_pnl / total_trades if total_trades > 0 else 0,
            'profit_factor': self.calculate_profit_factor(),
            'recovery_factor': abs(total_pnl / drawdown.min()) if drawdown.min() < 0 else float('inf'),
            'expectancy': self.calculate_expectancy()
        }
        
        return BacktestResult(
            total_pnl=total_pnl,
            win_rate=win_rate,
            sharpe_ratio=sharpe_ratio,
            max_drawdown=abs(drawdown.min()),
            trades=self.trades,
            equity_curve=equity_curve,
            drawdown_curve=drawdown,
            metrics=metrics
        )
    
    def calculate_profit_factor(self) -> float:
        """Calculate profit factor (gross profit / gross loss)"""
        gross_profit = sum(t.pnl for t in self.trades if t.pnl > 0)
        gross_loss = abs(sum(t.pnl for t in self.trades if t.pnl < 0))
        return gross_profit / gross_loss if gross_loss != 0 else float('inf')
    
    def calculate_expectancy(self) -> float:
        """Calculate system expectancy (average win * win rate - average loss * loss rate)"""
        if not self.trades:
            return 0
            
        winning_trades = [t.pnl for t in self.trades if t.pnl > 0]
        losing_trades = [t.pnl for t in self.trades if t.pnl < 0]
        
        avg_win = np.mean(winning_trades) if winning_trades else 0
        avg_loss = abs(np.mean(losing_trades)) if losing_trades else 0
        win_rate = len(winning_trades) / len(self.trades)
        
        return (avg_win * win_rate) - (avg_loss * (1 - win_rate)) 