package main

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

func registerSystemRoutes(router *gin.Engine) {
	system := router.Group("/system", authMiddleware())
	{
		system.GET("/info", getHostInfo)
		system.GET("/cpu", getCPUInfo)
		system.GET("/mem", getMemInfo)
		system.GET("/disk", getDiskInfo)
		system.GET("/network", getNetworkInfo)
		system.GET("/load", getLoadInfo)
		system.GET("/uptime", getUptime)
		system.GET("/processes", getProcesses)
	}
}

func getHostInfo(c *gin.Context) {
	hostInfo, _ := host.Info()
	c.JSON(200, hostInfo)
}

func getCPUInfo(c *gin.Context) {
	// Basic CPU info
	info, _ := cpu.Info()
	percent, _ := cpu.Percent(0, true)
	counts, _ := cpu.Counts(true)
	loadAvg, _ := load.Avg()

	// Fallback check
	if len(info) == 0 {
		c.JSON(500, gin.H{"error": "no CPU info available"})
		return
	}

	// Get temperature via `sensors` command (optional)
	temp := getTemperatureMap()

	cpuData := info[0]

	c.JSON(200, gin.H{
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

	re := regexp.MustCompile(`\+([0-9]+(?:\.[0-9])?)°C`)
	cpuIndex := 1

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Core ") ||
			strings.HasPrefix(line, "Package id") ||
			strings.HasPrefix(line, "Tctl:") {

			match := re.FindStringSubmatch(line)
			if len(match) > 1 {
				temp, err := strconv.ParseFloat(match[1], 64)
				if err == nil {
					key := "cpu" + strconv.Itoa(cpuIndex)
					temps[key] = temp
					cpuIndex++
				}
			}
		}
	}

	return temps
}

func getMemInfo(c *gin.Context) {
	memInfo, _ := mem.VirtualMemory()
	c.JSON(200, memInfo)
}

func getDiskInfo(c *gin.Context) {
	parts, _ := disk.Partitions(true)
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
	c.JSON(200, results)
}

func getNetworkInfo(c *gin.Context) {
	stats, _ := net.IOCounters(true)
	ifaces, _ := net.Interfaces()
	var result []map[string]any
	for _, iface := range ifaces {
		ifaceStats := map[string]any{
			"name":         iface.Name,
			"mtu":          iface.MTU,
			"hardwareAddr": iface.HardwareAddr,
			"flags":        iface.Flags,
			"addresses":    []string{},
		}
		for _, addr := range iface.Addrs {
			ifaceStats["addresses"] = append(ifaceStats["addresses"].([]string), addr.Addr)
		}
		for _, stat := range stats {
			if stat.Name == iface.Name {
				ifaceStats["bytesSent"] = stat.BytesSent
				ifaceStats["bytesRecv"] = stat.BytesRecv
				break
			}
		}
		result = append(result, ifaceStats)
	}
	c.JSON(200, result)
}

func getLoadInfo(c *gin.Context) {
	loadAvg, _ := load.Avg()
	c.JSON(200, loadAvg)
}

func getUptime(c *gin.Context) {
	uptime, _ := host.Uptime()
	c.JSON(200, gin.H{"uptime_seconds": uptime})
}

func getProcesses(c *gin.Context) {
	procs, _ := process.Processes()
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
	c.JSON(200, result)
}
