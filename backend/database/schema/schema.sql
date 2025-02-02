-- Market Data Tables
CREATE TABLE markets (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL UNIQUE,
    exchange VARCHAR(10) NOT NULL, -- 'hyperliquid' or 'dydx'
    base_currency VARCHAR(10) NOT NULL,
    quote_currency VARCHAR(10) NOT NULL,
    min_size DECIMAL(20,8) NOT NULL,
    price_precision INT NOT NULL,
    size_precision INT NOT NULL,
    maintenance_margin DECIMAL(5,4) NOT NULL,
    initial_margin DECIMAL(5,4) NOT NULL,
    max_leverage INT NOT NULL,
    funding_interval INT NOT NULL, -- in seconds
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT markets_symbol_idx UNIQUE (symbol, exchange)
);

CREATE TABLE trades (
    id BIGSERIAL PRIMARY KEY,
    market_id INT NOT NULL,
    price DECIMAL(20,8) NOT NULL,
    size DECIMAL(20,8) NOT NULL,
    side VARCHAR(4) NOT NULL,
    liquidation BOOLEAN NOT NULL DEFAULT FALSE,
    funding_rate DECIMAL(10,8),
    executed_at TIMESTAMP NOT NULL,
    CONSTRAINT trades_market_fk FOREIGN KEY (market_id) REFERENCES markets(id)
) PARTITION BY RANGE (executed_at);

-- Create monthly partitions for trades
CREATE TABLE trades_y2024m01 PARTITION OF trades
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
CREATE TABLE trades_y2024m02 PARTITION OF trades
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

-- Create indexes on trades partitions
CREATE INDEX trades_market_time_idx ON trades (market_id, executed_at);
CREATE INDEX trades_time_idx ON trades (executed_at);

-- Order Management Tables
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    client_order_id VARCHAR(50) NOT NULL,
    market_id INT NOT NULL,
    type VARCHAR(20) NOT NULL,
    side VARCHAR(4) NOT NULL,
    price DECIMAL(20,8),
    stop_price DECIMAL(20,8),
    size DECIMAL(20,8) NOT NULL,
    filled_size DECIMAL(20,8) NOT NULL DEFAULT 0,
    remaining_size DECIMAL(20,8) NOT NULL,
    leverage INT NOT NULL,
    reduce_only BOOLEAN NOT NULL DEFAULT FALSE,
    post_only BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    CONSTRAINT orders_market_fk FOREIGN KEY (market_id) REFERENCES markets(id),
    CONSTRAINT orders_client_id_idx UNIQUE (client_order_id)
) PARTITION BY RANGE (created_at);

-- Create monthly partitions for orders
CREATE TABLE orders_y2024m01 PARTITION OF orders
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
CREATE TABLE orders_y2024m02 PARTITION OF orders
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

-- Create indexes on orders partitions
CREATE INDEX orders_market_status_idx ON orders (market_id, status, created_at);
CREATE INDEX orders_status_time_idx ON orders (status, created_at);

-- Position Management Tables
CREATE TABLE positions (
    id BIGSERIAL PRIMARY KEY,
    market_id INT NOT NULL,
    side VARCHAR(4) NOT NULL,
    entry_price DECIMAL(20,8) NOT NULL,
    current_price DECIMAL(20,8) NOT NULL,
    size DECIMAL(20,8) NOT NULL,
    leverage INT NOT NULL,
    liquidation_price DECIMAL(20,8) NOT NULL,
    margin_used DECIMAL(20,8) NOT NULL,
    unrealized_pnl DECIMAL(20,8) NOT NULL,
    realized_pnl DECIMAL(20,8) NOT NULL,
    funding_fee DECIMAL(20,8) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    closed_at TIMESTAMP,
    CONSTRAINT positions_market_fk FOREIGN KEY (market_id) REFERENCES markets(id)
);

CREATE INDEX positions_market_status_idx ON positions (market_id, status);
CREATE INDEX positions_status_time_idx ON positions (status, created_at);

-- Risk Management Tables
CREATE TABLE risk_checks (
    id BIGSERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL,
    level VARCHAR(10) NOT NULL,
    status VARCHAR(10) NOT NULL,
    value DECIMAL(20,8) NOT NULL,
    threshold DECIMAL(20,8) NOT NULL,
    market_id INT,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT risk_checks_market_fk FOREIGN KEY (market_id) REFERENCES markets(id)
) PARTITION BY RANGE (created_at);

