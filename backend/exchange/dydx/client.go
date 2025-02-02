package dydx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

// Client defines the interface for interacting with dYdX
type Client interface {
	// Market Data
	GetMarkets(ctx context.Context) ([]Market, error)
	GetOrderbook(ctx context.Context, symbol string) (*Orderbook, error)
	GetTrades(ctx context.Context, symbol string, limit int) ([]Trade, error)
	GetFundingRate(ctx context.Context, symbol string) (*FundingRate, error)

	// Trading
	CreateOrder(ctx context.Context, req CreateOrderRequest) (*Order, error)
	CancelOrder(ctx context.Context, orderID string) error
	GetOrder(ctx context.Context, orderID string) (*Order, error)
	GetOpenOrders(ctx context.Context) ([]Order, error)

	// Account
	GetPositions(ctx context.Context) ([]Position, error)
	GetBalance(ctx context.Context) (*Balance, error)
	GetAccount(ctx context.Context) (*Account, error)
	GetLeverage(ctx context.Context, symbol string) (int, error)
	SetLeverage(ctx context.Context, symbol string, leverage int) error

	// WebSocket
	SubscribeOrderbook(symbol string, ch chan<- Orderbook) error
	SubscribeTrades(symbol string, ch chan<- Trade) error
	SubscribePositions(ch chan<- Position) error
	UnsubscribeAll() error

	// Close closes the client and all connections
	Close() error
}

// DefaultClient implements the Client interface
type DefaultClient struct {
	baseURL    string
	wsURL      string
	apiKey     string
	apiSecret  string
	passphrase string
	httpClient *http.Client
	wsConn     *websocket.Conn
	limiter    *rate.Limiter

	// Subscriptions
	subscriptions map[string][]chan interface{}
	subMutex      sync.RWMutex

	// Connection management
	done      chan struct{}
	reconnect chan struct{}
}

// ClientOption defines options for creating a new client
type ClientOption func(*DefaultClient)

// NewClient creates a new dYdX client
func NewClient(apiKey, apiSecret, passphrase string, opts ...ClientOption) Client {
	client := &DefaultClient{
		baseURL:       "https://api.dydx.exchange",
		wsURL:         "wss://api.dydx.exchange/v3/ws",
		apiKey:        apiKey,
		apiSecret:     apiSecret,
		passphrase:    passphrase,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
		limiter:       rate.NewLimiter(rate.Limit(5), 10), // 5 requests/second, burst of 10
		subscriptions: make(map[string][]chan interface{}),
		done:          make(chan struct{}),
		reconnect:     make(chan struct{}, 1),
	}

	for _, opt := range opts {
		opt(client)
	}

	go client.maintainConnection()
	return client
}

// WithBaseURL sets the base URL for the REST API
func WithBaseURL(url string) ClientOption {
	return func(c *DefaultClient) {
		c.baseURL = url
	}
}

// WithWSURL sets the WebSocket URL
func WithWSURL(url string) ClientOption {
	return func(c *DefaultClient) {
		c.wsURL = url
	}
}

// WithRateLimit sets the rate limiter configuration
func WithRateLimit(rps float64, burst int) ClientOption {
	return func(c *DefaultClient) {
		c.limiter = rate.NewLimiter(rate.Limit(rps), burst)
	}
}

func (c *DefaultClient) maintainConnection() {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			if err := c.ping(); err != nil {
				c.reconnect <- struct{}{}
			}
		case <-c.reconnect:
			c.wsConn.Close()
			if err := c.connect(); err != nil {
				time.Sleep(time.Second * 5)
				c.reconnect <- struct{}{}
			}
		}
	}
}

func (c *DefaultClient) connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.wsURL, nil)
	if err != nil {
		return fmt.Errorf("websocket dial error: %w", err)
	}
	c.wsConn = conn

	// Authenticate WebSocket connection
	if err := c.authenticate(); err != nil {
		c.wsConn.Close()
		return fmt.Errorf("websocket authentication error: %w", err)
	}

	go c.readPump()
	return nil
}

func (c *DefaultClient) authenticate() error {
	timestamp := time.Now().Unix()
	signature := c.sign(fmt.Sprintf("%d", timestamp))

	msg := struct {
		Type       string `json:"type"`
		APIKey     string `json:"apiKey"`
		Signature  string `json:"signature"`
		Timestamp  int64  `json:"timestamp"`
		Passphrase string `json:"passphrase"`
	}{
		Type:       "subscribe",
		APIKey:     c.apiKey,
		Signature:  signature,
		Timestamp:  timestamp,
		Passphrase: c.passphrase,
	}

	return c.wsConn.WriteJSON(msg)
}

func (c *DefaultClient) readPump() {
	for {
		_, message, err := c.wsConn.ReadMessage()
		if err != nil {
			c.reconnect <- struct{}{}
			return
		}

		var msg struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		c.subMutex.RLock()
		channels, exists := c.subscriptions[msg.Type]
		c.subMutex.RUnlock()

		if !exists {
			continue
		}

		for _, ch := range channels {
			select {
			case ch <- msg.Data:
			default:
			}
		}
	}
}

func (c *DefaultClient) ping() error {
	return c.wsConn.WriteMessage(websocket.PingMessage, nil)
}

func (c *DefaultClient) subscribe(topic string, ch interface{}) error {
	c.subMutex.Lock()
	defer c.subMutex.Unlock()

	if c.subscriptions[topic] == nil {
		c.subscriptions[topic] = make([]chan interface{}, 0)
	}
	c.subscriptions[topic] = append(c.subscriptions[topic], ch.(chan interface{}))

	msg := struct {
		Type string `json:"type"`
		Sub  string `json:"sub"`
	}{
		Type: "subscribe",
		Sub:  topic,
	}
	return c.wsConn.WriteJSON(msg)
}

// Close closes the client and all connections
func (c *DefaultClient) Close() error {
	close(c.done)
	return c.wsConn.Close()
}

// sign generates a signature for authentication
func (c *DefaultClient) sign(message string) string {
	// TODO: Implement dYdX signature generation
	return ""
}
