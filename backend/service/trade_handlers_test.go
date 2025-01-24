package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"solmeme-trader/dex"
	"solmeme-trader/models"
	"solmeme-trader/monitoring"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) SaveTrade(ctx context.Context, trade *models.Trade) error {
	args := m.Called(ctx, trade)
	return args.Error(0)
}

func (m *MockRepository) GetTrade(ctx context.Context, id string) (*models.Trade, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Trade), args.Error(1)
}

func (m *MockRepository) ListTrades(ctx context.Context) ([]*models.Trade, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Trade), args.Error(1)
}

func (m *MockRepository) UpdateTrade(ctx context.Context, trade *models.Trade) error {
	args := m.Called(ctx, trade)
	return args.Error(0)
}

func (m *MockRepository) DeleteTrade(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) SavePosition(ctx context.Context, position *models.Position) error {
	args := m.Called(ctx, position)
	return args.Error(0)
}

func (m *MockRepository) GetPosition(ctx context.Context, id string) (*models.Position, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Position), args.Error(1)
}

func (m *MockRepository) ListPositions(ctx context.Context) ([]*models.Position, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Position), args.Error(1)
}

func (m *MockRepository) UpdatePosition(ctx context.Context, position *models.Position) error {
	args := m.Called(ctx, position)
	return args.Error(0)
}

func (m *MockRepository) ClosePosition(ctx context.Context, id string, closePrice float64) error {
	args := m.Called(ctx, id, closePrice)
	return args.Error(0)
}

func (m *MockRepository) GetDailyStats(ctx context.Context, date time.Time) (*models.DailyStats, error) {
	args := m.Called(ctx, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DailyStats), args.Error(1)
}

func (m *MockRepository) GetDailyStatsRange(ctx context.Context, startDate, endDate time.Time) ([]*models.DailyStats, error) {
	args := m.Called(ctx, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DailyStats), args.Error(1)
}

func (m *MockRepository) GetHistoricalMarketData(ctx context.Context, tokenAddress string, limit int) ([]*models.MarketData, error) {
	args := m.Called(ctx, tokenAddress, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.MarketData), args.Error(1)
}

func (m *MockRepository) GetLatestAnalysis(ctx context.Context, tokenAddress string) (*models.MarketAnalysis, error) {
	args := m.Called(ctx, tokenAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.MarketAnalysis), args.Error(1)
}

func TestHandleExecuteTrade(t *testing.T) {
	mockDex := new(MockDexClient)
	mockRepo := new(MockRepository)
	mockMonitor := monitoring.NewMonitor()
	service := &Service{
		repo:      mockRepo,
		dexClient: mockDex,
		monitor:   mockMonitor,
	}

	tests := []struct {
		name           string
		request        TradeRequest
		mockMarketData *dex.MarketData
		mockError      error
		saveError      error
		expectedStatus int
		expectedError  string
	}{
		{
			name: "success",
			request: TradeRequest{
				TokenAddress: "SOL123",
				Type:        "market",
				Side:        "buy",
				Amount:      1.0,
				SlippageBps: 30,
			},
			mockMarketData: &dex.MarketData{
				TokenAddress: "SOL123",
				Price:       100.0,
				Volume24h:   1000.0,
				Timestamp:   time.Now(),
			},
			mockError:      nil,
			saveError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing token",
			request: TradeRequest{
				TokenAddress: "",
				Type:        "market",
				Side:        "buy",
				Amount:      1.0,
				SlippageBps: 30,
			},
			mockMarketData: nil,
			mockError:      nil,
			saveError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid trade parameters",
		},
		{
			name: "invalid amount",
			request: TradeRequest{
				TokenAddress: "SOL123",
				Type:        "market",
				Side:        "buy",
				Amount:      0,
				SlippageBps: 30,
			},
			mockMarketData: nil,
			mockError:      nil,
			saveError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid trade parameters",
		},
		{
			name: "market data error",
			request: TradeRequest{
				TokenAddress: "SOL123",
				Type:        "market",
				Side:        "buy",
				Amount:      1.0,
				SlippageBps: 30,
			},
			mockMarketData: nil,
			mockError:      assert.AnError,
			saveError:      nil,
			expectedStatus: http.StatusOK,
			expectedError:  "Failed to get market data",
		},
		{
			name: "save error",
			request: TradeRequest{
				TokenAddress: "SOL123",
				Type:        "market",
				Side:        "buy",
				Amount:      1.0,
				SlippageBps: 30,
			},
			mockMarketData: &dex.MarketData{
				TokenAddress: "SOL123",
				Price:       100.0,
				Volume24h:   1000.0,
				Timestamp:   time.Now(),
			},
			mockError:      nil,
			saveError:      assert.AnError,
			expectedStatus: http.StatusOK,
			expectedError:  "Failed to save trade",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockMarketData != nil || tt.mockError != nil {
				mockDex.On("GetMarketData", mock.Anything, tt.request.TokenAddress).Return(tt.mockMarketData, tt.mockError).Once()
			}

			if tt.mockMarketData != nil && tt.mockError == nil {
				mockRepo.On("SaveTrade", mock.Anything, mock.AnythingOfType("*models.Trade")).Return(tt.saveError).Once()
			}

			body, err := json.Marshal(tt.request)
			assert.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/v1/trade", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			service.handleExecuteTrade(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response TradeResponse
			err = json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)

			if tt.expectedError != "" {
				assert.Contains(t, response.Error, tt.expectedError)
				assert.False(t, response.Success)
			} else {
				assert.True(t, response.Success)
				assert.Empty(t, response.Error)
				assert.NotEmpty(t, response.TradeID)
			}

			mockDex.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
		})
	}
}
