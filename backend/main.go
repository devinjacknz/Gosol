package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"solmeme-trader/dex"
	"solmeme-trader/models"
	"solmeme-trader/monitoring"
	"solmeme-trader/repository/mongodb"
	"solmeme-trader/service"
	"solmeme-trader/trading"
	"solmeme-trader/trading/risk"
)

// Config represents application configuration
type Config struct {
	MongoDB struct {
		URI      string
		Database string
	}
	Trading struct {
		MaxPositions    int
		MaxPositionSize float64
		MaxLeverage     float64
		MaxDrawdown     float64
		InitialCapital  float64
		RiskPerTrade    float64
	}
}

// loadConfig loads configuration from environment variables
func loadConfig() (*Config, error) {
	return &Config{
		MongoDB: struct {
			URI      string
			Database string
		}{
			URI:      os.Getenv("MONGODB_URI"),
			Database: os.Getenv("MONGODB_DATABASE"),
		},
		Trading: struct {
			MaxPositions    int
			MaxPositionSize float64
			MaxLeverage     float64
			MaxDrawdown     float64
			InitialCapital  float64
			RiskPerTrade    float64
		}{
			MaxPositions:    5,
			MaxPositionSize: 1000,
			MaxLeverage:     5,
			MaxDrawdown:     0.2,
			InitialCapital:  10000,
			RiskPerTrade:    0.02,
		},
	}, nil
}

func main() {
	// Initialize context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize repository
	repo, err := mongodb.NewRepository(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}

	// Initialize DEX client
	dexClient := dex.NewDexClient()
	marketService := dex.NewMarketDataService(dexClient, 5*time.Minute)

	// Initialize monitoring
	monitor := monitoring.NewMonitor()
	monitoringAPI := monitoring.NewAPI(monitor)

	// Initialize risk manager
	riskConfig := risk.RiskConfig{
		MaxPositions:    cfg.Trading.MaxPositions,
		MaxPositionSize: cfg.Trading.MaxPositionSize,
		MaxLeverage:     cfg.Trading.MaxLeverage,
		MaxDrawdown:     cfg.Trading.MaxDrawdown,
		InitialCapital:  cfg.Trading.InitialCapital,
		RiskPerTrade:    cfg.Trading.RiskPerTrade,
	}
	riskManager := risk.NewRiskManager(riskConfig)

	// Initialize trade executor
	executor := trading.NewExecutor(repo, dexClient, riskManager, monitor)

	// Initialize service
	svc := service.NewService(repo, nil, marketService, monitor)

	// Initialize HTTP server
	mux := http.NewServeMux()
	monitoringAPI.RegisterRoutes(mux)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start HTTP server
	go func() {
		log.Printf("Starting server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Create channels for subscriptions
	marketDataChan := make(chan []byte, 100)
	tradeSignalChan := make(chan []byte, 100)

	// Subscribe to market data
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-marketDataChan:
				if err := svc.ProcessMarketData(ctx, string(msg)); err != nil {
					log.Printf("Failed to process market data: %v", err)
				}
			}
		}
	}()

	// Subscribe to trade signals
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-tradeSignalChan:
				var signal models.TradeSignalMessage
				if err := json.Unmarshal(msg, &signal); err != nil {
					log.Printf("Failed to unmarshal trade signal: %v", err)
					continue
				}

				if err := executor.ExecuteSignal(ctx, &signal); err != nil {
					log.Printf("Failed to execute trade signal: %v", err)
				}
			}
		}
	}()

	// Handle shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Failed to shutdown server: %v", err)
	}

	// Cancel context to stop goroutines
	cancel()
}
