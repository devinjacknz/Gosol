package repository

import (
	"context"
	"time"

	"github.com/leonzhao/trading-system/backend/models"
)

// Repository defines the interface for database operations
type Repository interface {
	// Trade operations
	SaveTrade(ctx context.Context, trade *models.Trade) error
	GetTradeByID(ctx context.Context, id string) (*models.Trade, error)
	ListTrades(ctx context.Context, filter *models.TradeFilter) ([]*models.Trade, error)
	UpdateTradeStatus(ctx context.Context, id string, status models.TradeStatus) error
	UpdateTrade(ctx context.Context, trade *models.Trade) error
	GetTradeStats(ctx context.Context, filter *models.TradeFilter) (*models.TradeStats, error)

	// Position operations
	SavePosition(ctx context.Context, position *models.Position) error
	GetPositionByID(ctx context.Context, id string) (*models.Position, error)
	ListPositions(ctx context.Context, filter *models.PositionFilter) ([]*models.Position, error)
	UpdatePosition(ctx context.Context, position *models.Position) error
	GetOpenPositions(ctx context.Context) ([]*models.Position, error)
	ClosePosition(ctx context.Context, id string, closePrice float64) error
	GetPositionStats(ctx context.Context, filter *models.PositionFilter) (*models.PositionStats, error)

	// Market data operations
	SaveMarketData(ctx context.Context, data *models.MarketData) error
	GetLatestMarketData(ctx context.Context, tokenAddress string) (*models.MarketData, error)
	GetHistoricalMarketData(ctx context.Context, tokenAddress string, limit int) ([]*models.MarketData, error)
	GetMarketStats(ctx context.Context, tokenAddress string) (*models.MarketStats, error)
	GetTechnicalIndicators(ctx context.Context, tokenAddress string) (*models.TechnicalIndicators, error)
	SaveAnalysisResult(ctx context.Context, result *models.AnalysisResult) error
	GetLatestAnalysis(ctx context.Context, tokenAddress string) (*models.AnalysisResult, error)

	// Daily stats operations
	SaveDailyStats(ctx context.Context, stats *models.DailyStats) error
	GetDailyStats(ctx context.Context, date time.Time) (*models.DailyStats, error)
	GetDailyStatsRange(ctx context.Context, startDate, endDate time.Time) ([]*models.DailyStats, error)

	// Health check
	Ping(ctx context.Context) error
}

// Options represents repository configuration options
type Options struct {
	URI            string
	Database       string
	Username       string
	Password       string
	Timeout        time.Duration
	ConnectTimeout time.Duration
	MaxConnections uint64
	MinConnections uint64
}

// DefaultOptions returns default repository options
func DefaultOptions() Options {
	return Options{
		URI:            "mongodb://localhost:27017",
		Database:       "solmeme_trader",
		Timeout:        10 * time.Second,
		ConnectTimeout: 5 * time.Second,
		MaxConnections: 100,
		MinConnections: 10,
	}
}

// Constants for position status
const (
	PositionStatusOpen       = "open"
	PositionStatusClosed     = "closed"
	PositionStatusLiquidated = "liquidated"
)

// Constants for trade status
const (
	TradeStatusPending   = "pending"
	TradeStatusExecuted  = "executed"
	TradeStatusFailed    = "failed"
	TradeStatusCancelled = "cancelled"
)

// Constants for trade side
const (
	TradeSideBuy  = "buy"
	TradeSideSell = "sell"
)

// Constants for trade type
const (
	TradeTypeMarket = "market"
	TradeTypeLimit  = "limit"
	TradeTypeStop   = "stop"
)

// Constants for risk management
const (
	DefaultLiquidationThreshold = 0.5  // 50% loss
	DefaultLeverageLimit        = 10.0 // 10x max leverage
	DefaultPositionSizeLimit    = 0.2  // 20% of portfolio
	DefaultStopLossLimit        = 0.1  // 10% loss
	DefaultTakeProfitLimit      = 0.2  // 20% gain
)
