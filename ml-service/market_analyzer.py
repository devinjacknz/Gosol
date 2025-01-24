import numpy as np
import pandas as pd
import pandas_ta as ta
from streaming_indicators import StreamingRSI, StreamingMACD, StreamingBB
from typing import List, Dict, Tuple

class MarketAnalyzer:
    def __init__(self):
        self.rsi = StreamingRSI(period=14)
        self.macd = StreamingMACD(fast=12, slow=26, signal=9)
        self.bb = StreamingBB(period=20, std_dev=2)
        
    def calculate_indicators(self, price_data: List[float], volume_data: List[float]) -> Dict:
        """Calculate technical indicators using streaming-indicators and pandas-ta"""
        df = pd.DataFrame({
            'price': price_data,
            'volume': volume_data
        })
        
        # Streaming indicators
        rsi_values = [self.rsi.add(price) for price in price_data]
        macd_values = [self.macd.add(price) for price in price_data]
        bb_values = [self.bb.add(price) for price in price_data]
        
        # Pandas-ta indicators
        df.ta.ema(length=20, append=True)
        df.ta.vwap(append=True)
        df.ta.volatility(append=True)
        
        latest_data = {
            'rsi': rsi_values[-1] if rsi_values[-1] is not None else 50,
            'macd': macd_values[-1].macd if macd_values[-1] is not None else 0,
            'macd_signal': macd_values[-1].signal if macd_values[-1] is not None else 0,
            'bb_upper': bb_values[-1].upper if bb_values[-1] is not None else price_data[-1],
            'bb_lower': bb_values[-1].lower if bb_values[-1] is not None else price_data[-1],
            'ema': df['EMA_20'].iloc[-1] if not pd.isna(df['EMA_20'].iloc[-1]) else price_data[-1],
            'vwap': df['VWAP'].iloc[-1] if not pd.isna(df['VWAP'].iloc[-1]) else price_data[-1],
            'volatility': df['VOLATILITY_20'].iloc[-1] if not pd.isna(df['VOLATILITY_20'].iloc[-1]) else 0
        }
        
        return latest_data
    
    def analyze_market_conditions(self, indicators: Dict, 
                                price_data: List[float], 
                                volume_data: List[float]) -> Dict:
        """Analyze market conditions based on technical indicators"""
        current_price = price_data[-1]
        avg_volume = np.mean(volume_data[-24:]) if len(volume_data) >= 24 else volume_data[-1]
        
        # Price trend analysis
        price_trend = self._calculate_trend(price_data[-20:])
        volume_trend = self._calculate_trend(volume_data[-20:])
        
        # Volatility analysis
        volatility = indicators['volatility']
        volatility_risk = self._assess_volatility_risk(volatility)
        
        # Support and resistance
        support, resistance = self._calculate_support_resistance(price_data)
        
        # Market strength indicators
        rsi_signal = self._interpret_rsi(indicators['rsi'])
        macd_signal = self._interpret_macd(indicators['macd'], indicators['macd_signal'])
        bb_signal = self._interpret_bollinger_bands(current_price, 
                                                  indicators['bb_upper'], 
                                                  indicators['bb_lower'])
        
        return {
            'trend': price_trend,
            'volume_trend': volume_trend,
            'volatility': volatility,
            'volatility_risk': volatility_risk,
            'support': support,
            'resistance': resistance,
            'rsi_signal': rsi_signal,
            'macd_signal': macd_signal,
            'bb_signal': bb_signal,
            'overall_sentiment': self._calculate_overall_sentiment(
                price_trend, volume_trend, rsi_signal, macd_signal, bb_signal
            )
        }
    
    def _calculate_trend(self, data: List[float]) -> float:
        """Calculate trend strength and direction"""
        if len(data) < 2:
            return 0
        return (data[-1] / data[0] - 1) * 100
    
    def _assess_volatility_risk(self, volatility: float) -> str:
        """Assess risk level based on volatility"""
        if volatility > 0.1:
            return "high"
        elif volatility > 0.05:
            return "medium"
        return "low"
    
    def _calculate_support_resistance(self, price_data: List[float]) -> Tuple[float, float]:
        """Calculate support and resistance levels"""
        if len(price_data) < 20:
            return price_data[-1] * 0.95, price_data[-1] * 1.05
        
        window = price_data[-20:]
        support = min(window)
        resistance = max(window)
        return support, resistance
    
    def _interpret_rsi(self, rsi: float) -> str:
        """Interpret RSI signal"""
        if rsi > 70:
            return "overbought"
        elif rsi < 30:
            return "oversold"
        return "neutral"
    
    def _interpret_macd(self, macd: float, signal: float) -> str:
        """Interpret MACD signal"""
        if macd > signal:
            return "bullish"
        elif macd < signal:
            return "bearish"
        return "neutral"
    
    def _interpret_bollinger_bands(self, price: float, upper: float, lower: float) -> str:
        """Interpret Bollinger Bands signal"""
        if price > upper:
            return "overbought"
        elif price < lower:
            return "oversold"
        return "neutral"
    
    def _calculate_overall_sentiment(self, price_trend: float, volume_trend: float,
                                   rsi_signal: str, macd_signal: str, bb_signal: str) -> str:
        """Calculate overall market sentiment"""
        signals = {
            'bullish': 0,
            'bearish': 0,
            'neutral': 0
        }
        
        # Price trend
        if price_trend > 1:
            signals['bullish'] += 1
        elif price_trend < -1:
            signals['bearish'] += 1
        else:
            signals['neutral'] += 1
            
        # Volume trend
        if volume_trend > 1:
            signals['bullish'] += 1
        elif volume_trend < -1:
            signals['bearish'] += 1
        else:
            signals['neutral'] += 1
            
        # RSI
        if rsi_signal == "oversold":
            signals['bullish'] += 1
        elif rsi_signal == "overbought":
            signals['bearish'] += 1
        else:
            signals['neutral'] += 1
            
        # MACD
        if macd_signal == "bullish":
            signals['bullish'] += 1
        elif macd_signal == "bearish":
            signals['bearish'] += 1
        else:
            signals['neutral'] += 1
            
        # Bollinger Bands
        if bb_signal == "oversold":
            signals['bullish'] += 1
        elif bb_signal == "overbought":
            signals['bearish'] += 1
        else:
            signals['neutral'] += 1
            
        # Determine overall sentiment
        max_signal = max(signals.items(), key=lambda x: x[1])
        return max_signal[0]
