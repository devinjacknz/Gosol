import aiohttp
import asyncio
from typing import Dict, Optional
from datetime import datetime, timezone

class DydxClient:
    def __init__(self, api_key: Optional[str] = None, api_secret: Optional[str] = None):
        if not api_key:
            raise ValueError("API key is required for dYdX client")
        self.api_key = api_key
        self.api_secret = api_secret
        self.base_url = "https://api.dydx.exchange"
        self.headers = {
            "DYDX-API-KEY": self.api_key,
            "DYDX-TIMESTAMP": str(int(datetime.now().timestamp() * 1000)),
            "DYDX-SIGNATURE": "",
            "DYDX-PASSPHRASE": "",
        }
        
    async def get_funding_rate(self, market: str) -> Dict:
        max_retries = 3
        retry_delay = 1
        for attempt in range(max_retries):
            try:
                async with aiohttp.ClientSession() as session:
                    async with session.get(
                        f"{self.base_url}/v3/markets/{market}",
                        headers=self.headers
                    ) as response:
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
            except Exception as e:
                if attempt == max_retries - 1:
                    raise Exception(f"Failed to fetch funding rate after {max_retries} attempts: {str(e)}")
                await asyncio.sleep(retry_delay)
    
    async def get_open_interest(self, market: str) -> float:
        max_retries = 3
        retry_delay = 1
        for attempt in range(max_retries):
            try:
                async with aiohttp.ClientSession() as session:
                    async with session.get(
                        f"{self.base_url}/v3/markets/{market}/stats",
                        headers=self.headers
                    ) as response:
                        if response.status != 200:
                            raise Exception(f"dYdX API error: {await response.text()}")
                        data = await response.json()
                        return float(data['markets']['openInterest'])
            except Exception as e:
                if attempt == max_retries - 1:
                    raise Exception(f"Failed to fetch open interest after {max_retries} attempts: {str(e)}")
                await asyncio.sleep(retry_delay)
