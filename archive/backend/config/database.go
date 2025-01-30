package config

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis channels
const (
	MarketDataChannel      = "market_data"
	TradeSignalChannel     = "trade_signals"
	AnalysisResultChannel  = "analysis_results"
)

// Database represents a Redis connection
type Database struct {
	client *redis.Client
}

// DatabaseConfig defines Redis connection parameters
type DatabaseConfig struct {
	Host     string
	Password string
	DB       int
}

// NewDatabase creates a new Redis connection
func NewDatabase(config DatabaseConfig) (*Database, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Host,
		Password: config.Password,
		DB:       config.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Database{client: client}, nil
}

// Subscribe subscribes to a Redis channel and returns a message channel
func (db *Database) Subscribe(ctx context.Context, channel string, msgChan chan []byte) error {
	pubsub := db.client.Subscribe(ctx, channel)
	
	// Start a goroutine to handle messages
	go func() {
		defer pubsub.Close()
		for {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				return
			}
			msgChan <- []byte(msg.Payload)
		}
	}()

	return nil
}

// Publish publishes a message to a Redis channel
func (db *Database) Publish(ctx context.Context, channel string, message []byte) error {
	return db.client.Publish(ctx, channel, message).Err()
}

// Close closes the Redis connection
func (db *Database) Close() error {
	return db.client.Close()
}
