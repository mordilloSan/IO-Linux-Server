// go-backend/docker/compose.go
package docker

import (
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

// Helper to run a docker compose command
func runComposeCommand(args ...string) ([]byte, error) {
	cmd := exec.Command("docker", append([]string{"compose"}, args...)...)
	return cmd.CombinedOutput()
}

func ComposeUp(c *gin.Context) {
	output, err := runComposeCommand("up", "-d")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to start compose project",
			"details": string(output),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Compose project started"})
}

func ComposeDown(c *gin.Context) {
	output, err := runComposeCommand("down")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to stop compose project",
			"details": string(output),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Compose project stopped"})
}

func ComposeRestart(c *gin.Context) {
	output, err := runComposeCommand("restart")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to restart compose project",
			"details": string(output),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Compose project restarted"})
}

func ComposeStatus(c *gin.Context) {
	output, err := runComposeCommand("ps")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get compose status",
			"details": string(output),
		})
		return
	}

	// split output into lines
	lines := strings.Split(string(output), "\n")
	c.JSON(http.StatusOK, gin.H{
		"status": lines,
	})
}

// RegisterDockerRoutes registers all docker-related routes
func RegisterDockerComposeRoutes(router *gin.Engine) {
	docker := router.Group("/docker")
	{
		compose := docker.Group("/compose")
		{
			compose.POST("/up", ComposeUp)
			compose.POST("/down", ComposeDown)
			compose.POST("/restart", ComposeRestart)
			compose.GET("/status", ComposeStatus)
		}
	}
}
