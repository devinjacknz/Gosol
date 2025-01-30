-- Drop indexes
DROP INDEX IF EXISTS idx_market_data_token_address;
DROP INDEX IF EXISTS idx_market_data_timestamp;
DROP INDEX IF EXISTS idx_trades_token_address;
DROP INDEX IF EXISTS idx_trades_wallet_address;
DROP INDEX IF EXISTS idx_trades_timestamp;
DROP INDEX IF EXISTS idx_analysis_results_token_address;
DROP INDEX IF EXISTS idx_analysis_results_timestamp;

-- Drop tables
DROP TABLE IF EXISTS market_data;
DROP TABLE IF EXISTS trades;
DROP TABLE IF EXISTS analysis_results;
