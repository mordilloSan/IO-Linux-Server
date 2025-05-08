package system

import (
	"fmt"
	"go-backend/internal/auth"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jaypipes/ghw/pkg/gpu"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

func RegisterSystemRoutes(router *gin.Engine) {
	system := router.Group("/system", auth.AuthMiddleware())
	{
		system.GET("/info", getHostInfo)
		system.GET("/cpu", getCPUInfo)
		system.GET("/mem", getMemInfo)
		system.GET("/disk", getDiskInfo)
		system.GET("/network", getNetworkInfo)
		system.GET("/load", getLoadInfo)
		system.GET("/uptime", getUptime)
		system.GET("/processes", getProcesses)
		system.GET("/baseboard", getBaseboardInfo)
		system.GET("/gpu", getGPUInfo)
	}
}

func getHostInfo(c *gin.Context) {
	hostInfo, err := host.Info()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get host info", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, hostInfo)
}

func getCPUInfo(c *gin.Context) {
	info, err := cpu.Info()
	if err != nil || len(info) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get CPU info", "details": err.Error()})
		return
	}
	percent, _ := cpu.Percent(0, true)
	counts, _ := cpu.Counts(true)
	loadAvg, _ := load.Avg()
	temp := getTemperatureMap()

	cpuData := info[0]
	c.JSON(http.StatusOK, gin.H{
		"vendorId":     cpuData.VendorID,
		"modelName":    cpuData.ModelName,
		"family":       cpuData.Family,
		"model":        cpuData.Model,
		"mhz":          cpuData.Mhz,
		"cores":        counts,
		"loadAverage":  loadAvg,
		"perCoreUsage": percent,
		"temperature":  temp,
	})
}

func getTemperatureMap() map[string]float64 {
	out, err := exec.Command("sensors").Output()
	if err != nil {
		return map[string]float64{}
	}

	temps := make(map[string]float64)
	lines := strings.Split(string(out), "\n")

	re := regexp.MustCompile(`\+([0-9]+(?:\.[0-9]+)?)°C`)
	currentAdapter := ""

	cpuIndex := 1
	mbIndex := 1

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Track adapter blocks (like acpitz-acpi-0, coretemp-isa-0000, etc.)
		if !strings.HasPrefix(line, " ") && !strings.Contains(line, ":") && line != "" {
			currentAdapter = strings.ToLower(line)
			continue
		}

		// Extract temperatures
		match := re.FindStringSubmatch(line)
		if len(match) < 2 {
			continue
		}

		tempVal, err := strconv.ParseFloat(match[1], 64)
		if err != nil {
			continue
		}

		lowerLine := strings.ToLower(line)

		// Heuristic for CPU temps
		if strings.HasPrefix(lowerLine, "core") ||
			strings.Contains(lowerLine, "package id") ||
			strings.Contains(lowerLine, "tctl") {
			key := "cpu" + strconv.Itoa(cpuIndex)
			temps[key] = tempVal
			cpuIndex++
			continue
		}

		// Heuristic for Motherboard temps
		if strings.Contains(currentAdapter, "acpitz") || // ACPI thermal zone
			strings.Contains(lowerLine, "mb") ||
			strings.Contains(lowerLine, "board") ||
			strings.Contains(lowerLine, "system") ||
			strings.Contains(lowerLine, "systin") ||
			strings.Contains(lowerLine, "temp1") {

			key := "mb" + strconv.Itoa(mbIndex)
			temps[key] = tempVal
			mbIndex++
			continue
		}
	}

	return temps
}

func getMemInfo(c *gin.Context) {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get memory info", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, memInfo)
}

func getDiskInfo(c *gin.Context) {
	parts, err := disk.Partitions(true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get disk partitions", "details": err.Error()})
		return
	}

	var results []map[string]any
	for _, p := range parts {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}
		results = append(results, map[string]any{
			"device":      p.Device,
			"mountpoint":  p.Mountpoint,
			"fstype":      p.Fstype,
			"total":       usage.Total,
			"used":        usage.Used,
			"free":        usage.Free,
			"usedPercent": usage.UsedPercent,
		})
	}
	c.JSON(http.StatusOK, results)
}

