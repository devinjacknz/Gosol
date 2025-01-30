package models

import (
	"time"

	"gorm.io/gorm"
)

// MarketDataRecord represents market data stored in the database
type MarketDataRecord struct {
	gorm.Model
	Symbol    string `gorm:"index:idx_market_data_symbol_timestamp"`
	Price     float64
	Volume    float64
	Timestamp time.Time `gorm:"index:idx_market_data_symbol_timestamp"`
}

// KlineRecord represents candlestick data stored in the database
type KlineRecord struct {
	gorm.Model
	Symbol    string `gorm:"index:idx_kline_symbol_interval_timestamp"`
	Interval  string `gorm:"index:idx_kline_symbol_interval_timestamp"`
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	Timestamp time.Time `gorm:"index:idx_kline_symbol_interval_timestamp"`
}

// TradeRecord represents executed trades stored in the database
type TradeRecord struct {
	gorm.Model
	UserID    string `gorm:"index:idx_trade_user_timestamp"`
	Symbol    string `gorm:"index"`
	Side      string
	Price     float64
	Amount    float64
	Fee       float64
	Total     float64
	Timestamp time.Time `gorm:"index:idx_trade_user_timestamp"`
}

// OrderRecord represents orders stored in the database
type OrderRecord struct {
	gorm.Model
	OrderID   string `gorm:"uniqueIndex"`
	UserID    string `gorm:"index:idx_order_user_status"`
	Symbol    string `gorm:"index"`
	Side      string
	Type      string
	Price     float64
	Amount    float64
	Filled    float64
	Status    string    `gorm:"index:idx_order_user_status"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
}

// PositionRecord represents user positions stored in the database
type PositionRecord struct {
	gorm.Model
	UserID        string `gorm:"uniqueIndex:idx_position_user_symbol"`
	Symbol        string `gorm:"uniqueIndex:idx_position_user_symbol"`
	Amount        float64
	EntryPrice    float64
	CurrentPrice  float64
	UnrealizedPnL float64
	RealizedPnL   float64
	UpdatedAt     time.Time
}

// StrategyRecord represents trading strategies stored in the database
type StrategyRecord struct {
	gorm.Model
	UserID      string `gorm:"index"`
	Name        string `gorm:"uniqueIndex"`
	Description string
	Config      string // JSON encoded configuration
	Status      string
	CreatedAt   time.Time `gorm:"index"`
	UpdatedAt   time.Time
}

// BacktestRecord represents backtest results stored in the database
type BacktestRecord struct {
	gorm.Model
	UserID         string `gorm:"index"`
	StrategyName   string `gorm:"index"`
	StartTime      time.Time
	EndTime        time.Time
	InitialCapital float64
	FinalCapital   float64
	TotalTrades    int
	WinRate        float64
	SharpeRatio    float64
	MaxDrawdown    float64
	Results        string    // JSON encoded detailed results
	CreatedAt      time.Time `gorm:"index"`
}

// RiskLimitRecord represents user risk limits stored in the database
type RiskLimitRecord struct {
	gorm.Model
	UserID         string `gorm:"uniqueIndex:idx_risk_user_symbol"`
	Symbol         string `gorm:"uniqueIndex:idx_risk_user_symbol"`
	MaxPosition    float64
	MaxLeverage    float64
	MaxDrawdown    float64
	DailyLossLimit float64
	UpdatedAt      time.Time
}

// UserSettingRecord represents user settings stored in the database
type UserSettingRecord struct {
	gorm.Model
	UserID    string `gorm:"uniqueIndex"`
	Settings  string // JSON encoded settings
	UpdatedAt time.Time
}

// MarketDataCache represents cached market data
type MarketDataCache struct {
	gorm.Model
	Symbol    string `gorm:"uniqueIndex"`
	Data      string // JSON encoded market data
	UpdatedAt time.Time
}

// OrderBookCache represents cached order book data
type OrderBookCache struct {
	gorm.Model
	Symbol    string `gorm:"uniqueIndex"`
	Data      string // JSON encoded order book
	UpdatedAt time.Time
}

// IndicatorCache represents cached technical indicator values
type IndicatorCache struct {
	gorm.Model
	Symbol    string `gorm:"uniqueIndex:idx_indicator_cache"`
	Indicator string `gorm:"uniqueIndex:idx_indicator_cache"`
	Interval  string `gorm:"uniqueIndex:idx_indicator_cache"`
	Values    string // JSON encoded indicator values
	UpdatedAt time.Time
}
