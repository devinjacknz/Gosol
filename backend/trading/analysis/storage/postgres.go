package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kwanRoshi/Gosol/backend/trading/analysis/streaming"
	_ "github.com/lib/pq"
)

// PostgresStorage implements Storage interface using PostgreSQL
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage creates a new PostgreSQL storage instance
func NewPostgresStorage(connStr string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create table if not exists
	if err := createTable(db); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &PostgresStorage{db: db}, nil
}

func createTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS indicator_values (
			id SERIAL PRIMARY KEY,
			indicator_name VARCHAR(50) NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			value DOUBLE PRECISION NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(indicator_name, timestamp)
		);
		CREATE INDEX IF NOT EXISTS idx_indicator_timestamp 
		ON indicator_values(indicator_name, timestamp);
	`
	_, err := db.Exec(query)
	return err
}

// Store stores an indicator value
func (s *PostgresStorage) Store(ctx context.Context, value streaming.IndicatorValue) error {
	query := `
		INSERT INTO indicator_values (indicator_name, timestamp, value)
		VALUES ($1, $2, $3)
		ON CONFLICT (indicator_name, timestamp)
		DO UPDATE SET value = EXCLUDED.value
	`
	_, err := s.db.ExecContext(ctx, query, value.Name, value.Timestamp, value.Value)
	if err != nil {
		return fmt.Errorf("failed to store value: %w", err)
	}
	return nil
}

// Query retrieves indicator values within a time range
func (s *PostgresStorage) Query(ctx context.Context, indicatorName string, timeRange TimeRange) ([]streaming.IndicatorValue, error) {
	query := `
		SELECT timestamp, value
		FROM indicator_values
		WHERE indicator_name = $1
		AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp ASC
	`
	rows, err := s.db.QueryContext(ctx, query, indicatorName, timeRange.Start, timeRange.End)
	if err != nil {
		return nil, fmt.Errorf("failed to query values: %w", err)
	}
	defer rows.Close()

	var results []streaming.IndicatorValue
	for rows.Next() {
		var value streaming.IndicatorValue
		value.Name = indicatorName
		if err := rows.Scan(&value.Timestamp, &value.Value); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, value)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no data found for indicator %s", indicatorName)
	}

	return results, nil
}

// GetLatest retrieves the latest value for an indicator
func (s *PostgresStorage) GetLatest(ctx context.Context, indicatorName string) (*streaming.IndicatorValue, error) {
	query := `
		SELECT timestamp, value
		FROM indicator_values
		WHERE indicator_name = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`
	var value streaming.IndicatorValue
	value.Name = indicatorName

	err := s.db.QueryRowContext(ctx, query, indicatorName).Scan(&value.Timestamp, &value.Value)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no data found for indicator %s", indicatorName)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest value: %w", err)
	}

	return &value, nil
}

// Close closes the database connection
func (s *PostgresStorage) Close() error {
	return s.db.Close()
}
