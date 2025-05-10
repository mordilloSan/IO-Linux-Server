package system

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
)

func getCurrentFrequencies() ([]float64, error) {
	var freqs []float64
	basePath := "/sys/devices/system/cpu"

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "cpu") {
			continue
		}

		cpuPath := filepath.Join(basePath, entry.Name(), "cpufreq", "scaling_cur_freq")
		data, err := os.ReadFile(cpuPath)
		if err != nil {
			continue // skip offline or inaccessible cores
		}

		kHzStr := strings.TrimSpace(string(data))
		kHz, err := strconv.ParseFloat(kHzStr, 64)
		if err != nil {
			continue
		}

		freqs = append(freqs, kHz/1000.0) // convert to MHz
	}

	return freqs, nil
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
	allTemps := getTemperatureMap()
	currentFreqs, _ := getCurrentFrequencies()

	// Filter only CPU-related temps: coreX and package
	cpuTemps := make(map[string]float64)
	for k, v := range allTemps {
		if strings.HasPrefix(k, "core") || k == "package" {
			cpuTemps[k] = v
		}
	}

	cpuData := info[0]
	c.JSON(http.StatusOK, gin.H{
		"vendorId":           cpuData.VendorID,
		"modelName":          cpuData.ModelName,
		"family":             cpuData.Family,
		"model":              cpuData.Model,
		"mhz":                cpuData.Mhz,
		"currentFrequencies": currentFreqs,
		"cores":              counts,
		"loadAverage":        loadAvg,
		"perCoreUsage":       percent,
		"temperature":        cpuTemps,
	})
}

func getLoadInfo(c *gin.Context) {
	loadAvg, err := load.Avg()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get load average", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, loadAvg)
}
