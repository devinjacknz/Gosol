package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/leonzhao/gosol/backend/trading/analysis/monitoring"
)

func main() {
	// Create monitoring service
	monitoringService := monitoring.NewService(monitoring.Config{
		Port: 2112, // Default Prometheus metrics port
	})

	// Start monitoring service
	go func() {
		if err := monitoringService.Start(); err != nil {
			log.Printf("Error starting monitoring service: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	ctx := context.Background()
	if err := monitoringService.Stop(ctx); err != nil {
		log.Printf("Error stopping monitoring service: %v", err)
	}
}
