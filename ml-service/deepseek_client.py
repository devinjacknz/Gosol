import os
import aiohttp
from typing import Dict, List
import json

class DeepseekClient:
    def __init__(self, api_key: str = None):
        self.api_key = api_key or os.getenv('DEEPSEEK_API_KEY')
        self.base_url = "https://api.deepseek.com/v1/chat"
        
    async def analyze_market_sentiment(self, token_data: Dict) -> Dict:
        """Analyze market sentiment using Deepseek's API"""
        headers = {
            "Authorization": f"Bearer {self.api_key}",
            "Content-Type": "application/json"
        }
        
        # Prepare market data for analysis
        prompt = self._create_analysis_prompt(token_data)
        
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
            "market_sentiment": "bullish/bearish/neutral",
            "risk_level": <number between 1-10>,
            "short_term_prediction": {
                "target_price": <predicted price>,
                "timeframe": "<timeframe in hours/days>",
                "key_levels": {
                    "support": <support price level>,
                    "resistance": <resistance price level>
                }
            },
            "key_factors": [
                "<factor 1>",
                "<factor 2>",
                "..."
            ],
            "trading_recommendation": "BUY/SELL/HOLD",
            "confidence": <number between 0-1>,
            "risk_analysis": {
                "market_manipulation_risk": "low/medium/high",
                "liquidity_risk": "low/medium/high",
                "volatility_risk": "low/medium/high"
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