type netStats struct {
	lastRecv uint64
	lastSent uint64
	lastTime int64
}

var netStatsMap = make(map[string]netStats)
var netStatsLock = sync.Mutex{}

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

func getNetworkInfo(c *gin.Context) {
	stats, _ := net.IOCounters(true)
	ifaces, _ := net.Interfaces()

	now := time.Now().Unix()
	result := []map[string]any{}

	netStatsLock.Lock()
	defer netStatsLock.Unlock()

	for _, iface := range ifaces {
		// Skip loopback and docker/veth
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

		// Try to fetch speed only for real interfaces
		speed, duplex, err := getInterfaceSpeed(iface.Name)
		if err == nil {
			if speed != "" {
				entry["linkSpeed"] = speed
			}
			if duplex != "" {
				entry["duplex"] = duplex
			}
		}

		// Add IP addresses
		for _, addr := range iface.Addrs {
			entry["addresses"] = append(entry["addresses"].([]string), addr.Addr)
		}

		// Add IO counters
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

	c.JSON(http.StatusOK, result)
}

func getLoadInfo(c *gin.Context) {
	loadAvg, err := load.Avg()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get load average", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, loadAvg)
}

func getUptime(c *gin.Context) {
	uptime, err := host.Uptime()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get uptime", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"uptime_seconds": uptime})
}

func getProcesses(c *gin.Context) {
	procs, err := process.Processes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get processes", "details": err.Error()})
		return
	}
	var result []map[string]any
	for _, p := range procs {
		name, _ := p.Name()
		cpu, _ := p.CPUPercent()
		mem, _ := p.MemoryPercent()
		result = append(result, map[string]any{
			"pid":         p.Pid,
			"name":        name,
			"cpu_percent": cpu,
			"mem_percent": mem,
		})
	}
	c.JSON(http.StatusOK, result)
}

func getBaseboardInfo(c *gin.Context) {
	basePath := "/sys/class/dmi/id"
	fields := map[string]string{
		"board_name":    "model",
		"board_vendor":  "manufacturer",
		"board_version": "version",
		"board_serial":  "serial",
		"bios_vendor":   "bios_vendor",
		"bios_version":  "bios_version",
		"bios_date":     "bios_date",
	}

	info := make(map[string]string)
	for file, label := range fields {
		content, err := os.ReadFile(filepath.Join(basePath, file))
		if err == nil {
			info[label] = strings.TrimSpace(string(content))
		}
	}

	if len(info) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to read motherboard info",
		})
		return
	}

	// Include motherboard temperatures
	tempMap := getTemperatureMap()
	socketTemps := make([]float64, 0)

	for key, value := range tempMap {
		if strings.HasPrefix(key, "mb") {
			socketTemps = append(socketTemps, value)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"baseboard": map[string]string{
			"manufacturer": info["manufacturer"],
			"model":        info["model"],
			"version":      info["version"],
			"serial":       info["serial"],
		},
		"bios": map[string]string{
			"vendor":  info["bios_vendor"],
			"version": info["bios_version"],
			"date":    info["bios_date"],
		},
		"temperatures": map[string]any{
			"socket": socketTemps,
		},
	})
}

type NetDelta struct {
	TxBps uint64
	RxBps uint64
}

func getGPUInfo(c *gin.Context) {
	gpuInfo, err := gpu.New()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to retrieve GPU information",
			"details": err.Error(),
		})
		return
	}

	var gpus []gin.H
	for _, card := range gpuInfo.GraphicsCards {
		gpus = append(gpus, gin.H{
			"address":      card.Address,
			"vendor":       card.DeviceInfo.Vendor.Name,
			"model":        card.DeviceInfo.Product.Name,
			"device_id":    card.DeviceInfo.Product.ID,
			"vendor_id":    card.DeviceInfo.Vendor.ID,
			"subsystem":    card.DeviceInfo.Subsystem.Name,
			"subsystem_id": card.DeviceInfo.Subsystem.ID,
			"revision":     card.DeviceInfo.Revision,
			"driver":       card.DeviceInfo.Driver,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"gpus": gpus,
	})
}
