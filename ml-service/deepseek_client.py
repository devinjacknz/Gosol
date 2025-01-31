import os
import aiohttp
import asyncio
import logging
from typing import Dict, List, Any, Optional
import json
from ollama_client import OllamaClient

logger = logging.getLogger(__name__)

class DeepseekClient:
    def __init__(self, api_key: Optional[str] = None):
        self.api_key = api_key or os.getenv('DEEPSEEK_API_KEY')
        self.base_url = "https://api.deepseek.com/v1/chat"
        self.timeout = aiohttp.ClientTimeout(total=120)
        try:
            self.ollama_client = OllamaClient()
        except Exception as e:
            logger.warning(f"Failed to initialize Ollama client: {e}")
            self.ollama_client = None
        
    async def analyze_market_sentiment(self, token_data: Dict) -> Dict[str, Any]:
        """Analyze market sentiment using Deepseek's API"""
        if not self.api_key:
            logger.warning("No DeepSeek API key found")
            raise ValueError("DeepSeek API key not configured")
            
        headers = {
            "Authorization": f"Bearer {self.api_key}",
            "Content-Type": "application/json"
        }
        
        # Prepare market data for analysis
        prompt = self._create_analysis_prompt(token_data)
        
        try:
            async with aiohttp.ClientSession(timeout=self.timeout) as session:
                async with session.post(
                    f"{self.base_url}/completions",
                    headers=headers,
                    json={
                        "model": "deepseek-coder-33b-instruct",
                        "messages": [
                            {"role": "system", "content": "You are a cryptocurrency market analyst specializing in Solana meme coins. You analyze market data and provide structured JSON responses."},
                            {"role": "user", "content": prompt}
                        ],
                        "temperature": 0.3,
                        "max_tokens": 1000,
                        "top_p": 0.8,
                        "stream": False,
                        "stop": ["</s>"]
                    }
                ) as response:
                    if response.status != 200:
                        raise Exception(f"Deepseek API error: {await response.text()}")
                    
                    result = await response.json()
                    return self._parse_analysis_response(result)
        except Exception as api_error:
            print(f"DeepSeek API error: {api_error}. Attempting fallback to Ollama...")
            if self.ollama_client and await self.ollama_client.is_available():
                try:
                    result = await asyncio.wait_for(
                        self.ollama_client.analyze_market_sentiment(token_data),
                        timeout=120
                    )
                    logger.info("Successfully used Ollama fallback for market analysis")
                    return result
                except asyncio.TimeoutError:
                    logger.error("Ollama analysis timed out")
                    raise api_error
                except Exception as ollama_error:
                    logger.error(f"Ollama fallback failed: {ollama_error}")
                    raise api_error
            logger.warning("Ollama not available for fallback")
            raise api_error
    
    def _create_analysis_prompt(self, token_data: Dict) -> str:
        """Create a prompt for market analysis"""
        return f"""
        Please analyze the following market data for a Solana meme coin:
        
        Price Metrics:
        - Current Price: {token_data.get('current_price')}
        - 24h Change: {token_data.get('price_change_24h')}%
        - 7d Change: {token_data.get('price_change_7d')}%
        
        Volume Metrics:
        - 24h Volume: {token_data.get('volume_24h')}
        - Volume Change: {token_data.get('volume_change')}%
        
        Market Metrics:
        - Market Cap: {token_data.get('market_cap')}
        - Holders: {token_data.get('holders')}
        
        Please analyze this data and provide a response in the following JSON format:
        {{
            "market_sentiment": "bullish",
            "risk_level": 5,
            "short_term_prediction": {{
                "target_price": "1.25",
                "timeframe": "24h",
                "key_levels": {{
                    "support": "1.20",
                    "resistance": "1.30"
                }}
            }},
            "key_factors": [
                "Strong volume increase",
                "Positive price momentum",
                "Growing holder base"
            ],
            "trading_recommendation": "BUY",
            "confidence": 0.75,
            "risk_analysis": {{
                "market_manipulation_risk": "medium",
                "liquidity_risk": "low",
                "volatility_risk": "high"
            }}
        }}

        Ensure your response is a valid JSON object with all fields properly formatted.
        """
    
    async def check_health(self) -> bool:
        """Check if Deepseek API is healthy and accessible"""
        if not self.api_key:
            return False
            
        try:
            headers = {
                "Authorization": f"Bearer {self.api_key}",
                "Content-Type": "application/json"
            }
            
            async with aiohttp.ClientSession(timeout=self.timeout) as session:
                async with session.post(
                    f"{self.base_url}/completions",
                    headers=headers,
                    json={
                        "model": "deepseek-coder-33b-instruct",
                        "messages": [{"role": "system", "content": "health check"}],
                        "max_tokens": 1
                    }
                ) as response:
                    return response.status == 200
        except Exception as e:
            logger.error(f"Deepseek health check failed: {e}")
            return False
            
    def _parse_analysis_response(self, response: Dict) -> Dict:
        """Parse the Deepseek API response into structured data"""
        try:
            content = response['choices'][0]['message']['content']
            content = content.strip()
            
            # Handle potential JSON within markdown code blocks
            if '```json' in content:
                content = content.split('```json')[1].split('```')[0].strip()
            elif '```' in content:
                content = content.split('```')[1].split('```')[0].strip()
                
            analysis = json.loads(content)
            
            # Validate numeric fields
            risk_level = analysis.get('risk_level')
            if not isinstance(risk_level, (int, float)) or risk_level < 0 or risk_level > 10:
                risk_level = 5.0
                
            confidence = analysis.get('confidence')
            if not isinstance(confidence, (int, float)) or confidence < 0 or confidence > 1:
                confidence = 0.5
                
            return {
                'sentiment': analysis.get('market_sentiment', 'neutral'),
                'risk_level': float(risk_level),
                'price_prediction': {
                    'target': analysis.get('short_term_prediction', {}).get('target_price'),
                    'timeframe': analysis.get('short_term_prediction', {}).get('timeframe'),
                    'support': analysis.get('short_term_prediction', {}).get('key_levels', {}).get('support'),
                    'resistance': analysis.get('short_term_prediction', {}).get('key_levels', {}).get('resistance')
                },
                'key_factors': analysis.get('key_factors', []),
                'recommendation': analysis.get('trading_recommendation', 'HOLD'),
                'confidence': float(confidence),
                'risk_analysis': {
                    'manipulation_risk': analysis.get('risk_analysis', {}).get('market_manipulation_risk', 'medium'),
                    'liquidity_risk': analysis.get('risk_analysis', {}).get('liquidity_risk', 'medium'),
                    'volatility_risk': analysis.get('risk_analysis', {}).get('volatility_risk', 'medium')
                }
            }
        except (KeyError, json.JSONDecodeError, ValueError) as e:
            logger.error(f"Failed to parse analysis response: {e}")
            return {
                'sentiment': 'neutral',
                'risk_level': 5.0,
                'price_prediction': {
                    'target': None,
                    'timeframe': None,
                    'support': None,
                    'resistance': None
                },
                'key_factors': [],
                'recommendation': 'HOLD',
                'confidence': 0.5,
                'risk_analysis': {
                    'manipulation_risk': 'medium',
                    'liquidity_risk': 'medium',
                    'volatility_risk': 'medium'
                }
            }
