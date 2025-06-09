package websocket

import (
	"encoding/json"
	"go-backend/internal/bridge"
	"go-backend/internal/logger"
	"go-backend/internal/session"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSMessage struct {
	Type      string          `json:"type"`
	RequestID string          `json:"requestId,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

type WSResponse struct {
	Type      string      `json:"type"`
	RequestID string      `json:"requestId,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// --- CHANNEL SUBSCRIPTION INFRASTRUCTURE ---

var (
	channelsMu         sync.Mutex
	channelSubscribers = make(map[string]map[*websocket.Conn]struct{})
)

func subscribe(conn *websocket.Conn, channel string) {
	channelsMu.Lock()
	defer channelsMu.Unlock()
	if channelSubscribers[channel] == nil {
		channelSubscribers[channel] = make(map[*websocket.Conn]struct{})
	}
	channelSubscribers[channel][conn] = struct{}{}
	logger.Infof("WebSocket subscribed to channel: %s", channel)
}

func unsubscribe(conn *websocket.Conn, channel string) {
	channelsMu.Lock()
	defer channelsMu.Unlock()
	if subs, exists := channelSubscribers[channel]; exists && subs != nil {
		delete(subs, conn)
		if len(subs) == 0 {
			delete(channelSubscribers, channel)
		}
	}
	logger.Infof("WebSocket unsubscribed from channel: %s", channel)
}

func removeConnFromAllChannels(conn *websocket.Conn) {
	channelsMu.Lock()
	defer channelsMu.Unlock()
	for channel, subs := range channelSubscribers {
		delete(subs, conn)
		if len(subs) == 0 {
			delete(channelSubscribers, channel)
		}
	}
}

func broadcastToChannel(channel string, msg WSResponse) {
	channelsMu.Lock()
	subs := channelSubscribers[channel]
	channelsMu.Unlock()
	for conn := range subs {
		_ = conn.WriteJSON(msg) // optionally handle write errors/log
	}
}

// --- MAIN HANDLER ---

func WebSocketHandler(c *gin.Context) {
	sess, valid := session.ValidateFromRequest(c.Request)
	if !valid || sess == nil {
		logger.Warnf("WebSocket unauthorized or missing session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Errorf("WS upgrade failed: %v", err)
		return
	}
	defer func() {
		removeConnFromAllChannels(conn)
		conn.Close()
	}()

	logger.Infof("WebSocket connected for user: %s (session: %s, privileged: %v)", sess.User.Name, sess.SessionID, sess.Privileged)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			logger.Warnf("WS disconnect: %v", err)
			break
		}
		logger.Infof("WS got message: %s", msg)
		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			_ = conn.WriteJSON(WSResponse{Type: "error", Error: "Invalid JSON"})
			continue
		}

		switch wsMsg.Type {

		case "subscribe":
			var payload struct {
				Channel string `json:"channel"`
			}
			if err := json.Unmarshal(wsMsg.Payload, &payload); err != nil || payload.Channel == "" {
				_ = conn.WriteJSON(WSResponse{Type: "error", Error: "Missing channel"})
				continue
			}
			subscribe(conn, payload.Channel)
			_ = conn.WriteJSON(WSResponse{Type: "subscribed", Data: payload.Channel})

		case "unsubscribe":
			var payload struct {
				Channel string `json:"channel"`
			}
			if err := json.Unmarshal(wsMsg.Payload, &payload); err != nil || payload.Channel == "" {
				_ = conn.WriteJSON(WSResponse{Type: "error", Error: "Missing channel"})
				continue
			}
			unsubscribe(conn, payload.Channel)
			_ = conn.WriteJSON(WSResponse{Type: "unsubscribed", Data: payload.Channel})

		case "getUserInfo":
			_ = conn.WriteJSON(WSResponse{
				Type:      "getUserInfo_response",
				RequestID: wsMsg.RequestID,
				Data:      sess.User,
			})

		case "bridgeCall":
			var payload struct {
				ReqType string   `json:"reqType"`
				Command string   `json:"command"`
				Args    []string `json:"args"`
			}
			if err := json.Unmarshal(wsMsg.Payload, &payload); err != nil {
				_ = conn.WriteJSON(WSResponse{Type: "error", Error: "Invalid bridgeCall payload"})
				continue
			}
			output, err := bridge.CallWithSession(sess, payload.ReqType, payload.Command, payload.Args)
			if err != nil {
				_ = conn.WriteJSON(WSResponse{
					Type:      wsMsg.Type + "_response",
					RequestID: wsMsg.RequestID,
					Error:     err.Error(),
					Data:      output,
				})
				continue
			}
			_ = conn.WriteJSON(WSResponse{
				Type:      wsMsg.Type + "_response",
				RequestID: wsMsg.RequestID,
				Data:      output,
			})

		default:
			_ = conn.WriteJSON(WSResponse{Type: "error", Error: "Unknown message type"})
		}
	}
}
