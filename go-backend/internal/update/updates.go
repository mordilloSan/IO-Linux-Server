package update

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"go-backend/internal/auth"
	"go-backend/internal/bridge"
	"go-backend/internal/logger"

	"github.com/gin-gonic/gin"
)

var validPackageName = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

func RegisterUpdateRoutes(router *gin.Engine) {
	system := router.Group("/system", auth.AuthMiddleware())
	{
		system.GET("/updates", getUpdatesHandler)
		system.POST("/update", updatePackageHandler)
		system.GET("/updates/update-history", getUpdateHistoryHandler)
		system.GET("/updates/changelog", getChangelogHandler)
		system.GET("/updates/settings", getUpdateSettings)
		system.POST("/updates/settings", postUpdateSettings)
	}
}

func getUpdatesHandler(c *gin.Context) {
	logger.Info.Println("🔍 Checking for system updates (D-Bus)...")

	output, err := bridge.Call("dbus", "GetUpdates", nil)

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
	var updates []map[string]interface{}

	// Defensive: If output is empty, treat as empty array
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		updates = []map[string]interface{}{}
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
		PackageName string `json:"package"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.PackageName) == "" {
		logger.Warning.Println("⚠️ Missing or invalid package name in update request.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request. 'package' field is required."})
		return
	}

	if !validPackageName.MatchString(req.PackageName) {
		logger.Warning.Printf("⚠️ Invalid package name submitted: %s", req.PackageName)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid package name"})
		return
	}

	logger.Info.Printf("📦 Triggering update for package: %s", req.PackageName)

	output, err := bridge.Call("command", "pkcon", []string{"update", "--noninteractive", req.PackageName})

	if err != nil {
		logger.Error.Printf("❌ Failed to update %s: %v\nOutput: %s", req.PackageName, err, output)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update package",
			"details": err.Error(),
			"output":  output,
		})
		return
	}

	logger.Info.Printf("✅ Package %s updated successfully.\nOutput:\n%s", req.PackageName, output)
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
