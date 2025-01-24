package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"solmeme-trader/models"
	"solmeme-trader/monitoring"
	"solmeme-trader/repository"
)

// Collection names
const (
	marketDataCollection = "market_data"
	positionCollection  = "positions"
	tradeCollection     = "trades"
	statsCollection     = "stats"
	analysisCollection  = "analysis"
	indicatorCollection = "indicators"
	eventCollection     = "events"
)

// MongoRepository implements repository.Repository
type MongoRepository struct {
	client   *mongo.Client
	database *mongo.Database
	options  repository.Options
}

// NewRepository creates a new MongoDB repository
func NewRepository(ctx context.Context, uri string, dbName string) (repository.Repository, error) {
	opts := repository.DefaultOptions()
	opts.URI = uri
	opts.Database = dbName

	clientOpts := options.Client().
		ApplyURI(uri).
		SetConnectTimeout(opts.ConnectTimeout).
		SetMaxConnecting(uint64(opts.MaxConnections)).
		SetMinPoolSize(uint64(opts.MinConnections)).
		SetMaxPoolSize(uint64(opts.MaxConnections))

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return &MongoRepository{
		client:   client,
		database: client.Database(dbName),
		options:  opts,
	}, nil
}

// Market data operations

func (r *MongoRepository) SaveMarketData(ctx context.Context, data *models.MarketData) error {
	_, err := r.database.Collection(marketDataCollection).InsertOne(ctx, data)
	return err
}

func (r *MongoRepository) GetMarketData(ctx context.Context, tokenAddress string) (*models.MarketData, error) {
	var data models.MarketData
	err := r.database.Collection(marketDataCollection).FindOne(ctx, bson.M{
		"token_address": tokenAddress,
	}).Decode(&data)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &data, err
}

