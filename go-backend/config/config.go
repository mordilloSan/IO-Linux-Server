package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	DockerAppsSubdir string `yaml:"docker_apps_subdir"`
}

var appConfig AppConfig

// LoadConfig reads config.yaml or uses defaults
func LoadConfig() error {
	file, err := os.Open("config.yaml")
	if err != nil {
		// No config.yaml found? Use defaults
		appConfig = AppConfig{
			DockerAppsSubdir: "docker",
		}
		return nil
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&appConfig)
	if err != nil {
		return err
	}

	// Fallback if docker_apps_subdir missing
	if appConfig.DockerAppsSubdir == "" {
		appConfig.DockerAppsSubdir = "docker"
	}

	return nil
}

// GetDockerAppsDir returns /home/username/docker
func GetDockerAppsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, appConfig.DockerAppsSubdir), nil
}

// EnsureDockerAppsDirExists creates the folder if missing
func EnsureDockerAppsDirExists() error {
	dockerDir, err := GetDockerAppsDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dockerDir, 0755)
}
