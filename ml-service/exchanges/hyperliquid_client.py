import aiohttp
import asyncio
from typing import Dict, Optional
from datetime import datetime, timezone

class HyperliquidClient:
    def __init__(self, api_key: Optional[str] = None, api_secret: Optional[str] = None):
        if not api_key:
            raise ValueError("API key is required for Hyperliquid client")
        self.api_key = api_key
        self.api_secret = api_secret
        self.base_url = "https://api.hyperliquid.xyz"
        self.mainnet_url = "https://api.hyperliquid.xyz/info"
        self.testnet_url = "https://api.hyperliquid-testnet.xyz/info"
        self.use_testnet = True  # Default to testnet for safety
        self.max_retries = 3
        self.retry_delay = 1  # seconds
        self.headers = {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": f"Bearer {api_key}"
        }
    
    async def get_funding_rate(self, market: str) -> Dict:
        url = self.testnet_url if self.use_testnet else self.mainnet_url
        for attempt in range(self.max_retries):
            try:
                async with aiohttp.ClientSession() as session:
                    async with session.get(url, headers=self.headers) as response:
                        if response.status != 200:
                            raise Exception(f"Hyperliquid API error: {await response.text()}")
                        data = await response.json()
                        market_data = next(
                            (m for m in data['universe'] if m['name'] == market),
                            None
                        )
                        if not market_data:
                            raise Exception(f"Market {market} not found")
                        
                        return {
                            'funding_rate': float(market_data['funding']),
                            'mark_price': float(market_data['markPrice']),
                            'index_price': float(market_data['indexPrice']),
                            'next_funding_time': datetime.now(timezone.utc)
                        }
            except Exception as e:
                if attempt == self.max_retries - 1:
                    raise Exception(f"Failed to fetch funding rate after {self.max_retries} attempts: {str(e)}")
                await asyncio.sleep(self.retry_delay)
        return {
            'funding_rate': 0.0,
            'mark_price': 0.0,
            'index_price': 0.0,
            'next_funding_time': datetime.now(timezone.utc)
        }
    
    async def get_open_interest(self, market: str) -> float:
        url = self.testnet_url if self.use_testnet else self.mainnet_url
        for attempt in range(self.max_retries):
            try:
                async with aiohttp.ClientSession() as session:
                    async with session.get(url, headers=self.headers) as response:
                        if response.status != 200:
                            raise Exception(f"Hyperliquid API error: {await response.text()}")
                        data = await response.json()
                        market_data = next(
                            (m for m in data['universe'] if m['name'] == market),
                            None
                        )
                        if not market_data:
                            raise Exception(f"Market {market} not found")
                        return float(market_data['openInterest'])
            except Exception as e:
                if attempt == self.max_retries - 1:
                    raise Exception(f"Failed to fetch open interest after {self.max_retries} attempts: {str(e)}")
                await asyncio.sleep(self.retry_delay)
        return 0.0
