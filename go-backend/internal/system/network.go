package system

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v4/net"
)

type netStats struct {
	lastRecv uint64
	lastSent uint64
	lastTime int64
}

var (
	netStatsMap  = make(map[string]netStats)
	netStatsLock = sync.Mutex{}
)

func getInterfaceSpeed(interfaceName string) (string, string, error) {
	speedPath := fmt.Sprintf("/sys/class/net/%s/speed", interfaceName)
	duplexPath := fmt.Sprintf("/sys/class/net/%s/duplex", interfaceName)

	speedBytes, err := os.ReadFile(speedPath)
	if err != nil {
		return "", "", err
	}
	duplexBytes, err := os.ReadFile(duplexPath)
	if err != nil {
		return strings.TrimSpace(string(speedBytes)), "", err
	}

	return strings.TrimSpace(string(speedBytes)), strings.TrimSpace(string(duplexBytes)), nil
}

func FetchNetworkInfo() ([]map[string]any, error) {
	stats, _ := net.IOCounters(true)
	ifaces, _ := net.Interfaces()

	now := time.Now().Unix()
	result := []map[string]any{}

	netStatsLock.Lock()
	defer netStatsLock.Unlock()

	for _, iface := range ifaces {
		if strings.HasPrefix(iface.Name, "lo") || strings.HasPrefix(iface.Name, "docker") || strings.HasPrefix(iface.Name, "veth") {
			continue
		}

		entry := map[string]any{
			"name":         iface.Name,
			"mtu":          iface.MTU,
			"hardwareAddr": iface.HardwareAddr,
			"flags":        iface.Flags,
			"addresses":    []string{},
			"bytesSent":    uint64(0),
			"bytesRecv":    uint64(0),
			"txSpeed":      float64(0),
			"rxSpeed":      float64(0),
			"linkSpeed":    "unknown",
			"duplex":       "unknown",
		}

		speed, duplex, err := getInterfaceSpeed(iface.Name)
		if err == nil {
			if speed != "" {
				entry["linkSpeed"] = speed
			}
			if duplex != "" {
				entry["duplex"] = duplex
			}
		}

		for _, addr := range iface.Addrs {
			entry["addresses"] = append(entry["addresses"].([]string), addr.Addr)
		}

		for _, s := range stats {
			if s.Name == iface.Name {
				entry["bytesSent"] = s.BytesSent
				entry["bytesRecv"] = s.BytesRecv

				prev := netStatsMap[iface.Name]
				interval := now - prev.lastTime
				if interval > 0 {
					entry["txSpeed"] = float64(s.BytesSent-prev.lastSent) / float64(interval)
					entry["rxSpeed"] = float64(s.BytesRecv-prev.lastRecv) / float64(interval)
				}

				netStatsMap[iface.Name] = netStats{
					lastRecv: s.BytesRecv,
					lastSent: s.BytesSent,
					lastTime: now,
				}
				break
			}
		}

		result = append(result, entry)
	}

	return result, nil
}

func getNetworkInfo(c *gin.Context) {
	data, err := FetchNetworkInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get network info", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
