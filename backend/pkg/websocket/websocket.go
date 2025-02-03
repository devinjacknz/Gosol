package websocket

import (
    "encoding/json"
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "github.com/devinjacknz/godydxhyber/backend/pkg/monitoring"
    "net/http"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    Subprotocols:    []string{"13"},
}

func HandleWebSocket(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer conn.Close()

    monitoring.IncrementWSConnections()
    defer monitoring.DecrementWSConnections()

    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            break
        }

        var msg struct {
            Type    string      `json:"type"`
            Channel string      `json:"channel,omitempty"`
            Data    interface{} `json:"data,omitempty"`
        }

        if err := json.Unmarshal(message, &msg); err != nil {
            continue
        }

        switch msg.Type {
        case "subscribe":
            monitoring.RecordMessage("subscribe", msg.Channel)
            response := map[string]string{
                "type":    "subscribed",
                "channel": msg.Channel,
            }
            if data, err := json.Marshal(response); err == nil {
                conn.WriteMessage(websocket.TextMessage, data)
            }
        case "ping":
            response := map[string]string{"type": "pong"}
            if data, err := json.Marshal(response); err == nil {
                conn.WriteMessage(websocket.TextMessage, data)
            }
        default:
            conn.WriteMessage(websocket.TextMessage, message)
        }
    }
}
