package websocket

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn      *websocket.Conn
	channel   chan any
	sessionID string
	done      chan struct{}
	closeOnce sync.Once
	lastPong  time.Time
}
