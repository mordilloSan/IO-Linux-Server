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
	network := router.Group("/network")
	{
		network.GET("/info", getNetworkInfo)
		network.POST("/set-dns", postSetDNS)
		network.POST("/set-gateway", postSetGateway)
		network.POST("/set-mtu", postSetMTU)
		network.POST("/set-ipv4-dhcp", postSetIPv4DHCP)
		network.POST("/set-ipv4-static", postSetIPv4Static)
		network.POST("/set-ipv6-dhcp", postSetIPv6DHCP)
		network.POST("/set-ipv6-static", postSetIPv6Static)
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

func postSetDNS(c *gin.Context) {
	var req struct {
		Interface string   `json:"interface"`
		DNS       []string `json:"dns"`
	}
	if err := c.BindJSON(&req); err != nil || req.Interface == "" || len(req.DNS) == 0 {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	err := dbus.SetDNS(req.Interface, req.DNS)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "ok"})
}

func postSetGateway(c *gin.Context) {
	var req struct {
		Interface string `json:"interface"`
		Gateway   string `json:"gateway"`
	}
	if err := c.BindJSON(&req); err != nil || req.Interface == "" || req.Gateway == "" {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	err := dbus.SetGateway(req.Interface, req.Gateway)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "ok"})
}

func postSetMTU(c *gin.Context) {
	var req struct {
		Interface string `json:"interface"`
		MTU       string `json:"mtu"`
	}
	if err := c.BindJSON(&req); err != nil || req.Interface == "" || req.MTU == "" {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	err := dbus.SetMTU(req.Interface, req.MTU)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "ok"})
}

func postSetIPv4DHCP(c *gin.Context) {
	var req struct {
		Interface string `json:"interface"`
	}
	if err := c.BindJSON(&req); err != nil || req.Interface == "" {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	err := dbus.SetIPv4DHCP(req.Interface)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "ok"})
}

func postSetIPv4Static(c *gin.Context) {
	var req struct {
		Interface   string `json:"interface"`
		AddressCIDR string `json:"address_cidr"`
	}
	if err := c.BindJSON(&req); err != nil || req.Interface == "" || req.AddressCIDR == "" {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	err := dbus.SetIPv4Static(req.Interface, req.AddressCIDR)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "ok"})
}

func postSetIPv6DHCP(c *gin.Context) {
	var req struct {
		Interface string `json:"interface"`
	}
	if err := c.BindJSON(&req); err != nil || req.Interface == "" {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	err := dbus.SetIPv6DHCP(req.Interface)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "ok"})
}

func postSetIPv6Static(c *gin.Context) {
	var req struct {
		Interface   string `json:"interface"`
		AddressCIDR string `json:"address_cidr"`
	}
	if err := c.BindJSON(&req); err != nil || req.Interface == "" || req.AddressCIDR == "" {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	err := dbus.SetIPv6Static(req.Interface, req.AddressCIDR)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "ok"})
}
