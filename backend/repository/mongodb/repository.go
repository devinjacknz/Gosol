package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"solmeme-trader/models"
	"solmeme-trader/repository"
)

// MongoRepository implements the Repository interface using MongoDB
type MongoRepository struct {
	client     *mongo.Client
	database   string
	trades     *mongo.Collection
	positions  *mongo.Collection
	marketData *mongo.Collection
	dailyStats *mongo.Collection
	analysis   *mongo.Collection
}

// NewRepository creates a new MongoDB repository
func NewRepository(ctx context.Context, opts repository.Options) (repository.Repository, error) {
	clientOpts := options.Client().
		ApplyURI(opts.URI).
		SetConnectTimeout(opts.ConnectTimeout).
		SetMaxPoolSize(opts.MaxConnections).
		SetMinPoolSize(opts.MinConnections).
		SetMaxConnecting(opts.MaxConnections)

	if opts.Username != "" && opts.Password != "" {
		clientOpts.SetAuth(options.Credential{
			Username: opts.Username,
			Password: opts.Password,
		})
	}

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	repo := &MongoRepository{
		client:     client,
		database:   opts.Database,
		trades:     client.Database(opts.Database).Collection("trades"),
		positions:  client.Database(opts.Database).Collection("positions"),
		marketData: client.Database(opts.Database).Collection("market_data"),
		dailyStats: client.Database(opts.Database).Collection("daily_stats"),
		analysis:   client.Database(opts.Database).Collection("analysis"),
	}

	return repo, nil
}

// SaveDailyStats saves daily trading statistics
func (r *MongoRepository) SaveDailyStats(ctx context.Context, stats *models.DailyStats) error {
	filter := bson.M{"date": stats.Date}
	update := bson.M{"$set": stats}
	opts := options.Update().SetUpsert(true)

	_, err := r.dailyStats.UpdateOne(ctx, filter, update, opts)
	return err
}

// GetDailyStats retrieves daily trading statistics for a specific date
func (r *MongoRepository) GetDailyStats(ctx context.Context, date time.Time) (*models.DailyStats, error) {
	// Normalize date to start of day
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	var stats models.DailyStats
	err := r.dailyStats.FindOne(ctx, bson.M{"date": startOfDay}).Decode(&stats)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Return empty stats for the day if none exist
			return &models.DailyStats{
				Date: startOfDay,
			}, nil
		}
		return nil, err
	}

	return &stats, nil
}

