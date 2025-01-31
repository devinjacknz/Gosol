import pandas as pd
import numpy as np
from typing import Dict, List, Optional, Union, Any
from dataclasses import dataclass
from loguru import logger

def safe_fill_series(series: pd.Series, method: str = 'ffill', fill_value: Any = None) -> pd.Series:
    try:
        if not isinstance(series, pd.Series):
            return pd.Series([], dtype=float)
        
        series = pd.to_numeric(series, errors='coerce')
        
        if method == 'ffill':
            series = series.ffill()
        elif method == 'bfill':
            series = series.bfill()
        elif method == 'zero':
            series = series.fillna(0.0)
        
        if fill_value is not None:
            series = series.replace([np.inf, -np.inf, np.nan], fill_value)
        else:
            series = series.replace([np.inf, -np.inf], np.nan)
            if method not in ['ffill', 'bfill', 'zero']:
                series = series.fillna(method=method)
        
        return series
    except Exception as e:
        logger.error(f"Error in safe_fill_series: {e}")
        return pd.Series([], dtype=float)

def safe_compare_series(series: pd.Series, value: Union[int, float, pd.Series], operator: str) -> pd.Series:
    try:
        series = pd.to_numeric(series, errors='coerce')
        if isinstance(value, pd.Series):
            value = pd.to_numeric(value, errors='coerce')
        elif isinstance(value, (int, float)):
            value = float(value)
        else:
            return pd.Series(False, index=series.index)
        
        result = pd.Series(False, index=series.index)
        valid_mask = ~(series.isna() | (isinstance(value, pd.Series) & value.isna()))
        
        if operator == '>':
            result[valid_mask] = series[valid_mask] > value[valid_mask] if isinstance(value, pd.Series) else series[valid_mask] > value
        elif operator == '<':
            result[valid_mask] = series[valid_mask] < value[valid_mask] if isinstance(value, pd.Series) else series[valid_mask] < value
        elif operator == '>=':
            result[valid_mask] = series[valid_mask] >= value[valid_mask] if isinstance(value, pd.Series) else series[valid_mask] >= value
        elif operator == '<=':
            result[valid_mask] = series[valid_mask] <= value[valid_mask] if isinstance(value, pd.Series) else series[valid_mask] <= value
        elif operator == '==':
            result[valid_mask] = series[valid_mask] == value[valid_mask] if isinstance(value, pd.Series) else series[valid_mask] == value
        elif operator == '!=':
            result[valid_mask] = series[valid_mask] != value[valid_mask] if isinstance(value, pd.Series) else series[valid_mask] != value
            
        return result
    except Exception as e:
        logger.error(f"Error in safe_compare_series: {e}")
        return pd.Series(False, index=series.index)
from streaming_indicators import StreamingRSI, StreamingMACD, StreamingBB, StreamingSMA, StreamingEMA

# Configure logger
logger.add(
    "logs/technical_analysis.log",
    rotation="500 MB",
    retention="10 days",
    level="INFO",
    format="{time:YYYY-MM-DD HH:mm:ss} | {level} | {message}"
)



