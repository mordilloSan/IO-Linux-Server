package main

import (
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

func registerUpdateRoutes(router *gin.Engine) {
	system := router.Group("/system", authMiddleware())
	{
		system.GET("/updates", getUpdatesHandler)
		system.POST("/update", updatePackageHandler)
		system.GET("/updates/update-history", getUpdateHistoryHandler)
	}
}

type UpdateGroup struct {
	Name      string   `json:"name"`
	Version   string   `json:"version"`
	Severity  string   `json:"severity"`
	Packages  []string `json:"packages"`
	Changelog string   `json:"changelog"`
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

	distro, _ := getDistroID()

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
			changelog := getChangelog(distro, name)
			groupMap[key] = &UpdateGroup{
				Name:      name,
				Version:   version,
				Severity:  severity,
				Packages:  []string{name},
				Changelog: changelog,
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
