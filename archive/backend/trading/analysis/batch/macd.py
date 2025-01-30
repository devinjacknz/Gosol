import sys
import json
import pandas as pd
import numpy as np

def calculate_macd(prices, fast=12, slow=26, signal=9):
    """计算MACD指标"""
    if len(prices) < slow:
        return {
            "error": f"需要至少{slow}个数据点，当前{len(prices)}个"
        }
    
    close_prices = np.array(prices, dtype=np.float64)
    # 计算EMA
    ema_fast = pd.Series(close_prices).ewm(span=fast, adjust=False).mean()
    ema_slow = pd.Series(close_prices).ewm(span=slow, adjust=False).mean()
    macd_line = ema_fast - ema_slow
    signal_line = macd_line.ewm(span=signal, adjust=False).mean()
    
    # 返回原始计算结果（包含NaN）
    return {
        "macd": macd_line.tolist(),
        "signal": signal_line.tolist()
    }

if __name__ == "__main__":
    try:
        # 从命令行参数获取输入数据
        input_data = json.loads(sys.argv[1])
        results = calculate_macd(input_data)
        print(json.dumps(results))
    except Exception as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(1)
