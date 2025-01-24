package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"solmeme-trader/dex"
	"solmeme-trader/monitoring"
	"solmeme-trader/repository"
	"solmeme-trader/repository/mongodb"
	"solmeme-trader/service"
)

func main() {
	ctx := context.Background()

	// Initialize MongoDB repository
	repo, err := mongodb.NewRepository(ctx, repository.Options{
		URI:            os.Getenv("MONGODB_URI"),
		Database:       os.Getenv("MONGODB_DATABASE"),
		Username:       os.Getenv("MONGODB_USERNAME"),
		Password:       os.Getenv("MONGODB_PASSWORD"),
		ConnectTimeout: 5 * time.Second,
		Timeout:       10 * time.Second,
		MaxConnections: 100,
		MinConnections: 10,
	})
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Initialize DEX clients
	dexClient := dex.NewDexClient(
		os.Getenv("RAYDIUM_API_URL"),
		os.Getenv("JUPITER_API_URL"),
	)

	// Initialize monitoring
	monitor := monitoring.NewMonitor()
	monitorAPI := monitoring.NewAPI(monitor)

	// Initialize market data service
	marketService := dexClient.GetMarketService()

	// Initialize services
	svc := service.NewService(repo, dexClient, marketService, monitor)

	// Create router
	mux := http.NewServeMux()

	// Register API routes
	svc.Routes(mux)

	// Register monitoring routes
	monitorAPI.RegisterRoutes(mux)

	// Create server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
