package auth

import (
	"go-backend/internal/logger"
	"go-backend/internal/session"
	"go-backend/internal/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/msteinert/pam"
)

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

	// 1. Authenticate with PAM
	if err := pamAuth(req.Username, req.Password); err != nil {
		logger.Warning.Printf("❌ Authentication failed for user: %s", req.Username)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication failed"})
		return
	}

	// 3. Create session
	sessionID := uuid.New().String()
	user := utils.User{ID: req.Username, Name: req.Username}
	session.CreateSession(sessionID, user, sessionDuration)

	c.SetCookie("session_id", sessionID, int(sessionDuration.Seconds()), "/", "", false, true)
	logger.Info.Printf("✅ User %s logged in, session ID: %s", req.Username, sessionID)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func meHandler(c *gin.Context) {
	user := c.MustGet("user").(utils.User)
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func logoutHandler(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err == nil {
		session.DeleteSession(sessionID) // <--- this also kills the bridge!
		c.SetCookie("session_id", "", -1, "/", "", false, true)
		logger.Info.Printf("👋 Logged out session: %s", sessionID)
	}
	c.Status(http.StatusOK)
}