-- Create monthly partitions for risk_checks
CREATE TABLE risk_checks_y2024m01 PARTITION OF risk_checks
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
CREATE TABLE risk_checks_y2024m02 PARTITION OF risk_checks
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

-- Create indexes on risk_checks partitions
CREATE INDEX risk_checks_type_time_idx ON risk_checks (type, created_at);
CREATE INDEX risk_checks_market_time_idx ON risk_checks (market_id, created_at);

-- Funding Rate Tables
CREATE TABLE funding_rates (
    id BIGSERIAL PRIMARY KEY,
    market_id INT NOT NULL,
    rate DECIMAL(10,8) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    CONSTRAINT funding_rates_market_fk FOREIGN KEY (market_id) REFERENCES markets(id)
) PARTITION BY RANGE (timestamp);

-- Create monthly partitions for funding_rates
CREATE TABLE funding_rates_y2024m01 PARTITION OF funding_rates
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
CREATE TABLE funding_rates_y2024m02 PARTITION OF funding_rates
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

-- Create indexes on funding_rates partitions
CREATE INDEX funding_rates_market_time_idx ON funding_rates (market_id, timestamp);

-- Maintenance Functions
CREATE OR REPLACE FUNCTION create_next_month_partition()
RETURNS void AS $$
DECLARE
    next_month DATE;
    partition_name TEXT;
    create_sql TEXT;
BEGIN
    next_month := date_trunc('month', now()) + interval '1 month';
    
    -- Create trades partition
    partition_name := 'trades_y' || to_char(next_month, 'YYYY') || 'm' || to_char(next_month, 'MM');
    create_sql := format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF trades FOR VALUES FROM (%L) TO (%L)',
        partition_name,
        next_month,
        next_month + interval '1 month'
    );
    EXECUTE create_sql;
    
    -- Create orders partition
    partition_name := 'orders_y' || to_char(next_month, 'YYYY') || 'm' || to_char(next_month, 'MM');
    create_sql := format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF orders FOR VALUES FROM (%L) TO (%L)',
        partition_name,
        next_month,
        next_month + interval '1 month'
    );
    EXECUTE create_sql;
    
    -- Create risk_checks partition
    partition_name := 'risk_checks_y' || to_char(next_month, 'YYYY') || 'm' || to_char(next_month, 'MM');
    create_sql := format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF risk_checks FOR VALUES FROM (%L) TO (%L)',
        partition_name,
        next_month,
        next_month + interval '1 month'
    );
    EXECUTE create_sql;
    
    -- Create funding_rates partition
    partition_name := 'funding_rates_y' || to_char(next_month, 'YYYY') || 'm' || to_char(next_month, 'MM');
    create_sql := format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF funding_rates FOR VALUES FROM (%L) TO (%L)',
        partition_name,
        next_month,
        next_month + interval '1 month'
    );
    EXECUTE create_sql;
END;
$$ LANGUAGE plpgsql;

-- Create function to archive old data
CREATE OR REPLACE FUNCTION archive_old_data(older_than INTERVAL)
RETURNS void AS $$
DECLARE
    archive_date TIMESTAMP;
    partition_name TEXT;
    archive_sql TEXT;
BEGIN
    archive_date := date_trunc('month', now() - older_than);
    
    -- Archive trades
    partition_name := 'trades_y' || to_char(archive_date, 'YYYY') || 'm' || to_char(archive_date, 'MM');
    archive_sql := format(
        'CREATE TABLE IF NOT EXISTS archived_%I (LIKE %I INCLUDING ALL)',
        partition_name,
        partition_name
    );
    EXECUTE archive_sql;
    
    archive_sql := format(
        'INSERT INTO archived_%I SELECT * FROM %I',
        partition_name,
        partition_name
    );
    EXECUTE archive_sql;
    
    -- Repeat for orders and risk_checks
    -- ... similar logic for other tables
END;
$$ LANGUAGE plpgsql;

-- Create maintenance function
CREATE OR REPLACE FUNCTION perform_maintenance()
RETURNS void AS $$
BEGIN
    -- Create next month's partitions
    PERFORM create_next_month_partition();
    
    -- Archive data older than 3 months
    PERFORM archive_old_data(interval '3 months');
    
    -- Analyze tables for query optimization
    ANALYZE markets;
    ANALYZE trades;
    ANALYZE orders;
    ANALYZE positions;
    ANALYZE risk_checks;
END;
$$ LANGUAGE plpgsql; 