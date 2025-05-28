package dbus

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/shirou/gopsutil/v4/net"
)

type NMInterfaceInfo struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"` // ethernet, wifi, loopback, etc.
	MAC          string   `json:"mac"`
	MTU          uint32   `json:"mtu"`
	Speed        string   `json:"speed"`  // from /sys/class/net/<iface>/speed
	Duplex       string   `json:"duplex"` // from /sys/class/net/<iface>/duplex
	State        uint32   `json:"state"`
	IP4Addresses []string `json:"ipv4"`
	IP6Addresses []string `json:"ipv6"`
	RxSpeed      float64  `json:"rx_speed"`
	TxSpeed      float64  `json:"tx_speed"`
	DNS          []string `json:"dns"`
	Gateway      string   `json:"gateway"`
}

var (
	lastNetStats  = make(map[string]net.IOCountersStat)
	lastTimestamp int64
)

func mapDeviceType(devType uint32) string {
	switch devType {
	case 1:
		return "ethernet"
	case 2:
		return "wifi"
	case 5:
		return "bt"
	case 6:
		return "olpc-mesh"
	case 7:
		return "wimax"
	case 8:
		return "modem"
	case 9:
		return "infiniband"
	case 10:
		return "bond"
	case 11:
		return "vlan"
	case 14:
		return "bridge"
	case 17:
		return "tun"
	case 27:
		return "ovs-bridge"
	default:
		return "unknown"
	}
}

func GetNetworkInfo() ([]NMInterfaceInfo, error) {
	var results []NMInterfaceInfo

	snapshots, _ := net.IOCounters(true)
	snapshotMap := make(map[string]net.IOCountersStat)
	for _, s := range snapshots {
		snapshotMap[s.Name] = s
	}

	err := RetryOnceIfClosed(nil, func() error {
		conn, err := dbus.SystemBus()
		if err != nil {
			return fmt.Errorf("failed to connect to system bus: %w", err)
		}
		defer conn.Close()

		nm := conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")

		var devicePaths []dbus.ObjectPath
		if err := nm.Call("org.freedesktop.NetworkManager.GetDevices", 0).Store(&devicePaths); err != nil {
			return fmt.Errorf("GetDevices failed: %w", err)
		}

		for _, devPath := range devicePaths {
			dev := conn.Object("org.freedesktop.NetworkManager", devPath)

			props := make(map[string]dbus.Variant)
			if err := dev.Call("org.freedesktop.DBus.Properties.GetAll", 0, "org.freedesktop.NetworkManager.Device").Store(&props); err != nil {
				continue
			}

			name, _ := props["Interface"].Value().(string)
			mac, _ := props["HwAddress"].Value().(string)

			devType := uint32(0)
			if v, ok := props["DeviceType"]; ok {
				if cast, ok := v.Value().(uint32); ok {
					devType = cast
				}
			}
			ifaceType := mapDeviceType(devType)

			mtu := uint32(0)
			if v, ok := props["Mtu"]; ok && v.Value() != nil {
				if cast, ok := v.Value().(uint32); ok {
					mtu = cast
				}
			}

			state := uint32(0)
			if v, ok := props["State"]; ok && v.Value() != nil {
				if cast, ok := v.Value().(uint32); ok {
					state = cast
				}
			}

			speed := "unknown"
			duplex := "unknown"
			if name != "" {
				speedPath := fmt.Sprintf("/sys/class/net/%s/speed", name)
				duplexPath := fmt.Sprintf("/sys/class/net/%s/duplex", name)

				if b, err := os.ReadFile(speedPath); err == nil {
					speed = strings.TrimSpace(string(b)) + " Mbps"
				}
				if b, err := os.ReadFile(duplexPath); err == nil {
					duplex = strings.TrimSpace(string(b))
				}
			}

			var ip4s, ip6s, dns []string
			gateway := ""

			if ip4Path, ok := props["Ip4Config"].Value().(dbus.ObjectPath); ok && ip4Path != "/" {
				ip4Obj := conn.Object("org.freedesktop.NetworkManager", ip4Path)
				var ip4Props map[string]dbus.Variant
				if err := ip4Obj.Call("org.freedesktop.DBus.Properties.GetAll", 0, "org.freedesktop.NetworkManager.IP4Config").Store(&ip4Props); err == nil {
					if addresses, ok := ip4Props["Addresses"].Value().([][]uint32); ok {
						for _, addr := range addresses {
							ip := fmt.Sprintf("%d.%d.%d.%d/%d",
								byte(addr[0]), byte(addr[0]>>8), byte(addr[0]>>16), byte(addr[0]>>24), addr[1])
							ip4s = append(ip4s, ip)
						}
					}
					if dnsAddrs, ok := ip4Props["Nameservers"].Value().([]uint32); ok {
						for _, ip := range dnsAddrs {
							dns = append(dns, fmt.Sprintf("%d.%d.%d.%d", byte(ip), byte(ip>>8), byte(ip>>16), byte(ip>>24)))
						}
					}
					if gw, ok := ip4Props["Gateway"].Value().(string); ok {
						gateway = gw
					}
				}
			}

			if ip6Path, ok := props["Ip6Config"].Value().(dbus.ObjectPath); ok && ip6Path != "/" {
				ip6Obj := conn.Object("org.freedesktop.NetworkManager", ip6Path)
				var ip6Props map[string]dbus.Variant
				if err := ip6Obj.Call("org.freedesktop.DBus.Properties.GetAll", 0, "org.freedesktop.NetworkManager.IP6Config").Store(&ip6Props); err == nil {
					if addresses, ok := ip6Props["Addresses"].Value().([][]interface{}); ok {
						for _, tuple := range addresses {
							addrBytes, _ := tuple[0].([]byte)
							prefix, _ := tuple[1].(uint32)
							if len(addrBytes) == 16 {
								parts := make([]string, 8)
								for i := 0; i < 8; i++ {
									parts[i] = fmt.Sprintf("%02x%02x", addrBytes[2*i], addrBytes[2*i+1])
								}
								ip6s = append(ip6s, fmt.Sprintf("%s/%d", strings.Join(parts, ":"), prefix))
							}
						}
					}
					if dns6, ok := ip6Props["Nameservers"].Value().([][]byte); ok {
						for _, addr := range dns6 {
							if len(addr) == 16 {
								parts := make([]string, 8)
								for i := 0; i < 8; i++ {
									parts[i] = fmt.Sprintf("%02x%02x", addr[2*i], addr[2*i+1])
								}
								dns = append(dns, strings.Join(parts, ":"))
							}
						}
					}
					if gw, ok := ip6Props["Gateway"].Value().(string); ok && gateway == "" {
						gateway = gw
					}
				}
			}

			rxSpeed := 0.0
			txSpeed := 0.0
			if snapshot, ok := snapshotMap[name]; ok {
				now := time.Now().Unix()
				interval := now - lastTimestamp
				if interval < 1 {
					interval = 1
				}
				if prev, ok := lastNetStats[name]; ok {
					rxSpeed = float64(snapshot.BytesRecv-prev.BytesRecv) / float64(interval)
					txSpeed = float64(snapshot.BytesSent-prev.BytesSent) / float64(interval)
				}
				lastNetStats[name] = snapshot
				lastTimestamp = now
			}

			results = append(results, NMInterfaceInfo{
				Name:         name,
				Type:         ifaceType,
				MAC:          mac,
				MTU:          mtu,
				Speed:        speed,
				Duplex:       duplex,
				State:        state,
				IP4Addresses: ip4s,
				IP6Addresses: ip6s,
				RxSpeed:      rxSpeed,
				TxSpeed:      txSpeed,
				DNS:          dns,
				Gateway:      gateway,
			})
		}
		return nil
	})

	return results, err
}

