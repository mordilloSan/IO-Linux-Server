package auth

import (
	"net/http"

	"go-backend/internal/logger"
	"go-backend/internal/session"
	"go-backend/internal/utils"

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

			logger.Debug.Printf("CORS allowed: %s %s", c.Request.Method, origin)
		} else if origin != "" {
			logger.Debug.Printf("CORS denied: %s %s", c.Request.Method, origin)
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
		user, sessionID, valid, privileged := session.ValidateFromRequest(c.Request)
		if !valid {
			logger.Warning.Println("⚠️  Unauthorized request or expired session")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set("user", user)
		c.Set("session_id", sessionID)
		c.Set("privileged", privileged)
		c.Next()
	}
}
