import aiohttp
from typing import Dict, Optional
from datetime import datetime, timezone

class HyperliquidClient:
    def __init__(self, api_key: Optional[str] = None, api_secret: Optional[str] = None):
        self.api_key = api_key
        self.api_secret = api_secret
        self.base_url = "https://api.hyperliquid.xyz"
    
    async def get_funding_rate(self, market: str) -> Dict:
        async with aiohttp.ClientSession() as session:
            async with session.get(f"{self.base_url}/info") as response:
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
    
    async def get_open_interest(self, market: str) -> float:
        async with aiohttp.ClientSession() as session:
            async with session.get(f"{self.base_url}/info") as response:
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
