import os
import aiohttp
import asyncio
from typing import Dict, List, Any, Optional
import json
from ollama_client import OllamaClient

class DeepseekClient:
    def __init__(self, api_key: Optional[str] = None):
        self.api_key = api_key or os.getenv('DEEPSEEK_API_KEY')
        self.base_url = "https://api.deepseek.com/v1/chat"
        self.ollama_client = OllamaClient()
        
    async def analyze_market_sentiment(self, token_data: Dict) -> Dict[str, Any]:
        """Analyze market sentiment using Deepseek's API"""
        headers = {
            "Authorization": f"Bearer {self.api_key}",
            "Content-Type": "application/json"
        }
        
        # Prepare market data for analysis
        prompt = self._create_analysis_prompt(token_data)
        
        try:
            async with aiohttp.ClientSession() as session:
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
            if await self.ollama_client.is_available():
                try:
                    return await asyncio.wait_for(
                        self.ollama_client.analyze_market_sentiment(token_data),
                        timeout=120
                    )
                except Exception as ollama_error:
                    print(f"Ollama fallback failed: {ollama_error}")
                    raise api_error
            print("Ollama not available for fallback")
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
        {
            "market_sentiment": "bullish",  // Example: use "bullish", "bearish", or "neutral"
            "risk_level": 5,  // Example: use a number between 1-10
            "short_term_prediction": {
                "target_price": "1.25",  // Example price prediction
                "timeframe": "24h",  // Example: "4h", "24h", "7d"
                "key_levels": {
                    "support": "1.20",  // Example support level
                    "resistance": "1.30"  // Example resistance level
                }
            },
            "key_factors": [
                "Strong volume increase",  // Example factors
                "Positive price momentum",
                "Growing holder base"
            ],
            "trading_recommendation": "BUY",  // Example: use "BUY", "SELL", or "HOLD"
            "confidence": 0.75,  // Example: use a number between 0 and 1
            "risk_analysis": {
                "market_manipulation_risk": "medium",  // Example: use "low", "medium", or "high"
                "liquidity_risk": "low",  // Example: use "low", "medium", or "high"
                "volatility_risk": "high"  // Example: use "low", "medium", or "high"
            }
        }

        Ensure your response is a valid JSON object with all fields properly formatted.
        """
    
    def _parse_analysis_response(self, response: Dict) -> Dict:
        """Parse the Deepseek API response into structured data"""
        try:
            content = response['choices'][0]['message']['content']
            # Parse the JSON response from the content
            analysis = json.loads(content)
            return {
                'sentiment': analysis.get('market_sentiment', 'neutral'),
                'risk_level': float(analysis.get('risk_level', 5)),
                'price_prediction': {
                    'target': analysis.get('short_term_prediction', {}).get('target_price'),
                    'timeframe': analysis.get('short_term_prediction', {}).get('timeframe'),
                    'support': analysis.get('short_term_prediction', {}).get('key_levels', {}).get('support'),
                    'resistance': analysis.get('short_term_prediction', {}).get('key_levels', {}).get('resistance')
                },
                'key_factors': analysis.get('key_factors', []),
                'recommendation': analysis.get('trading_recommendation', 'HOLD'),
                'confidence': float(analysis.get('confidence', 0.5)),
                'risk_analysis': {
                    'manipulation_risk': analysis.get('risk_analysis', {}).get('market_manipulation_risk', 'medium'),
                    'liquidity_risk': analysis.get('risk_analysis', {}).get('liquidity_risk', 'medium'),
                    'volatility_risk': analysis.get('risk_analysis', {}).get('volatility_risk', 'medium')
                }
            }
        except (KeyError, json.JSONDecodeError) as e:
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
