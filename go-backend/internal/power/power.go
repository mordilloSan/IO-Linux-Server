package power

import (
	"go-backend/internal/auth"
	"go-backend/internal/bridge"
	"go-backend/internal/logger"
	"go-backend/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Middleware to ensure user is authenticated and admin
func requireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := c.Get("user")
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if u, ok := user.(utils.User); !ok || !u.IsAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin required"})
			return
		}
		c.Next()
	}
}

// RegisterPowerRoutes sets up reboot and poweroff endpoints
func RegisterPowerRoutes(r *gin.Engine) {
	group := r.Group("/power")
	group.Use(auth.AuthMiddleware(), requireAdmin())

	group.POST("/reboot", func(c *gin.Context) {
		if err := bridge.RebootSystem(); err != nil {
			logger.Error.Println("❌ Reboot failed:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "reboot failed", "detail": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "rebooting..."})
	})

	group.POST("/shutdown", func(c *gin.Context) {
		if err := bridge.PowerOffSystem(); err != nil {
			logger.Error.Println("❌ Shutdown failed:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "shutdown failed", "detail": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "shutting down..."})
	})
}
