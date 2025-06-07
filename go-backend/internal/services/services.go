package services

import (
	"encoding/json"
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
		logger.Warnf("Unauthorized attempt to %s (missing/invalid session)", action)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}
	serviceName := c.Param("name")

	// --- Validate input service name ---
	if !validServiceName.MatchString(serviceName) {
		logger.Warnf("Invalid service name for %s: %q by user: %s", action, serviceName, user.Name)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid service name"})
		return
	}
	logger.Infof("User %s requested %s on %s (session: %s)", user.Name, action, serviceName, sessionID)

	_, err := bridge.CallWithSession(sessionID, user.Name, "dbus", action, []string{serviceName})
	if err != nil {
		logger.Errorf("Failed to %s %s via bridge (user: %s, session: %s): %v", action, serviceName, user.Name, sessionID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Infof("%s on %s succeeded for user %s (session: %s)", action, serviceName, user.Name, sessionID)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func getServiceStatus(c *gin.Context) {
	user, sessionID, valid, _ := session.ValidateFromRequest(c.Request)
	if !valid {
		logger.Warnf("Unauthorized service status request (missing/invalid session)")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	logger.Infof("User %s requested service status (session: %s)", user.Name, sessionID)

	output, err := bridge.CallWithSession(sessionID, user.Name, "dbus", "ListServices", nil)
	if err != nil {
		logger.Errorf("Failed to list services via bridge (user: %s, session: %s): %v", user.Name, sessionID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var resp struct {
		Status string          `json:"status"`
		Output json.RawMessage `json:"output"`
		Error  string          `json:"error"`
	}
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		logger.Errorf("Failed to decode bridge response (user: %s): %v", user.Name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "decode bridge response"})
		return
	}

	if resp.Status != "ok" {
		logger.Warnf("Bridge returned error for service status (user: %s): %v", user.Name, resp.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": resp.Error})
		return
	}

	logger.Debugf("Returned service status to user %s", user.Name)
	c.Data(http.StatusOK, "application/json", resp.Output)
}

func getServiceDetail(c *gin.Context) {
	user, sessionID, valid, _ := session.ValidateFromRequest(c.Request)
	if !valid {
		logger.Warnf("Unauthorized detail request for %q (missing/invalid session)", c.Param("name"))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}
	serviceName := c.Param("name")
	logger.Infof("%s requested detail for %s (session: %s)", user.Name, serviceName, sessionID)

	output, err := bridge.CallWithSession(sessionID, user.Name, "dbus", "GetServiceInfo", []string{serviceName})
	if err != nil {
		logger.Errorf("Failed to get info for %s via bridge (user: %s, session: %s): %v", serviceName, user.Name, sessionID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var resp struct {
		Status string          `json:"status"`
		Output json.RawMessage `json:"output"`
		Error  string          `json:"error"`
	}
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		logger.Errorf("Failed to decode bridge response for %s (user: %s): %v", serviceName, user.Name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "decode bridge response"})
		return
	}
	if resp.Status != "ok" {
		logger.Warnf("Bridge returned error for %s (user: %s): %v", serviceName, user.Name, resp.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": resp.Error})
		return
	}
	logger.Debugf("Returned detail for %s to user %s", serviceName, user.Name)
	c.Data(http.StatusOK, "application/json", resp.Output)
}
