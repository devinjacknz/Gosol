package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/devinjacknz/godydxhyber/backend/logger"
	"github.com/devinjacknz/godydxhyber/backend/middleware"
)

var log = logger.NewLogger()

func main() {
	log.Info("Starting backend service...")

	// Initialize context
	ctx := context.Background()
	
	// Initialize Redis client if URL is provided
	var rdb *redis.Client
	redisURL := os.Getenv("REDIS_URL")
	if redisURL != "" {
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Warnf("Invalid Redis URL: %v, using default configuration", err)
			opt = &redis.Options{
				Addr: "redis:6379",
				DB:   0,
			}
		}
		rdb = redis.NewClient(opt)
		
		// Ping Redis to verify connection with timeout
		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		if err := rdb.Ping(ctx).Err(); err != nil {
			log.Warnf("Failed to connect to Redis: %v", err)
			rdb = nil
		} else {
			log.Info("Successfully connected to Redis")
		}
	} else {
		log.Info("No Redis URL provided, running without Redis")
	}
	
	if rdb != nil {
		defer rdb.Close()
	}

	// Initialize Gin router with custom middleware
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	
	// Use custom middleware
	r.Use(middleware.RecoverMiddleware())
	r.Use(middleware.DebugMiddleware())
	r.Use(gin.Logger())
	
	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	// Debug endpoints
	debug := r.Group("/debug")
	{
		debug.GET("/stats", middleware.RequestStatsEndpoint)
		debug.GET("/memory", middleware.MemoryStatsEndpoint)
		debug.GET("/stack", middleware.StackTraceEndpoint)
		debug.GET("/pprof/*any", middleware.PprofEndpoint)
		debug.GET("/goroutines", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"count": runtime.NumGoroutine(),
			})
		})
		debug.GET("/env", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"env":     os.Environ(),
				"version": runtime.Version(),
				"arch":    runtime.GOARCH,
				"os":      runtime.GOOS,
			})
		})
	}

	// Health check endpoint that handles both GET and HEAD
	r.Any("/health", func(c *gin.Context) {
		if c.Request.Method == "HEAD" {
			c.Status(http.StatusOK)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"redis":  "connected",
		})
	})

	// WebSocket upgrader
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in development
		},
	}

	// WebSocket endpoint with market data streaming
	r.GET("/ws/:symbol", func(c *gin.Context) {
		symbol := c.Param("symbol")
		log.Infof("New WebSocket connection request for symbol: %s", symbol)

		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true // Allow all origins in development
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Errorf("Failed to upgrade connection: %v", err)
			return
		}
		defer conn.Close()

		// Configure WebSocket
		conn.SetReadLimit(512) // Set message size limit
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		// Send initial market data
		marketData := map[string]interface{}{
			"type": "market_data",
			"data": map[string]interface{}{
				"symbol": symbol,
				"price": 50000.0,
				"volume": 1000.0,
				"timestamp": time.Now().Unix(),
				"funding_rate": 0.0001,
				"next_funding_time": time.Now().Add(8 * time.Hour).Unix(),
				"orderBook": map[string][][]float64{
					"asks": {{50100.0, 1.5}, {50200.0, 2.0}, {50300.0, 1.0}},
					"bids": {{49900.0, 2.0}, {49800.0, 3.0}, {49700.0, 1.5}},
				},
				"trades": []map[string]interface{}{
					{
						"id": "t1",
						"price": 50000.0,
						"amount": 0.5,
						"side": "buy",
						"timestamp": time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
					},
				},
				"change24h": 2.5,
			},
		}

		if err := conn.WriteJSON(marketData); err != nil {
			log.Errorf("Failed to send initial market data: %v", err)
			return
		}

		log.Infof("WebSocket connection established for symbol: %s", symbol)

		// Start ping ticker
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		// Start market data update ticker
		marketDataTicker := time.NewTicker(1 * time.Second)
		defer marketDataTicker.Stop()

		// Handle WebSocket connection
		go func() {
			for {
				select {
				case <-ticker.C:
					if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
						log.Warnf("Failed to write ping message: %v", err)
						return
					}
				case <-marketDataTicker.C:
					marketData := map[string]interface{}{
						"type": "market_data",
						"data": map[string]interface{}{
							"symbol": symbol,
							"price": 50000.0 + float64(time.Now().Unix()%100-50),
							"volume": 1000.0 + float64(time.Now().Unix()%200),
							"timestamp": time.Now().Unix(),
							"funding_rate": 0.0001,
							"next_funding_time": time.Now().Add(8 * time.Hour).Unix(),
							"orderBook": map[string][][]float64{
								"asks": {{50100.0, 1.5}, {50200.0, 2.0}, {50300.0, 1.0}},
								"bids": {{49900.0, 2.0}, {49800.0, 3.0}, {49700.0, 1.5}},
							},
							"trades": []map[string]interface{}{
								{
									"id": "t" + time.Now().Format("20060102150405"),
									"price": 50000.0 + float64(time.Now().Unix()%100-50),
									"amount": 0.5,
									"side": "buy",
									"timestamp": time.Now().Format(time.RFC3339),
								},
							},
							"change24h": 2.5,
						},
					}
					if err := conn.WriteJSON(marketData); err != nil {
						log.Errorf("Failed to send market data update: %v", err)
						return
					}
				}
			}
		}()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Errorf("WebSocket error: %v", err)
				} else {
					log.Infof("WebSocket connection closed: %v", err)
				}
				break
			}

			var msg map[string]interface{}
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Errorf("Failed to parse message: %v", err)
				continue
			}

			if msg["type"] == "subscribe" {
				response := map[string]interface{}{
					"type": "subscribed",
					"symbol": symbol,
				}
				if err := conn.WriteJSON(response); err != nil {
					log.Errorf("Failed to send subscription confirmation: %v", err)
					break
				}
				log.Infof("Client subscribed to %s", symbol)
			}
		}
	})

	// Market data endpoints with versioned API group
	v1 := r.Group("/api/v1")
	{
		v1.GET("/market-data/:symbol", func(c *gin.Context) {
			symbol := c.Param("symbol")
			
			// Generate dynamic market data
			marketData := map[string]interface{}{
				"symbol": symbol,
				"price": 50000.0 + float64(time.Now().Unix()%100-50),
				"volume": 1000.0 + float64(time.Now().Unix()%200),
				"timestamp": time.Now().Unix(),
				"funding_rate": 0.0001,
				"next_funding_time": time.Now().Add(8 * time.Hour).Unix(),
				"orderBook": map[string][][]float64{
					"asks": {{50100.0, 1.5}, {50200.0, 2.0}, {50300.0, 1.0}},
					"bids": {{49900.0, 2.0}, {49800.0, 3.0}, {49700.0, 1.5}},
				},
				"trades": []map[string]interface{}{
					{
						"id": "t" + time.Now().Format("20060102150405"),
						"price": 50000.0 + float64(time.Now().Unix()%100-50),
						"amount": 0.5,
						"side": "buy",
						"timestamp": time.Now().Format(time.RFC3339),
					},
				},
				"change24h": 2.5,
			}
			
			c.JSON(http.StatusOK, marketData)
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Infof("Backend service listening on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
