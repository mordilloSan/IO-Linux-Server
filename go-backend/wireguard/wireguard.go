package wireguard

import (
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"go-backend/auth"
	"go-backend/logger"

	"github.com/gin-gonic/gin"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func RegisterWireguardRoutes(router *gin.Engine) {
	system := router.Group("/wireguard", auth.AuthMiddleware())
	{
		system.POST("/setup", SetupInterfaceHandler)
		system.GET("/interface/:name", GetInterfaceDetails)
	}
}

func SetupInterfaceHandler(c *gin.Context) {
	var input SetupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Warning.Printf("Invalid WireGuard setup input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := SetupInterface(input.Name, input.Endpoint, input.ListenPort, input.NumPeers); err != nil {
		logger.Error.Printf("Failed to setup WireGuard interface %s: %v", input.Name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info.Printf("WireGuard interface %s configured with %d peers", input.Name, input.NumPeers)
	c.JSON(http.StatusOK, gin.H{
		"message": "WireGuard interface setup successfully",
		"name":    input.Name,
		"peers":   input.NumPeers,
	})
}

func GetInterfaceDetails(c *gin.Context) {
	name := c.Param("name")

	iface, err := GetInterface(name)
	if err != nil {
		logger.Warning.Printf("Failed to retrieve WireGuard interface %s: %v", name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"interface": iface})
}

func GenerateKeyPair() (string, string, error) {
	priv, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return "", "", err
	}
	pub := priv.PublicKey()
	return priv.String(), pub.String(), nil
}

func SetupInterface(name, endpoint string, listenPort, numPeers int) error {
	logger.Info.Printf("Creating WireGuard interface: %s", name)

	if err := CreateInterface(name); err != nil {
		return err
	}
	if err := SetListenPort(name, listenPort); err != nil {
		return err
	}

	privKey, _, err := GenerateKeyPair()
	if err != nil {
		return err
	}
	if err := SetPrivateKey(name, privKey); err != nil {
		return err
	}

	for i := 0; i < numPeers; i++ {
		peerPubKey := fmt.Sprintf("PEER%d_PUBLIC_KEY", i)
		peerAllowedIPs := fmt.Sprintf("10.0.0.%d/32", i+2)

		if err := AddPeer(name, peerPubKey, []string{peerAllowedIPs}); err != nil {
			logger.Warning.Printf("Failed to add peer %d to %s: %v", i, name, err)
			return err
		}
		logger.Debug.Printf("Added peer %d to %s with IP %s", i, name, peerAllowedIPs)
	}

	return nil
}

func SetListenPort(name string, port int) error {
	return exec.Command("wg", "set", name, "listen-port", fmt.Sprintf("%d", port)).Run()
}

func CreateInterface(name string) error {
	wg := &netlink.GenericLink{
		LinkAttrs: netlink.LinkAttrs{Name: name},
		LinkType:  "wireguard",
	}
	if err := netlink.LinkAdd(wg); err != nil {
		return err
	}
	return netlink.LinkSetUp(wg)
}

func AddPeer(name, pubkey string, allowedIPs []string) error {
	cmd := exec.Command("wg", "set", name,
		"peer", pubkey,
		"allowed-ips", fmt.Sprintf("%s", allowedIPs),
	)
	return cmd.Run()
}

func SetPrivateKey(name, privateKey string) error {
	return exec.Command("wg", "set", name, "private-key", privateKey).Run()
}

func ListInterfaces() ([]WGInterface, error) {
	client, err := wgctrl.New()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	devices, err := client.Devices()
	if err != nil {
		return nil, err
	}

	var result []WGInterface
	for _, dev := range devices {
		result = append(result, convertDevice(dev))
	}
	return result, nil
}

func GetInterface(name string) (*WGInterface, error) {
	client, err := wgctrl.New()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	dev, err := client.Device(name)
	if err != nil {
		return nil, err
	}

	iface := convertDevice(dev)
	return &iface, nil
}

func convertDevice(dev *wgtypes.Device) WGInterface {
	var peers []WGPeer
	for _, peer := range dev.Peers {
		var allowed []string
		for _, ip := range peer.AllowedIPs {
			allowed = append(allowed, ip.String())
		}

		lastHandshake := "never"
		if !peer.LastHandshakeTime.IsZero() {
			lastHandshake = peer.LastHandshakeTime.Format(time.RFC3339)
		}

		endpoint := ""
		if peer.Endpoint != nil {
			endpoint = peer.Endpoint.String()
		}

		peers = append(peers, WGPeer{
			PublicKey:     peer.PublicKey.String(),
			Endpoint:      endpoint,
			AllowedIPs:    allowed,
			LastHandshake: lastHandshake,
		})
	}

	return WGInterface{
		Name:       dev.Name,
		PublicKey:  dev.PublicKey.String(),
		ListenPort: dev.ListenPort,
		Peers:      peers,
	}
}
