package update

import (
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	"go-backend/auth"
	"go-backend/logger"

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
	}
}

func getUpdatesHandler(c *gin.Context) {
	logger.Info.Println("🔍 Checking for system updates...")
	cmd := exec.Command("pkcon", "get-updates")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// pkcon returns exit code 5 if no updates
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 5 {
			logger.Info.Println("✅ No updates available.")
			c.JSON(http.StatusOK, gin.H{"updates": []UpdateGroup{}})
			return
		}
		logger.Error.Printf("❌ Failed to get updates: %v\nOutput: %s", err, string(output))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get updates",
			"details": err.Error(),
		})
		return
	}

	logger.Info.Printf("✅ Retrieved update list:\n%s", string(output))
	groupMap := make(map[string]*UpdateGroup)
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if line == "" || !strings.Contains(line, "-") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		severity := fields[0]
		rawID := fields[1]
		parts := strings.Split(rawID, "-")
		if len(parts) < 3 {
			continue
		}
		name := strings.Join(parts[:len(parts)-2], "-")
		version := parts[len(parts)-2]
		key := version + "|" + severity

		if group, exists := groupMap[key]; exists {
			group.Packages = append(group.Packages, name)
		} else {
			groupMap[key] = &UpdateGroup{
				Name:     name,
				Version:  version,
				Severity: severity,
				Packages: []string{name},
			}
		}
	}

	var updates []UpdateGroup
	for _, group := range groupMap {
		group.Name = strings.Join(group.Packages, ", ")
		updates = append(updates, *group)
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

	cmd := exec.Command("pkexec", "pkcon", "update", "--noninteractive", req.PackageName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		logger.Error.Printf("❌ Failed to update %s: %v\nOutput: %s", req.PackageName, err, string(output))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update package",
			"details": err.Error(),
			"output":  string(output),
		})
		return
	}

	logger.Info.Printf("✅ Package %s updated successfully.\nOutput:\n%s", req.PackageName, string(output))
	c.JSON(http.StatusOK, gin.H{
		"message": "Package update triggered",
		"output":  string(output),
	})
}
