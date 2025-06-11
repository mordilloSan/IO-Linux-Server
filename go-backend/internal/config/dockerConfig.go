package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go-backend/internal/logger"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	DockerAppsSubdir string `yaml:"docker_apps_subdir"`
}

var appConfig AppConfig

// LoadDockerConfig reads config.yaml or applies default values
func LoadDockerConfig() error {
	file, err := os.Open("/etc/linuxio/dockerConfig.yaml")
	if err != nil {
		logger.Warnf("No config.yaml found, using defaults")
		appConfig = AppConfig{
			DockerAppsSubdir: "docker",
		}
		return nil
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&appConfig); err != nil {
		logger.Errorf("Failed to parse config.yaml: %v", err)
		return err
	}

	// Fallback if value is empty
	if appConfig.DockerAppsSubdir == "" {
		logger.Warnf("docker_apps_subdir missing, falling back to default: 'docker'")
		appConfig.DockerAppsSubdir = "docker"
	}

	logger.Infof("Config loaded. DockerAppsSubdir: %s", appConfig.DockerAppsSubdir)
	return nil
}

// GetDockerAppsDir returns absolute path to user's docker apps folder
func GetDockerAppsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		logger.Errorf("Failed to get user home directory: %v", err)
		return "", err
	}
	return filepath.Join(home, appConfig.DockerAppsSubdir), nil
}

// EnsureDockerAppsDirExists creates the folder if it doesn't exist
func EnsureDockerAppsDirExists() error {
	dockerDir, err := GetDockerAppsDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dockerDir, 0755); err != nil {
		logger.Errorf("Failed to create docker apps directory: %v", err)
		return err
	}
	logger.Infof("Docker apps directory ensured at: %s", dockerDir)
	return nil
}

// IsDockerInstalled returns true if 'docker' is available in PATH
func IsDockerInstalled() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// IsDockerDaemonRunning returns true if 'docker info' runs without error
func IsDockerDaemonRunning() bool {
	cmd := exec.Command("docker", "info")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Warnf("Docker daemon check failed: %v (output: %s)", err, strings.TrimSpace(string(output)))
		return false
	}
	return true
}

// EnsureDockerAvailable logs warnings or errors if Docker is not usable
func EnsureDockerAvailable() error {
	if !IsDockerInstalled() {
		logger.Errorf("Docker is not installed or not in PATH")
		return fmt.Errorf("docker not installed")
	}
	if !IsDockerDaemonRunning() {
		logger.Errorf("Docker daemon is not running or permission denied")
		return fmt.Errorf("docker daemon not running")
	}
	logger.Infof("Docker is installed and running")
	return nil
}
