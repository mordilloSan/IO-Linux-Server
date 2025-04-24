// websocket/websocket.go
package websocket

import (
	"go-backend/session"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var clients = make(map[string]*Client)
var clientsMux = make(chan func())

func init() {
	go func() {
		for fn := range clientsMux {
			fn()
		}
	}()
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.channel)
		close(c.done)
		c.conn.Close()
	})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	conn      *websocket.Conn
	channel   chan any
	sessionID string
	done      chan struct{}
	closeOnce sync.Once
}

func RegisterWebSocketRoutes(router *gin.Engine) {
	router.GET("/ws/system", func(c *gin.Context) {
		user, sessionID, ok := session.ValidateFromRequest(c.Request)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		client := &Client{
			conn:      conn,
			channel:   make(chan any),
			sessionID: sessionID,
			done:      make(chan struct{}),
		}

		_ = user

		clientsMux <- func() {
			clients[client.sessionID] = client
		}

		go sendLoop(client)
		receiveLoop(client)
	})
}

func sendLoop(client *Client) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !session.IsValid(client.sessionID) {
				client.conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "Session expired"))
				client.Close()
				return
			}

			select {
			case client.channel <- map[string]any{
				"timestamp": time.Now().Format(time.RFC3339),
				"message":   "system heartbeat",
			}:
			case <-client.done:
				return
			}
		case msg, ok := <-client.channel:
			if !ok {
				return
			}
			client.conn.WriteJSON(msg)
		case <-client.done:
			return
		}
	}
}

func receiveLoop(client *Client) {
	defer client.Close()
	for {
		_, _, err := client.conn.ReadMessage()
		if err != nil {
			return
		}
	}
}

func CloseClientBySession(sessionID string) {
	clientsMux <- func() {
		if client, ok := clients[sessionID]; ok {
			client.conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Logged out"))
			client.Close()
			delete(clients, sessionID)
		}
	}
}
