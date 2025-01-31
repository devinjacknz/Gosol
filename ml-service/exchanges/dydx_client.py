import aiohttp
from typing import Dict, Optional
from datetime import datetime, timezone

class DydxClient:
    def __init__(self, api_key: Optional[str] = None, api_secret: Optional[str] = None):
        self.api_key = api_key
        self.api_secret = api_secret
        self.base_url = "https://api.dydx.exchange"
        
    async def get_funding_rate(self, market: str) -> Dict:
        async with aiohttp.ClientSession() as session:
            async with session.get(f"{self.base_url}/v3/markets/{market}") as response:
                if response.status != 200:
                    raise Exception(f"dYdX API error: {await response.text()}")
                data = await response.json()
                return {
                    'funding_rate': float(data['market']['nextFundingRate']),
                    'mark_price': float(data['market']['oraclePrice']),
                    'index_price': float(data['market']['indexPrice']),
                    'next_funding_time': datetime.fromtimestamp(
                        int(data['market']['nextFundingAt']),
                        tz=timezone.utc
                    )
                }
    
    async def get_open_interest(self, market: str) -> float:
        async with aiohttp.ClientSession() as session:
            async with session.get(f"{self.base_url}/v3/markets/{market}/stats") as response:
                if response.status != 200:
                    raise Exception(f"dYdX API error: {await response.text()}")
                data = await response.json()
                return float(data['markets']['openInterest'])
