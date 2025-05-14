package utils

import (
	"go-backend/internal/websocket/internalstate"
)

// BroadcastToChannel sends payload to all clients subscribed to a channel.
func BroadcastToChannel(channel string, payload any) {
	internalstate.WithClients(func(clients map[string]*internalstate.Client) {
		for _, client := range clients {
			if client.Channel == channel {
				select {
				case client.Send <- payload:
				default:
					// drop if buffer is full
				}
			}
		}
	})
}
