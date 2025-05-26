package services

import (
	"net/http"
	"regexp"

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
		system.GET("/services/:name", getServiceDetail)

		system.POST("/services/:name/start", startService)
		system.POST("/services/:name/stop", stopService)
		system.POST("/services/:name/restart", restartService)
		system.POST("/services/:name/reload", reloadService)
		system.POST("/services/:name/enable", enableService)
		system.POST("/services/:name/disable", disableService)
		system.POST("/services/:name/mask", maskService)
		system.POST("/services/:name/unmask", unmaskService)
	}
}

func startService(c *gin.Context)   { serviceAction(c, "StartService") }
func stopService(c *gin.Context)    { serviceAction(c, "StopService") }
func restartService(c *gin.Context) { serviceAction(c, "RestartService") }
func reloadService(c *gin.Context)  { serviceAction(c, "ReloadService") }
func enableService(c *gin.Context)  { serviceAction(c, "EnableService") }
func disableService(c *gin.Context) { serviceAction(c, "DisableService") }
func maskService(c *gin.Context)    { serviceAction(c, "MaskService") }
func unmaskService(c *gin.Context)  { serviceAction(c, "UnmaskService") }

var validServiceName = regexp.MustCompile(`^[\w.-]+\.service$`)

// Generic handler for service actions
func serviceAction(c *gin.Context, action string) {

	user, sessionID, valid, _ := session.ValidateFromRequest(c.Request)
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}
	serviceName := c.Param("name")

	// --- Validate input service name ---
	if !validServiceName.MatchString(serviceName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid service name"})
		return
	}
	_, err := bridge.CallWithSession(sessionID, user.Name, "dbus", action, []string{serviceName})
	if err != nil {
		logger.Error.Printf("Failed to %s via bridge: %v", action, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func getServiceStatus(c *gin.Context) {
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

func getServiceDetail(c *gin.Context) {
	user, sessionID, valid, _ := session.ValidateFromRequest(c.Request)
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}
	serviceName := c.Param("name")
	output, err := bridge.CallWithSession(sessionID, user.Name, "dbus", "GetServiceInfo", []string{serviceName})
	if err != nil {
		logger.Error.Printf("Failed to get service info via bridge: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/json", []byte(output))
}
