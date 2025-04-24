package services

import (
	"net/http"
	"os/exec"
	"strings"

	"go-backend/auth"

	"github.com/gin-gonic/gin"
)

type ServiceStatus struct {
	Name        string `json:"name"`
	LoadState   string `json:"load_state"`
	ActiveState string `json:"active_state"`
	SubState    string `json:"sub_state"`
}

func RegisterServiceRoutes(router *gin.Engine) {
	system := router.Group("/system", auth.AuthMiddleware())
	{
		system.GET("/services/status", getServiceStatus)
	}
}

func getServiceStatus(c *gin.Context) {
	cmd := exec.Command("systemctl", "list-units", "--type=service", "--all", "--no-pager", "--no-legend")
	out, err := cmd.Output()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	lines := strings.Split(string(out), "\n")
	services := []ServiceStatus{}
	failedCount := 0

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		svc := ServiceStatus{
			Name:        fields[0],
			LoadState:   fields[1],
			ActiveState: fields[2],
			SubState:    fields[3],
		}

		if svc.ActiveState == "failed" {
			failedCount++
		}

		services = append(services, svc)
	}

	c.JSON(http.StatusOK, gin.H{
		"units":    len(services),
		"failed":   failedCount,
		"services": services,
	})
}
