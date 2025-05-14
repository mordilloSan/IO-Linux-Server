// internal/websocket/control/control.go
package control

import (
	"go-backend/internal/websocket/internalstate"
	"log"

	"github.com/gorilla/websocket"
)

// CloseClientBySession safely closes a client session
func CloseClientBySession(sessionID string) {
	internalstate.WithClients(func(clients map[string]*internalstate.Client) {
		if client, ok := clients[sessionID]; ok {
			select {
			case client.Send <- websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Logged out"):
			default:
			}
			internalstate.RemoveClient(sessionID)
			log.Printf("[ws] Closed session: %s", sessionID)
		}
	})
}
