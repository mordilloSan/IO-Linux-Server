package main

import (
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
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
	info, _ := cpu.Info()
	usage, _ := cpu.Percent(0, false)
	c.JSON(200, gin.H{"info": info, "usage": usage})
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
