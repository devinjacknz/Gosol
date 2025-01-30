package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/leonzhao/trading-system/backend/models"
)

// SaveTrade saves a trade to the database
func (r *MongoRepository) SaveTrade(ctx context.Context, trade *models.Trade) error {
	if trade.ID == "" {
		trade.ID = primitive.NewObjectID().Hex()
	}
	trade.UpdateTime = time.Now()
	
	_, err := r.trades.InsertOne(ctx, trade)
	return err
}

// GetTradeByID retrieves a trade by ID
func (r *MongoRepository) GetTradeByID(ctx context.Context, id string) (*models.Trade, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var trade models.Trade
	err = r.trades.FindOne(ctx, bson.M{"_id": objectID}).Decode(&trade)
	if err != nil {
		return nil, err
	}

	return &trade, nil
}

// ListTrades lists trades based on filter
func (r *MongoRepository) ListTrades(ctx context.Context, filter *models.TradeFilter) ([]*models.Trade, error) {
	query := bson.M{}
	if filter.TokenAddress != "" {
		query["token_address"] = filter.TokenAddress
	}
	if len(filter.Type) > 0 {
		query["type"] = bson.M{"$in": filter.Type}
	}
	if len(filter.Side) > 0 {
		query["side"] = bson.M{"$in": filter.Side}
	}
	if len(filter.Status) > 0 {
		query["status"] = bson.M{"$in": filter.Status}
	}
	if filter.StartTime != nil {
		query["timestamp"] = bson.M{"$gte": filter.StartTime}
	}
	if filter.EndTime != nil {
		if _, ok := query["timestamp"]; ok {
			query["timestamp"].(bson.M)["$lte"] = filter.EndTime
		} else {
			query["timestamp"] = bson.M{"$lte": filter.EndTime}
		}
	}
	if filter.MinAmount != nil {
		query["amount"] = bson.M{"$gte": filter.MinAmount}
	}
	if filter.MaxAmount != nil {
		if _, ok := query["amount"]; ok {
			query["amount"].(bson.M)["$lte"] = filter.MaxAmount
		} else {
			query["amount"] = bson.M{"$lte": filter.MaxAmount}
		}
	}

	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}})

	cursor, err := r.trades.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var trades []*models.Trade
	if err = cursor.All(ctx, &trades); err != nil {
		return nil, err
	}

	return trades, nil
}

// UpdateTradeStatus updates the status of a trade
func (r *MongoRepository) UpdateTradeStatus(ctx context.Context, id string, status models.TradeStatus) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":      status,
			"update_time": time.Now(),
		},
	}

	_, err = r.trades.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

// GetTradeStats retrieves trade statistics
func (r *MongoRepository) GetTradeStats(ctx context.Context, filter *models.TradeFilter) (*models.TradeStats, error) {
	trades, err := r.ListTrades(ctx, filter)
	if err != nil {
		return nil, err
	}

	stats := &models.TradeStats{
		LastTradeTime: time.Now(),
	}

	for _, trade := range trades {
		stats.TotalTrades++
		stats.TotalVolume += trade.Value

		switch trade.Status {
		case models.TradeExecuted:
			stats.SuccessfulTrades++
			stats.TotalFees += trade.Fee
		case models.TradeFailed:
			stats.FailedTrades++
		}
	}

	if stats.TotalTrades > 0 {
		stats.AverageAmount = stats.TotalVolume / float64(stats.TotalTrades)
		if stats.SuccessfulTrades > 0 {
			stats.AverageFee = stats.TotalFees / float64(stats.SuccessfulTrades)
		}
	}

	return stats, nil
}
