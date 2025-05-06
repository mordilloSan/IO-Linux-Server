package auth

import (
	"net/http"
	"time"

	"go-backend/logger"
	"go-backend/session"
	"go-backend/utils"

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
		cookie, err := c.Cookie("session_id")
		if err != nil || cookie == "" {
			logger.Warning.Println("⚠️  Unauthorized request: missing session_id cookie")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var sess session.Session
		var exists bool
		done := make(chan bool)

		session.SessionMux <- func() {
			sess, exists = session.Sessions[cookie]
			done <- true
		}
		<-done

		if !exists || sess.ExpiresAt.Before(time.Now()) {
			logger.Warning.Printf("⚠️  Session expired or invalid: %s", cookie)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session expired"})
			return
		}

		c.Set("user", sess.User)
		c.Next()
	}
}


// Ensure user is admin
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := c.Get("user")
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if u, ok := user.(utils.User); !ok || !u.IsAdmin {
			logger.Warning.Printf("⚠️  Admin access denied for user: %+v", user)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin privileges required"})
			return
		}
		c.Next()
	}
}