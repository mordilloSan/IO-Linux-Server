package dbus

import (
	"fmt"
	"os"
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

func GetNetworkInterfaces() ([]NMInterfaceInfo, error) {
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

			var ip4s, ip6s []string
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
			})
		}
		return nil
	})

	return results, err
}
