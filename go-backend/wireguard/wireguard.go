package wireguard

import (
	"fmt"
	"go-backend/auth" // Make sure your auth middleware is implemented
	"net/http"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// RegisterWireguardRoutes registers the routes for WireGuard setup and interface details
func RegisterWireguardRoutes(router *gin.Engine) {
	system := router.Group("/wireguard", auth.AuthMiddleware()) // Assuming auth middleware is implemented
	{
		system.POST("/setup", SetupInterfaceHandler)        // Setup the WireGuard interface
		system.GET("/interface/:name", GetInterfaceDetails) // Get specific interface details
	}
}

// SetupInterfaceHandler handles user input and configures the WireGuard interface
func SetupInterfaceHandler(c *gin.Context) {
	var input SetupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := SetupInterface(input.Name, input.Endpoint, input.ListenPort, input.NumPeers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "WireGuard interface setup successfully",
		"name":    input.Name,
		"peers":   input.NumPeers,
	})
}

// GetInterfaceDetails retrieves the WireGuard interface details and its peers
func GetInterfaceDetails(c *gin.Context) {
	name := c.Param("name")

	iface, err := GetInterface(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"interface": iface,
	})
}

// GenerateKeyPair generates a new key pair (private & public) for WireGuard
func GenerateKeyPair() (privateKey string, publicKey string, err error) {
	priv, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return "", "", err
	}
	pub := priv.PublicKey()
	return priv.String(), pub.String(), nil
}

// SetupInterface configures the WireGuard interface with peers and the provided details
func SetupInterface(name, endpoint string, listenPort, numPeers int) error {
	// Create the WireGuard interface
	if err := CreateInterface(name); err != nil {
		return err
	}

	// Set the listen port
	if err := SetListenPort(name, listenPort); err != nil {
		return err
	}

	// Generate the private key for the interface
	privKey, _, err := GenerateKeyPair()
	if err != nil {
		return err
	}

	// Set the private key on the interface
	if err := SetPrivateKey(name, privKey); err != nil {
		return err
	}

	// Add the peer(s)
	for i := 0; i < numPeers; i++ {
		// Generate random public key for each peer (simplified here)
		peerPubKey := fmt.Sprintf("PEER%d_PUBLIC_KEY", i)
		peerAllowedIPs := fmt.Sprintf("10.0.0.%d/32", i+2) // Example IPs (10.0.0.2/32, 10.0.0.3/32, etc.)

		if err := AddPeer(name, peerPubKey, []string{peerAllowedIPs}); err != nil {
			return err
		}
	}

	return nil
}

// SetListenPort configures the listen port for the WireGuard interface
func SetListenPort(name string, port int) error {
	// Normally this would configure routing to specific ports, but we can assume it's done here
	cmd := exec.Command("wg", "set", name, "listen-port", fmt.Sprintf("%d", port))
	return cmd.Run()
}

// CreateInterface creates a WireGuard interface
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

// AddPeer adds a peer to the WireGuard interface
func AddPeer(name, pubkey string, allowedIPs []string) error {
	cmd := exec.Command("wg", "set", name,
		"peer", pubkey,
		"allowed-ips", fmt.Sprintf("%s", allowedIPs),
	)
	return cmd.Run()
}

// SetPrivateKey sets the private key on the WireGuard interface
func SetPrivateKey(name, privateKey string) error {
	cmd := exec.Command("wg", "set", name, "private-key", privateKey)
	return cmd.Run()
}

// ListInterfaces lists all the WireGuard interfaces
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

// GetInterface retrieves details of a specific WireGuard interface by name
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

// convertDevice converts a wireguard device to WGInterface
func convertDevice(dev *wgtypes.Device) WGInterface {
	var peers []WGPeer
	for _, peer := range dev.Peers {
		var allowed []string
		for _, ip := range peer.AllowedIPs {
			allowed = append(allowed, ip.String())
		}

		// Get the last handshake and endpoint status for each peer
		lastHandshake := "never"
		if !peer.LastHandshakeTime.IsZero() {
			lastHandshake = peer.LastHandshakeTime.Format(time.RFC3339)
		}

		peers = append(peers, WGPeer{
			PublicKey:     peer.PublicKey.String(),
			Endpoint:      peer.Endpoint.String(),
			AllowedIPs:    allowed,
			LastHandshake: lastHandshake, // Add this field
		})
	}

	// Return the device interface details
	return WGInterface{
		Name:       dev.Name,
		PublicKey:  dev.PublicKey.String(),
		ListenPort: dev.ListenPort,
		Peers:      peers,
	}
}
