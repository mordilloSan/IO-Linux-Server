package system

import (
	"go-backend/internal/auth"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/process"
)

func RegisterSystemRoutes(router *gin.Engine) {
	system := router.Group("/system", auth.AuthMiddleware())
	{
		system.GET("/info", getHostInfo)
		system.GET("/cpu", getCPUInfo)
		system.GET("/mem", getMemInfo)
		system.GET("/fs", getFsInfo)
		system.GET("/disk", getDriveInfo)
		system.GET("/network", getNetworkInfo)
		system.GET("/load", getLoadInfo)
		system.GET("/uptime", getUptime)
		system.GET("/processes", getProcesses)
		system.GET("/baseboard", getBaseboardInfo)
		system.GET("/gpu", getGPUInfo)
		system.GET("/sensors", getSensorData)
		system.GET("/smart/:device", getSmartInfo)
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

func getFsInfo(c *gin.Context) {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list processes", "details": err.Error()})
		return
	}

	// Preload info concurrently
	type ProcInfo struct {
		Pid    int32   `json:"pid"`
		Name   string  `json:"name"`
		CPU    float64 `json:"cpu_percent"`
		Memory float32 `json:"mem_percent"`
	}

	var wg sync.WaitGroup
	ch := make(chan ProcInfo, len(procs))

	for _, p := range procs {
		wg.Add(1)
		go func(p *process.Process) {
			defer wg.Done()
			name, _ := p.Name()
			cpu, _ := p.CPUPercent()
			mem, _ := p.MemoryPercent()
			ch <- ProcInfo{Pid: p.Pid, Name: name, CPU: cpu, Memory: mem}
		}(p)
	}

	wg.Wait()
	close(ch)

	var result []any
	for info := range ch {
		result = append(result, info)
	}

	c.JSON(http.StatusOK, result)
}