func SetDNS(iface string, dns []string) error {
	// Use nmcli for simplicity (you can D-Bus it if you prefer)
	dnsArg := ""
	for i, d := range dns {
		if i > 0 {
			dnsArg += ","
		}
		dnsArg += d
	}
	cmd := exec.Command("nmcli", "con", "mod", iface, "ipv4.dns", dnsArg)
	if err := cmd.Run(); err != nil {
		return err
	}
	// Apply
	cmd = exec.Command("nmcli", "con", "up", iface)
	return cmd.Run()
}

func SetGateway(iface, gateway string) error {
	cmd := exec.Command("nmcli", "con", "mod", iface, "ipv4.gateway", gateway)
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("nmcli", "con", "up", iface)
	return cmd.Run()
}

func SetIPv4DHCP(iface string) error {
	cmd := exec.Command("nmcli", "con", "mod", iface, "ipv4.method", "auto")
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("nmcli", "con", "up", iface)
	return cmd.Run()
}

func SetIPv4Static(iface, addressCIDR string) error {
	cmd := exec.Command("nmcli", "con", "mod", iface, "ipv4.addresses", addressCIDR, "ipv4.method", "manual")
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("nmcli", "con", "up", iface)
	return cmd.Run()
}

func SetIPv6DHCP(iface string) error {
	cmd := exec.Command("nmcli", "con", "mod", iface, "ipv6.method", "auto")
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("nmcli", "con", "up", iface)
	return cmd.Run()
}

func SetIPv6Static(iface, addressCIDR string) error {
	cmd := exec.Command("nmcli", "con", "mod", iface, "ipv6.addresses", addressCIDR, "ipv6.method", "manual")
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("nmcli", "con", "up", iface)
	return cmd.Run()
}

func SetMTU(iface, mtu string) error {
	cmd := exec.Command("nmcli", "con", "mod", iface, "802-3-ethernet.mtu", mtu)
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("nmcli", "con", "up", iface)
	return cmd.Run()
}
