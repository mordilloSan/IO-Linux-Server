package main

import (
	"fmt"
	"os"
	"strings"
)

func getDistroID() (string, error) {
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