func (r *MongoRepository) GetMarketDataHistory(ctx context.Context, tokenAddress string, start, end time.Time) ([]*models.MarketData, error) {
	cursor, err := r.database.Collection(marketDataCollection).Find(ctx, bson.M{
		"token_address": tokenAddress,
		"timestamp": bson.M{
			"$gte": start,
			"$lte": end,
		},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var data []*models.MarketData
	if err := cursor.All(ctx, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (r *MongoRepository) GetLatestMarketData(ctx context.Context, tokenAddress string) (*models.MarketData, error) {
	var data models.MarketData
	err := r.database.Collection(marketDataCollection).FindOne(ctx, bson.M{
		"token_address": tokenAddress,
	}, options.FindOne().SetSort(bson.M{"timestamp": -1})).Decode(&data)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &data, err
}

// Position operations

func (r *MongoRepository) SavePosition(ctx context.Context, position *models.Position) error {
	_, err := r.database.Collection(positionCollection).InsertOne(ctx, position)
	return err
}

func (r *MongoRepository) UpdatePosition(ctx context.Context, position *models.Position) error {
	_, err := r.database.Collection(positionCollection).ReplaceOne(ctx, bson.M{
		"id": position.ID,
	}, position)
	return err
}

func (r *MongoRepository) GetPosition(ctx context.Context, id string) (*models.Position, error) {
	var position models.Position
	err := r.database.Collection(positionCollection).FindOne(ctx, bson.M{
		"id": id,
	}).Decode(&position)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &position, err
}

func (r *MongoRepository) GetOpenPositions(ctx context.Context) ([]*models.Position, error) {
	cursor, err := r.database.Collection(positionCollection).Find(ctx, bson.M{
		"status": models.PositionStatusOpen,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var positions []*models.Position
	if err := cursor.All(ctx, &positions); err != nil {
		return nil, err
	}
	return positions, nil
}

func (r *MongoRepository) GetPositionsByToken(ctx context.Context, tokenAddress string) ([]*models.Position, error) {
	cursor, err := r.database.Collection(positionCollection).Find(ctx, bson.M{
		"token_address": tokenAddress,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var positions []*models.Position
	if err := cursor.All(ctx, &positions); err != nil {
		return nil, err
	}
	return positions, nil
}

func (r *MongoRepository) GetPositionHistory(ctx context.Context, tokenAddress string, start, end time.Time) ([]*models.Position, error) {
	cursor, err := r.database.Collection(positionCollection).Find(ctx, bson.M{
		"token_address": tokenAddress,
		"open_time": bson.M{
			"$gte": start,
			"$lte": end,
		},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var positions []*models.Position
	if err := cursor.All(ctx, &positions); err != nil {
		return nil, err
	}
	return positions, nil
}

// Trade operations

func (r *MongoRepository) SaveTrade(ctx context.Context, trade *models.Trade) error {
	_, err := r.database.Collection(tradeCollection).InsertOne(ctx, trade)
	return err
}

func (r *MongoRepository) UpdateTrade(ctx context.Context, trade *models.Trade) error {
	_, err := r.database.Collection(tradeCollection).ReplaceOne(ctx, bson.M{
		"id": trade.ID,
	}, trade)
	return err
}

func (r *MongoRepository) UpdateTradeStatus(ctx context.Context, tradeID string, status string, reason *string) error {
	update := bson.M{
		"$set": bson.M{
			"status":      status,
			"update_time": time.Now(),
		},
	}
	if reason != nil {
		update["$set"].(bson.M)["reason"] = *reason
	}

	_, err := r.database.Collection(tradeCollection).UpdateOne(ctx, bson.M{
		"id": tradeID,
	}, update)
	return err
}

func (r *MongoRepository) GetTrade(ctx context.Context, id string) (*models.Trade, error) {
	var trade models.Trade
	err := r.database.Collection(tradeCollection).FindOne(ctx, bson.M{
		"id": id,
	}).Decode(&trade)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &trade, err
}

func (r *MongoRepository) GetTradesByToken(ctx context.Context, tokenAddress string) ([]*models.Trade, error) {
	cursor, err := r.database.Collection(tradeCollection).Find(ctx, bson.M{
		"token_address": tokenAddress,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var trades []*models.Trade
	if err := cursor.All(ctx, &trades); err != nil {
		return nil, err
	}
	return trades, nil
}

func (r *MongoRepository) GetTradeHistory(ctx context.Context, tokenAddress string, start, end time.Time) ([]*models.Trade, error) {
	cursor, err := r.database.Collection(tradeCollection).Find(ctx, bson.M{
		"token_address": tokenAddress,
		"timestamp": bson.M{
			"$gte": start,
			"$lte": end,
		},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var trades []*models.Trade
	if err := cursor.All(ctx, &trades); err != nil {
		return nil, err
	}
	return trades, nil
}

func (r *MongoRepository) GetTradeStats(ctx context.Context, tokenAddress string) (int, int, float64, error) {
	pipeline := mongo.Pipeline{
		bson.D{{"$match", bson.M{
			"token_address": tokenAddress,
			"status":       models.TradeStatusCompleted,
		}}},
		bson.D{{"$group", bson.M{
			"_id": nil,
			"total_trades": bson.M{"$sum": 1},
			"winning_trades": bson.M{
				"$sum": bson.M{
					"$cond": []interface{}{
						bson.M{"$gt": []interface{}{"$realized_pnl", 0}},
						1,
						0,
					},
				},
			},
			"total_profit": bson.M{
				"$sum": bson.M{
					"$cond": []interface{}{
						bson.M{"$gt": []interface{}{"$realized_pnl", 0}},
						"$realized_pnl",
						0,
					},
				},
			},
			"total_loss": bson.M{
				"$sum": bson.M{
					"$cond": []interface{}{
						bson.M{"$lt": []interface{}{"$realized_pnl", 0}},
						bson.M{"$abs": "$realized_pnl"},
						0,
					},
				},
			},
		}}},
	}

	cursor, err := r.database.Collection(tradeCollection).Aggregate(ctx, pipeline)
	if err != nil {
		return 0, 0, 0, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		TotalTrades   int     `bson:"total_trades"`
		WinningTrades int     `bson:"winning_trades"`
		TotalProfit   float64 `bson:"total_profit"`
		TotalLoss     float64 `bson:"total_loss"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return 0, 0, 0, err
	}

	if len(results) == 0 {
		return 0, 0, 0, nil
	}

	profitFactor := 0.0
	if results[0].TotalLoss > 0 {
		profitFactor = results[0].TotalProfit / results[0].TotalLoss
	}

	return results[0].TotalTrades, results[0].WinningTrades, profitFactor, nil
}

// Stats operations

func (r *MongoRepository) SaveDailyStats(ctx context.Context, stats *models.DailyStats) error {
	_, err := r.database.Collection(statsCollection).InsertOne(ctx, stats)
	return err
}

func (r *MongoRepository) GetDailyStats(ctx context.Context, date time.Time) (*models.DailyStats, error) {
	var stats models.DailyStats
	err := r.database.Collection(statsCollection).FindOne(ctx, bson.M{
		"date": date,
	}).Decode(&stats)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &stats, err
}

func (r *MongoRepository) GetDailyStatsRange(ctx context.Context, start, end time.Time) ([]*models.DailyStats, error) {
	cursor, err := r.database.Collection(statsCollection).Find(ctx, bson.M{
		"date": bson.M{
			"$gte": start,
			"$lte": end,
		},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stats []*models.DailyStats
	if err := cursor.All(ctx, &stats); err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *MongoRepository) CalculateCurrentProfit(ctx context.Context, tokenAddress string) (float64, error) {
	pipeline := mongo.Pipeline{
		bson.D{{"$match", bson.M{
			"token_address": tokenAddress,
			"status":       models.TradeStatusCompleted,
		}}},
		bson.D{{"$group", bson.M{
			"_id":    nil,
			"profit": bson.M{"$sum": "$realized_pnl"},
		}}},
	}

	cursor, err := r.database.Collection(tradeCollection).Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Profit float64 `bson:"profit"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return 0, err
	}

	if len(results) == 0 {
		return 0, nil
	}

	return results[0].Profit, nil
}

// Analysis operations

func (r *MongoRepository) SaveAnalysisResult(ctx context.Context, result *models.AnalysisResult) error {
	_, err := r.database.Collection(analysisCollection).InsertOne(ctx, result)
	return err
}

func (r *MongoRepository) GetAnalysisResult(ctx context.Context, tokenAddress string) (*models.AnalysisResult, error) {
	var result models.AnalysisResult
	err := r.database.Collection(analysisCollection).FindOne(ctx, bson.M{
		"token_address": tokenAddress,
	}, options.FindOne().SetSort(bson.M{"timestamp": -1})).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &result, err
}

func (r *MongoRepository) GetAnalysisHistory(ctx context.Context, tokenAddress string, start, end time.Time) ([]*models.AnalysisResult, error) {
	cursor, err := r.database.Collection(analysisCollection).Find(ctx, bson.M{
		"token_address": tokenAddress,
		"timestamp": bson.M{
			"$gte": start,
			"$lte": end,
		},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*models.AnalysisResult
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// Technical indicators

func (r *MongoRepository) SaveTechnicalIndicators(ctx context.Context, tokenAddress string, indicators *models.TechnicalIndicators) error {
	doc := bson.M{
		"token_address": tokenAddress,
		"indicators":    indicators,
		"timestamp":     time.Now(),
	}
	_, err := r.database.Collection(indicatorCollection).InsertOne(ctx, doc)
	return err
}

func (r *MongoRepository) GetTechnicalIndicators(ctx context.Context, tokenAddress string) (*models.TechnicalIndicators, error) {
	var doc struct {
		Indicators *models.TechnicalIndicators `bson:"indicators"`
	}
	err := r.database.Collection(indicatorCollection).FindOne(ctx, bson.M{
		"token_address": tokenAddress,
	}, options.FindOne().SetSort(bson.M{"timestamp": -1})).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return doc.Indicators, err
}

func (r *MongoRepository) GetIndicatorsHistory(ctx context.Context, tokenAddress string, start, end time.Time) ([]*models.TechnicalIndicators, error) {
	cursor, err := r.database.Collection(indicatorCollection).Find(ctx, bson.M{
		"token_address": tokenAddress,
		"timestamp": bson.M{
			"$gte": start,
			"$lte": end,
		},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []struct {
		Indicators *models.TechnicalIndicators `bson:"indicators"`
	}
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	indicators := make([]*models.TechnicalIndicators, len(docs))
	for i, doc := range docs {
		indicators[i] = doc.Indicators
	}
	return indicators, nil
}

// Monitoring operations

func (r *MongoRepository) SaveEvent(ctx context.Context, event *monitoring.Event) error {
	_, err := r.database.Collection(eventCollection).InsertOne(ctx, event)
	return err
}

func (r *MongoRepository) GetEvents(ctx context.Context, eventType string, start, end time.Time) ([]*monitoring.Event, error) {
	cursor, err := r.database.Collection(eventCollection).Find(ctx, bson.M{
		"type": eventType,
		"timestamp": bson.M{
			"$gte": start,
			"$lte": end,
		},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []*monitoring.Event
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *MongoRepository) GetEventsByToken(ctx context.Context, tokenAddress string, start, end time.Time) ([]*monitoring.Event, error) {
	cursor, err := r.database.Collection(eventCollection).Find(ctx, bson.M{
		"token": tokenAddress,
		"timestamp": bson.M{
			"$gte": start,
			"$lte": end,
		},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []*monitoring.Event
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}
	return events, nil
}

// Utility operations

func (r *MongoRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx, nil)
}

func (r *MongoRepository) Close() error {
	return r.client.Disconnect(context.Background())
}
