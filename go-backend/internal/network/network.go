package network

import (
	"encoding/json"
	"go-backend/cmd/bridge/dbus"
	"go-backend/internal/bridge"
	"go-backend/internal/session"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterNetworkRoutes(router *gin.Engine) {
	system := router.Group("/network")
	{
		system.GET("/info", getNetworkInfo)

	}
}

func getNetworkInfo(c *gin.Context) {
	user, sessionID, valid, _ := session.ValidateFromRequest(c.Request)
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	rawResp, err := bridge.CallWithSession(sessionID, user.ID, "dbus", "GetNetworkInfo", nil)
	if err != nil {
		c.JSON(500, gin.H{
			"error":  "bridge call failed",
			"detail": err.Error(),
			"output": rawResp,
		})
		return
	}
	var resp bridge.BridgeResponse
	if err := json.Unmarshal([]byte(rawResp), &resp); err != nil {
		c.JSON(500, gin.H{
			"error":  "invalid bridge response",
			"detail": err.Error(),
			"output": rawResp,
		})
		return
	}
	if resp.Status != "ok" {
		c.JSON(500, gin.H{
			"error":  resp.Error,
			"output": string(resp.Output),
		})
		return
	}
	var data []dbus.NMInterfaceInfo
	if err := json.Unmarshal(resp.Output, &data); err != nil {
		c.JSON(500, gin.H{
			"error":  "invalid output structure",
			"detail": err.Error(),
			"output": string(resp.Output),
		})
		return
	}
	c.JSON(200, data)
}
