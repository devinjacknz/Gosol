# Solana Meme Coin Trading Agent

A sophisticated trading agent for Solana meme coins with ML-powered analysis using Deepseek.

## Features

- React-based dashboard for real-time trading monitoring
- Dual wallet system (A/B wallets) for profit management
- Go-based trading core for high-performance execution
- Python ML service with Deepseek integration
- Automatic/Manual trading mode switch
- Real-time DEX trading on Solana
- Performance metrics and analytics

## Project Structure

```
.
├── frontend/         # React TypeScript dashboard
├── backend/         # Go trading core
└── ml-service/      # Python ML service with Deepseek
```

## Setup Instructions

### Frontend (React Dashboard)
```bash
cd frontend
npm install
npm start
```

### Backend (Go Trading Core)
```bash
cd backend
go mod tidy
go run main.go
```

### ML Service (Python)
```bash
cd ml-service
source venv/bin/activate  # On Windows: .\venv\Scripts\activate
pip install -r requirements.txt
python main.py
```

## Configuration

- Set up wallet configurations in the dashboard
- Configure trading parameters through the UI
- Set up API keys for DEX interactions

## Security Notes

- Private keys are stored securely and encrypted
- B wallet receives profits automatically from A wallet
- All transactions are logged and monitored

## License

MIT License 