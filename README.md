# SolMeme Trader

An automated trading system for Solana meme tokens, featuring real-time market analysis, risk management, and multi-DEX integration.

## Project Structure

```
.
├── backend/             # Go backend service
│   ├── config/         # Configuration management
│   ├── dex/           # DEX integration clients
│   ├── models/        # Data models
│   ├── monitoring/    # System monitoring
│   ├── repository/    # Data persistence layer
│   ├── service/       # Business logic layer
│   ├── trading/       # Trading core components
│   └── tests/         # Integration tests
├── frontend/           # React frontend
│   ├── public/        # Static assets
│   └── src/           # Source code
└── ml-service/         # Python ML service
```

## Features

- Real-time market data monitoring
- Technical analysis and trading signals
- Risk management and position sizing
- Multi-DEX support (Raydium, Jupiter)
- Performance analytics
- Web interface for monitoring and control

## Prerequisites

- Go 1.21+
- Node.js 18+
- Python 3.9+
- MongoDB
- Redis

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/solmeme-trader.git
cd solmeme-trader
```

2. Install backend dependencies:
```bash
cd backend
go mod download
```

3. Install frontend dependencies:
```bash
cd frontend
npm install
```

4. Install ML service dependencies:
```bash
cd ml-service
pip install -r requirements.txt
```

## Configuration

1. Backend configuration:
```bash
cp backend/.env.example backend/.env
# Edit backend/.env with your settings
```

2. Frontend configuration:
```bash
cp frontend/.env.example frontend/.env
# Edit frontend/.env with your settings
```

3. ML service configuration:
```bash
cp ml-service/.env.example ml-service/.env
# Edit ml-service/.env with your settings
```

## Running the Application

1. Start the backend service:
```bash
cd backend
go run main.go
```

2. Start the frontend development server:
```bash
cd frontend
npm start
```

3. Start the ML service:
```bash
cd ml-service
python main.py
```

## Development

### Backend Development

```bash
cd backend
go test ./...        # Run tests
go run main.go      # Run development server
```

### Frontend Development

```bash
cd frontend
npm test            # Run tests
npm start           # Start development server
```

### ML Service Development

```bash
cd ml-service
python -m pytest    # Run tests
python main.py      # Start development server
```

## Testing

- Backend: `go test ./...`
- Frontend: `npm test`
- ML Service: `python -m pytest`
- Integration Tests: `make test-integration`

## Deployment

1. Build the services:
```bash
make build
```

2. Run with Docker:
```bash
docker-compose up -d
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Solana](https://solana.com/)
- [Raydium](https://raydium.io/)
- [Jupiter](https://jup.ag/)
