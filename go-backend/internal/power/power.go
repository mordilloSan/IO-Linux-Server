package power

import (
	"go-backend/internal/auth"
	"go-backend/internal/bridge"
	"go-backend/internal/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterPowerRoutes sets up reboot and poweroff endpoints
func RegisterPowerRoutes(r *gin.Engine) {
	group := r.Group("/power")
	group.Use(auth.AuthMiddleware())

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
