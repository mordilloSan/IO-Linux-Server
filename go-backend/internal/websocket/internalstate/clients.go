package internalstate

import (
	"sync"
)

type Client struct {
	Send    chan any
	Channel string
}

var (
	clients     = make(map[string]*Client)
	clientsLock sync.RWMutex
)

func WithClients(fn func(map[string]*Client)) {
	clientsLock.RLock()
	defer clientsLock.RUnlock()
	fn(clients)
}

func SetClient(sessionID string, client *Client) {
	clientsLock.Lock()
	defer clientsLock.Unlock()
	clients[sessionID] = client
}

func RemoveClient(sessionID string) {
	clientsLock.Lock()
	defer clientsLock.Unlock()
	delete(clients, sessionID)
}
