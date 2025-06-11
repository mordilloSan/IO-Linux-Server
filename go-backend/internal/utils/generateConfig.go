package utils

import (
	embed "go-backend"
	"go-backend/internal/logger"
	"os"
	"path/filepath"
)

// Ensures file at `path` exists; if not, writes `defaultContent` to it.
func EnsureDefaultFile(path string, defaultContent []byte) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			logger.Errorf("Failed to create directory for %s: %v", path, err)
			return err
		}
		if err := os.WriteFile(path, defaultContent, 0644); err != nil {
			logger.Errorf("Failed to write default file %s: %v", path, err)
			return err
		}
		logger.Infof("✅ Generated default file: %s", path)
	}
	return nil
}

func EnsureStartupDefaults() error {

	if err := EnsureDefaultFile("/etc/linuxio/themeConfig.json", embed.DefaultThemeConfig); err != nil {
		return err
	}
	if err := EnsureDefaultFile("/etc/linuxio/dockerConfig.yaml", embed.DefaultDockerConfig); err != nil {
		return err
	}
	// ...add more files as needed...
	return nil
}
