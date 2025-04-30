package websocket

import (
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
		c.conn.Close()
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
			lastPong:  time.Now(),
		}

		conn.SetPongHandler(func(appData string) error {
			client.lastPong = time.Now()
			return nil
		})

		_ = user

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
				client.conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "Session expired"))
				client.Close()
				return
			}

			if time.Since(client.lastPong) > 10*time.Second {
				client.conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, "pong timeout"))
				client.Close()
				return
			}

			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				client.Close()
				return
			}

		case <-heartbeatTicker.C:
			if !client.Send(map[string]any{
				"timestamp": time.Now().Format(time.RFC3339),
				"message":   "system heartbeat",
			}) {
				client.Close()
				return
			}

		case msg, ok := <-client.channel:
			if !ok {
				return
			}
			err := client.conn.WriteJSON(msg)
			if err != nil {
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
