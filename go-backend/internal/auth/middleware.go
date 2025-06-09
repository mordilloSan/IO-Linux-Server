package auth

import (
	"net/http"

	"github.com/mordilloSan/LinuxIO/go-backend/internal/logger"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/session"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

func CorsMiddleware() gin.HandlerFunc {
	devOrigin := "http://localhost:" + utils.GetDevPort()

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if origin == devOrigin {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type")
			c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

			logger.Debugf("CORS allowed: %s %s", c.Request.Method, origin)
		} else if origin != "" {
			logger.Debugf("CORS denied: %s %s", c.Request.Method, origin)
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sess, valid := session.ValidateFromRequest(c.Request)
		if !valid || sess == nil {
			logger.Warnf("⚠️  Unauthorized request or expired session")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set("session", sess)
		// For compatibility with older code, you can also set these (optional):
		c.Set("user", sess.User)
		c.Set("session_id", sess.SessionID)
		c.Set("privileged", sess.Privileged)
		c.Next()
	}
}
