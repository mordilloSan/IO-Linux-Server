package websocket

import (
	"context"
	"encoding/json"
	"go-backend/internal/logger"
	"go-backend/internal/session"
	"go-backend/internal/websocket/broadcast"
	"go-backend/internal/websocket/internalstate"
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

var clientsMux = make(chan func())

var (
	channelCounts       = make(map[string]int)
	sessionChannelMap   = make(map[string]string)
	channelCountsMux    sync.Mutex
	dashboardCancelFunc context.CancelFunc
)

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

		internalstate.SetClient(sessionID, &internalstate.Client{
			Send:    client.send,
			Channel: "",
		})

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

			if time.Since(client.lastPong) > 60*time.Second {
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
			logger.Info.Printf("[ws] Read error for %s: %v", client.sessionID, err)
			return
		}

		var msg map[string]string
		if err := json.Unmarshal(data, &msg); err != nil {
			logger.Warning.Printf("[ws] Invalid JSON from %s: %s", client.sessionID, string(data))
			continue
		}

		switch msg["action"] {
		case "subscribe":
			handleSubscription(client.sessionID, msg["channel"])
			client.channel = msg["channel"]
			internalstate.SetClient(client.sessionID, &internalstate.Client{
				Send:    client.send,
				Channel: msg["channel"],
			})
			logger.Info.Printf("[ws] %s subscribed to channel: %s", client.sessionID, client.channel)
		}
	}
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.send)
		close(c.done)
		_ = c.conn.Close()
		handleUnsubscribe(c.sessionID)
		internalstate.RemoveClient(c.sessionID)
	})
}

func handleSubscription(sessionID, channel string) {
	channelCountsMux.Lock()
	defer channelCountsMux.Unlock()

	prev, exists := sessionChannelMap[sessionID]
	if exists && prev == channel {
		return
	}

	if exists {
		channelCounts[prev]--
		if prev == "dashboard" && channelCounts[prev] == 0 {
			logger.Info.Println("[ws] Stopping dashboard broadcaster (last subscriber left)")
			if dashboardCancelFunc != nil {
				dashboardCancelFunc()
				dashboardCancelFunc = nil
			}
		}
	}

	sessionChannelMap[sessionID] = channel
	channelCounts[channel]++

	if channel == "dashboard" && channelCounts[channel] == 1 {
		logger.Info.Println("[ws] Starting dashboard broadcaster (first subscriber)")
		var ctx context.Context
		ctx, dashboardCancelFunc = context.WithCancel(context.Background())
		go broadcast.StartDashboardBroadcaster(ctx)
	}
}

func handleUnsubscribe(sessionID string) {
	channelCountsMux.Lock()
	defer channelCountsMux.Unlock()

	prev, exists := sessionChannelMap[sessionID]
	if !exists {
		return
	}

	delete(sessionChannelMap, sessionID)
	channelCounts[prev]--
	if prev == "dashboard" && channelCounts[prev] == 0 {
		logger.Info.Println("[ws] Stopping dashboard broadcaster (last subscriber left)")
		if dashboardCancelFunc != nil {
			dashboardCancelFunc()
			dashboardCancelFunc = nil
		}
	}
}
