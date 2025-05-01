package websocket

import (
	"go-backend/logger"
	"go-backend/session"
	"net/http"
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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.channel)
		close(c.done)
		_ = c.conn.Close()
		logger.Info.Printf("[ws] Closed connection for session %s", c.sessionID)
	})
}

func (c *Client) Send(msg any) bool {
	return safeSend(c.channel, msg)
}

func safeSend(c chan any, msg any) bool {
	select {
	case c <- msg:
		return true
	default:
		return false
	}
}

func RegisterWebSocketRoutes(router *gin.Engine) {
	router.GET("/ws/system", func(c *gin.Context) {
		user, sessionID, ok := session.ValidateFromRequest(c.Request)
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

		logger.Info.Printf("[ws] Client connected: %s (%s)", user.Name, sessionID)

		client := &Client{
			conn:      conn,
			channel:   make(chan any),
			sessionID: sessionID,
			done:      make(chan struct{}),
			lastPong:  time.Now(),
		}

		conn.SetPongHandler(func(appData string) error {
			client.lastPong = time.Now()
			return nil
		})

		clientsMux <- func() {
			clients[client.sessionID] = client
		}

		go sendLoop(client)
		receiveLoop(client)
	})
}

func sendLoop(client *Client) {
	pingTicker := time.NewTicker(2 * time.Second)
	heartbeatTicker := time.NewTicker(5 * time.Second)
	defer pingTicker.Stop()
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-client.done:
			return

		case <-pingTicker.C:
			if !session.IsValid(client.sessionID) {
				logger.Warning.Printf("[ws] Session expired: %s", client.sessionID)
				_ = client.conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "Session expired"))
				client.Close()
				return
			}

			if time.Since(client.lastPong) > 10*time.Second {
				logger.Warning.Printf("[ws] Pong timeout: %s", client.sessionID)
				_ = client.conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, "pong timeout"))
				client.Close()
				return
			}

			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Error.Printf("[ws] Ping failed for %s: %v", client.sessionID, err)
				client.Close()
				return
			}

		case <-heartbeatTicker.C:
			if !client.Send(map[string]any{
				"timestamp": time.Now().Format(time.RFC3339),
				"message":   "system heartbeat",
			}) {
				logger.Warning.Printf("[ws] Heartbeat send failed for %s", client.sessionID)
				client.Close()
				return
			}

		case msg, ok := <-client.channel:
			if !ok {
				return
			}
			if err := client.conn.WriteJSON(msg); err != nil {
				logger.Error.Printf("[ws] WriteJSON failed for %s: %v", client.sessionID, err)
				client.Close()
				return
			}
		}
	}
}

func receiveLoop(client *Client) {
	defer client.Close()
	for {
		_, _, err := client.conn.ReadMessage()
		if err != nil {
			logger.Info.Printf("[ws] ReadMessage closed for %s: %v", client.sessionID, err)
			return
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
