package services

import (
	"net/http"

	"go-backend/internal/auth"
	"go-backend/internal/bridge"
	"go-backend/internal/logger"
	"go-backend/internal/session"

	"github.com/gin-gonic/gin"
)

func RegisterServiceRoutes(router *gin.Engine) {
	system := router.Group("/system", auth.AuthMiddleware())
	{
		system.GET("/services/status", getServiceStatus)
	}
}

func getServiceStatus(c *gin.Context) {
	// Get session and user from context/middleware (example, adjust as needed)

	// Extract session info
	user, sessionID, valid, _ := session.ValidateFromRequest(c.Request)
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	output, err := bridge.CallWithSession(sessionID, user.Name, "dbus", "ListServices", nil)

	// Call the bridge

	if err != nil {
		logger.Error.Printf("Failed to list services via bridge: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// output is a JSON string, so decode and forward
	c.Data(http.StatusOK, "application/json", []byte(output))
}
