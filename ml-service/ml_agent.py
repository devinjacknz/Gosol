import asyncio
import numpy as np
import pandas as pd
from typing import Dict, Optional
from datetime import datetime
from agent_system import BaseAgent, TradeSignal, AgentConfig
import logging

logger = logging.getLogger(__name__)

class DeepSeekAgent(BaseAgent):
    """使用DeepSeek模型的交易Agent"""
    
    def __init__(self, config: AgentConfig):
        super().__init__(config)
        from deepseek_client import DeepseekClient
        from ollama_client import OllamaClient
        
        # Initialize clients
        try:
            self.deepseek_client = DeepseekClient()
            self.ollama_client = OllamaClient()
            logger.info("AI clients initialized successfully")
        except Exception as e:
            logger.error(f"Error initializing AI clients: {str(e)}")
            raise
    
    def _prepare_market_data(self, data: pd.DataFrame) -> str:
        """准备市场数据文本"""
        # 计算技术指标
        data['sma20'] = data['close'].rolling(window=20).mean()
        data['sma50'] = data['close'].rolling(window=50).mean()
        data['rsi'] = self._calculate_rsi(data['close'])
        data['atr'] = self._calculate_atr(data)
        data['volume_sma'] = data['volume'].rolling(window=20).mean()
        
        # 获取最近的数据点
        recent_data = data.tail(5)
        trend_data = data.tail(20)
        
        # 计算市场趋势
        trend = "上升" if trend_data['close'].iloc[-1] > trend_data['close'].iloc[0] else "下降"
        volatility = data['close'].pct_change().std() * np.sqrt(252)
        volume_trend = "放大" if recent_data['volume'].mean() > data['volume_sma'].iloc[-1] else "减小"
        
        # 构建提示文本
        prompt = f"""分析以下{self.config.symbol}市场数据并提供交易建议：

市场概况：
- 当前趋势：{trend}
- 波动率：{volatility:.2%}
- 成交量趋势：{volume_trend}
- RSI：{data['rsi'].iloc[-1]:.2f}
- 20日均线：{data['sma20'].iloc[-1]:.2f}
- 50日均线：{data['sma50'].iloc[-1]:.2f}

最近5根K线数据：
"""
        
        for _, row in recent_data.iterrows():
            prompt += f"""
时间：{row.name}
开盘：{row['open']:.2f}
最高：{row['high']:.2f}
最低：{row['low']:.2f}
收盘：{row['close']:.2f}
成交量：{row['volume']:.2f}
"""
        
        prompt += """
请基于以下因素进行分析：
1. 价格趋势和动量
2. 支撑和阻力位
3. 成交量变化
4. 技术指标信号
5. 市场波动性
6. 潜在风险

请按以下格式提供分析结果：
DECISION: [BUY/SELL/HOLD] - 交易决策
CONFIDENCE: [0.0-1.0] - 信心水平
REASON: [详细分析原因]
RISK_LEVEL: [LOW/MEDIUM/HIGH] - 风险等级
STOP_LOSS: [建议止损价格]
TAKE_PROFIT: [建议止盈价格]
"""
        
        return prompt
    
    def _parse_model_output(self, output: str) -> Dict:
        """解析模型输出"""
        try:
            lines = output.strip().split('\n')
            result = {
                'decision': None,
                'confidence': 0.0,
                'reason': '',
                'risk_level': 'HIGH',
                'stop_loss': None,
                'take_profit': None
            }
            
            for line in lines:
                if line.startswith('DECISION:'):
                    result['decision'] = line.split(':')[1].strip()
                elif line.startswith('CONFIDENCE:'):
                    result['confidence'] = float(line.split(':')[1].strip())
                elif line.startswith('REASON:'):
                    result['reason'] = line.split(':')[1].strip()
                elif line.startswith('RISK_LEVEL:'):
                    result['risk_level'] = line.split(':')[1].strip()
                elif line.startswith('STOP_LOSS:'):
                    try:
                        result['stop_loss'] = float(line.split(':')[1].strip())
                    except:
                        pass
                elif line.startswith('TAKE_PROFIT:'):
                    try:
                        result['take_profit'] = float(line.split(':')[1].strip())
                    except:
                        pass
            
            return result
        except Exception as e:
            logger.error(f"Error parsing model output: {str(e)}")
            return {
                'decision': None,
                'confidence': 0.0,
                'reason': f'Error parsing output: {str(e)}',
                'risk_level': 'HIGH',
                'stop_loss': None,
                'take_profit': None
            }
    
    async def analyze(self, data: pd.DataFrame) -> Optional[TradeSignal]:
        """分析市场数据并生成交易信号"""
        if len(data) < 50:
            return None
        
        try:
            # 准备输入数据
            prompt = self._prepare_market_data(data)
            
            # Prepare market data for analysis
            market_data = {
                'prompt': prompt,
                'current_price': data['close'].iloc[-1],
                'price_change_24h': data['close'].pct_change(24).iloc[-1] * 100,
                'price_change_7d': data['close'].pct_change(7).iloc[-1] * 100,
                'volume_24h': data['volume'].iloc[-1],
                'volume_change': (data['volume'].iloc[-1] / data['volume_sma'].iloc[-1] - 1) * 100,
                'market_cap': data['close'].iloc[-1] * data['volume'].iloc[-1],
                'holders': 5000
            }
            
            # Try Ollama first, then fallback to DeepSeek
            try:
                if await self.ollama_client.is_available():
                    logger.info("Using local Ollama model for analysis...")
                    prediction = await self.ollama_client.analyze_market_sentiment(market_data)
                else:
                    logger.warning("Ollama not available, falling back to DeepSeek API...")
                    prediction = await self.deepseek_client.analyze_market_sentiment(market_data)
            except Exception as e:
                logger.error(f"Model analysis failed: {str(e)}")
                return None
            
            if not prediction:
                return None
            
            # 计算技术指标
            data['atr'] = self._calculate_atr(data)
            current_price = data['close'].iloc[-1]
            atr = data['atr'].iloc[-1]
            
            # 根据预测生成信号
            if prediction and prediction.get('sentiment') in ['bullish', 'bearish'] and \
               prediction.get('confidence', 0) >= self.config.confidence_threshold:
                
                direction = 'buy' if prediction['sentiment'] == 'bullish' else 'sell'
                
                # 计算止损止盈
                if direction == 'buy':
                    stop_loss = current_price - (atr * self.config.parameters['stop_loss_atr'])
                    take_profit = current_price + (atr * self.config.parameters['take_profit_atr'])
                else:
                    stop_loss = current_price + (atr * self.config.parameters['stop_loss_atr'])
                    take_profit = current_price - (atr * self.config.parameters['take_profit_atr'])
                
                # 创建交易信号
                signal = TradeSignal(
                    symbol=self.config.symbol,
                    direction=direction,
                    price=current_price,
                    stop_loss=stop_loss,
                    take_profit=take_profit,
                    size=self._calculate_position_size(current_price, stop_loss),
                    confidence=prediction['confidence'],
                    agent_name=self.config.name,
                    timestamp=data.index[-1],
                    metadata={
                        'strategy': 'deepseek_ml',
                        'reason': prediction['reason'],
                        'atr': atr
                    }
                )
                
                self.signals.append(signal)
                return signal
            
            return None
            
        except Exception as e:
            logger.error(f"Error in DeepSeek analysis: {str(e)}")
            return None
    
    def _calculate_atr(self, data: pd.DataFrame, period: int = 14) -> pd.Series:
        """计算ATR指标"""
        high = data['high']
        low = data['low']
        close = data['close']
        
        tr1 = high - low
        tr2 = abs(high - close.shift())
        tr3 = abs(low - close.shift())
        
        tr = pd.concat([tr1, tr2, tr3], axis=1).max(axis=1)
        atr = tr.rolling(window=period).mean()
        
        return atr
    
    def _calculate_rsi(self, prices: pd.Series, period: int = 14) -> pd.Series:
        """计算RSI指标"""
        delta = prices.astype(float).diff()
        gain = (delta.where(delta > 0, 0.0)).rolling(window=period).mean()
        loss = (-delta.where(delta < 0, 0.0)).rolling(window=period).mean()
        rs = gain / loss.replace(0, float('inf'))  # Avoid division by zero
        return 100 - (100 / (1 + rs))          