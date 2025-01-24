package service

import (
	"net/http"

	"solmeme-trader/dex"
	"solmeme-trader/monitoring"
	"solmeme-trader/repository"
)

// Service handles business logic
type Service struct {
	repo      repository.Repository
	dexClient dex.DexClient
	monitor   monitoring.IMonitor
}

// NewService creates a new service
func NewService(repo repository.Repository, dexClient dex.DexClient, monitor monitoring.IMonitor) *Service {
	return &Service{
		repo:      repo,
		dexClient: dexClient,
		monitor:   monitor,
	}
}

// Routes registers all service routes
func (s *Service) Routes(mux *http.ServeMux) {
	// Trade routes
	mux.HandleFunc("/api/v1/trade", s.handleExecuteTrade)

	// Market data routes
	mux.HandleFunc("/api/v1/market/data", s.handleGetMarketData)
	mux.HandleFunc("/api/v1/market/orderbook", s.handleGetOrderBook)
	mux.HandleFunc("/api/v1/market/quote", s.handleGetQuote)
}
