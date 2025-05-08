package utils

import (
	"fmt"
	"go-backend/internal/logger"
	"os"
	"strings"
)

// GetDistroID reads /etc/os-release and extracts ID_LIKE
func GetDistroID() (string, error) {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		logger.Error.Printf("❌ Failed to read /etc/os-release: %v", err)
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ID_LIKE=") {
			idLike := strings.Trim(strings.TrimPrefix(line, "ID_LIKE="), "\"")
			logger.Debug.Printf("✅ Detected distro ID_LIKE: %s", idLike)
			return idLike, nil
		}
	}

	logger.Warning.Println("⚠️ ID_LIKE not found in /etc/os-release")
	return "", fmt.Errorf("ID_LIKE not found")
}

// GetDevPort returns the development port from env or defaults to 3000
func GetDevPort() string {
	port := os.Getenv("VITE_DEV_PORT")
	if port == "" {
		port = "3000"
		logger.Warning.Println("⚠️ VITE_DEV_PORT not set, defaulting to 3000")
	} else {
		logger.Debug.Printf("🔧 VITE_DEV_PORT detected: %s", port)
	}
	return port
}
