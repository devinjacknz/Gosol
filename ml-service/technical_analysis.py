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
        high = data.high.astype(float)
        low = data.low.astype(float)
        close = data.close.astype(float)
        
        # ADX calculation
        high_diff = high.diff()
        low_diff = low.diff()
        
        pos_dm = pd.Series(np.where((high_diff > 0) & (high_diff > -low_diff), high_diff, 0))
        neg_dm = pd.Series(np.where((low_diff < 0) & (-low_diff > high_diff), -low_diff, 0))
        
        tr1 = high - low
        tr2 = abs(high - close.shift())
        tr3 = abs(low - close.shift())
        tr = pd.concat([tr1, tr2, tr3], axis=1).max(axis=1)
        
        pos_di = 100 * (pos_dm.ewm(span=period).mean() / tr.ewm(span=period).mean())
        neg_di = 100 * (neg_dm.ewm(span=period).mean() / tr.ewm(span=period).mean())
        
        dx = 100 * abs(pos_di - neg_di) / (pos_di + neg_di)
        adx = dx.ewm(span=period).mean()
        
        # Aroon calculation
        high_period = high.rolling(window=period).apply(lambda x: (period - x.argmax()) / period * 100)
        low_period = low.rolling(window=period).apply(lambda x: (period - x.argmin()) / period * 100)
        
        # CCI calculation
        tp = (high + low + close) / 3
        tp_sma = tp.rolling(window=period).mean()
        mad = tp.rolling(window=period).apply(lambda x: abs(x - x.mean()).mean())
        cci = (tp - tp_sma) / (0.015 * mad)
        
        return {
            'adx': float(adx.iloc[-1]),
            'aroon_up': float(high_period.iloc[-1]),
            'aroon_down': float(low_period.iloc[-1]),
            'cci': float(cci.iloc[-1])
        }
    
    @staticmethod
    def calculate_momentum_indicators(data: pd.Series, period: int = 14) -> Dict[str, float]:
        """Calculate momentum indicators"""
        close = data.close.astype(float)
        high = data.high.astype(float)
        low = data.low.astype(float)
        
        # ROC calculation
        roc = ((close - close.shift(period)) / close.shift(period)) * 100
        
        # Momentum calculation
        mom = close - close.shift(period)
        
        # Williams %R calculation
        highest_high = high.rolling(window=period).max()
        lowest_low = low.rolling(window=period).min()
        willr = -100 * (highest_high - close) / (highest_high - lowest_low)
        
        return {
            'roc': float(roc.iloc[-1]),
            'momentum': float(mom.iloc[-1]),
            'willr': float(willr.iloc[-1])
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
        alpha = 2.0 / (10 + 1)
        ema_hl = high_low.ewm(alpha=alpha, adjust=False).mean()
        chaikin_vol = ema_hl / ema_hl.shift(10)
        
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
        clv = ((close - close.low) - (close.high - close)) / (close.high - close.low)
        ad = (clv * volume).cumsum()
        
        # Money Flow Index
        tp = (close.high + close.low + close) / 3
        rmf = tp * volume
        
        positive_flow = (tp > tp.shift(1)).astype(float) * rmf
        negative_flow = (tp < tp.shift(1)).astype(float) * rmf
        
        positive_mf = positive_flow.rolling(window=14).sum()
        negative_mf = negative_flow.rolling(window=14).sum()
        
        mfi = 100 - (100 / (1 + positive_mf / negative_mf))
        
        # Volume Rate of Change
        vroc = ((volume - volume.shift(14)) / volume.shift(14)) * 100
        
        return {
            'ad': ad.iloc[-1],
            'mfi': mfi.iloc[-1],
            'vroc': vroc.iloc[-1]
        }
    
    @staticmethod
    def identify_patterns(data: pd.DataFrame) -> Dict[str, bool]:
        """Identify candlestick patterns"""
        open_price = data['open'].astype(float)
        high = data['high'].astype(float)
        low = data['low'].astype(float)
        close = data['close'].astype(float)
        
        body = close - open_price
        body_size = abs(body)
        upper_shadow = high - pd.DataFrame([open_price, close]).max(axis=0)
        lower_shadow = pd.DataFrame([open_price, close]).min(axis=0) - low
        
        avg_price = (high + low) / 2
        doji_threshold = 0.001
        
        patterns = {}
        
        # Doji pattern
        patterns['doji'] = bool(body_size.iloc[-1] <= (avg_price.iloc[-1] * doji_threshold))
        
        # Engulfing pattern
        prev_body = close.shift(1) - open_price.shift(1)
        bull_engulf = (body > 0) & (open_price < close.shift(1)) & (close > open_price.shift(1))
        bear_engulf = (body < 0) & (open_price > close.shift(1)) & (close < open_price.shift(1))
        patterns['engulfing'] = bool((bull_engulf | bear_engulf).iloc[-1])
        
        # Hammer pattern
        patterns['hammer'] = bool((lower_shadow > (2 * body_size)).iloc[-1] & (upper_shadow < body_size).iloc[-1])
        
        # Shooting Star pattern
        patterns['shooting_star'] = bool((upper_shadow > (2 * body_size)).iloc[-1] & (lower_shadow < body_size).iloc[-1])
        
        # Morning/Evening Star patterns
        patterns['morning_star'] = bool(
            (body.iloc[-3] < 0) &  # First day: long black
            (body_size.iloc[-2] < body_size.iloc[-3] * 0.3) &  # Second day: small body
            (body.iloc[-1] > 0) &  # Third day: long white
            (close.iloc[-1] > (open_price.iloc[-3] + close.iloc[-3]) / 2)  # Close above midpoint
        )
        
        patterns['evening_star'] = bool(
            (body.iloc[-3] > 0) &  # First day: long white
            (body_size.iloc[-2] < body_size.iloc[-3] * 0.3) &  # Second day: small body
            (body.iloc[-1] < 0) &  # Third day: long black
            (close.iloc[-1] < (open_price.iloc[-3] + close.iloc[-3]) / 2)  # Close below midpoint
        )
        
        return patterns   