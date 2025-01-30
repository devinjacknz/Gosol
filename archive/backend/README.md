# Solmeme Trader Backend

A trading bot for Solana meme tokens that integrates with DEXes like Jupiter and Raydium.

## Features

- Real-time market data monitoring
- Automated trading with configurable strategies
- Risk management and position sizing
- Integration with multiple DEXes
- Technical analysis indicators
- AI-powered market analysis
- PostgreSQL for persistent storage
- Redis for real-time data and pub/sub

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL 15
- Redis 7
- Solana CLI tools

## Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd backend
```

2. Install dependencies:
```bash
make deps
```

3. Copy the environment file and configure your settings:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Start the required services:
```bash
make docker-up
```

5. Run database migrations:
```bash
make migrate-up
```

## Development

Start the development server:
```bash
make dev
```

Run tests:
```bash
make test
```

Run linter:
```bash
make lint
```

## Project Structure

```
backend/
├── config/         # Configuration and initialization
├── dex/           # DEX integration (Jupiter, Raydium)
├── migrations/    # Database migrations
├── models/        # Data models and types
├── repository/    # Database operations
├── service/       # Business logic
├── trading/       # Trading execution and monitoring
├── main.go        # Application entry point
└── docker-compose.yml
```

## API Endpoints

### Market Data
- `GET /api/market-data/:token` - Get latest market data
- `GET /api/analysis/:token` - Get market analysis

### Trading
- `GET /api/status` - Get trading status
- `GET /api/config` - Get trading configuration
- `POST /api/config` - Update trading configuration
- `POST /api/toggle-trading` - Enable/disable trading
- `GET /api/wallet-balance` - Get wallet balance
- `POST /api/transfer-profit` - Transfer profits to wallet B

## Configuration

### Environment Variables

- `POSTGRES_*` - PostgreSQL connection settings
- `REDIS_*` - Redis connection settings
- `SOLANA_RPC_ENDPOINT` - Solana RPC endpoint
- `PORT` - Server port (default: 8080)
- `GIN_MODE` - Gin framework mode (debug/release)

### Trading Parameters

- `MAX_AMOUNT` - Maximum trade size in SOL
- `MIN_AMOUNT` - Minimum trade size in SOL
- `STOP_LOSS` - Stop loss percentage
- `TAKE_PROFIT` - Take profit percentage
- `MAX_SLIPPAGE` - Maximum allowed slippage
- `RETRY_ATTEMPTS` - Number of retry attempts for failed trades
- `RETRY_DELAY` - Delay between retries
- `RISK_PER_TRADE` - Maximum risk per trade as percentage
- `MAX_OPEN_TRADES` - Maximum number of concurrent trades

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
