-- Create market_data table
CREATE TABLE market_data (
    id SERIAL PRIMARY KEY,
    token_address TEXT NOT NULL,
    price DOUBLE PRECISION NOT NULL,
    volume_24h DOUBLE PRECISION NOT NULL,
    market_cap DOUBLE PRECISION NOT NULL,
    holders INTEGER NOT NULL,
    liquidity DOUBLE PRECISION NOT NULL,
    price_impact DOUBLE PRECISION NOT NULL,
    order_book JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create trades table
CREATE TABLE trades (
    id SERIAL PRIMARY KEY,
    token_address TEXT NOT NULL,
    wallet_address TEXT NOT NULL,
    type TEXT NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    price DOUBLE PRECISION NOT NULL,
    close_price DOUBLE PRECISION,
    status TEXT NOT NULL,
    tx_hash TEXT,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Create analysis_results table
CREATE TABLE analysis_results (
    id SERIAL PRIMARY KEY,
    token_address TEXT NOT NULL,
    prediction DOUBLE PRECISION NOT NULL,
    confidence DOUBLE PRECISION NOT NULL,
    sentiment TEXT NOT NULL,
    risk_level INTEGER NOT NULL,
    technical_indicators JSONB NOT NULL,
    deepseek_analysis JSONB NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Create indexes
CREATE INDEX idx_market_data_token_address ON market_data(token_address);
CREATE INDEX idx_market_data_timestamp ON market_data(timestamp);
CREATE INDEX idx_trades_token_address ON trades(token_address);
CREATE INDEX idx_trades_wallet_address ON trades(wallet_address);
CREATE INDEX idx_trades_timestamp ON trades(timestamp);
CREATE INDEX idx_analysis_results_token_address ON analysis_results(token_address);
CREATE INDEX idx_analysis_results_timestamp ON analysis_results(timestamp);
