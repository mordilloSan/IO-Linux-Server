package utils

import (
	"fmt"
	"os"
	"strings"
)

func GetDistroID() (string, error) {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ID_LIKE=") {
			return strings.Trim(strings.TrimPrefix(line, "ID_LIKE="), "\""), nil
		}
	}
	return "", fmt.Errorf("ID_LIKE not found")
}

func GetDevPort() string {
	port := os.Getenv("VITE_DEV_PORT")
	if port == "" {
		port = "3000"
	}
	return port
}
