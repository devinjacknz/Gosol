import pandas as pd
import numpy as np
from typing import Dict, List, Optional, Union
from dataclasses import dataclass
import logging
from streaming_indicators import StreamingRSI, StreamingMACD, StreamingBB, StreamingSMA, StreamingEMA

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class TechnicalAnalysis:
    """Technical analysis indicators calculation using streaming indicators"""
    
    def __init__(self):
        self.streaming_sma = {}  # Dict to store SMA calculators for different periods
        self.streaming_ema = {}  # Dict to store EMA calculators for different periods
        self.streaming_rsi = StreamingRSI()
        self.streaming_macd = StreamingMACD()
        self.streaming_bb = StreamingBB()
    
    def sma(self, data: pd.DataFrame, period: int = 20) -> pd.Series:
        """Simple Moving Average using streaming calculator"""
        if period not in self.streaming_sma:
            self.streaming_sma[period] = StreamingSMA(period)
        
        result = pd.Series(index=data.index)
        for idx, row in data.iterrows():
            result[idx] = self.streaming_sma[period].add(row['price'])
        return result
    
    def ema(self, data: pd.DataFrame, period: int = 20) -> pd.Series:
        """Exponential Moving Average using streaming calculator"""
        if period not in self.streaming_ema:
            self.streaming_ema[period] = StreamingEMA(period)
        
        result = pd.Series(index=data.index)
        for idx, row in data.iterrows():
            result[idx] = self.streaming_ema[period].add(row['price'])
        return result
    
    def rsi(self, data: pd.DataFrame, period: int = 14) -> pd.Series:
        """Relative Strength Index using streaming calculator"""
        result = pd.Series(index=data.index)
        for idx, row in data.iterrows():
            result[idx] = self.streaming_rsi.add(row['price'])
        return result
    
    def macd(self, data: pd.DataFrame, fastperiod: int = 12, 
             slowperiod: int = 26, signalperiod: int = 9) -> Dict[str, pd.Series]:
        """Moving Average Convergence Divergence using streaming calculator"""
        macd_line = pd.Series(index=data.index)
        signal_line = pd.Series(index=data.index)
        histogram = pd.Series(index=data.index)
        
        for idx, row in data.iterrows():
            macd_val, signal_val = self.streaming_macd.add(row['price'])
            macd_line[idx] = macd_val
            signal_line[idx] = signal_val
            histogram[idx] = macd_val - signal_val
        
        return {
            'macd': macd_line,
            'signal': signal_line,
            'histogram': histogram
        }
    
    def bollinger_bands(self, data: pd.DataFrame, period: int = 20, 
                       num_std: float = 2.0) -> Dict[str, pd.Series]:
        """Bollinger Bands using streaming calculator"""
        upper = pd.Series(index=data.index)
        middle = pd.Series(index=data.index)
        lower = pd.Series(index=data.index)
        
        for idx, row in data.iterrows():
            up, mid, low = self.streaming_bb.add(row['price'])
            upper[idx] = up
            middle[idx] = mid
            lower[idx] = low
        
        return {
            'upper': upper,
            'middle': middle,
            'lower': lower
        }
    
    def atr(self, data: pd.DataFrame, period: int = 14) -> pd.Series:
        """Average True Range using pandas"""
        high = data['high']
        low = data['low']
        close = data['price']
        
        tr1 = high - low
        tr2 = abs(high - close.shift())
        tr3 = abs(low - close.shift())
        tr = pd.concat([tr1, tr2, tr3], axis=1).max(axis=1)
        
        return tr.ewm(span=period, adjust=False).mean()
    
    def adx(self, data: pd.DataFrame, period: int = 14) -> pd.Series:
        """Average Directional Index"""
        high = data['high']
        low = data['low']
        close = data['price']
        
        # Calculate +DM and -DM
        high_diff = high.diff()
        low_diff = low.diff()
        
        pos_dm = ((high_diff > 0) & (high_diff > -low_diff)) * high_diff
        neg_dm = ((low_diff < 0) & (-low_diff > high_diff)) * -low_diff
        
        # Calculate TR
        tr = self.atr(data, period)
        
        # Calculate +DI and -DI
        pos_di = 100 * (pos_dm.rolling(window=period).mean() / tr)
        neg_di = 100 * (neg_dm.rolling(window=period).mean() / tr)
        
        # Calculate DX and ADX
        dx = 100 * abs(pos_di - neg_di) / (pos_di + neg_di)
        adx = dx.rolling(window=period).mean()
        
        return adx
    
    def obv(self, data: pd.DataFrame) -> pd.Series:
        """On Balance Volume using pandas"""
        return (np.sign(data['price'].diff()) * data['volume']).cumsum()
    
    def mfi(self, data: pd.DataFrame, period: int = 14) -> pd.Series:
        """Money Flow Index using pandas"""
        typical_price = (data['high'] + data['low'] + data['price']) / 3
        money_flow = typical_price * data['volume']
        
        pos_flow = money_flow.where(typical_price > typical_price.shift(), 0)
        neg_flow = money_flow.where(typical_price < typical_price.shift(), 0)
        
        pos_mf = pos_flow.ewm(span=period, adjust=False).mean()
        neg_mf = neg_flow.ewm(span=period, adjust=False).mean()
        
        return 100 - (100 / (1 + pos_mf / neg_mf))
    
    def cci(self, data: pd.DataFrame, period: int = 20) -> pd.Series:
        """Commodity Channel Index"""
        typical_price = (data['high'] + data['low'] + data['price']) / 3
        sma = typical_price.rolling(window=period).mean()
        mean_deviation = abs(typical_price - sma).rolling(window=period).mean()
        
        return (typical_price - sma) / (0.015 * mean_deviation)
    
    def stochastic(self, data: pd.DataFrame, k_period: int = 14, 
                  d_period: int = 3) -> Dict[str, pd.Series]:
        """Stochastic Oscillator"""
        low_min = data['low'].rolling(window=k_period).min()
        high_max = data['high'].rolling(window=k_period).max()
        
        k = 100 * (data['price'] - low_min) / (high_max - low_min)
        d = k.rolling(window=d_period).mean()
        
        return {
            'k': k,
            'd': d
        }
    
    def williams_r(self, data: pd.DataFrame, period: int = 14) -> pd.Series:
        """Williams %R"""
        high = data['high'].rolling(window=period).max()
        low = data['low'].rolling(window=period).min()
        r = -100 * (high - data['price']) / (high - low)
        return r
    
    def momentum(self, data: pd.DataFrame, period: int = 10) -> pd.Series:
        """Momentum"""
        return data['price'].diff(period)
    
    def vwap(self, data: pd.DataFrame) -> pd.Series:
        """Volume Weighted Average Price using pandas"""
        return (data['price'] * data['volume']).cumsum() / data['volume'].cumsum()
    
    def support_resistance(self, data: pd.DataFrame, period: int = 20,
                         threshold: float = 0.05) -> Dict[str, List[float]]:
        """Calculate support and resistance levels using pandas"""
        highs = data['high'].rolling(window=period).max()
        lows = data['low'].rolling(window=period).min()
        
        potential_support = lows[lows.shift(1) > lows].dropna()
        potential_resistance = highs[highs.shift(1) < highs].dropna()
        
        def group_levels(levels: pd.Series, threshold: float) -> List[float]:
            if len(levels) == 0:
                return []
            
            grouped = []
            current_group = [levels.iloc[0]]
            
            for level in levels.iloc[1:]:
                if abs(level - np.mean(current_group)) / np.mean(current_group) <= threshold:
                    current_group.append(level)
                else:
                    grouped.append(np.mean(current_group))
                    current_group = [level]
            
            if current_group:
                grouped.append(np.mean(current_group))
            
            return sorted(grouped)
        
        return {
            'support': group_levels(potential_support, threshold),
            'resistance': group_levels(potential_resistance, threshold)
        }
    
    def pivot_points(self, data: pd.DataFrame) -> Dict[str, float]:
        """Calculate pivot points (Floor Pivot Points)"""
        high = data['high'].iloc[-1]
        low = data['low'].iloc[-1]
        close = data['price'].iloc[-1]
        
        pivot = (high + low + close) / 3
        r1 = 2 * pivot - low
        r2 = pivot + (high - low)
        r3 = high + 2 * (pivot - low)
        s1 = 2 * pivot - high
        s2 = pivot - (high - low)
        s3 = low - 2 * (high - pivot)
        
        return {
            'pivot': pivot,
            'r1': r1,
            'r2': r2,
            'r3': r3,
            's1': s1,
            's2': s2,
            's3': s3
        }
    
    @staticmethod
    def calculate_trend_strength(data: pd.Series, period: int = 14) -> Dict[str, float]:
        """Calculate trend strength indicators"""
        # ADX (Average Directional Index)
        adx = talib.ADX(data.high, data.low, data.close, timeperiod=period)
        
        # Aroon Indicator
        aroon_up, aroon_down = talib.AROON(data.high, data.low, timeperiod=period)
        
        # CCI (Commodity Channel Index)
        cci = talib.CCI(data.high, data.low, data.close, timeperiod=period)
        
        return {
            'adx': adx.iloc[-1],
            'aroon_up': aroon_up.iloc[-1],
            'aroon_down': aroon_down.iloc[-1],
            'cci': cci.iloc[-1]
        }
    
    @staticmethod
    def calculate_momentum_indicators(data: pd.Series, period: int = 14) -> Dict[str, float]:
        """Calculate momentum indicators"""
        # ROC (Rate of Change)
        roc = talib.ROC(data, timeperiod=period)
        
        # MOM (Momentum)
        mom = talib.MOM(data, timeperiod=period)
        
        # Williams %R
        willr = talib.WILLR(data.high, data.low, data.close, timeperiod=period)
        
        return {
            'roc': roc.iloc[-1],
            'momentum': mom.iloc[-1],
            'willr': willr.iloc[-1]
        }
    
    @staticmethod
    def calculate_volatility_indicators(data: pd.Series) -> Dict[str, float]:
        """Calculate volatility indicators"""
        # Standard Deviation
        std = data.rolling(window=20).std()
        
        # Historical Volatility
        returns = np.log(data / data.shift(1))
        hist_vol = returns.rolling(window=20).std() * np.sqrt(252)
        
        # Chaikin Volatility
        high_low = data.high - data.low
        chaikin_vol = talib.EMA(high_low, timeperiod=10) / \
                     talib.EMA(high_low, timeperiod=10).shift(10)
        
        return {
            'std': std.iloc[-1],
            'hist_vol': hist_vol.iloc[-1],
            'chaikin_vol': chaikin_vol.iloc[-1]
        }
    
    @staticmethod
    def calculate_volume_indicators(close: pd.Series,
                                  volume: pd.Series) -> Dict[str, float]:
        """Calculate volume-based indicators"""
        # Chaikin A/D Line
        ad = talib.AD(close.high, close.low, close.close, volume)
        
        # Money Flow Index
        mfi = talib.MFI(close.high, close.low, close.close, volume, timeperiod=14)
        
        # Volume Rate of Change
        vroc = talib.ROC(volume, timeperiod=14)
        
        return {
            'ad': ad.iloc[-1],
            'mfi': mfi.iloc[-1],
            'vroc': vroc.iloc[-1]
        }
    
    @staticmethod
    def identify_patterns(data: pd.DataFrame) -> Dict[str, bool]:
        """Identify candlestick patterns"""
        patterns = {
            'doji': talib.CDLDOJI(data.open, data.high, data.low, data.close),
            'engulfing': talib.CDLENGULFING(data.open, data.high, data.low, data.close),
            'hammer': talib.CDLHAMMER(data.open, data.high, data.low, data.close),
            'shooting_star': talib.CDLSHOOTINGSTAR(data.open, data.high, data.low, data.close),
            'morning_star': talib.CDLMORNINGSTAR(data.open, data.high, data.low, data.close),
            'evening_star': talib.CDLEVENINGSTAR(data.open, data.high, data.low, data.close)
        }
        
        return {name: bool(pattern.iloc[-1]) for name, pattern in patterns.items()} 