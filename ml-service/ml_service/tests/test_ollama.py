import asyncio
import sys
import os
from typing import Dict, Any, Optional

# Add the current directory to Python path
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

from ml_service.ollama_client import OllamaClient
from deepseek_client import DeepseekClient

# Set up DeepSeek API key for testing
os.environ['DEEPSEEK_API_KEY'] = 'sk-4ff47d34c52948edab6c9d0e7745b75b'

async def test_market_analysis():
    try:
        print("Initializing clients...")
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
        
        # Test Ollama availability
        print("Checking Ollama availability...")
        is_available = await ollama.is_available()
        print(f"Ollama available: {is_available}")
        
        if not is_available:
            print("Warning: Ollama service not available!")
            return
        
        # Test Ollama client with timeout
        print("Testing Ollama client...")
        try:
            result = await asyncio.wait_for(
                ollama.analyze_market_sentiment(test_data),
                timeout=120  # Increase timeout to match client
            )
            print("Ollama result:", result)
        except asyncio.TimeoutError:
            print("Error: Ollama request timed out after 120 seconds")
            return
        except Exception as e:
            print(f"Error during Ollama request: {str(e)}")
            return
        
        # Test DeepSeek client with fallback
        print("\nTesting DeepSeek client with fallback...")
        try:
            result = await asyncio.wait_for(
                deepseek.analyze_market_sentiment(test_data),
                timeout=30
            )
            print("DeepSeek result:", result)
        except asyncio.TimeoutError:
            print("Error: DeepSeek request timed out after 30 seconds")
        except Exception as e:
            print(f"Error during DeepSeek request: {str(e)}")
    except Exception as e:
        print(f"Test failed with error: {str(e)}")

if __name__ == "__main__":
    asyncio.run(test_market_analysis())
