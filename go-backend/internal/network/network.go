package network

import (
	"encoding/json"
	"go-backend/cmd/bridge/dbus"
	"go-backend/internal/bridge"
	"go-backend/internal/logger"
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
		logger.Warning.Printf("[network] Unauthorized getNetworkInfo")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}
	logger.Info.Printf("[network] %s requested network info (session: %s)", user.ID, sessionID)

	rawResp, err := bridge.CallWithSession(sessionID, user.ID, "dbus", "GetNetworkInfo", nil)
	if err != nil {
		logger.Error.Printf("[network] Bridge call failed: %v", err)
		c.JSON(500, gin.H{"error": "bridge call failed", "detail": err.Error(), "output": rawResp})
		return
	}
	var resp bridge.BridgeResponse
	if err := json.Unmarshal([]byte(rawResp), &resp); err != nil {
		logger.Error.Printf("[network] Invalid bridge response: %v", err)
		c.JSON(500, gin.H{"error": "invalid bridge response", "detail": err.Error(), "output": rawResp})
		return
	}
	if resp.Status != "ok" {
		logger.Warning.Printf("[network] Bridge returned error: %v", resp.Error)
		c.JSON(500, gin.H{"error": resp.Error, "output": string(resp.Output)})
		return
	}
	var data []dbus.NMInterfaceInfo
	if err := json.Unmarshal(resp.Output, &data); err != nil {
		logger.Error.Printf("[network] Invalid output structure: %v", err)
		c.JSON(500, gin.H{"error": "invalid output structure", "detail": err.Error(), "output": string(resp.Output)})
		return
	}
	logger.Debug.Printf("[network] Successfully returned %d interfaces to %s", len(data), user.ID)
	c.JSON(200, data)
}

