package power

import (
	"fmt"
	"go-backend/internal/auth"
	"go-backend/internal/bridge"
	"go-backend/internal/session"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterPowerRoutes(r *gin.Engine) {
	group := r.Group("/power")
	group.Use(auth.AuthMiddleware())

	group.POST("/reboot", func(c *gin.Context) {
		// Extract session info
		user, sessionID, valid, _ := session.ValidateFromRequest(c.Request)
		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			return
		}

		output, err := bridge.CallWithSession(sessionID, user.ID, "dbus", "Reboot", nil) // D-Bus
		if err != nil {
			fmt.Printf("[power] Reboot failed: %+v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "reboot failed",
				"detail": err.Error(),
				"output": output,
			})
			return
		}
		fmt.Println("[power] Reboot triggered successfully")
		c.JSON(http.StatusOK, gin.H{"message": "rebooting...", "output": output})
	})

	group.POST("/shutdown", func(c *gin.Context) {
		// Extract session info
		user, sessionID, valid, _ := session.ValidateFromRequest(c.Request)
		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			return
		}

		output, err := bridge.CallWithSession(sessionID, user.ID, "dbus", "PowerOff", nil) // D-Bus
		if err != nil {
			fmt.Printf("[power] Shutdown failed: %+v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "shutdown failed",
				"detail": err.Error(),
				"output": output,
			})
			return
		}
		fmt.Println("[power] Shutdown triggered successfully")
		c.JSON(http.StatusOK, gin.H{"message": "shutting down...", "output": output})
	})
}
