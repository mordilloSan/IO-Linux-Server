// go-backend/docker/compose.go
package docker

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"go-backend/config"

	"github.com/gin-gonic/gin"
)

// Safe project name: only letters, numbers, dashes, underscores
var validProjectName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Validate project name format
func isValidProjectName(name string) bool {
	return validProjectName.MatchString(name)
}

// run docker compose inside specific directory
func runComposeCommandInDir(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("docker", append([]string{"compose"}, args...)...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

// Build full project directory
func getComposeProjectDir(project string) (string, error) {
	baseDir, err := config.GetDockerAppsDir()
	if err != nil {
		return "", err
	}
	projectDir := filepath.Join(baseDir, project)
	return projectDir, nil
}

// Ensure compose.yaml exists (optional but good)
func checkComposeFileExists(dir string) bool {
	composePathYaml := filepath.Join(dir, "compose.yaml")
	composePathYml := filepath.Join(dir, "docker-compose.yml")
	if _, err := os.Stat(composePathYaml); err == nil {
		return true
	}
	if _, err := os.Stat(composePathYml); err == nil {
		return true
	}
	return false
}

// ---- Handlers ----

func ComposeUp(c *gin.Context) {
	project := c.Param("project")

	if !isValidProjectName(project) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project name"})
		return
	}

	projectDir, err := getComposeProjectDir(project)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project directory"})
		return
	}

	if !checkComposeFileExists(projectDir) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No compose.yaml found in project directory"})
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

	if !isValidProjectName(project) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project name"})
		return
	}

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

	if !isValidProjectName(project) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project name"})
		return
	}

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

	if !isValidProjectName(project) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project name"})
		return
	}

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

// List all projects
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