func postSetDNS(c *gin.Context) {
	var req struct {
		Interface string   `json:"interface"`
		DNS       []string `json:"dns"`
	}
	user, sessionID, valid, _ := session.ValidateFromRequest(c.Request)
	if !valid {
		logger.Warning.Printf("[network] Unauthorized set-dns")
		c.JSON(401, gin.H{"error": "invalid session"})
		return
	}
	if err := c.BindJSON(&req); err != nil || req.Interface == "" || len(req.DNS) == 0 {
		logger.Warning.Printf("[network] Bad request for set-dns: %+v", req)
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	logger.Info.Printf("[network] %s sets DNS on %s: %v", user.ID, req.Interface, req.DNS)
	err := dbus.SetDNS(req.Interface, req.DNS)
	if err != nil {
		logger.Error.Printf("[network] Failed to set DNS on %s: %v", req.Interface, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	logger.Info.Printf("[network] Set DNS on %s to %v (user: %s, session: %s)", req.Interface, req.DNS, user.ID, sessionID)
	c.JSON(200, gin.H{"status": "ok"})
}

func postSetGateway(c *gin.Context) {
	var req struct {
		Interface string `json:"interface"`
		Gateway   string `json:"gateway"`
	}
	user, sessionID, valid, _ := session.ValidateFromRequest(c.Request)
	if !valid {
		logger.Warning.Printf("[network] Unauthorized set-gateway")
		c.JSON(401, gin.H{"error": "invalid session"})
		return
	}
	if err := c.BindJSON(&req); err != nil || req.Interface == "" || req.Gateway == "" {
		logger.Warning.Printf("[network] Bad request for set-gateway: %+v", req)
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	logger.Info.Printf("[network] %s sets gateway on %s: %s", user.ID, req.Interface, req.Gateway)
	err := dbus.SetGateway(req.Interface, req.Gateway)
	if err != nil {
		logger.Error.Printf("[network] Failed to set gateway on %s: %v", req.Interface, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	logger.Info.Printf("[network] Set gateway on %s to %s (user: %s, session: %s)", req.Interface, req.Gateway, user.ID, sessionID)
	c.JSON(200, gin.H{"status": "ok"})
}

func postSetMTU(c *gin.Context) {
	var req struct {
		Interface string `json:"interface"`
		MTU       string `json:"mtu"`
	}
	user, sessionID, valid, _ := session.ValidateFromRequest(c.Request)
	if !valid {
		logger.Warning.Printf("[network] Unauthorized set-mtu")
		c.JSON(401, gin.H{"error": "invalid session"})
		return
	}
	if err := c.BindJSON(&req); err != nil || req.Interface == "" || req.MTU == "" {
		logger.Warning.Printf("[network] Bad request for set-mtu: %+v", req)
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	logger.Info.Printf("[network] %s sets MTU on %s: %s", user.ID, req.Interface, req.MTU)
	err := dbus.SetMTU(req.Interface, req.MTU)
	if err != nil {
		logger.Error.Printf("[network] Failed to set MTU on %s: %v", req.Interface, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	logger.Info.Printf("[network] Set MTU on %s to %s (user: %s, session: %s)", req.Interface, req.MTU, user.ID, sessionID)
	c.JSON(200, gin.H{"status": "ok"})
}

func postSetIPv4DHCP(c *gin.Context) {
	var req struct {
		Interface string `json:"interface"`
	}
	user, sessionID, valid, _ := session.ValidateFromRequest(c.Request)
	if !valid {
		logger.Warning.Printf("[network] Unauthorized set-ipv4-dhcp")
		c.JSON(401, gin.H{"error": "invalid session"})
		return
	}
	if err := c.BindJSON(&req); err != nil || req.Interface == "" {
		logger.Warning.Printf("[network] Bad request for set-ipv4-dhcp: %+v", req)
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	logger.Info.Printf("[network] %s requests IPv4 DHCP on %s", user.ID, req.Interface)
	err := dbus.SetIPv4DHCP(req.Interface)
	if err != nil {
		logger.Error.Printf("[network] Failed to set IPv4 DHCP on %s: %v", req.Interface, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	logger.Info.Printf("[network] Set IPv4 DHCP on %s (user: %s, session: %s)", req.Interface, user.ID, sessionID)
	c.JSON(200, gin.H{"status": "ok"})
}

func postSetIPv4Static(c *gin.Context) {
	var req struct {
		Interface   string `json:"interface"`
		AddressCIDR string `json:"address_cidr"`
	}
	user, sessionID, valid, _ := session.ValidateFromRequest(c.Request)
	if !valid {
		logger.Warning.Printf("[network] Unauthorized set-ipv4-static")
		c.JSON(401, gin.H{"error": "invalid session"})
		return
	}
	if err := c.BindJSON(&req); err != nil || req.Interface == "" || req.AddressCIDR == "" {
		logger.Warning.Printf("[network] Bad request for set-ipv4-static: %+v", req)
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	logger.Info.Printf("[network] %s sets IPv4 static on %s: %s", user.ID, req.Interface, req.AddressCIDR)
	err := dbus.SetIPv4Static(req.Interface, req.AddressCIDR)
	if err != nil {
		logger.Error.Printf("[network] Failed to set IPv4 static on %s: %v", req.Interface, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	logger.Info.Printf("[network] Set IPv4 static on %s to %s (user: %s, session: %s)", req.Interface, req.AddressCIDR, user.ID, sessionID)
	c.JSON(200, gin.H{"status": "ok"})
}

func postSetIPv6DHCP(c *gin.Context) {
	var req struct {
		Interface string `json:"interface"`
	}
	user, sessionID, valid, _ := session.ValidateFromRequest(c.Request)
	if !valid {
		logger.Warning.Printf("[network] Unauthorized set-ipv6-dhcp")
		c.JSON(401, gin.H{"error": "invalid session"})
		return
	}
	if err := c.BindJSON(&req); err != nil || req.Interface == "" {
		logger.Warning.Printf("[network] Bad request for set-ipv6-dhcp: %+v", req)
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	logger.Info.Printf("[network] %s requests IPv6 DHCP on %s", user.ID, req.Interface)
	err := dbus.SetIPv6DHCP(req.Interface)
	if err != nil {
		logger.Error.Printf("[network] Failed to set IPv6 DHCP on %s: %v", req.Interface, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	logger.Info.Printf("[network] Set IPv6 DHCP on %s (user: %s, session: %s)", req.Interface, user.ID, sessionID)
	c.JSON(200, gin.H{"status": "ok"})
}

func postSetIPv6Static(c *gin.Context) {
	var req struct {
		Interface   string `json:"interface"`
		AddressCIDR string `json:"address_cidr"`
	}
	user, sessionID, valid, _ := session.ValidateFromRequest(c.Request)
	if !valid {
		logger.Warning.Printf("[network] Unauthorized set-ipv6-static")
		c.JSON(401, gin.H{"error": "invalid session"})
		return
	}
	if err := c.BindJSON(&req); err != nil || req.Interface == "" || req.AddressCIDR == "" {
		logger.Warning.Printf("[network] Bad request for set-ipv6-static: %+v", req)
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	logger.Info.Printf("[network] %s sets IPv6 static on %s: %s", user.ID, req.Interface, req.AddressCIDR)
	err := dbus.SetIPv6Static(req.Interface, req.AddressCIDR)
	if err != nil {
		logger.Error.Printf("[network] Failed to set IPv6 static on %s: %v", req.Interface, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	logger.Info.Printf("[network] Set IPv6 static on %s to %s (user: %s, session: %s)", req.Interface, req.AddressCIDR, user.ID, sessionID)
	c.JSON(200, gin.H{"status": "ok"})
}
