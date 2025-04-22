package main

import (
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

type Update struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Severity string `json:"severity"`
}

func registerUpdateRoutes(router *gin.Engine) {
	system := router.Group("/system", authMiddleware())
	{
		system.GET("/updates", getUpdatesHandler)
		system.POST("/update", updatePackageHandler)
		system.GET("/updates/update-history", getUpdateHistoryHandler)
	}
}

func getUpdatesHandler(c *gin.Context) {
	cmd := exec.Command("pkcon", "get-updates")
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get updates",
			"details": err.Error(),
		})
		return
	}

	var updates []Update
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

		// Try to safely parse name-version-arch (or name-version if arch is missing)
		parts := strings.Split(rawID, "-")
		if len(parts) < 2 {
			continue
		}

		version := parts[len(parts)-2]
		name := strings.Join(parts[:len(parts)-2], "-")

		if name == "" || version == "" {
			continue
		}

		updates = append(updates, Update{
			Name:     name,
			Version:  version,
			Severity: severity,
		})
	}

	c.JSON(http.StatusOK, gin.H{"updates": updates})
}

func updatePackageHandler(c *gin.Context) {
	var req struct {
		PackageName string `json:"package"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.PackageName) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request. 'package' field is required.",
		})
		return
	}

	cmd := exec.Command("pkexec", "pkcon", "update", req.PackageName)
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
