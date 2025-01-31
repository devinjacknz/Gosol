#!/bin/bash

# 设置环境变量
export EXCHANGE_API_KEY="your_api_key_here"
export EXCHANGE_API_SECRET="your_api_secret_here"
export USE_TESTNET="true"

# 创建必要的目录
mkdir -p database logs models

# 检查Python环境
command -v python3 >/dev/null 2>&1 || { echo "Python3 is required but not installed. Aborting." >&2; exit 1; }

# 安装基础依赖
echo "Installing basic dependencies..."
pip3 install ccxt pandas numpy websockets pandas-ta streaming-indicators

# 安装机器学习依赖
echo "Installing ML dependencies..."
pip3 install torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cpu
pip3 install transformers accelerate

# 下载模型（如果需要）
echo "Checking for DeepSeek model..."
if [ ! -d "models/deepseek-coder-1.5b-base" ]; then
    echo "Downloading DeepSeek model..."
    python3 -c "
from transformers import AutoModelForCausalLM, AutoTokenizer
model_name = 'deepseek-ai/deepseek-coder-1.5b-base'
tokenizer = AutoTokenizer.from_pretrained(model_name)
model = AutoModelForCausalLM.from_pretrained(model_name)
tokenizer.save_pretrained('models/deepseek-coder-1.5b-base')
model.save_pretrained('models/deepseek-coder-1.5b-base')
"
fi

# 运行系统
echo "Starting trading system..."
python3 trading_system.py   