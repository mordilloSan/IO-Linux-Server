package websocket

import (
	"encoding/json"
	"go-backend/internal/logger"
	"go-backend/internal/session"
	"net/http"

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

func WebSocketHandler(c *gin.Context) {
	user, sessionID, valid, privileged := session.ValidateFromRequest(c.Request)
	if !valid {
		logger.Warning.Printf("WebSocket unauthorized: %s", sessionID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error.Printf("WS upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	logger.Info.Printf("WebSocket connected for user: %s (session: %s, privileged: %v)", user.Name, sessionID, privileged)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			logger.Warning.Printf("WS disconnect: %v", err)
			break
		}
		logger.Info.Printf("WS got message: %s", msg)
		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			_ = conn.WriteJSON(WSResponse{
				Type:  "error",
				Error: "Invalid JSON",
			})
			continue
		}

		var resp WSResponse
		resp.Type = wsMsg.Type + "_response"
		resp.RequestID = wsMsg.RequestID

		// Route by message type
		switch wsMsg.Type {

		case "getUserInfo":
			resp.Data = user

		// Example: Add other API routes here
		// case "getNetworkInterfaces":
		//     data, err := network.GetInterfaces(user, sessionID, privileged)
		//     if err != nil {
		//         resp.Error = err.Error()
		//     } else {
		//         resp.Data = data
		//     }

		default:
			resp.Type = "error"
			resp.Error = "Unknown message type"
		}

		if err := conn.WriteJSON(resp); err != nil {
			logger.Warning.Printf("WS write error: %v", err)
			break
		}
	}
}
