package auth

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/msteinert/pam"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

const sessionDuration = 6 * time.Hour

func RegisterAuthRoutes(router *gin.Engine) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", loginHandler)
		auth.GET("/me", AuthMiddleware(), meHandler)
		auth.GET("/logout", AuthMiddleware(), logoutHandler)
	}
}

func pamAuth(username, password string) error {
	t, err := pam.StartFunc("login", username, func(s pam.Style, msg string) (string, error) {
		return password, nil
	})
	if err != nil {
		return err
	}
	return t.Authenticate(0)
}

func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := pamAuth(req.Username, req.Password); err != nil {
		log.Printf("❌ Auth failed for user: %s", req.Username)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication failed"})
		return
	}

	sessionID := uuid.New().String()
	session := Session{
		User:      User{ID: req.Username, Name: req.Username},
		ExpiresAt: time.Now().Add(sessionDuration),
	}
	SessionMux <- func() { Sessions[sessionID] = session }

	c.SetCookie("session_id", sessionID, int(sessionDuration.Seconds()), "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func meHandler(c *gin.Context) {
	user := c.MustGet("user").(User)
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func logoutHandler(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err == nil {
		SessionMux <- func() { delete(Sessions, sessionID) }
		c.SetCookie("session_id", "", -1, "/", "", false, true)
	}
	c.Status(http.StatusOK)
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("session_id")
		if err != nil || cookie == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var session Session
		var exists bool
		done := make(chan bool)
		SessionMux <- func() {
			session, exists = Sessions[cookie]
			done <- true
		}
		<-done

		if !exists || session.ExpiresAt.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session expired"})
			return
		}

		c.Set("user", session.User)
		c.Next()
	}
}
