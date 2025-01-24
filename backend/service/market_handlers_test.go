package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"solmeme-trader/dex"
	"solmeme-trader/monitoring"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDexClient is a mock implementation of DexClient
type MockDexClient struct {
	mock.Mock
}

func (m *MockDexClient) GetMarketData(ctx context.Context, tokenAddress string) (*dex.MarketData, error) {
	args := m.Called(ctx, tokenAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dex.MarketData), args.Error(1)
}

func (m *MockDexClient) GetOrderBook(ctx context.Context, tokenAddress string) (*dex.OrderBook, error) {
	args := m.Called(ctx, tokenAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dex.OrderBook), args.Error(1)
}

func (m *MockDexClient) GetQuote(ctx context.Context, tokenAddress string, amount float64) (float64, error) {
	args := m.Called(ctx, tokenAddress, amount)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockDexClient) ExecuteTrade(ctx context.Context, tokenAddress string, amount float64, side string) error {
	args := m.Called(ctx, tokenAddress, amount, side)
	return args.Error(0)
}

func (m *MockDexClient) CancelTrade(ctx context.Context, tradeID string) error {
	args := m.Called(ctx, tradeID)
	return args.Error(0)
}

func (m *MockDexClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestHandleGetMarketData(t *testing.T) {
	mockDex := new(MockDexClient)
	mockMonitor := monitoring.NewMonitor()
	service := &Service{
		dexClient: mockDex,
		monitor:   mockMonitor,
	}

	tests := []struct {
		name           string
		tokenAddress   string
		mockData       *dex.MarketData
		mockError      error
		expectedStatus int
		expectedError  string
	}{
		{
			name:         "success",
			tokenAddress: "SOL123",
			mockData: &dex.MarketData{
				TokenAddress: "SOL123",
				Price:       100.0,
				Volume24h:   1000.0,
				Timestamp:   time.Now(),
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing token",
			tokenAddress:   "",
			mockData:       nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Token address is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockData != nil {
				mockDex.On("GetMarketData", mock.Anything, tt.tokenAddress).Return(tt.mockData, tt.mockError)
			}

			req := httptest.NewRequest("GET", "/api/v1/market/data?token="+tt.tokenAddress, nil)
			w := httptest.NewRecorder()

			service.handleGetMarketData(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response MarketDataResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
			}

			mockDex.AssertExpectations(t)
		})
	}
}

func TestHandleGetOrderBook(t *testing.T) {
	mockDex := new(MockDexClient)
	mockMonitor := monitoring.NewMonitor()
	service := &Service{
		dexClient: mockDex,
		monitor:   mockMonitor,
	}

	tests := []struct {
		name           string
		tokenAddress   string
		mockData       *dex.OrderBook
		mockError      error
		expectedStatus int
		expectedError  string
	}{
		{
			name:         "success",
			tokenAddress: "SOL123",
			mockData: &dex.OrderBook{
				Bids: []dex.OrderBookItem{{Price: 100.0, Amount: 1.0}},
				Asks: []dex.OrderBookItem{{Price: 101.0, Amount: 1.0}},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing token",
			tokenAddress:   "",
			mockData:       nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Token address is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockData != nil {
				mockDex.On("GetOrderBook", mock.Anything, tt.tokenAddress).Return(tt.mockData, tt.mockError)
			}

			req := httptest.NewRequest("GET", "/api/v1/market/orderbook?token="+tt.tokenAddress, nil)
			w := httptest.NewRecorder()

			service.handleGetOrderBook(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response MarketDataResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
			}

			mockDex.AssertExpectations(t)
		})
	}
}

func TestHandleGetQuote(t *testing.T) {
	mockDex := new(MockDexClient)
	mockMonitor := monitoring.NewMonitor()
	service := &Service{
		dexClient: mockDex,
		monitor:   mockMonitor,
	}

	tests := []struct {
		name           string
		tokenAddress   string
		amount         string
		mockPrice      float64
		mockError      error
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "success",
			tokenAddress:   "SOL123",
			amount:        "1.0",
			mockPrice:     100.0,
			mockError:     nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing token",
			tokenAddress:   "",
			amount:        "1.0",
			mockPrice:     0,
			mockError:     nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Token address is required",
		},
		{
			name:           "missing amount",
			tokenAddress:   "SOL123",
			amount:        "",
			mockPrice:     0,
			mockError:     nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Amount is required",
		},
		{
			name:           "invalid amount",
			tokenAddress:   "SOL123",
			amount:        "invalid",
			mockPrice:     0,
			mockError:     nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid amount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockPrice > 0 {
				mockDex.On("GetQuote", mock.Anything, tt.tokenAddress, mock.AnythingOfType("float64")).Return(tt.mockPrice, tt.mockError)
			}

			url := "/api/v1/market/quote?token=" + tt.tokenAddress
			if tt.amount != "" {
				url += "&amount=" + tt.amount
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			service.handleGetQuote(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response MarketDataResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
			}

			mockDex.AssertExpectations(t)
		})
	}
}
