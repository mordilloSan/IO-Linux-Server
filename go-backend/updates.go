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

	lines := strings.Split(string(output), "\n")
	var updates []gin.H

	for _, line := range lines {
		if strings.HasPrefix(line, "Normal") || strings.HasPrefix(line, "Important") || strings.HasPrefix(line, "Security") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				severity := fields[0]
				rawID := fields[1]
				summary := strings.Join(fields[2:], " ")

				// Split rawID like name-version-arch
				parts := strings.Split(rawID, "-")
				if len(parts) >= 3 {
					arch := parts[len(parts)-1]
					version := parts[len(parts)-2]
					name := strings.Join(parts[:len(parts)-2], "-")

					updates = append(updates, gin.H{
						"name":     name,
						"version":  version,
						"arch":     arch,
						"summary":  summary,
						"severity": severity,
					})
				} else {
					// fallback if split failed
					updates = append(updates, gin.H{
						"packageID": rawID,
						"summary":   summary,
						"severity":  severity,
					})
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"updates": updates,
	})
}
