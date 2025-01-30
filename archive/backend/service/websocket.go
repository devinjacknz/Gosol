package service

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocketService handles WebSocket connections
type WebSocketService struct {
	// Upgrader for WebSocket connections
	upgrader websocket.Upgrader

	// Connected clients
	clients map[*websocket.Conn]bool

	// Broadcast channel
	broadcast chan []byte

	// Mutex for thread safety
	mu sync.RWMutex
}

// NewWebSocketService creates a new WebSocket service
func NewWebSocketService() *WebSocketService {
	return &WebSocketService{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte),
	}
}

// Start starts the WebSocket service
func (s *WebSocketService) Start() {
	go s.handleBroadcast()
}

// HandleConnection handles a new WebSocket connection
func (s *WebSocketService) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Register client
	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	// Handle client messages
	go s.handleClient(conn)
}

// handleClient handles messages from a client
func (s *WebSocketService) handleClient(conn *websocket.Conn) {
	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
		conn.Close()
	}()

	for {
		// Read message
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle message
		s.handleMessage(conn, message)
	}
}

// handleMessage processes a message from a client
func (s *WebSocketService) handleMessage(conn *websocket.Conn, message []byte) {
	var msg struct {
		Type    string         `json:"type"`
		Channel string         `json:"channel"`
		Data    map[string]any `json:"data"`
	}

	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Failed to parse message: %v", err)
		return
	}

	switch msg.Type {
	case "subscribe":
		// Handle subscription
		s.handleSubscription(conn, msg.Channel, msg.Data)
	case "unsubscribe":
		// Handle unsubscription
		s.handleUnsubscription(conn, msg.Channel)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// handleSubscription handles a subscription request
func (s *WebSocketService) handleSubscription(conn *websocket.Conn, channel string, data map[string]any) {
	// TODO: Implement subscription logic based on channel type
	// Example: subscribe to market data, klines, order book, etc.
}

// handleUnsubscription handles an unsubscription request
func (s *WebSocketService) handleUnsubscription(conn *websocket.Conn, channel string) {
	// TODO: Implement unsubscription logic
}

// handleBroadcast handles broadcasting messages to all clients
func (s *WebSocketService) handleBroadcast() {
	for message := range s.broadcast {
		s.mu.RLock()
		for conn := range s.clients {
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("Failed to send message: %v", err)
				conn.Close()
				delete(s.clients, conn)
			}
		}
		s.mu.RUnlock()
	}
}

// Broadcast sends a message to all connected clients
func (s *WebSocketService) Broadcast(message []byte) {
	s.broadcast <- message
}

// BroadcastJSON sends a JSON message to all connected clients
func (s *WebSocketService) BroadcastJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	s.Broadcast(data)
	return nil
}
