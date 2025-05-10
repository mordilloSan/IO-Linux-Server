package system

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v4/mem"
)

func getMemInfo(c *gin.Context) {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get memory info", "details": err.Error()})
		return
	}

	// ZFS ARC Cache (if available)
	var arc uint64
	if data, err := os.ReadFile("/proc/spl/kstat/zfs/arcstats"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "size") {
				fields := strings.Fields(line)
				if len(fields) >= 3 {
					if val, err := strconv.ParseUint(fields[2], 10, 64); err == nil {
						arc = val
					}
				}
				break
			}
		}
	}

	dockerUsed, _ := getDockerMemoryUsage()

	c.JSON(http.StatusOK, gin.H{
		"system": memInfo,
		"docker": gin.H{"used": dockerUsed},
		"zfs":    gin.H{"arc": arc},
	})
}

func getDockerMemoryUsage() (uint64, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return 0, err
	}
	defer cli.Close()

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return 0, err
	}

	var total uint64
	for _, container := range containers {
		statsResp, err := cli.ContainerStatsOneShot(context.Background(), container.ID)
		if err != nil {
			continue
		}
		defer statsResp.Body.Close()

		var stats struct {
			MemoryStats struct {
				Usage uint64 `json:"usage"`
			} `json:"memory_stats"`
		}
		if err := json.NewDecoder(statsResp.Body).Decode(&stats); err != nil {
			continue
		}

		total += stats.MemoryStats.Usage
	}

	return total, nil
}
