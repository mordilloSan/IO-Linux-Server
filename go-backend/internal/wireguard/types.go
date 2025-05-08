package wireguard

// SetupInput - Structure to bind the input data
type SetupInput struct {
	Name       string `json:"name"`
	Endpoint   string `json:"endpoint"`
	ListenPort int    `json:"listenPort"`
	NumPeers   int    `json:"numPeers"`
}

type WGInterface struct {
	Name       string   `json:"name"`
	PublicKey  string   `json:"publicKey"`
	ListenPort int      `json:"listenPort"`
	Peers      []WGPeer `json:"peers"`
}

type WGPeer struct {
	PublicKey     string   `json:"publicKey"`
	Endpoint      string   `json:"endpoint,omitempty"`
	AllowedIPs    []string `json:"allowedIPs"`
	LastHandshake string   `json:"lastHandshake,omitempty"` // Add last handshake
}
