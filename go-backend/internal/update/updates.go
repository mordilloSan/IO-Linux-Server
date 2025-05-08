package update

import (
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	"go-backend/internal/auth"
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
	}
}

func getUpdatesHandler(c *gin.Context) {
	logger.Info.Println("🔍 Checking for system updates...")

	cmd := exec.Command("pkcon", "get-updates")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 5 {
			logger.Info.Println("✅ No updates available.")
			c.JSON(http.StatusOK, gin.H{"updates": []any{}})
			return
		}
		logger.Error.Printf("❌ Failed to get updates: %v\nOutput: %s", err, string(output))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get updates", "details": err.Error()})
		return
	}

	lines := strings.Split(string(output), "\n")
	logger.Debug.Printf("📦 Raw pkcon output:\n%s", string(output))

	type UpdateItem struct {
		Name     string `json:"name"`
		Version  string `json:"version"`
		Severity string `json:"severity"`
	}

	var updates []UpdateItem

	for _, line := range lines {
		if severity, name, version, ok := parseUpdateLine(line); ok {
			updates = append(updates, UpdateItem{
				Name:     name,
				Version:  version,
				Severity: normalizeSeverity(severity),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"updates": updates})
}

func parseUpdateLine(line string) (string, string, string, bool) {
	if !strings.Contains(line, "(") || !strings.Contains(line, "-") {
		return "", "", "", false
	}

	// Remove trailing repo info in parentheses
	parts := strings.SplitN(line, "(", 2)
	if len(parts) < 1 {
		return "", "", "", false
	}
	line = strings.TrimSpace(parts[0])

	fields := strings.Fields(line)
	if len(fields) < 2 {
		return "", "", "", false
	}

	// Multi-word severities (like "Bug fix", "Security update")
	severity := fields[0]
	if len(fields) > 2 && (strings.ToLower(fields[1]) == "fix" || strings.ToLower(fields[1]) == "update") {
		severity += " " + fields[1]
		fields = append([]string{severity}, fields[2:]...)
	} else {
		fields = append([]string{severity}, fields[1:]...)
	}

	if len(fields) < 2 {
		return "", "", "", false
	}

	packageID := fields[1]

	// Strip arch suffix (e.g. .amd64)
	if idx := strings.LastIndex(packageID, "."); idx != -1 {
		packageID = packageID[:idx]
	}

	parts = strings.Split(packageID, "-")
	if len(parts) < 3 {
		return "", "", "", false
	}

	version := parts[len(parts)-2] + "-" + parts[len(parts)-1]
	name := strings.Join(parts[:len(parts)-2], "-")

	return severity, name, version, true
}

func normalizeSeverity(raw string) string {
	switch strings.ToLower(strings.ReplaceAll(raw, " ", "")) {
	case "important":
		return "🔴 Critical"
	case "securityupdate", "security":
		return "🟠 Security"
	case "bugfix", "bug":
		return "🟡 Bugfix"
	case "enhancement":
		return "🟢 Enhancement"
	default:
		return raw
	}
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
