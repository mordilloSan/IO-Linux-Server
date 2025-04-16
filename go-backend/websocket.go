// websocket_routes.go
package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func registerWebSocketRoutes(router *gin.Engine) {
	router.GET("/ws/system", func(c *gin.Context) {
		// ✅ Auth: Check for session_id
		cookie, err := c.Request.Cookie("session_id")
		if err != nil || cookie.Value == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var session Session
		var exists bool
		done := make(chan bool)

		sessionMux <- func() {
			session, exists = sessions[cookie.Value]
			done <- true
		}
		<-done

		if !exists || session.ExpiresAt.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session expired"})
			return
		}

		// ✅ Upgrade to WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		var prevIOCounters []net.IOCountersStat

		for {
			cpuPercent, _ := cpu.Percent(0, false)
			memStats, _ := mem.VirtualMemory()
			ioCounters, _ := net.IOCounters(true)

			networkDeltas := make(map[string]map[string]uint64)
			for i, iface := range ioCounters {
				if i < len(prevIOCounters) {
					networkDeltas[iface.Name] = map[string]uint64{
						"bytesSent": iface.BytesSent - prevIOCounters[i].BytesSent,
						"bytesRecv": iface.BytesRecv - prevIOCounters[i].BytesRecv,
					}
				}
			}
			prevIOCounters = ioCounters

			data := map[string]any{
				"cpu":     cpuPercent,
				"memory":  memStats,
				"network": networkDeltas,
			}

			conn.WriteJSON(data)
			time.Sleep(2 * time.Second)
		}
	})
}
