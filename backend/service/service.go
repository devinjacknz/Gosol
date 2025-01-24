package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"solmeme-trader/dex"
	"solmeme-trader/models"
	"solmeme-trader/monitoring"
	"solmeme-trader/repository"
)

// Service handles business logic
type Service struct {
	repo          repository.Repository
	db            interface{} // For pubsub
	marketService *dex.MarketDataService
	monitor       *monitoring.Monitor
}

// NewService creates a new service
func NewService(repo repository.Repository, db interface{}, marketService *dex.MarketDataService, monitor *monitoring.Monitor) *Service {
	return &Service{
		repo:          repo,
		db:            db,
		marketService: marketService,
		monitor:       monitor,
	}
}

// ProcessMarketData processes market data
func (s *Service) ProcessMarketData(ctx context.Context, msg string) error {
	start := time.Now()

	// Parse market data
	var data models.MarketData
	if err := json.Unmarshal([]byte(msg), &data); err != nil {
		s.monitor.RecordEvent(ctx, monitoring.Event{
			Type:      monitoring.MetricMarketData,
			Severity:  monitoring.SeverityError,
			Message:   "Failed to parse market data",
			Details:   map[string]interface{}{"error": err.Error()},
			Token:     &data.TokenAddress,
			Timestamp: time.Now(),
		})
		return fmt.Errorf("failed to parse market data: %w", err)
	}

	// Save market data
	if err := s.repo.SaveMarketData(ctx, &data); err != nil {
		s.monitor.RecordEvent(ctx, monitoring.Event{
			Type:      monitoring.MetricMarketData,
			Severity:  monitoring.SeverityError,
			Message:   "Failed to save market data",
			Details:   map[string]interface{}{"error": err.Error()},
			Token:     &data.TokenAddress,
			Timestamp: time.Now(),
		})
		return fmt.Errorf("failed to save market data: %w", err)
	}

	// Update trade status if needed
	trades, err := s.repo.GetTradesByToken(ctx, data.TokenAddress)
	if err != nil {
		s.monitor.RecordEvent(ctx, monitoring.Event{
			Type:      monitoring.MetricTrading,
			Severity:  monitoring.SeverityError,
			Message:   "Failed to get trades",
			Details:   map[string]interface{}{"error": err.Error()},
			Token:     &data.TokenAddress,
			Timestamp: time.Now(),
		})
		return fmt.Errorf("failed to get trades: %w", err)
	}

	for _, trade := range trades {
		if trade.Status == models.TradeStatusPending {
			reason := "Market conditions met"
			if err := s.repo.UpdateTradeStatus(ctx, trade.ID, models.TradeStatusCompleted, &reason); err != nil {
				s.monitor.RecordEvent(ctx, monitoring.Event{
					Type:      monitoring.MetricTrading,
					Severity:  monitoring.SeverityError,
					Message:   "Failed to update trade status",
					Details:   map[string]interface{}{"error": err.Error()},
					Token:     &data.TokenAddress,
					Timestamp: time.Now(),
				})
			}
		}
	}

	// Record successful processing
	s.monitor.RecordMarketData(ctx, &data, time.Since(start))

	return nil
}

// GetMarketData gets market data for a token
func (s *Service) GetMarketData(ctx context.Context, tokenAddress string) (*models.MarketData, error) {
	return s.repo.GetLatestMarketData(ctx, tokenAddress)
}

// GetMarketDataHistory gets market data history for a token
func (s *Service) GetMarketDataHistory(ctx context.Context, tokenAddress string, start, end time.Time) ([]*models.MarketData, error) {
	return s.repo.GetMarketDataHistory(ctx, tokenAddress, start, end)
}

// GetTradingStats gets trading statistics
func (s *Service) GetTradingStats(ctx context.Context, tokenAddress string) (*models.DailyStats, error) {
	// Get trade stats
	totalTrades, winningTrades, profitFactor, err := s.repo.GetTradeStats(ctx, tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get trade stats: %w", err)
	}

	// Calculate current profit
	currentProfit, err := s.repo.CalculateCurrentProfit(ctx, tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate profit: %w", err)
	}

	// Get latest market data
	marketData, err := s.repo.GetLatestMarketData(ctx, tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get market data: %w", err)
	}

	// Create stats
	stats := &models.DailyStats{
		Date:          time.Now(),
		TotalTrades:   totalTrades,
		WinningTrades: winningTrades,
		RealizedPnL:   currentProfit,
		Volume:        marketData.Volume24h,
		WinRate:       float64(winningTrades) / float64(totalTrades),
		ProfitFactor:  profitFactor,
	}

	return stats, nil
}

// GetEvents gets monitoring events
func (s *Service) GetEvents(ctx context.Context, eventType string, start, end time.Time) ([]*monitoring.Event, error) {
	return s.repo.GetEvents(ctx, eventType, start, end)
}

// GetEventsByToken gets monitoring events for a token
func (s *Service) GetEventsByToken(ctx context.Context, tokenAddress string, start, end time.Time) ([]*monitoring.Event, error) {
	return s.repo.GetEventsByToken(ctx, tokenAddress, start, end)
}
