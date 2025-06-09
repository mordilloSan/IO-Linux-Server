package system

import (
	"encoding/json"
	"net/http"

	"go-backend/internal/bridge"
	"go-backend/internal/logger"
	"go-backend/internal/session"

	"github.com/gin-gonic/gin"
)

func getDiskInfo(c *gin.Context) {
	sess, valid := session.ValidateFromRequest(c.Request)
	if !valid || sess == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	output, err := bridge.CallWithSession(sess, "system", "get_drive_info", nil)
	if err != nil {
		logger.Errorf("Failed to get drive info via bridge: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var resp struct {
		Status string          `json:"status"`
		Output json.RawMessage `json:"output"`
		Error  string          `json:"error"`
	}
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		logger.Errorf("Failed to decode bridge response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "decode bridge response"})
		return
	}

	if resp.Status != "ok" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": resp.Error})
		return
	}

	c.Data(http.StatusOK, "application/json", resp.Output)
}
