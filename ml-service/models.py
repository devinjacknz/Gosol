from dataclasses import dataclass
from typing import List, Optional
from datetime import datetime

@dataclass
class PredictionRequest:
    token_address: str
    price_history: List[float]
    volume_history: List[float]
    timestamp: int
    market_cap: float
    holders: int

@dataclass
class PredictionResponse:
    sentiment: str
    risk_level: float
    recommendation: str
    confidence: float
    price_prediction: Optional[float] = None
    support_level: Optional[float] = None
    resistance_level: Optional[float] = None
    timestamp: int = int(datetime.now().timestamp())
