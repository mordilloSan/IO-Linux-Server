package websocket

import (
	"encoding/json"
	"go-backend/internal/logger"
	"go-backend/internal/session"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn      *websocket.Conn
	send      chan any
	sessionID string
	channel   string
	done      chan struct{}
	closeOnce sync.Once
	lastPong  time.Time
}

var clients = make(map[string]*Client)
var clientsMux = make(chan func())

func init() {
	go func() {
		for fn := range clientsMux {
			fn()
		}
	}()
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func RegisterWebSocketRoutes(router *gin.Engine) {
	router.GET("/ws", func(c *gin.Context) {
		_, sessionID, ok := session.ValidateFromRequest(c.Request)
		if !ok {
			logger.Warning.Println("[ws] Unauthorized connection attempt")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Error.Printf("[ws] Upgrade failed: %v", err)
			return
		}
		defer conn.Close()

		client := &Client{
			conn:      conn,
			send:      make(chan any),
			sessionID: sessionID,
			channel:   "",
			done:      make(chan struct{}),
			lastPong:  time.Now(),
		}

		conn.SetPongHandler(func(string) error {
			client.lastPong = time.Now()
			return nil
		})

		clientsMux <- func() {
			clients[sessionID] = client
		}

		go sendLoop(client)
		receiveLoop(client)
	})
}

func sendLoop(client *Client) {
	pingTicker := time.NewTicker(2 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case <-client.done:
			return

		case <-pingTicker.C:
			if !session.IsValid(client.sessionID) {
				client.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "Session expired"))
				client.Close()
				return
			}

			if time.Since(client.lastPong) > 10*time.Second {
				client.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Pong timeout"))
				client.Close()
				return
			}

			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				client.Close()
				return
			}

		case msg, ok := <-client.send:
			if !ok {
				return
			}
			if err := client.conn.WriteJSON(msg); err != nil {
				client.Close()
				return
			}
		}
	}
}

func receiveLoop(client *Client) {
	defer client.Close()
	for {
		_, data, err := client.conn.ReadMessage()
		if err != nil {
			return
		}
		logger.Info.Printf("[ws] Received raw: %s", string(data))
		var msg map[string]string
		if err := json.Unmarshal(data, &msg); err != nil {
			logger.Warning.Printf("[ws] Invalid JSON from %s: %s", client.sessionID, string(data))
			continue
		}

		switch msg["action"] {
		case "subscribe":
			client.channel = msg["channel"]
			logger.Info.Printf("[ws] %s subscribed to channel: %s", client.sessionID, client.channel)
		}

	}
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.send)
		close(c.done)
		_ = c.conn.Close()
		clientsMux <- func() {
			delete(clients, c.sessionID)
		}
	})
}

func BroadcastToChannel(channel string, payload any) {
	clientsMux <- func() {
		for _, client := range clients {
			if client.channel == channel {
				select {
				case client.send <- payload:
				default:
					logger.Warning.Printf("[ws] Dropped message to %s (slow consumer)", client.sessionID)
				}
			}
		}
	}
}

func CloseClientBySession(sessionID string) {
	clientsMux <- func() {
		if client, ok := clients[sessionID]; ok {
			logger.Info.Printf("[ws] Closing client due to logout: %s", sessionID)
			_ = client.conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Logged out"))
			client.Close()
			delete(clients, sessionID)
		}
	}
}

// This function is called in a separate goroutine to broadcast
// a message to the "dashboard" channel every 5 seconds
// use --> go websocket.Tester()
func Tester() {
	for {
		logger.Info.Println("[ws] Broadcasting to dashboard channel")
		time.Sleep(5 * time.Second)
		BroadcastToChannel("dashboard", map[string]any{
			"type":      "heartbeat",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

// RunBroadcasterMultiWithIntervals allows broadcasting multiple named sources to a channel,
// each with their own interval.
func RunBroadcasterMultiWithIntervals(channel string, funcs map[string]struct {
	Fn       func() (any, error)
	Interval time.Duration
}) {
	for dataType, cfg := range funcs {
		go func(dataType string, cfg struct {
			Fn       func() (any, error)
			Interval time.Duration
		}) {
			ticker := time.NewTicker(cfg.Interval)
			defer ticker.Stop()

			for range ticker.C {
				payload, err := cfg.Fn()
				if err != nil {
					logger.Error.Printf("[ws] Error fetching %s: %v", dataType, err)
					continue
				}
				logger.Info.Printf("[ws] Sending %s update on channel '%s'", dataType, channel)

				BroadcastToChannel(channel, map[string]interface{}{
					"type":    dataType,
					"channel": channel,
					"payload": payload,
				})
			}
		}(dataType, cfg)
	}
}
