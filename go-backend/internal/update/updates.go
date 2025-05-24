package update

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"go-backend/internal/auth"
	"go-backend/internal/bridge"
	"go-backend/internal/logger"
	"go-backend/internal/session"

	"github.com/gin-gonic/gin"
)

var validPackageName = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

func RegisterUpdateRoutes(router *gin.Engine) {
	system := router.Group("/system", auth.AuthMiddleware())
	{
		system.GET("/updates", getUpdatesHandler)
		system.POST("/update", updatePackageHandler)
		system.GET("/updates/update-history", getUpdateHistoryHandler)
		system.GET("/updates/settings", getUpdateSettings)
		system.POST("/updates/settings", postUpdateSettings)
	}
}

func getUpdatesHandler(c *gin.Context) {
	logger.Info.Println("🔍 Checking for system updates (D-Bus)...")

	// Extract session info
	user, sessionID, valid, _ := session.ValidateFromRequest(c.Request)
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	output, err := bridge.CallWithSession(sessionID, user.ID, "dbus", "GetUpdates", nil)

	if err != nil {
		logger.Error.Printf("❌ Failed to get updates: %v\nOutput: %s", err, output)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get updates",
			"details": err.Error(),
			"output":  output,
		})
		return
	}

	// Output is JSON from the bridge, unmarshal it
	var updates []Update

	// Defensive: If output is empty, treat as empty array
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		updates = []Update{}
	} else if err := json.Unmarshal([]byte(trimmed), &updates); err != nil {
		logger.Error.Printf("❌ Failed to decode updates JSON: %v\nOutput: %s", err, output)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to decode updates JSON",
			"details": err.Error(),
			"output":  output,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"updates": updates})
}

func updatePackageHandler(c *gin.Context) {
	var req struct {
		PackageID string `json:"package"` // Now this must be the *full* PackageKit ID
	}

	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.PackageID) == "" {
		logger.Warning.Println("⚠️ Missing or invalid package id in update request.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request. 'package' field is required."})
		return
	}

	// Defensive: Optionally check if this looks like a PackageKit package_id (e.g., contains semicolons)
	if !strings.Contains(req.PackageID, ";") {
		logger.Warning.Printf("⚠️ Invalid package_id submitted: %s", req.PackageID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid package_id"})
		return
	}

	logger.Info.Printf("📦 Triggering update for package: %s", req.PackageID)

	// Extract session info
	user, sessionID, valid, _ := session.ValidateFromRequest(c.Request)
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	output, err := bridge.CallWithSession(sessionID, user.ID, "dbus", "InstallPackage", []string{req.PackageID})

	if err != nil {
		logger.Error.Printf("❌ Failed to update %s: %v\nOutput: %s", req.PackageID, err, output)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update package",
			"details": err.Error(),
			"output":  output,
		})
		return
	}

	logger.Info.Printf("✅ Package %s updated successfully.\nOutput:\n%s", req.PackageID, output)
	c.JSON(http.StatusOK, gin.H{
		"message": "Package update triggered",
		"output":  output,
	})
}

// GET /system/updates/settings
func getUpdateSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"enabled":   true,
		"frequency": "daily",
		"lastRun":   "2025-05-15T12:34:00Z",
	})
}

// POST /system/updates/settings
func postUpdateSettings(c *gin.Context) {
	var req struct {
		Enabled   bool   `json:"enabled"`
		Frequency string `json:"frequency"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid settings"})
		return
	}

	// Save logic here...

	c.Status(http.StatusNoContent)
}