class TechnicalAnalysis:
    """Technical analysis indicators calculation using streaming indicators"""
    
    @classmethod
    def safe_float(cls, x: Union[pd.Series, float, int], default: float = 0.0) -> float:
        """Safely convert value to float with robust error handling"""
        try:
            if isinstance(x, pd.Series):
                val = float(x.iloc[-1]) if not x.empty else default
            else:
                val = float(x)
            return default if pd.isna(val) else val
        except (IndexError, ValueError, TypeError):
            return default

    def __init__(self):
        try:
            self.streaming_sma = {}
            self.streaming_ema = {}
            self.streaming_rsi = StreamingRSI()
            self.streaming_macd = StreamingMACD()
            self.streaming_bb = StreamingBB()
        except Exception as e:
            logger.error(f"Failed to initialize streaming indicators: {e}")
            raise RuntimeError("Failed to initialize technical analysis system")
    
    def sma(self, data: pd.DataFrame, period: int = 20) -> pd.Series:
        """Simple Moving Average using streaming calculator"""
        if not isinstance(data, pd.DataFrame):
            raise TypeError("data must be a pandas DataFrame")
        if 'price' not in data.columns:
            raise ValueError("DataFrame must contain 'price' column")
        if not isinstance(period, int) or period <= 0:
            raise ValueError("period must be a positive integer")
            
        try:
            if period not in self.streaming_sma:
                self.streaming_sma[period] = StreamingSMA(period)
            
            result = pd.Series(index=data.index, dtype=float)
            for idx, row in data.iterrows():
                try:
                    price = float(row['price'])
                    if pd.isna(price):
                        logger.warning(f"NaN value detected at index {idx}")
                        continue
                    result[idx] = self.streaming_sma[period].add(price)
                except (ValueError, TypeError) as e:
                    logger.error(f"Error processing price at index {idx}: {e}")
                    continue
            return result
        except Exception as e:
            logger.error(f"Error calculating SMA: {e}")
            raise
    
    def ema(self, data: pd.DataFrame, period: int = 20) -> pd.Series:
        """Exponential Moving Average using streaming calculator"""
        if not isinstance(data, pd.DataFrame):
            raise TypeError("data must be a pandas DataFrame")
        if 'price' not in data.columns:
            raise ValueError("DataFrame must contain 'price' column")
        if not isinstance(period, int) or period <= 0:
            raise ValueError("period must be a positive integer")
            
        try:
            if period not in self.streaming_ema:
                self.streaming_ema[period] = StreamingEMA(period)
            
            result = pd.Series(index=data.index, dtype=float)
            for idx, row in data.iterrows():
                try:
                    price = float(row['price'])
                    if pd.isna(price):
                        logger.warning(f"NaN value detected at index {idx}")
                        continue
                    result[idx] = self.streaming_ema[period].add(price)
                except (ValueError, TypeError) as e:
                    logger.error(f"Error processing price at index {idx}: {e}")
                    continue
            return result
        except Exception as e:
            logger.error(f"Error calculating EMA: {e}")
            raise
    
    def rsi(self, data: pd.DataFrame, period: int = 14) -> pd.Series:
        """Relative Strength Index using streaming calculator"""
        if not isinstance(data, pd.DataFrame):
            raise TypeError("data must be a pandas DataFrame")
        if 'price' not in data.columns:
            raise ValueError("DataFrame must contain 'price' column")
        if not isinstance(period, int) or period <= 0:
            raise ValueError("period must be a positive integer")
            
        try:
            result = pd.Series(index=data.index, dtype=float)
            for idx, row in data.iterrows():
                try:
                    price = float(row['price'])
                    if pd.isna(price):
                        logger.warning(f"NaN value detected at index {idx}")
                        continue
                    result[idx] = self.streaming_rsi.add(price)
                except (ValueError, TypeError) as e:
                    logger.error(f"Error processing price at index {idx}: {e}")
                    continue
            
            # Validate RSI values are within expected range
            result = result.clip(0, 100)
            return result
        except Exception as e:
            logger.error(f"Error calculating RSI: {e}")
            raise
    
    def macd(self, data: pd.DataFrame, fastperiod: int = 12, 
             slowperiod: int = 26, signalperiod: int = 9) -> Dict[str, pd.Series]:
        """Moving Average Convergence Divergence using streaming calculator"""
        if not isinstance(data, pd.DataFrame):
            raise TypeError("data must be a pandas DataFrame")
        if 'price' not in data.columns:
            raise ValueError("DataFrame must contain 'price' column")
        if not all(isinstance(p, int) and p > 0 for p in [fastperiod, slowperiod, signalperiod]):
            raise ValueError("All periods must be positive integers")
        if fastperiod >= slowperiod:
            raise ValueError("fastperiod must be less than slowperiod")
            
        try:
            macd_line = pd.Series(index=data.index, dtype=float)
            signal_line = pd.Series(index=data.index, dtype=float)
            histogram = pd.Series(index=data.index, dtype=float)
            
            for idx, row in data.iterrows():
                try:
                    price = float(row['price'])
                    if pd.isna(price):
                        logger.warning(f"NaN value detected at index {idx}")
                        continue
                    macd_val, signal_val = self.streaming_macd.add(price)
                    if not pd.isna(macd_val) and not pd.isna(signal_val):
                        macd_line[idx] = macd_val
                        signal_line[idx] = signal_val
                        histogram[idx] = macd_val - signal_val
                except (ValueError, TypeError) as e:
                    logger.error(f"Error processing price at index {idx}: {e}")
                    continue
            
            return {
                'macd': macd_line,
                'signal': signal_line,
                'histogram': histogram
            }
        except Exception as e:
            logger.error(f"Error calculating MACD: {e}")
            raise
    
    def bollinger_bands(self, data: pd.DataFrame, period: int = 20, 
                       num_std: float = 2.0) -> Dict[str, pd.Series]:
        """Bollinger Bands using streaming calculator"""
        if not isinstance(data, pd.DataFrame):
            raise TypeError("data must be a pandas DataFrame")
        if 'price' not in data.columns:
            raise ValueError("DataFrame must contain 'price' column")
        if not isinstance(period, int) or period <= 0:
            raise ValueError("period must be a positive integer")
        if not isinstance(num_std, (int, float)) or num_std <= 0:
            raise ValueError("num_std must be a positive number")
            
        try:
            upper = pd.Series(index=data.index, dtype=float)
            middle = pd.Series(index=data.index, dtype=float)
            lower = pd.Series(index=data.index, dtype=float)
            
            for idx, row in data.iterrows():
                try:
                    price = float(row['price'])
                    if pd.isna(price):
                        logger.warning(f"NaN value detected at index {idx}")
                        continue
                    up, mid, low = self.streaming_bb.add(price)
                    if all(not pd.isna(x) for x in (up, mid, low)):
                        upper[idx] = up
                        middle[idx] = mid
                        lower[idx] = low
                except (ValueError, TypeError) as e:
                    logger.error(f"Error processing price at index {idx}: {e}")
                    continue
            
            # Validate band relationships
            if not (upper >= middle).all() or not (middle >= lower).all():
                logger.warning("Bollinger Bands values are not in expected order")
            
            return {
                'upper': upper,
                'middle': middle,
                'lower': lower
            }
        except Exception as e:
            logger.error(f"Error calculating Bollinger Bands: {e}")
            raise
    
    def atr(self, data: pd.DataFrame, period: int = 14) -> pd.Series:
        """Average True Range using pandas"""
        if not isinstance(data, pd.DataFrame):
            raise TypeError("data must be a pandas DataFrame")
        if not all(col in data.columns for col in ['high', 'low', 'price']):
            raise ValueError("DataFrame must contain 'high', 'low', and 'price' columns")
        if not isinstance(period, int) or period <= 0:
            raise ValueError("period must be a positive integer")
            
        try:
            high = pd.to_numeric(data['high'], errors='coerce')
            low = pd.to_numeric(data['low'], errors='coerce')
            close = pd.to_numeric(data['price'], errors='coerce')
            
            if high.isna().any() or low.isna().any() or close.isna().any():
                logger.warning("NaN values detected in price data")
            
            tr1 = high - low
            tr2 = abs(high - close.shift())
            tr3 = abs(low - close.shift())
            tr = pd.concat([tr1, tr2, tr3], axis=1).max(axis=1)
            
            atr = tr.ewm(span=period, adjust=False).mean()
            return pd.Series(atr, dtype=float)
        except Exception as e:
            logger.error(f"Error calculating ATR: {e}")
            raise
    
    def adx(self, data: pd.DataFrame, period: int = 14) -> pd.Series:
        """Average Directional Index"""
        if not isinstance(data, pd.DataFrame):
            raise TypeError("data must be a pandas DataFrame")
        if not all(col in data.columns for col in ['high', 'low', 'price']):
            raise ValueError("DataFrame must contain 'high', 'low', and 'price' columns")
        if not isinstance(period, int) or period <= 0:
            raise ValueError("period must be a positive integer")
        if len(data) < period + 1:
            raise ValueError(f"DataFrame must contain at least {period + 1} rows")
            
        try:
            high = pd.to_numeric(data['high'], errors='coerce')
            low = pd.to_numeric(data['low'], errors='coerce')
            close = pd.to_numeric(data['price'], errors='coerce')
            
            if high.isna().any() or low.isna().any() or close.isna().any():
                logger.warning("NaN values detected in price data")
            
            # Calculate +DM and -DM
            high_diff = pd.Series(high.diff(), dtype=float)
            low_diff = pd.Series(low.diff(), dtype=float)
            
            pos_dm = pd.Series(np.where((high_diff > 0) & (high_diff > -low_diff), high_diff, 0), index=high.index)
            neg_dm = pd.Series(np.where((low_diff < 0) & (-low_diff > high_diff), -low_diff, 0), index=low.index)
            
            # Calculate TR
            tr = self.atr(data, period)
            
            # Calculate +DI and -DI with zero division protection
            pos_di = 100 * (pos_dm.rolling(window=period).mean() / tr.replace(0, float('inf')))
            neg_di = 100 * (neg_dm.rolling(window=period).mean() / tr.replace(0, float('inf')))
            
            # Calculate DX and ADX with zero division protection
            dx = 100 * abs(pos_di - neg_di) / (pos_di + neg_di).replace(0, float('inf'))
            adx = dx.rolling(window=period).mean()
            
            # Clip values to valid range
            adx = adx.clip(0, 100)
            return pd.Series(adx, dtype=float)
        except Exception as e:
            logger.error(f"Error calculating ADX: {e}")
            raise
    
    def obv(self, data: pd.DataFrame) -> pd.Series:
        """On Balance Volume using pandas"""
        if not isinstance(data, pd.DataFrame):
            raise TypeError("data must be a pandas DataFrame")
        if not all(col in data.columns for col in ['price', 'volume']):
            raise ValueError("DataFrame must contain 'price' and 'volume' columns")
            
        try:
            price = pd.to_numeric(data['price'], errors='coerce')
            volume = pd.to_numeric(data['volume'], errors='coerce')
            
            if price.isna().any() or volume.isna().any():
                logger.warning("NaN values detected in price or volume data")
            
            price_diff = price.diff()
            obv_values = (np.sign(price_diff) * volume).cumsum()
            obv = pd.Series(obv_values, index=data.index, dtype=float)
            
            # Handle NaN values without using fillna method
            obv = obv.replace([np.inf, -np.inf], np.nan)
            last_valid = None
            for i in range(len(obv)):
                if pd.isna(obv.iloc[i]):
                    obv.iloc[i] = last_valid if last_valid is not None else 0
                else:
                    last_valid = obv.iloc[i]
            
            return obv
        except Exception as e:
            logger.error(f"Error calculating OBV: {e}")
            raise
    
    def mfi(self, data: pd.DataFrame, period: int = 14) -> pd.Series:
        """Money Flow Index using pandas"""
        if not isinstance(data, pd.DataFrame):
            raise TypeError("data must be a pandas DataFrame")
        if not all(col in data.columns for col in ['high', 'low', 'price', 'volume']):
            raise ValueError("DataFrame must contain 'high', 'low', 'price', and 'volume' columns")
        if not isinstance(period, int) or period <= 0:
            raise ValueError("period must be a positive integer")
            
        try:
            high = pd.to_numeric(data['high'], errors='coerce')
            low = pd.to_numeric(data['low'], errors='coerce')
            close = pd.to_numeric(data['price'], errors='coerce')
            volume = pd.to_numeric(data['volume'], errors='coerce')
            
            if high.isna().any() or low.isna().any() or close.isna().any() or volume.isna().any():
                logger.warning("NaN values detected in price or volume data")
            
            typical_price = (high + low + close) / 3
            money_flow = typical_price * volume
            
            pos_flow = money_flow.where(typical_price > typical_price.shift(), 0)
            neg_flow = money_flow.where(typical_price < typical_price.shift(), 0)
            
            pos_mf = pos_flow.ewm(span=period, adjust=False).mean()
            neg_mf = neg_flow.ewm(span=period, adjust=False).mean()
            
            mfi = 100 - (100 / (1 + pos_mf / neg_mf.replace(0, float('inf'))))
            return pd.Series(mfi, dtype=float).clip(0, 100)
        except Exception as e:
            logger.error(f"Error calculating MFI: {e}")
            raise
    
    def cci(self, data: pd.DataFrame, period: int = 20) -> pd.Series:
        """Commodity Channel Index"""
        if not isinstance(data, pd.DataFrame):
            raise TypeError("data must be a pandas DataFrame")
        if not all(col in data.columns for col in ['high', 'low', 'price']):
            raise ValueError("DataFrame must contain 'high', 'low', and 'price' columns")
        if not isinstance(period, int) or period <= 0:
            raise ValueError("period must be a positive integer")
            
        try:
            high = pd.to_numeric(data['high'], errors='coerce')
            low = pd.to_numeric(data['low'], errors='coerce')
            close = pd.to_numeric(data['price'], errors='coerce')
            
            if high.isna().any() or low.isna().any() or close.isna().any():
                logger.warning("NaN values detected in price data")
            
            typical_price = (high + low + close) / 3
            sma = typical_price.rolling(window=period).mean()
            mean_deviation = abs(typical_price - sma).rolling(window=period).mean()
            
            # Avoid division by zero
            mean_deviation = mean_deviation.replace(0, float('inf'))
            cci = (typical_price - sma) / (0.015 * mean_deviation)
            
            return pd.Series(cci, dtype=float)
        except Exception as e:
            logger.error(f"Error calculating CCI: {e}")
            raise
    
    def stochastic(self, data: pd.DataFrame, k_period: int = 14, 
                  d_period: int = 3) -> Dict[str, pd.Series]:
        """Stochastic Oscillator"""
        if not isinstance(data, pd.DataFrame):
            raise TypeError("data must be a pandas DataFrame")
        if not all(col in data.columns for col in ['high', 'low', 'price']):
            raise ValueError("DataFrame must contain 'high', 'low', and 'price' columns")
        if not all(isinstance(p, int) and p > 0 for p in [k_period, d_period]):
            raise ValueError("k_period and d_period must be positive integers")
            
        try:
            high = pd.to_numeric(data['high'], errors='coerce')
            low = pd.to_numeric(data['low'], errors='coerce')
            close = pd.to_numeric(data['price'], errors='coerce')
            
            if high.isna().any() or low.isna().any() or close.isna().any():
                logger.warning("NaN values detected in price data")
            
            low_min = low.rolling(window=k_period).min()
            high_max = high.rolling(window=k_period).max()
            
            # Avoid division by zero
            denominator = high_max - low_min
            denominator = denominator.replace(0, float('inf'))
            
            k = 100 * (close - low_min) / denominator
            d = k.rolling(window=d_period).mean()
            
            # Clip values to valid range
            k = k.clip(0, 100)
            d = d.clip(0, 100)
            
            return {
                'k': pd.Series(k, dtype=float),
                'd': pd.Series(d, dtype=float)
            }
        except Exception as e:
            logger.error(f"Error calculating Stochastic Oscillator: {e}")
            raise
    
    def williams_r(self, data: pd.DataFrame, period: int = 14) -> pd.Series:
        """Williams %R"""
        if not isinstance(data, pd.DataFrame):
            raise TypeError("data must be a pandas DataFrame")
        if not all(col in data.columns for col in ['high', 'low', 'price']):
            raise ValueError("DataFrame must contain 'high', 'low', and 'price' columns")
        if not isinstance(period, int) or period <= 0:
            raise ValueError("period must be a positive integer")
            
        try:
            high = pd.to_numeric(data['high'], errors='coerce')
            low = pd.to_numeric(data['low'], errors='coerce')
            close = pd.to_numeric(data['price'], errors='coerce')
            
            if high.isna().any() or low.isna().any() or close.isna().any():
                logger.warning("NaN values detected in price data")
            
            highest_high = high.rolling(window=period).max()
            lowest_low = low.rolling(window=period).min()
            
            # Avoid division by zero
            denominator = highest_high - lowest_low
            denominator = denominator.replace(0, float('inf'))
            
            r = -100 * (highest_high - close) / denominator
            return pd.Series(r.clip(-100, 0), dtype=float)
        except Exception as e:
            logger.error(f"Error calculating Williams %R: {e}")
            raise
    
    def momentum(self, data: pd.DataFrame, period: int = 10) -> pd.Series:
        """Momentum"""
        if not isinstance(data, pd.DataFrame):
            raise TypeError("data must be a pandas DataFrame")
        if 'price' not in data.columns:
            raise ValueError("DataFrame must contain 'price' column")
        if not isinstance(period, int) or period <= 0:
            raise ValueError("period must be a positive integer")
            
        try:
            price = pd.to_numeric(data['price'], errors='coerce')
            
            if price.isna().any():
                logger.warning("NaN values detected in price data")
            
            momentum = price.diff(period)
            return pd.Series(momentum, dtype=float)
        except Exception as e:
            logger.error(f"Error calculating Momentum: {e}")
            raise
    
    def vwap(self, data: pd.DataFrame) -> pd.Series:
        """Volume Weighted Average Price using pandas"""
        if not isinstance(data, pd.DataFrame):
            raise TypeError("data must be a pandas DataFrame")
        if not all(col in data.columns for col in ['price', 'volume']):
            raise ValueError("DataFrame must contain 'price' and 'volume' columns")
        if len(data) == 0:
            raise ValueError("DataFrame cannot be empty")
            
        try:
            price = pd.to_numeric(data['price'], errors='coerce')
            volume = pd.to_numeric(data['volume'], errors='coerce')
            
            if price.isna().any() or volume.isna().any():
                logger.warning("NaN values detected in price or volume data")
                # Forward fill NaN values, then backward fill any remaining NaNs
                price = safe_fill_series(price, method='ffill')
                volume = safe_fill_series(volume, method='ffill', fill_value=0.0)
            
            # Handle negative volumes
            if (volume < 0).any():
                logger.warning("Negative volume values detected, converting to absolute values")
                volume = volume.abs()
            
            typical_price = price
            cumulative_pv = (typical_price * volume).cumsum()
            cumulative_volume = volume.cumsum()
            
            # Avoid division by zero while maintaining NaN propagation for invalid data
            vwap = pd.Series(index=data.index, dtype=float)
            mask = cumulative_volume > 0
            vwap[mask] = cumulative_pv[mask] / cumulative_volume[mask]
            vwap[~mask] = price[~mask]  # Use price when no volume data is available
            
            # Ensure VWAP stays within reasonable bounds
            price_std = price.std()
            if not pd.isna(price_std):
                mean_price = price.mean()
                lower_bound = mean_price - 3 * price_std
                upper_bound = mean_price + 3 * price_std
                vwap = vwap.clip(lower=lower_bound, upper=upper_bound)
            
            return pd.Series(vwap, dtype=float)
        except Exception as e:
            logger.error(f"Error calculating VWAP: {e}")
            raise
    
    def support_resistance(self, data: pd.DataFrame, period: int = 20,
                         threshold: float = 0.05) -> Dict[str, List[float]]:
        """Calculate support and resistance levels using pandas"""
        if not isinstance(data, pd.DataFrame):
            raise TypeError("data must be a pandas DataFrame")
        if not all(col in data.columns for col in ['high', 'low']):
            raise ValueError("DataFrame must contain 'high' and 'low' columns")
        if not isinstance(period, int) or period <= 0:
            raise ValueError("period must be a positive integer")
        if not isinstance(threshold, float) or threshold <= 0:
            raise ValueError("threshold must be a positive float")
            
        try:
            high = pd.to_numeric(data['high'], errors='coerce')
            low = pd.to_numeric(data['low'], errors='coerce')
            
            if high.isna().any() or low.isna().any():
                logger.warning("NaN values detected in price data")
            
            highs = high.rolling(window=period).max()
            lows = low.rolling(window=period).min()
            
            potential_support = lows[lows.shift(1) > lows].dropna()
            potential_resistance = highs[highs.shift(1) < highs].dropna()
            
            def group_levels(levels: pd.Series, threshold: float) -> List[float]:
                if len(levels) == 0:
                    return []
                
                try:
                    grouped = []
                    current_group = [float(levels.iloc[0])]
                    
                    for level in levels.iloc[1:]:
                        level_float = float(level)
                        group_mean = np.mean(current_group)
                        if abs(level_float - group_mean) / group_mean <= threshold:
                            current_group.append(level_float)
                        else:
                            grouped.append(float(np.mean(current_group)))
                            current_group = [level_float]
                    
                    if current_group:
                        grouped.append(float(np.mean(current_group)))
                    
                    return sorted(grouped)
                except Exception as e:
                    logger.error(f"Error in group_levels: {e}")
                    return []
            
            return {
                'support': group_levels(potential_support, threshold),
                'resistance': group_levels(potential_resistance, threshold)
            }
        except Exception as e:
            logger.error(f"Error calculating support/resistance levels: {e}")
            return {'support': [], 'resistance': []}
    
    def pivot_points(self, data: pd.DataFrame) -> Dict[str, float]:
        """Calculate pivot points (Floor Pivot Points) with robust error handling"""
        if not isinstance(data, pd.DataFrame):
            logger.error("Invalid input: data must be a pandas DataFrame")
            return {'pivot': 0.0, 'r1': 0.0, 'r2': 0.0, 'r3': 0.0,
                    's1': 0.0, 's2': 0.0, 's3': 0.0}
        if not all(col in data.columns for col in ['high', 'low', 'price']):
            logger.error("Missing required columns: high, low, price")
            return {'pivot': 0.0, 'r1': 0.0, 'r2': 0.0, 'r3': 0.0,
                    's1': 0.0, 's2': 0.0, 's3': 0.0}
        if len(data) < 1:
            logger.error("Empty DataFrame: must contain at least one row")
            return {'pivot': 0.0, 'r1': 0.0, 'r2': 0.0, 'r3': 0.0,
                    's1': 0.0, 's2': 0.0, 's3': 0.0}
            
        try:
            high = pd.to_numeric(data['high'].iloc[-1], errors='coerce')
            low = pd.to_numeric(data['low'].iloc[-1], errors='coerce')
            close = pd.to_numeric(data['price'].iloc[-1], errors='coerce')
            
            if any(pd.isna([high, low, close])):
                logger.warning("NaN values detected in price data")
                return {
                    'pivot': 0.0, 'r1': 0.0, 'r2': 0.0, 'r3': 0.0,
                    's1': 0.0, 's2': 0.0, 's3': 0.0
                }
            
            pivot = (high + low + close) / 3
            r1 = 2 * pivot - low
            r2 = pivot + (high - low)
            r3 = high + 2 * (pivot - low)
            s1 = 2 * pivot - high
            s2 = pivot - (high - low)
            s3 = low - 2 * (high - pivot)
            
            return {
                'pivot': float(pivot),
                'r1': float(r1),
                'r2': float(r2),
                'r3': float(r3),
                's1': float(s1),
                's2': float(s2),
                's3': float(s3)
            }
        except Exception as e:
            logger.error(f"Error calculating pivot points: {e}")
            return {
                'pivot': 0.0, 'r1': 0.0, 'r2': 0.0, 'r3': 0.0,
                's1': 0.0, 's2': 0.0, 's3': 0.0
            }
    
    @classmethod
    def calculate_trend_strength(cls, data: pd.DataFrame, period: int = 14) -> Dict[str, float]:
        """Calculate trend strength indicators with robust error handling and validation"""
        if not isinstance(data, pd.DataFrame):
            logger.error("Invalid input: data must be a pandas DataFrame")
            return {'adx': 0.0, 'aroon_up': 0.0, 'aroon_down': 0.0, 'cci': 0.0}
        if not all(col in data.columns for col in ['high', 'low', 'price']):
            logger.error("Missing required columns: high, low, price")
            return {'adx': 0.0, 'aroon_up': 0.0, 'aroon_down': 0.0, 'cci': 0.0}
        if not isinstance(period, int) or period <= 0:
            logger.error("Invalid period: must be a positive integer")
            return {'adx': 0.0, 'aroon_up': 0.0, 'aroon_down': 0.0, 'cci': 0.0}
        if len(data) < period:
            logger.warning(f"Insufficient data: {len(data)} rows < {period} period")
            return {'adx': 0.0, 'aroon_up': 0.0, 'aroon_down': 0.0, 'cci': 0.0}
            
        try:
            high = pd.to_numeric(data['high'], errors='coerce')
            low = pd.to_numeric(data['low'], errors='coerce')
            close = pd.to_numeric(data['price'], errors='coerce')
            
            if high.isna().any() or low.isna().any() or close.isna().any():
                logger.warning("NaN values detected in price data")
                # Forward fill NaN values, then backward fill any remaining NaNs
                high = safe_fill_series(high, method='ffill')
                low = safe_fill_series(low, method='ffill')
                close = safe_fill_series(close, method='ffill')
            
            # ADX calculation with improved type safety
            high_diff = pd.to_numeric(high.diff(), errors='coerce')
            low_diff = pd.to_numeric(low.diff(), errors='coerce')
            
            # Calculate positive and negative directional movement
            pos_dm_mask = (safe_compare_series(high_diff, 0, '>')) & (safe_compare_series(high_diff, -low_diff, '>'))
            neg_dm_mask = (safe_compare_series(low_diff, 0, '<')) & (safe_compare_series(-low_diff, high_diff, '>'))
            
            pos_dm = pd.Series(0.0, index=high_diff.index)
            neg_dm = pd.Series(0.0, index=low_diff.index)
            pos_dm[pos_dm_mask] = high_diff[pos_dm_mask]
            neg_dm[neg_dm_mask] = -low_diff[neg_dm_mask]
            
            # Calculate true range with type safety
            tr1 = pd.to_numeric(high - low, errors='coerce')
            tr2 = pd.to_numeric(abs(high - close.shift()), errors='coerce')
            tr3 = pd.to_numeric(abs(low - close.shift()), errors='coerce')
            tr = pd.concat([tr1, tr2, tr3], axis=1).max(axis=1)
            
            # Calculate smoothed values
            tr_ema = tr.ewm(span=period, adjust=False).mean()
            tr_ema = tr_ema.replace(0, float('inf'))
            
            pos_di = pd.to_numeric(100 * (pos_dm.ewm(span=period, adjust=False).mean() / tr_ema), errors='coerce')
            neg_di = pd.to_numeric(100 * (neg_dm.ewm(span=period, adjust=False).mean() / tr_ema), errors='coerce')
            
            di_sum = pd.to_numeric(pos_di + neg_di, errors='coerce')
            di_sum = di_sum.replace(0, float('inf'))
            
            dx = pd.to_numeric(100 * abs(pos_di - neg_di) / di_sum, errors='coerce')
            adx = dx.ewm(span=period, adjust=False).mean().clip(0, 100)
            
            # Aroon calculation with improved type safety
            def safe_rolling_func(x, func):
                try:
                    return pd.to_numeric((period - func(x)) / period * 100, errors='coerce')
                except Exception:
                    return 50.0  # Neutral value for error cases
                    
            high_period = pd.to_numeric(high.rolling(window=period, min_periods=1).apply(
                lambda x: safe_rolling_func(x, np.argmax)), errors='coerce')
            low_period = pd.to_numeric(low.rolling(window=period, min_periods=1).apply(
                lambda x: safe_rolling_func(x, np.argmin)), errors='coerce')
            
            # CCI calculation with improved type safety
            tp = pd.to_numeric((high + low + close) / 3, errors='coerce')
            tp_sma = pd.to_numeric(tp.rolling(window=period, min_periods=1).mean(), errors='coerce')
            mad = pd.to_numeric(tp.rolling(window=period, min_periods=1).apply(
                lambda x: abs(x - x.mean()).mean()), errors='coerce')
            mad = mad.replace(0, float('inf'))
            cci = pd.to_numeric((tp - tp_sma) / (0.015 * mad), errors='coerce')
            
            # Ensure all values are within valid ranges and handle NaN values
            adx = pd.to_numeric(adx, errors='coerce').fillna(50.0).clip(0, 100)
            high_period = high_period.fillna(50.0).clip(0, 100)
            low_period = low_period.fillna(50.0).clip(0, 100)
            cci = cci.fillna(0.0).clip(-100, 100)
            
            # Get latest values with improved NaN protection and bounds checking
            return {
                'adx': cls.safe_float(adx.iloc[-1] if not adx.empty else 50.0, 50.0),
                'aroon_up': cls.safe_float(high_period.iloc[-1] if not high_period.empty else 50.0, 50.0),
                'aroon_down': cls.safe_float(low_period.iloc[-1] if not low_period.empty else 50.0, 50.0),
                'cci': cls.safe_float(cci.iloc[-1] if not cci.empty else 0.0, 0.0)
            }
        except Exception as e:
            logger.error(f"Error calculating trend strength: {e}", exc_info=True)
            return {
                'adx': 50.0,  # Neutral ADX value
                'aroon_up': 50.0,  # Neutral Aroon value
                'aroon_down': 50.0,  # Neutral Aroon value
                'cci': 0.0  # Neutral CCI value
            }
    
    @classmethod
    def calculate_momentum_indicators(cls, data: pd.DataFrame, period: int = 14) -> Dict[str, float]:
        """Calculate momentum indicators with robust error handling and validation"""
        if not isinstance(data, pd.DataFrame):
            logger.error("Invalid input: data must be a pandas DataFrame")
            return {'roc': 0.0, 'momentum': 0.0, 'willr': -50.0}
        if not all(col in data.columns for col in ['price', 'high', 'low']):
            logger.error("Missing required columns: price, high, low")
            return {'roc': 0.0, 'momentum': 0.0, 'willr': -50.0}
        if not isinstance(period, int) or period <= 0:
            logger.error("Invalid period: must be a positive integer")
            return {'roc': 0.0, 'momentum': 0.0, 'willr': -50.0}
        if len(data) < period:
            logger.warning(f"Insufficient data: {len(data)} rows < {period} period")
            return {'roc': 0.0, 'momentum': 0.0, 'willr': -50.0}
            
        try:
            close = pd.to_numeric(data['price'], errors='coerce')
            high = pd.to_numeric(data['high'], errors='coerce')
            low = pd.to_numeric(data['low'], errors='coerce')
            
            if close.isna().any() or high.isna().any() or low.isna().any():
                logger.warning("NaN values detected in price data")
                close = safe_fill_series(close, method='ffill')
                high = safe_fill_series(high, method='ffill')
                low = safe_fill_series(low, method='ffill')
            
            # ROC calculation with improved type safety
            shifted_close = pd.to_numeric(close.shift(period), errors='coerce')
            roc = pd.Series(index=close.index, dtype=float)
            roc_mask = safe_compare_series(shifted_close, 0, '!=') & (~shifted_close.isna())
            roc[roc_mask] = ((close[roc_mask] - shifted_close[roc_mask]) / shifted_close[roc_mask]) * 100
            roc = safe_fill_series(roc, fill_value=0.0).clip(-100, 100)
            
            # Momentum calculation with type safety
            mom = pd.to_numeric(close - close.shift(period), errors='coerce')
            mom = safe_fill_series(mom, fill_value=0.0)
            price_std = cls.safe_float(close.std(), 1.0)
            mom = mom.clip(-3 * price_std, 3 * price_std)
            
            # Williams %R calculation with type safety
            highest_high = pd.to_numeric(high.rolling(window=period).max(), errors='coerce')
            lowest_low = pd.to_numeric(low.rolling(window=period).min(), errors='coerce')
            willr = pd.Series(index=close.index, dtype=float)
            
            denominator = pd.to_numeric(highest_high - lowest_low, errors='coerce')
            denom_mask = safe_compare_series(denominator, 0, '>')
            willr[denom_mask] = -100 * (highest_high[denom_mask] - close[denom_mask]) / denominator[denom_mask]
            willr[~denom_mask] = -50
            willr = willr.clip(-100, 0)
            
            return {
                'roc': cls.safe_float(roc),
                'momentum': cls.safe_float(mom),
                'willr': cls.safe_float(willr, -50.0)
            }
        except Exception as e:
            logger.error(f"Error calculating momentum indicators: {e}")
            return {'roc': 0.0, 'momentum': 0.0, 'willr': -50.0}
    
    @classmethod
    def calculate_volatility_indicators(cls, data: pd.DataFrame, window: int = 20) -> Dict[str, float]:
        """Calculate volatility indicators with robust error handling and validation"""
        if not isinstance(data, pd.DataFrame):
            logger.error("Invalid input: data must be a pandas DataFrame")
            return {'std': 0.0, 'hist_vol': 0.0, 'chaikin_vol': 1.0}
        if not all(col in data.columns for col in ['price', 'high', 'low']):
            logger.error("Missing required columns: price, high, low")
            return {'std': 0.0, 'hist_vol': 0.0, 'chaikin_vol': 1.0}
        if len(data) < window:
            logger.warning(f"Insufficient data: {len(data)} rows < {window} window")
            return {'std': 0.0, 'hist_vol': 0.0, 'chaikin_vol': 1.0}
            
        try:
            price = pd.to_numeric(data['price'], errors='coerce')
            high = pd.to_numeric(data['high'], errors='coerce')
            low = pd.to_numeric(data['low'], errors='coerce')
            
            if price.isna().any() or high.isna().any() or low.isna().any():
                logger.warning("NaN values detected in price data")
                price = safe_fill_series(price, method='ffill')
                high = safe_fill_series(high, method='ffill')
                low = safe_fill_series(low, method='ffill')
            
            price_mean = cls.safe_float(price.mean(), 0.0)
            price_std = cls.safe_float(price.std(), 0.0)
            if price_std > 0:
                price = price.clip(price_mean - 3 * price_std, price_mean + 3 * price_std)
            
            std = price.rolling(window=window, min_periods=1).std()
            
            returns = pd.Series(0.0, index=data.index)
            shifted_price = safe_fill_series(price.shift(1), fill_value=price.iloc[0] if not price.empty else 0.0)
            price_mask = safe_compare_series(shifted_price, 0, '>') & (~shifted_price.isna()) & (~price.isna())
            returns[price_mask] = np.log(price[price_mask] / shifted_price[price_mask])
            hist_vol = returns.rolling(window=window, min_periods=1).std() * np.sqrt(252)
            
            high_low = pd.to_numeric(high - low, errors='coerce')
            ema_hl = high_low.ewm(span=10, adjust=False).mean()
            shifted_ema = ema_hl.shift(10)
            
            chaikin_vol = pd.Series(1.0, index=data.index)
            ema_mask = safe_compare_series(shifted_ema, 0, '>')
            chaikin_vol[ema_mask] = ema_hl[ema_mask] / shifted_ema[ema_mask]
            
            std_val = cls.safe_float(std.iloc[-1], 0.0)
            hist_vol_val = cls.safe_float(hist_vol.iloc[-1], 0.0)
            chaikin_vol_val = cls.safe_float(chaikin_vol.iloc[-1], 1.0)
            price_upper = cls.safe_float(price.iloc[-1], 1.0)
            
            return {
                'std': min(max(std_val, 0.0), price_upper),
                'hist_vol': min(max(hist_vol_val, 0.0), 5.0),
                'chaikin_vol': min(max(chaikin_vol_val, 0.1), 10.0)
            }
        except Exception as e:
            logger.error(f"Error calculating volatility indicators: {e}")
            return {'std': 0.0, 'hist_vol': 0.0, 'chaikin_vol': 1.0}
    
    @classmethod
    def calculate_volume_indicators(cls, data: pd.DataFrame, period: int = 14) -> Dict[str, float]:
        """Calculate volume indicators with robust error handling and validation"""
        if not isinstance(data, pd.DataFrame):
            logger.error("Invalid input: data must be a pandas DataFrame")
            return {'ad': 0.0, 'mfi': 50.0, 'vroc': 0.0}
        if not all(col in data.columns for col in ['price', 'high', 'low', 'volume']):
            logger.error("Missing required columns: price, high, low, volume")
            return {'ad': 0.0, 'mfi': 50.0, 'vroc': 0.0}
        if len(data) < period:
            logger.warning(f"Insufficient data: {len(data)} rows < {period} period")
            return {'ad': 0.0, 'mfi': 50.0, 'vroc': 0.0}
            
        try:
            close = pd.to_numeric(data['price'], errors='coerce')
            high = pd.to_numeric(data['high'], errors='coerce')
            low = pd.to_numeric(data['low'], errors='coerce')
            volume = pd.to_numeric(data['volume'], errors='coerce')
            
            if close.isna().any() or high.isna().any() or low.isna().any() or volume.isna().any():
                logger.warning("NaN values detected in price or volume data")
                close = safe_fill_series(close, method='ffill')
                high = safe_fill_series(high, method='ffill')
                low = safe_fill_series(low, method='ffill')
                volume = safe_fill_series(volume, method='ffill', fill_value=0.0)
            
            volume = pd.to_numeric(volume.abs(), errors='coerce')
            
            high_low = pd.to_numeric(high - low, errors='coerce')
            clv = pd.Series(0.0, index=data.index)
            mask = safe_compare_series(high_low, 0, '>')
            
            # Calculate CLV with type safety
            numerator = pd.to_numeric((close[mask] - low[mask]) - (high[mask] - close[mask]), errors='coerce')
            denominator = high_low[mask].replace(0, float('inf'))
            clv[mask] = numerator / denominator
            
            # Calculate Accumulation/Distribution with type safety
            ad = pd.to_numeric((clv * volume).cumsum(), errors='coerce')
            
            # Calculate Typical Price with type safety
            tp = pd.to_numeric((high + low + close) / 3, errors='coerce')
            rmf = pd.to_numeric(tp * volume, errors='coerce')
            
            positive_flow = pd.Series(0.0, index=data.index)
            negative_flow = pd.Series(0.0, index=data.index)
            
            price_diff = pd.to_numeric(tp.diff(), errors='coerce')
            pos_mask = safe_compare_series(price_diff, 0, '>')
            neg_mask = safe_compare_series(price_diff, 0, '<')
            positive_flow[pos_mask] = rmf[pos_mask]
            negative_flow[neg_mask] = rmf[neg_mask]
            
            # Calculate Money Flow sums with type safety
            pos_mf = pd.to_numeric(positive_flow.rolling(window=period, min_periods=1).sum(), errors='coerce')
            neg_mf = pd.to_numeric(negative_flow.rolling(window=period, min_periods=1).sum(), errors='coerce')
            
            # Calculate Money Flow Index with type safety
            mfi = pd.Series(50.0, index=data.index)
            neg_mask = safe_compare_series(neg_mf, 0, '>')
            
            # Handle division by zero and type safety in MFI calculation
            denominator = neg_mf[neg_mask].replace(0, float('inf'))
            ratio = pd.to_numeric(pos_mf[neg_mask] / denominator, errors='coerce')
            mfi[neg_mask] = pd.to_numeric(100 - (100 / (1 + ratio)), errors='coerce')
            mfi = mfi.clip(0, 100)
            
            # Calculate Volume Rate of Change with type safety
            vroc = pd.Series(0.0, index=data.index)
            shifted_volume = pd.to_numeric(volume.shift(period), errors='coerce')
            vol_mask = safe_compare_series(shifted_volume, 0, '>')
            
            # Handle division by zero in VROC calculation
            denominator = shifted_volume[vol_mask].replace(0, float('inf'))
            vroc[vol_mask] = pd.to_numeric(
                ((volume[vol_mask] - shifted_volume[vol_mask]) / denominator) * 100,
                errors='coerce'
            )
            vroc = vroc.clip(-100, 100)
            
            return {
                'ad': cls.safe_float(ad.iloc[-1] if not ad.empty else 0.0),
                'mfi': cls.safe_float(mfi.iloc[-1] if not mfi.empty else 50.0),
                'vroc': cls.safe_float(vroc.iloc[-1] if not vroc.empty else 0.0)
            }
        except Exception as e:
            logger.error(f"Error calculating volume indicators: {e}")
            return {'ad': 0.0, 'mfi': 50.0, 'vroc': 0.0}
    
    @classmethod
    def identify_patterns(cls, data: pd.DataFrame) -> Dict[str, bool]:
        """Identify candlestick patterns with robust error handling"""
        try:
            if not isinstance(data, pd.DataFrame):
                raise TypeError("data must be a pandas DataFrame")
            if not all(col in data.columns for col in ['open', 'high', 'low', 'price']):
                raise ValueError("DataFrame must contain 'open', 'high', 'low', and 'price' columns")
            if len(data) < 3:
                raise ValueError("DataFrame must contain at least 3 rows for pattern identification")
            open_price = pd.to_numeric(data['open'], errors='coerce')
            high = pd.to_numeric(data['high'], errors='coerce')
            low = pd.to_numeric(data['low'], errors='coerce')
            close = pd.to_numeric(data['price'], errors='coerce')
            
            if open_price.isna().any() or high.isna().any() or low.isna().any() or close.isna().any():
                logger.warning("NaN values detected in price data")
                return {
                    'doji': False, 'engulfing': False, 'hammer': False,
                    'shooting_star': False, 'morning_star': False, 'evening_star': False
                }
            
            body = pd.to_numeric(close - open_price, errors='coerce')
            body_size = body.abs()
            
            price_max = pd.DataFrame([open_price, close]).max(axis=0)
            price_min = pd.DataFrame([open_price, close]).min(axis=0)
            upper_shadow = pd.to_numeric(high - price_max, errors='coerce')
            lower_shadow = pd.to_numeric(price_min - low, errors='coerce')
            
            avg_price = pd.to_numeric((high + low) / 2, errors='coerce')
            doji_threshold = 0.001
            
            patterns = {}
            
            # Doji pattern detection
            doji_size = cls.safe_float(body_size.iloc[-1], 0.0)
            doji_avg = cls.safe_float(avg_price.iloc[-1], 0.0)
            patterns['doji'] = bool(doji_size <= (doji_avg * doji_threshold)) if doji_avg > 0 else False
            
            # Engulfing pattern detection
            prev_close = safe_fill_series(close.shift(1), method='ffill')
            prev_open = safe_fill_series(open_price.shift(1), method='ffill')
            
            bull_engulf = safe_compare_series(body, 0, '>') & safe_compare_series(open_price, prev_close, '<') & safe_compare_series(close, prev_open, '>')
            bear_engulf = safe_compare_series(body, 0, '<') & safe_compare_series(open_price, prev_close, '>') & safe_compare_series(close, prev_open, '<')
            patterns['engulfing'] = bool((bull_engulf | bear_engulf).iloc[-1]) if not body.empty else False
            
            # Hammer pattern detection
            hammer_lower = safe_compare_series(lower_shadow, body_size * 2, '>')
            hammer_upper = safe_compare_series(upper_shadow, body_size, '<')
            patterns['hammer'] = bool((hammer_lower & hammer_upper).iloc[-1]) if not body_size.empty else False
            
            # Shooting star pattern detection
            star_upper = safe_compare_series(upper_shadow, body_size * 2, '>')
            star_lower = safe_compare_series(lower_shadow, body_size, '<')
            patterns['shooting_star'] = bool((star_upper & star_lower).iloc[-1]) if not body_size.empty else False
            
            # Morning Star pattern detection
            if len(body) >= 3:
                body_last_3 = pd.to_numeric(body.iloc[-3:], errors='coerce')
                body_size_last_3 = pd.to_numeric(body_size.iloc[-3:], errors='coerce')
                
                morning_first_down = safe_compare_series(body_last_3.iloc[0:1], 0, '<').iloc[0]
                morning_small_body = safe_compare_series(body_size_last_3.iloc[1:2], body_size_last_3.iloc[0] * 0.3, '<').iloc[0]
                morning_last_up = safe_compare_series(body_last_3.iloc[2:3], 0, '>').iloc[0]
                morning_close_above = safe_compare_series(close.iloc[-1:], (open_price.iloc[-1] + close.iloc[-1]) / 2, '>').iloc[0]
                
                patterns['morning_star'] = bool(
                    morning_first_down and morning_small_body and 
                    morning_last_up and morning_close_above
                )
            else:
                patterns['morning_star'] = False
            
            # Evening Star pattern detection
            if len(body) >= 3:
                body_last_3 = pd.to_numeric(body.iloc[-3:], errors='coerce')
                body_size_last_3 = pd.to_numeric(body_size.iloc[-3:], errors='coerce')
                
                evening_first_up = safe_compare_series(body_last_3.iloc[0:1], 0, '>').iloc[0]
                evening_small_body = safe_compare_series(body_size_last_3.iloc[1:2], body_size_last_3.iloc[0] * 0.3, '<').iloc[0]
                evening_last_down = safe_compare_series(body_last_3.iloc[2:3], 0, '<').iloc[0]
                evening_close_below = safe_compare_series(close.iloc[-1:], (open_price.iloc[-1] + close.iloc[-1]) / 2, '<').iloc[0]
                
                patterns['evening_star'] = bool(
                    evening_first_up and evening_small_body and 
                    evening_last_down and evening_close_below
                )
            else:
                patterns['evening_star'] = False
            
            return patterns
        except Exception as e:
            logger.error(f"Error identifying patterns: {e}")
            return {
                'doji': False, 'engulfing': False, 'hammer': False,
                'shooting_star': False, 'morning_star': False, 'evening_star': False
            }                                       