// GetDailyStatsRange retrieves daily trading statistics for a date range
func (r *MongoRepository) GetDailyStatsRange(ctx context.Context, startDate, endDate time.Time) ([]*models.DailyStats, error) {
	// Normalize dates to start of day
	startOfDay := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endOfDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())

	filter := bson.M{
		"date": bson.M{
			"$gte": startOfDay,
			"$lte": endOfDay,
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "date", Value: 1}})

	cursor, err := r.dailyStats.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stats []*models.DailyStats
	if err = cursor.All(ctx, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// SaveMarketData saves market data
func (r *MongoRepository) SaveMarketData(ctx context.Context, data *models.MarketData) error {
	filter := bson.M{"token_address": data.TokenAddress}
	update := bson.M{"$set": data}
	opts := options.Update().SetUpsert(true)

	_, err := r.marketData.UpdateOne(ctx, filter, update, opts)
	return err
}

// GetLatestMarketData retrieves the latest market data for a token
func (r *MongoRepository) GetLatestMarketData(ctx context.Context, tokenAddress string) (*models.MarketData, error) {
	var data models.MarketData
	err := r.marketData.FindOne(ctx, bson.M{"token_address": tokenAddress}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// GetHistoricalMarketData retrieves historical market data for a token
func (r *MongoRepository) GetHistoricalMarketData(ctx context.Context, tokenAddress string, limit int) ([]*models.MarketData, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.marketData.Find(ctx, bson.M{"token_address": tokenAddress}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var data []*models.MarketData
	if err = cursor.All(ctx, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// GetMarketStats retrieves market statistics for a token
func (r *MongoRepository) GetMarketStats(ctx context.Context, tokenAddress string) (*models.MarketStats, error) {
	var stats models.MarketStats
	err := r.marketData.FindOne(ctx, bson.M{
		"token_address": tokenAddress,
	}).Decode(&stats)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// GetTechnicalIndicators retrieves technical indicators for a token
func (r *MongoRepository) GetTechnicalIndicators(ctx context.Context, tokenAddress string) (*models.TechnicalIndicators, error) {
	var indicators models.TechnicalIndicators
	err := r.analysis.FindOne(
		ctx,
		bson.M{"token_address": tokenAddress},
		options.FindOne().SetSort(bson.M{"timestamp": -1}),
	).Decode(&indicators)
	if err == mongo.ErrNoDocuments {
		return &models.TechnicalIndicators{}, nil
	}
	return &indicators, err
}

// SaveAnalysisResult saves market analysis results
func (r *MongoRepository) SaveAnalysisResult(ctx context.Context, result *models.AnalysisResult) error {
	filter := bson.M{"token_address": result.TokenAddress}
	update := bson.M{"$set": result}
	opts := options.Update().SetUpsert(true)

	_, err := r.analysis.UpdateOne(ctx, filter, update, opts)
	return err
}

// GetLatestAnalysis retrieves the latest analysis for a token
func (r *MongoRepository) GetLatestAnalysis(ctx context.Context, tokenAddress string) (*models.AnalysisResult, error) {
	var result models.AnalysisResult
	err := r.analysis.FindOne(ctx, bson.M{
		"token_address": tokenAddress,
	}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ClosePosition closes a position with the given ID and close price
func (r *MongoRepository) ClosePosition(ctx context.Context, id string, closePrice float64) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":        repository.PositionStatusClosed,
			"current_price": closePrice,
			"close_time":   time.Now(),
			"last_updated": time.Now(),
		},
	}

	_, err = r.positions.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

// UpdateTrade updates a trade in the database
func (r *MongoRepository) UpdateTrade(ctx context.Context, trade *models.Trade) error {
	objectID, err := primitive.ObjectIDFromHex(trade.ID)
	if err != nil {
		return err
	}

	trade.UpdateTime = time.Now()
	_, err = r.trades.ReplaceOne(ctx, bson.M{"_id": objectID}, trade)
	return err
}

// SavePosition saves a position to the database
func (r *MongoRepository) SavePosition(ctx context.Context, position *models.Position) error {
	if position.ID == "" {
		position.ID = primitive.NewObjectID().Hex()
	}
	position.LastUpdated = time.Now()
	
	_, err := r.positions.InsertOne(ctx, position)
	return err
}

// GetPositionByID retrieves a position by ID
func (r *MongoRepository) GetPositionByID(ctx context.Context, id string) (*models.Position, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var position models.Position
	err = r.positions.FindOne(ctx, bson.M{"_id": objectID}).Decode(&position)
	if err != nil {
		return nil, err
	}

	return &position, nil
}

// ListPositions lists positions based on filter
func (r *MongoRepository) ListPositions(ctx context.Context, filter *models.PositionFilter) ([]*models.Position, error) {
	query := bson.M{}
	if filter.TokenAddress != "" {
		query["token_address"] = filter.TokenAddress
	}
	if filter.Side != "" {
		query["side"] = filter.Side
	}
	if filter.Status != "" {
		query["status"] = filter.Status
	}
	if filter.StartTime != nil {
		query["open_time"] = bson.M{"$gte": filter.StartTime}
	}
	if filter.EndTime != nil {
		if _, ok := query["open_time"]; ok {
			query["open_time"].(bson.M)["$lte"] = filter.EndTime
		} else {
			query["open_time"] = bson.M{"$lte": filter.EndTime}
		}
	}

	cursor, err := r.positions.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var positions []*models.Position
	if err = cursor.All(ctx, &positions); err != nil {
		return nil, err
	}

	return positions, nil
}

// GetOpenPositions retrieves all open positions
func (r *MongoRepository) GetOpenPositions(ctx context.Context) ([]*models.Position, error) {
	return r.ListPositions(ctx, &models.PositionFilter{
		Status: repository.PositionStatusOpen,
	})
}

// UpdatePosition updates a position in the database
func (r *MongoRepository) UpdatePosition(ctx context.Context, position *models.Position) error {
	objectID, err := primitive.ObjectIDFromHex(position.ID)
	if err != nil {
		return err
	}

	position.LastUpdated = time.Now()
	_, err = r.positions.ReplaceOne(ctx, bson.M{"_id": objectID}, position)
	return err
}

// GetPositionStats retrieves position statistics
func (r *MongoRepository) GetPositionStats(ctx context.Context, filter *models.PositionFilter) (*models.PositionStats, error) {
	positions, err := r.ListPositions(ctx, filter)
	if err != nil {
		return nil, err
	}

	stats := &models.PositionStats{
		LastUpdated: time.Now(),
	}

	for _, pos := range positions {
		stats.TotalPositions++
		if pos.Status == repository.PositionStatusOpen {
			stats.OpenPositions++
			stats.UnrealizedPnL += pos.UnrealizedPnL
			stats.TotalValue += pos.Value
		} else {
			stats.ClosedPositions++
			stats.RealizedPnL += pos.RealizedPnL
		}
	}

	return stats, nil
}

// Ping checks the database connection
func (r *MongoRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx, nil)
}
