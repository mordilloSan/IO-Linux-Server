package system

import (
	"encoding/json"
	"fmt"
	"go-backend/cmd/bridge/dbus"
	"go-backend/internal/bridge"
	"go-backend/internal/session"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getNetworkInfo(c *gin.Context) {
	user, sessionID, valid, _ := session.ValidateFromRequest(c.Request)
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	output, err := bridge.CallWithSession(sessionID, user.ID, "dbus", "GetNetworkInfo", nil)
	if err != nil {
		fmt.Printf("[network] Failed: %+v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "failed to get network interfaces",
			"detail": err.Error(),
			"output": output,
		})
		return
	}

	var data []dbus.NMInterfaceInfo
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "invalid bridge response",
			"detail": err.Error(),
			"output": output,
		})
		return
	}
	c.JSON(http.StatusOK, data)
}
