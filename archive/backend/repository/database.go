package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/leonzhao/trading-system/backend/models"
)

// Database represents the database connection
type Database struct {
	db *gorm.DB
}

// NewDatabase creates a new database connection
func NewDatabase(dsn string) (*Database, error) {
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate schemas
	if err := db.AutoMigrate(
		&models.MarketDataRecord{},
		&models.KlineRecord{},
		&models.TradeRecord{},
		&models.OrderRecord{},
		&models.PositionRecord{},
		&models.StrategyRecord{},
		&models.BacktestRecord{},
		&models.RiskLimitRecord{},
		&models.UserSettingRecord{},
		&models.MarketDataCache{},
		&models.OrderBookCache{},
		&models.IndicatorCache{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &Database{db: db}, nil
}

// Market Data Operations

func (d *Database) SaveMarketData(ctx context.Context, data *models.MarketDataRecord) error {
	return d.db.WithContext(ctx).Create(data).Error
}

func (d *Database) GetMarketData(ctx context.Context, symbol string, start, end time.Time) ([]*models.MarketDataRecord, error) {
	var records []*models.MarketDataRecord
	err := d.db.WithContext(ctx).
		Where("symbol = ? AND timestamp BETWEEN ? AND ?", symbol, start, end).
		Order("timestamp ASC").
		Find(&records).Error
	return records, err
}

// Kline Operations

func (d *Database) SaveKline(ctx context.Context, kline *models.KlineRecord) error {
	return d.db.WithContext(ctx).Create(kline).Error
}

func (d *Database) GetKlines(ctx context.Context, symbol, interval string, limit int) ([]*models.KlineRecord, error) {
	var records []*models.KlineRecord
	err := d.db.WithContext(ctx).
		Where("symbol = ? AND interval = ?", symbol, interval).
		Order("timestamp DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

// Trade Operations

func (d *Database) SaveTrade(ctx context.Context, trade *models.TradeRecord) error {
	return d.db.WithContext(ctx).Create(trade).Error
}

func (d *Database) GetUserTrades(ctx context.Context, userID string, limit int) ([]*models.TradeRecord, error) {
	var records []*models.TradeRecord
	err := d.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

// Order Operations

func (d *Database) SaveOrder(ctx context.Context, order *models.OrderRecord) error {
	return d.db.WithContext(ctx).Create(order).Error
}

func (d *Database) UpdateOrder(ctx context.Context, order *models.OrderRecord) error {
	return d.db.WithContext(ctx).Save(order).Error
}

func (d *Database) GetOrder(ctx context.Context, orderID string) (*models.OrderRecord, error) {
	var record models.OrderRecord
	err := d.db.WithContext(ctx).Where("order_id = ?", orderID).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (d *Database) GetUserOrders(ctx context.Context, userID string, status string) ([]*models.OrderRecord, error) {
	var records []*models.OrderRecord
	query := d.db.WithContext(ctx).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at DESC").Find(&records).Error
	return records, err
}

// Position Operations

func (d *Database) SavePosition(ctx context.Context, position *models.PositionRecord) error {
	return d.db.WithContext(ctx).Create(position).Error
}

func (d *Database) UpdatePosition(ctx context.Context, position *models.PositionRecord) error {
	return d.db.WithContext(ctx).Save(position).Error
}

func (d *Database) GetPosition(ctx context.Context, userID, symbol string) (*models.PositionRecord, error) {
	var record models.PositionRecord
	err := d.db.WithContext(ctx).
		Where("user_id = ? AND symbol = ?", userID, symbol).
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// Strategy Operations

func (d *Database) SaveStrategy(ctx context.Context, strategy *models.StrategyRecord) error {
	return d.db.WithContext(ctx).Create(strategy).Error
}

func (d *Database) UpdateStrategy(ctx context.Context, strategy *models.StrategyRecord) error {
	return d.db.WithContext(ctx).Save(strategy).Error
}

func (d *Database) GetStrategy(ctx context.Context, name string) (*models.StrategyRecord, error) {
	var record models.StrategyRecord
	err := d.db.WithContext(ctx).Where("name = ?", name).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (d *Database) GetUserStrategies(ctx context.Context, userID string) ([]*models.StrategyRecord, error) {
	var records []*models.StrategyRecord
	err := d.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&records).Error
	return records, err
}

// Backtest Operations

func (d *Database) SaveBacktest(ctx context.Context, backtest *models.BacktestRecord) error {
	return d.db.WithContext(ctx).Create(backtest).Error
}

func (d *Database) UpdateBacktest(ctx context.Context, backtest *models.BacktestRecord) error {
	return d.db.WithContext(ctx).Save(backtest).Error
}

func (d *Database) GetBacktest(ctx context.Context, id uint) (*models.BacktestRecord, error) {
	var record models.BacktestRecord
	err := d.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (d *Database) GetUserBacktests(ctx context.Context, userID string) ([]*models.BacktestRecord, error) {
	var records []*models.BacktestRecord
	err := d.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&records).Error
	return records, err
}

// Risk Limit Operations

func (d *Database) SaveRiskLimit(ctx context.Context, limit *models.RiskLimitRecord) error {
	return d.db.WithContext(ctx).Create(limit).Error
}

func (d *Database) UpdateRiskLimit(ctx context.Context, limit *models.RiskLimitRecord) error {
	return d.db.WithContext(ctx).Save(limit).Error
}

func (d *Database) GetRiskLimit(ctx context.Context, userID, symbol string) (*models.RiskLimitRecord, error) {
	var record models.RiskLimitRecord
	err := d.db.WithContext(ctx).
		Where("user_id = ? AND symbol = ?", userID, symbol).
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// Cache Operations

func (d *Database) SaveMarketDataCache(ctx context.Context, cache *models.MarketDataCache) error {
	return d.db.WithContext(ctx).Create(cache).Error
}

func (d *Database) GetMarketDataCache(ctx context.Context, symbol string) (*models.MarketDataCache, error) {
	var record models.MarketDataCache
	err := d.db.WithContext(ctx).
		Where("symbol = ?", symbol).
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (d *Database) SaveOrderBookCache(ctx context.Context, cache *models.OrderBookCache) error {
	return d.db.WithContext(ctx).Create(cache).Error
}

func (d *Database) GetOrderBookCache(ctx context.Context, symbol string) (*models.OrderBookCache, error) {
	var record models.OrderBookCache
	err := d.db.WithContext(ctx).
		Where("symbol = ?", symbol).
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (d *Database) SaveIndicatorCache(ctx context.Context, cache *models.IndicatorCache) error {
	return d.db.WithContext(ctx).Create(cache).Error
}

func (d *Database) GetIndicatorCache(ctx context.Context, symbol, indicator, interval string) (*models.IndicatorCache, error) {
	var record models.IndicatorCache
	err := d.db.WithContext(ctx).
		Where("symbol = ? AND indicator = ? AND interval = ?", symbol, indicator, interval).
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// Cleanup Operations

func (d *Database) CleanupOldData(ctx context.Context, before time.Time) error {
	// Delete old market data
	if err := d.db.WithContext(ctx).
		Where("timestamp < ?", before).
		Delete(&models.MarketDataRecord{}).Error; err != nil {
		return err
	}

	// Delete old klines
	if err := d.db.WithContext(ctx).
		Where("timestamp < ?", before).
		Delete(&models.KlineRecord{}).Error; err != nil {
		return err
	}

	// Delete old trades
	if err := d.db.WithContext(ctx).
		Where("timestamp < ?", before).
		Delete(&models.TradeRecord{}).Error; err != nil {
		return err
	}

	return nil
}

// Performance Optimization

func (d *Database) Vacuum() error {
	return d.db.Exec("VACUUM ANALYZE").Error
}

func (d *Database) ReindexTables() error {
	tables := []string{
		"market_data_records",
		"kline_records",
		"trade_records",
		"order_records",
		"position_records",
	}

	for _, table := range tables {
		if err := d.db.Exec(fmt.Sprintf("REINDEX TABLE %s", table)).Error; err != nil {
			return err
		}
	}

	return nil
}
