package main

import (
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
    "github.com/devinjacknz/godydxhyber/backend/pkg/monitoring"
    "github.com/devinjacknz/godydxhyber/backend/pkg/websocket"
)

func main() {
    r := gin.Default()

    // Configure CORS
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
        AllowCredentials: true,
    }))

    // Health check endpoint
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // WebSocket endpoint
    r.GET("/ws", websocket.HandleWebSocket)

    // Setup monitoring
    monitoring.Setup(r)

    // Start server
    r.Run(":8080")
}
