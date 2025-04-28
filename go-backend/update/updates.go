package update

import (
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	"go-backend/auth"

	"github.com/gin-gonic/gin"
)

// Define a strict regex: only letters, numbers, dashes, underscores, dots
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

type UpdateGroup struct {
	Name     string   `json:"name"`
	Version  string   `json:"version"`
	Severity string   `json:"severity"`
	Packages []string `json:"packages"`
}

func getUpdatesHandler(c *gin.Context) {
	cmd := exec.Command("pkcon", "get-updates")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Treat exit code 5 ("no updates") as success
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 5 {
			c.JSON(http.StatusOK, gin.H{"updates": []UpdateGroup{}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get updates",
			"details": err.Error(),
		})
		return
	}

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request. 'package' field is required."})
		return
	}

	// Validate input strictly
	if !validPackageName.MatchString(req.PackageName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid package name"})
		return
	}

	safePackageName := req.PackageName

	cmd := exec.Command("pkexec", "pkcon", "update", "--noninteractive", safePackageName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update package",
			"details": err.Error(),
			"output":  string(output),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Package update triggered",
		"output":  string(output),
	})
}
