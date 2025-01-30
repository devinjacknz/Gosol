import numpy as np
import pandas as pd
from typing import Dict, List, Optional, Tuple
from dataclasses import dataclass
from datetime import datetime, timedelta
from technical_analysis import TechnicalAnalysis
from streaming_indicators import StreamingRSI, StreamingMACD, StreamingBB

@dataclass
class MarketState:
    trend: str  # 'uptrend', 'downtrend', 'sideways'
    strength: float  # 0 to 1
    volatility: float
    volume_profile: str  # 'increasing', 'decreasing', 'stable'
    support_levels: List[float]
    resistance_levels: List[float]
    key_levels: List[float]
    timestamp: datetime
    indicators: Dict[str, float]  # 存储实时指标值
    signals: Dict[str, str]  # 存储交易信号

class MarketAnalyzer:
    def __init__(self):
        self.ta = TechnicalAnalysis()
        self.last_state: Optional[MarketState] = None
        
        # 初始化流式指标
        self.streaming_rsi = StreamingRSI()
        self.streaming_macd = StreamingMACD()
        self.streaming_bb = StreamingBB()
        
        # 缓存最近的价格数据
        self.price_cache = []
        self.max_cache_size = 1000
    
    def update_price(self, price: float, volume: float, timestamp: datetime) -> None:
        """更新价格数据并维护缓存"""
        self.price_cache.append({
            'price': price,
            'volume': volume,
            'timestamp': timestamp
        })
        if len(self.price_cache) > self.max_cache_size:
            self.price_cache.pop(0)
    
    def analyze_market(self, data: pd.DataFrame, timeframe: str = '1h') -> MarketState:
        """使用多个指标和技术分析市场状态"""
        
        # 趋势分析
        trend, strength = self._analyze_trend(data)
        
        # 波动性分析
        volatility = self._analyze_volatility(data)
        
        # 成交量分析
        volume_profile = self._analyze_volume(data)
        
        # 支撑/阻力位
        support_resistance = self.ta.support_resistance(data)
        support_levels = support_resistance['support']
        resistance_levels = support_resistance['resistance']
        
        # 关键价格水平
        key_levels = self._identify_key_levels(data)
        
        # 计算实时指标
        indicators = self._calculate_real_time_indicators(data.iloc[-1])
        
        # 生成交易信号
        signals = self._generate_signals(data.iloc[-1], indicators)
        
        # 创建市场状态
        state = MarketState(
            trend=trend,
            strength=strength,
            volatility=volatility,
            volume_profile=volume_profile,
            support_levels=support_levels,
            resistance_levels=resistance_levels,
            key_levels=key_levels,
            timestamp=datetime.now(),
            indicators=indicators,
            signals=signals
        )
        
        self.last_state = state
        return state
    
    def _analyze_trend(self, data: pd.DataFrame) -> Tuple[str, float]:
        """分析市场趋势和强度"""
        
        # 使用流式指标计算趋势
        macd_val, signal_val = self.streaming_macd.add(data['price'].iloc[-1])
        rsi_val = self.streaming_rsi.add(data['price'].iloc[-1])
        bb_upper, bb_middle, bb_lower = self.streaming_bb.add(data['price'].iloc[-1])
        
        # 趋势判断
        trend = 'sideways'
        if macd_val > signal_val and data['price'].iloc[-1] > bb_middle:
            trend = 'uptrend'
        elif macd_val < signal_val and data['price'].iloc[-1] < bb_middle:
            trend = 'downtrend'
        
        # 趋势强度计算
        strength = min(abs(macd_val - signal_val) / abs(signal_val), 1.0) if signal_val != 0 else 0.5
        
        return trend, strength
    
    def _analyze_volatility(self, data: pd.DataFrame) -> float:
        """分析市场波动性"""
        # 使用布林带宽度作为波动性指标
        _, bb_middle, _ = self.streaming_bb.add(data['price'].iloc[-1])
        bb_width = (bb_upper - bb_lower) / bb_middle if bb_middle != 0 else 0
        
        # 标准化波动性到0-1范围
        return min(bb_width, 1.0)
    
    def _analyze_volume(self, data: pd.DataFrame) -> str:
        """分析成交量趋势"""
        recent_volume = data['volume'].tail(20)
        avg_volume = recent_volume.mean()
        current_volume = recent_volume.iloc[-1]
        
        if current_volume > avg_volume * 1.2:
            return 'increasing'
        elif current_volume < avg_volume * 0.8:
            return 'decreasing'
        else:
            return 'stable'
    
    def _identify_key_levels(self, data: pd.DataFrame) -> List[float]:
        """识别关键价格水平"""
        pivot_points = self.ta.pivot_points(data)
        return [
            pivot_points['pivot'],
            pivot_points['r1'],
            pivot_points['r2'],
            pivot_points['s1'],
            pivot_points['s2']
        ]
    
    def _calculate_real_time_indicators(self, latest_data: pd.Series) -> Dict[str, float]:
        """计算实时技术指标"""
        return {
            'rsi': self.streaming_rsi.get_value(),
            'macd': self.streaming_macd.get_value()[0],
            'macd_signal': self.streaming_macd.get_value()[1],
            'bb_upper': self.streaming_bb.get_value()[0],
            'bb_middle': self.streaming_bb.get_value()[1],
            'bb_lower': self.streaming_bb.get_value()[2]
        }
    
    def _generate_signals(self, latest_data: pd.Series, indicators: Dict[str, float]) -> Dict[str, str]:
        """生成交易信号"""
        signals = {}
        
        # RSI信号
        rsi = indicators['rsi']
        if rsi > 70:
            signals['rsi'] = 'overbought'
        elif rsi < 30:
            signals['rsi'] = 'oversold'
        else:
            signals['rsi'] = 'neutral'
        
        # MACD信号
        macd = indicators['macd']
        macd_signal = indicators['macd_signal']
        if macd > macd_signal:
            signals['macd'] = 'bullish'
        else:
            signals['macd'] = 'bearish'
        
        # 布林带信号
        price = latest_data['price']
        bb_upper = indicators['bb_upper']
        bb_lower = indicators['bb_lower']
        if price > bb_upper:
            signals['bollinger'] = 'overbought'
        elif price < bb_lower:
            signals['bollinger'] = 'oversold'
        else:
            signals['bollinger'] = 'neutral'
        
        return signals
    
    def get_market_summary(self) -> Dict[str, any]:
        """获取市场概况"""
        if not self.last_state:
            return {}
        
        return {
            'trend': self.last_state.trend,
            'strength': self.last_state.strength,
            'volatility': self.last_state.volatility,
            'volume_profile': self.last_state.volume_profile,
            'indicators': self.last_state.indicators,
            'signals': self.last_state.signals,
            'timestamp': self.last_state.timestamp
        }
    
    def get_market_context(self, data: pd.DataFrame) -> Dict:
        """Get comprehensive market context"""
        
        # Analyze current market state
        current_state = self.analyze_market(data)
        
        # Calculate additional metrics
        rsi = self.ta.rsi(data)
        bb = self.ta.bollinger_bands(data)
        mfi = self.ta.mfi(data)
        
        context = {
            'state': {
                'trend': current_state.trend,
                'strength': current_state.strength,
                'volatility': current_state.volatility,
                'volume_profile': current_state.volume_profile
            },
            'indicators': {
                'rsi': rsi.iloc[-1],
                'bollinger_width': (bb['upper'].iloc[-1] - bb['lower'].iloc[-1]) / bb['middle'].iloc[-1],
                'mfi': mfi.iloc[-1]
            },
            'levels': {
                'support': current_state.support_levels,
                'resistance': current_state.resistance_levels,
                'key_levels': current_state.key_levels
            },
            'risk_metrics': self._calculate_risk_metrics(data)
        }
        
        return context
    
    def _calculate_risk_metrics(self, data: pd.DataFrame) -> Dict:
        """Calculate various risk metrics"""
        
        returns = data['price'].pct_change().dropna()
        
        metrics = {
            'volatility': returns.std() * np.sqrt(252),  # Annualized volatility
            'var_95': np.percentile(returns, 5),  # 95% Value at Risk
            'max_drawdown': self._calculate_max_drawdown(data['price']),
            'sharpe_ratio': self._calculate_sharpe_ratio(returns),
            'sortino_ratio': self._calculate_sortino_ratio(returns)
        }
        
        return metrics
    
    def _calculate_max_drawdown(self, prices: pd.Series) -> float:
        """Calculate maximum drawdown"""
        peak = prices.expanding(min_periods=1).max()
        drawdown = (prices - peak) / peak
        return abs(drawdown.min())
    
    def _calculate_sharpe_ratio(self, returns: pd.Series, risk_free_rate: float = 0.02) -> float:
        """Calculate Sharpe ratio"""
        excess_returns = returns - risk_free_rate/252
        if excess_returns.std() == 0:
            return 0
        return np.sqrt(252) * excess_returns.mean() / excess_returns.std()
    
    def _calculate_sortino_ratio(self, returns: pd.Series, risk_free_rate: float = 0.02) -> float:
        """Calculate Sortino ratio"""
        excess_returns = returns - risk_free_rate/252
        downside_returns = excess_returns[excess_returns < 0]
        if len(downside_returns) == 0 or downside_returns.std() == 0:
            return 0
        return np.sqrt(252) * excess_returns.mean() / downside_returns.std()
    
    def detect_market_regime(self, data: pd.DataFrame) -> str:
        """Detect current market regime"""
        
        # Calculate indicators
        volatility = self._analyze_volatility(data)
        trend, strength = self._analyze_trend(data)
        volume_profile = self._analyze_volume(data)
        
        # Define regimes
        if trend == 'uptrend' and strength > 0.7:
            if volatility < 0.3:
                return 'strong_uptrend'
            else:
                return 'volatile_uptrend'
        elif trend == 'downtrend' and strength > 0.7:
            if volatility < 0.3:
                return 'strong_downtrend'
            else:
                return 'volatile_downtrend'
        elif volatility > 0.7:
            return 'high_volatility'
        elif volume_profile == 'decreasing' and volatility < 0.3:
            return 'low_volatility_consolidation'
        else:
            return 'ranging'
    
    def get_trading_opportunities(self, data: pd.DataFrame) -> List[Dict]:
        """Identify potential trading opportunities"""
        
        opportunities = []
        current_price = data['price'].iloc[-1]
        
        # Get market context
        context = self.get_market_context(data)
        regime = self.detect_market_regime(data)
        
        # Analyze price action near key levels
        for level in context['levels']['key_levels']:
            distance = abs(current_price - level) / current_price
            
            if distance < 0.01:  # Within 1% of key level
                opportunity = {
                    'type': 'key_level_test',
                    'price': current_price,
                    'level': level,
                    'regime': regime,
                    'context': context['state'],
                    'confidence': self._calculate_opportunity_confidence(data, level)
                }
                opportunities.append(opportunity)
        
        # Analyze indicator signals
        signals = self._analyze_indicator_signals(data)
        for signal in signals:
            opportunity = {
                'type': 'indicator_signal',
                'price': current_price,
                'signal': signal,
                'regime': regime,
                'context': context['state'],
                'confidence': signal['confidence']
            }
            opportunities.append(opportunity)
        
        return opportunities
    
    def _analyze_indicator_signals(self, data: pd.DataFrame) -> List[Dict]:
        """Analyze various indicators for trading signals"""
        
        signals = []
        
        # RSI signals
        rsi = self.ta.rsi(data)
        if rsi.iloc[-1] < 30:
            signals.append({
                'indicator': 'RSI',
                'type': 'oversold',
                'value': rsi.iloc[-1],
                'confidence': 0.7
            })
        elif rsi.iloc[-1] > 70:
            signals.append({
                'indicator': 'RSI',
                'type': 'overbought',
                'value': rsi.iloc[-1],
                'confidence': 0.7
            })
        
        # MACD signals
        macd = self.ta.macd(data)
        if (macd['histogram'].iloc[-2] < 0 and macd['histogram'].iloc[-1] > 0):
            signals.append({
                'indicator': 'MACD',
                'type': 'bullish_cross',
                'value': macd['histogram'].iloc[-1],
                'confidence': 0.6
            })
        elif (macd['histogram'].iloc[-2] > 0 and macd['histogram'].iloc[-1] < 0):
            signals.append({
                'indicator': 'MACD',
                'type': 'bearish_cross',
                'value': macd['histogram'].iloc[-1],
                'confidence': 0.6
            })
        
        return signals
    
    def _calculate_opportunity_confidence(self, data: pd.DataFrame, level: float) -> float:
        """Calculate confidence score for a trading opportunity"""
        
        # Base confidence on multiple factors
        confidence = 0.5  # Start with neutral confidence
        
        # Factor 1: Historical significance of the level
        touches = self._count_level_touches(data, level)
        confidence += min(touches * 0.1, 0.3)  # Max 0.3 from historical significance
        
        # Factor 2: Current market conditions
        regime = self.detect_market_regime(data)
        if regime in ['strong_uptrend', 'strong_downtrend']:
            confidence += 0.1
        elif regime == 'high_volatility':
            confidence -= 0.2
        
        # Factor 3: Volume confirmation
        volume_profile = self._analyze_volume(data)
        if volume_profile == 'increasing':
            confidence += 0.1
        elif volume_profile == 'decreasing':
            confidence -= 0.1
        
        return min(max(confidence, 0.0), 1.0)
    
    def _count_level_touches(self, data: pd.DataFrame, level: float) -> int:
        """Count how many times price has touched a level"""
        threshold = level * 0.001  # 0.1% threshold
        touches = 0
        
        for price in data['price']:
            if abs(price - level) <= threshold:
                touches += 1
        
        return touches
