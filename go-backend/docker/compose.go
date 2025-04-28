// go-backend/docker/compose.go
package docker

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go-backend/config"

	"github.com/gin-gonic/gin"
)

// Helper: run docker compose inside a specific directory
func runComposeCommandInDir(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("docker", append([]string{"compose"}, args...)...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

// Helper: get full project path
func getComposeProjectDir(project string) (string, error) {
	baseDir, err := config.GetDockerAppsDir()
	if err != nil {
		return "", err
	}
	projectDir := filepath.Join(baseDir, project)
	return projectDir, nil
}

func ComposeUp(c *gin.Context) {
	project := c.Param("project")
	projectDir, err := getComposeProjectDir(project)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project directory"})
		return
	}

	output, err := runComposeCommandInDir(projectDir, "up", "-d")
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
	project := c.Param("project")
	projectDir, err := getComposeProjectDir(project)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project directory"})
		return
	}

	output, err := runComposeCommandInDir(projectDir, "down")
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
	project := c.Param("project")
	projectDir, err := getComposeProjectDir(project)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project directory"})
		return
	}

	output, err := runComposeCommandInDir(projectDir, "restart")
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
	project := c.Param("project")
	projectDir, err := getComposeProjectDir(project)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project directory"})
		return
	}

	output, err := runComposeCommandInDir(projectDir, "ps")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get compose status",
			"details": string(output),
		})
		return
	}

	lines := strings.Split(string(output), "\n")
	c.JSON(http.StatusOK, gin.H{"status": lines})
}

// List all compose projects (list directories)
func ListComposeProjects(c *gin.Context) {
	baseDir, err := config.GetDockerAppsDir()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get base directory"})
		return
	}

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list projects"})
		return
	}

	var projects []string
	for _, entry := range entries {
		if entry.IsDir() {
			projects = append(projects, entry.Name())
		}
	}

	c.JSON(http.StatusOK, gin.H{"projects": projects})
}

// Register Docker Compose routes
func RegisterDockerComposeRoutes(router *gin.Engine) {
	docker := router.Group("/docker/compose")
	{
		docker.GET("/projects", ListComposeProjects)
		docker.POST("/:project/up", ComposeUp)
		docker.POST("/:project/down", ComposeDown)
		docker.POST("/:project/restart", ComposeRestart)
		docker.GET("/:project/status", ComposeStatus)
	}
}
