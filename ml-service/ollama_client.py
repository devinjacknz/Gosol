import aiohttp
import asyncio
import os
from typing import Dict, Any, Optional
import json
from loguru import logger

class OllamaClient:
    def __init__(self, model: str = "deepseek-r1:1.5b"):
        self.base_url = os.getenv("OLLAMA_HOST", "http://host.docker.internal:11434")
        self.model = model
        self.max_retries = 3
        self.retry_delay = 1
        self.initialized = False
        self.model = model
        self.max_retries = 3
        self.retry_delay = 1

    async def analyze_market_sentiment(self, token_data: Dict) -> Dict[str, Any]:
        print("Creating analysis prompt...")
        prompt = self._create_analysis_prompt(token_data)
        print("Prompt created successfully")
        
        timeout = aiohttp.ClientTimeout(total=120)  # Increase timeout to 2 minutes
        print(f"Sending request to Ollama at {self.base_url}/api/chat...")
        
        headers = {
            'Content-Type': 'application/json',
            'Accept': 'application/x-ndjson'
        }
        
        default_response = {
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
        
        for attempt in range(self.max_retries):
            try:
                async with aiohttp.ClientSession(timeout=timeout) as session:
                    print("Making POST request to Ollama...")
                    async with session.post(
                        f"{self.base_url}/api/chat",
                        headers=headers,
                        json={
                            "model": self.model,
                            "messages": [
                                {"role": "system", "content": "You are a cryptocurrency market analyst specializing in Solana meme coins. You analyze market data and provide structured JSON responses."},
                                {"role": "user", "content": prompt}
                            ],
                            "stream": True
                        }
                    ) as response:
                        print(f"Received response with status: {response.status}")
                        if response.status != 200:
                            error_text = await response.text()
                            print(f"Error response from Ollama: {error_text}")
                            raise Exception(f"Ollama API error: {error_text}")
                        
                        print("Reading streaming response...")
                        full_response = ""
                        async for line in response.content:
                            if not line:
                                continue
                            try:
                                chunk = json.loads(line)
                                if chunk.get("done", False):
                                    break
                                content = chunk.get("message", {}).get("content", "")
                                if content:
                                    full_response += content
                                    print("Received chunk:", content[:50], "...")
                                    await asyncio.sleep(0.1)  # Small delay between chunks
                            except json.JSONDecodeError as e:
                                print(f"Failed to decode JSON: {e}")
                                continue
                        
                        print("Parsing complete response...")
                        # Extract JSON from markdown response
                        json_start = full_response.find('```json\n')
                        if json_start == -1:
                            json_start = full_response.find('{')
                            if json_start == -1:
                                raise ValueError("No JSON found in response")
                            json_end = full_response.rfind('}') + 1
                        else:
                            json_start += 8  # Length of '```json\n'
                            json_end = full_response.find('\n```', json_start)
                            if json_end == -1:
                                raise ValueError("Malformed JSON response")
                        json_str = full_response[json_start:json_end].strip()
                        return self._parse_analysis_response({"response": json_str})
            except Exception as e:
                print(f"Attempt {attempt + 1} failed: {str(e)}")
                if attempt == self.max_retries - 1:
                    print(f"Failed after {self.max_retries} attempts: {str(e)}")
                    return default_response
                await asyncio.sleep(self.retry_delay)
        return default_response

    def _create_analysis_prompt(self, token_data: Dict) -> str:
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
            "market_sentiment": "bullish/bearish/neutral",
            "risk_level": "number between 1-10",
            "short_term_prediction": {{
                "target_price": "predicted price",
                "timeframe": "timeframe in hours/days",
                "key_levels": {{
                    "support": "support price level",
                    "resistance": "resistance price level"
                }}
            }},
            "key_factors": [
                "factor 1",
                "factor 2",
                "..."
            ],
            "trading_recommendation": "BUY/SELL/HOLD",
            "confidence": "number between 0-1",
            "risk_analysis": {{
                "market_manipulation_risk": "low/medium/high",
                "liquidity_risk": "low/medium/high",
                "volatility_risk": "low/medium/high"
            }}
        }}

        Ensure your response is a valid JSON object with all fields properly formatted.
        """

    def _parse_analysis_response(self, response: Dict) -> Dict:
        try:
            content = response['response']
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

    async def get_model_info(self) -> Dict[str, Any]:
        """Get detailed information about the loaded model and pull if needed"""
        try:
            async with aiohttp.ClientSession() as session:
                # Try to get model info
                async with session.post(
                    f"{self.base_url}/api/show",
                    json={"name": self.model}
                ) as response:
                    if response.status == 200:
                        data = await response.json()
                        return {
                            "status": "loaded",
                            "size": data.get("size", "unknown"),
                            "modified": data.get("modified", "unknown"),
                            "latency_ms": data.get("response_ms", 0)
                        }
                    
                    # Model not found, try to pull it
                    async with session.post(
                        f"{self.base_url}/api/pull",
                        json={"name": self.model}
                    ) as pull_response:
                        if pull_response.status == 200:
                            # Wait for model to be ready
                            await asyncio.sleep(2)
                            return {
                                "status": "loaded",
                                "size": "unknown",
                                "modified": "unknown",
                                "latency_ms": 0
                            }
                        return {
                            "status": "not_loaded",
                            "error": f"Model pull failed with status {pull_response.status}"
                        }
        except Exception as e:
            return {
                "status": "error",
                "error": str(e)
            }

    async def check_health(self) -> bool:
        """Check if Ollama service is healthy and model is available"""
        try:
            if not self.initialized:
                # Initialize connection on first health check
                async with aiohttp.ClientSession() as session:
                    async with session.get(f"{self.base_url}/api/health") as response:
                        if response.status == 200:
                            self.initialized = True
                        else:
                            return False

            # Check if model is available
            model_info = await self.get_model_info()
            if model_info.get("status") == "loaded":
                return True

            # Try to pull model if not loaded
            async with aiohttp.ClientSession() as session:
                async with session.post(
                    f"{self.base_url}/api/pull",
                    json={"name": self.model}
                ) as pull_response:
                    if pull_response.status == 200:
                        await asyncio.sleep(2)  # Wait for model to be ready
                        return True
            return False
        except Exception as e:
            logger.error(f"Ollama health check failed: {e}")
            return False
            
    async def is_available(self) -> bool:
        """Deprecated: Use check_health() instead"""
        return await self.check_health()
