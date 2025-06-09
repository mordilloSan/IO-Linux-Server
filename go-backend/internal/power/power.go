package power

import (
	"net/http"

	"github.com/mordilloSan/LinuxIO/go-backend/internal/auth"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/bridge"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/logger"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/session"

	"github.com/gin-gonic/gin"
)

func RegisterPowerRoutes(r *gin.Engine) {
	group := r.Group("/power")
	group.Use(auth.AuthMiddleware())

	group.POST("/reboot", func(c *gin.Context) {
		sess, valid := session.ValidateFromRequest(c.Request)
		if !valid || sess == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			return
		}

		output, err := bridge.CallWithSession(sess, "dbus", "Reboot", nil)
		if err != nil {
			logger.Errorf("Reboot failed: %+v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "reboot failed",
				"detail": err.Error(),
				"output": output,
			})
			return
		}
		logger.Infof("Reboot triggered successfully for user %s (session: %s)", sess.User.ID, sess.SessionID)
		c.JSON(http.StatusOK, gin.H{"message": "rebooting...", "output": output})
	})

	group.POST("/shutdown", func(c *gin.Context) {
		sess, valid := session.ValidateFromRequest(c.Request)
		if !valid || sess == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			return
		}

		output, err := bridge.CallWithSession(sess, "dbus", "PowerOff", nil)
		if err != nil {
			logger.Errorf("Shutdown failed: %+v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "shutdown failed",
				"detail": err.Error(),
				"output": output,
			})
			return
		}
		logger.Infof("Shutdown triggered successfully for user %s (session: %s)", sess.User.ID, sess.SessionID)
		c.JSON(http.StatusOK, gin.H{"message": "shutting down...", "output": output})
	})
}
