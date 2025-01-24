package repository

import (
	"context"
	"time"

	"solmeme-trader/models"
	"solmeme-trader/monitoring"
)

// Repository defines database operations
type Repository interface {
	// Market data operations
	SaveMarketData(ctx context.Context, data *models.MarketData) error
	GetMarketData(ctx context.Context, tokenAddress string) (*models.MarketData, error)
	GetMarketDataHistory(ctx context.Context, tokenAddress string, start, end time.Time) ([]*models.MarketData, error)
	GetLatestMarketData(ctx context.Context, tokenAddress string) (*models.MarketData, error)

	// Position operations
	SavePosition(ctx context.Context, position *models.Position) error
	UpdatePosition(ctx context.Context, position *models.Position) error
	GetPosition(ctx context.Context, id string) (*models.Position, error)
	GetOpenPositions(ctx context.Context) ([]*models.Position, error)
	GetPositionsByToken(ctx context.Context, tokenAddress string) ([]*models.Position, error)
	GetPositionHistory(ctx context.Context, tokenAddress string, start, end time.Time) ([]*models.Position, error)

	// Trade operations
	SaveTrade(ctx context.Context, trade *models.Trade) error
	UpdateTrade(ctx context.Context, trade *models.Trade) error
	UpdateTradeStatus(ctx context.Context, tradeID string, status string, reason *string) error
	GetTrade(ctx context.Context, id string) (*models.Trade, error)
	GetTradesByToken(ctx context.Context, tokenAddress string) ([]*models.Trade, error)
	GetTradeHistory(ctx context.Context, tokenAddress string, start, end time.Time) ([]*models.Trade, error)
	GetTradeStats(ctx context.Context, tokenAddress string) (int, int, float64, error)

	// Stats operations
	SaveDailyStats(ctx context.Context, stats *models.DailyStats) error
	GetDailyStats(ctx context.Context, date time.Time) (*models.DailyStats, error)
	GetDailyStatsRange(ctx context.Context, start, end time.Time) ([]*models.DailyStats, error)
	CalculateCurrentProfit(ctx context.Context, tokenAddress string) (float64, error)

	// Analysis operations
	SaveAnalysisResult(ctx context.Context, result *models.AnalysisResult) error
	GetAnalysisResult(ctx context.Context, tokenAddress string) (*models.AnalysisResult, error)
	GetAnalysisHistory(ctx context.Context, tokenAddress string, start, end time.Time) ([]*models.AnalysisResult, error)

	// Technical indicators
	SaveTechnicalIndicators(ctx context.Context, tokenAddress string, indicators *models.TechnicalIndicators) error
	GetTechnicalIndicators(ctx context.Context, tokenAddress string) (*models.TechnicalIndicators, error)
	GetIndicatorsHistory(ctx context.Context, tokenAddress string, start, end time.Time) ([]*models.TechnicalIndicators, error)

	// Monitoring operations
	SaveEvent(ctx context.Context, event *monitoring.Event) error
	GetEvents(ctx context.Context, eventType string, start, end time.Time) ([]*monitoring.Event, error)
	GetEventsByToken(ctx context.Context, tokenAddress string, start, end time.Time) ([]*monitoring.Event, error)

	// Utility operations
	Ping(ctx context.Context) error
	Close() error
}

// Options defines repository configuration options
type Options struct {
	URI             string        `json:"uri"`
	Database        string        `json:"database"`
	ConnectTimeout  time.Duration `json:"connect_timeout"`
	QueryTimeout    time.Duration `json:"query_timeout"`
	MaxConnections  int           `json:"max_connections"`
	MinConnections  int           `json:"min_connections"`
	MaxIdleTime     time.Duration `json:"max_idle_time"`
	RetryAttempts   int           `json:"retry_attempts"`
	RetryInterval   time.Duration `json:"retry_interval"`
	ConnectRetries  int          `json:"connect_retries"`
	ConnectInterval time.Duration `json:"connect_interval"`
}

// DefaultOptions returns default repository options
func DefaultOptions() Options {
	return Options{
		ConnectTimeout:  10 * time.Second,
		QueryTimeout:    30 * time.Second,
		MaxConnections:  100,
		MinConnections:  10,
		MaxIdleTime:     5 * time.Minute,
		RetryAttempts:   3,
		RetryInterval:   1 * time.Second,
		ConnectRetries:  5,
		ConnectInterval: 5 * time.Second,
	}
}
