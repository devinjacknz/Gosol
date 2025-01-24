-- Create market_data table
CREATE TABLE IF NOT EXISTS market_data (
    id BIGSERIAL PRIMARY KEY,
    token_address TEXT NOT NULL,
    price DOUBLE PRECISION NOT NULL,
    volume_24h DOUBLE PRECISION NOT NULL,
    market_cap DOUBLE PRECISION NOT NULL,
    liquidity DOUBLE PRECISION NOT NULL,
    price_impact DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT market_data_token_address_timestamp_unique UNIQUE (token_address, timestamp)
);

-- Create trades table
CREATE TABLE IF NOT EXISTS trades (
    id BIGSERIAL PRIMARY KEY,
    token_address TEXT NOT NULL,
    wallet_address TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('buy', 'sell')),
    amount DOUBLE PRECISION NOT NULL,
    price DOUBLE PRECISION NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'completed', 'failed', 'cancelled')),
    tx_hash TEXT,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Create analysis_results table
CREATE TABLE IF NOT EXISTS analysis_results (
    id BIGSERIAL PRIMARY KEY,
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
CREATE INDEX IF NOT EXISTS market_data_token_address_idx ON market_data (token_address);
CREATE INDEX IF NOT EXISTS market_data_timestamp_idx ON market_data (timestamp DESC);

CREATE INDEX IF NOT EXISTS trades_token_address_idx ON trades (token_address);
CREATE INDEX IF NOT EXISTS trades_wallet_address_idx ON trades (wallet_address);
CREATE INDEX IF NOT EXISTS trades_timestamp_idx ON trades (timestamp DESC);
CREATE INDEX IF NOT EXISTS trades_status_idx ON trades (status);

CREATE INDEX IF NOT EXISTS analysis_results_token_address_idx ON analysis_results (token_address);
CREATE INDEX IF NOT EXISTS analysis_results_timestamp_idx ON analysis_results (timestamp DESC);

-- Add comments
COMMENT ON TABLE market_data IS 'Real-time market data for tokens';
COMMENT ON TABLE trades IS 'Trade records for buy/sell operations';
COMMENT ON TABLE analysis_results IS 'Market analysis results including technical and AI-powered analysis';

COMMENT ON COLUMN market_data.token_address IS 'Token mint address';
COMMENT ON COLUMN market_data.price IS 'Current token price in SOL';
COMMENT ON COLUMN market_data.volume_24h IS '24-hour trading volume in SOL';
COMMENT ON COLUMN market_data.market_cap IS 'Market capitalization in SOL';
COMMENT ON COLUMN market_data.liquidity IS 'Available liquidity in SOL';
COMMENT ON COLUMN market_data.price_impact IS 'Price impact percentage for standard trade size';

COMMENT ON COLUMN trades.token_address IS 'Token mint address';
COMMENT ON COLUMN trades.wallet_address IS 'Trader wallet address';
COMMENT ON COLUMN trades.type IS 'Trade type: buy or sell';
COMMENT ON COLUMN trades.amount IS 'Trade amount in SOL';
COMMENT ON COLUMN trades.price IS 'Trade price in SOL';
COMMENT ON COLUMN trades.status IS 'Trade status: pending, completed, failed, or cancelled';
COMMENT ON COLUMN trades.tx_hash IS 'Solana transaction hash';

COMMENT ON COLUMN analysis_results.token_address IS 'Token mint address';
COMMENT ON COLUMN analysis_results.prediction IS 'Price movement prediction (-1 to 1)';
COMMENT ON COLUMN analysis_results.confidence IS 'Prediction confidence (0 to 1)';
COMMENT ON COLUMN analysis_results.sentiment IS 'Market sentiment: bullish, bearish, or neutral';
COMMENT ON COLUMN analysis_results.risk_level IS 'Risk level (1-5)';
COMMENT ON COLUMN analysis_results.technical_indicators IS 'Technical analysis indicators in JSON format';
COMMENT ON COLUMN analysis_results.deepseek_analysis IS 'AI-powered analysis results in JSON format';
