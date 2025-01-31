import asyncio
import sys
import os
from typing import Dict, Any, Optional

# Add the current directory to Python path
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

from ollama_client import OllamaClient
from deepseek_client import DeepseekClient

# Set up DeepSeek API key for testing
os.environ['DEEPSEEK_API_KEY'] = 'sk-4ff47d34c52948edab6c9d0e7745b75b'

async def test_market_analysis():
    ollama = OllamaClient()
    deepseek = DeepseekClient()
    
    test_data = {
        'current_price': 1.23,
        'price_change_24h': 5.2,
        'price_change_7d': -2.1,
        'volume_24h': 1000000,
        'volume_change': 15.3,
        'market_cap': 10000000,
        'holders': 5000
    }
    
    # Test Ollama client
    print("Testing Ollama client...")
    result = await ollama.analyze_market_sentiment(test_data)
    print("Ollama result:", result)
    
    # Test DeepSeek client with fallback
    print("\nTesting DeepSeek client with fallback...")
    result = await deepseek.analyze_market_sentiment(test_data)
    print("DeepSeek result:", result)

if __name__ == "__main__":
    asyncio.run(test_market_analysis())
