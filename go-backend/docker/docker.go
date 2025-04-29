package docker

import (
	"context"
	"encoding/json"
	"go-backend/auth"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
)

// Helper to get a docker client
func getClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv)
}

func ListContainers(c *gin.Context) {
	cli, err := getClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cli.Close()

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type Metrics struct {
		CPUPercent float64 `json:"cpu_percent"`
		MemUsage   uint64  `json:"mem_usage"`
		NetInput   uint64  `json:"net_input"`   // bytes in
		NetOutput  uint64  `json:"net_output"`  // bytes out
		BlockRead  uint64  `json:"block_read"`  // bytes read
		BlockWrite uint64  `json:"block_write"` // bytes written
	}

	type ContainerWithMetrics struct {
		types.Container
		Metrics *Metrics `json:"metrics,omitempty"`
	}

	var enriched []ContainerWithMetrics

	for _, ctr := range containers {
		metrics := &Metrics{}

		statsResp, err := cli.ContainerStatsOneShot(context.Background(), ctr.ID)
		if err == nil {
			var stats struct {
				CPUStats struct {
					CPUUsage struct {
						TotalUsage  uint64   `json:"total_usage"`
						PercpuUsage []uint64 `json:"percpu_usage"`
					} `json:"cpu_usage"`
					SystemCPUUsage uint64 `json:"system_cpu_usage"`
				} `json:"cpu_stats"`
				MemoryStats struct {
					Usage uint64 `json:"usage"`
				} `json:"memory_stats"`
				Networks map[string]struct {
					RxBytes uint64 `json:"rx_bytes"`
					TxBytes uint64 `json:"tx_bytes"`
				} `json:"networks"`
				BlkioStats struct {
					IoServiceBytesRecursive []struct {
						Op    string `json:"op"`
						Value uint64 `json:"value"`
					} `json:"io_service_bytes_recursive"`
				} `json:"blkio_stats"`
			}

			if err := json.NewDecoder(statsResp.Body).Decode(&stats); err == nil {
				// CPU %
				cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage)
				systemDelta := float64(stats.CPUStats.SystemCPUUsage)
				if systemDelta > 0 && len(stats.CPUStats.CPUUsage.PercpuUsage) > 0 {
					metrics.CPUPercent = (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
				}

				// Memory
				metrics.MemUsage = stats.MemoryStats.Usage

				// Network I/O (sum all interfaces)
				for _, net := range stats.Networks {
					metrics.NetInput += net.RxBytes
					metrics.NetOutput += net.TxBytes
				}

				// Block I/O
				for _, entry := range stats.BlkioStats.IoServiceBytesRecursive {
					switch entry.Op {
					case "Read":
						metrics.BlockRead += entry.Value
					case "Write":
						metrics.BlockWrite += entry.Value
					}
				}

			}
			statsResp.Body.Close()
		}

		enriched = append(enriched, ContainerWithMetrics{
			Container: ctr,
			Metrics:   metrics,
		})
	}

	c.JSON(http.StatusOK, enriched)
}

func StartContainer(c *gin.Context) {
	id := c.Param("id")
	cli, err := getClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cli.Close()

	if err := cli.ContainerStart(context.Background(), id, container.StartOptions{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func StopContainer(c *gin.Context) {
	id := c.Param("id")
	cli, err := getClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cli.Close()

	timeout := container.StopOptions{}
	if err := cli.ContainerStop(context.Background(), id, timeout); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func RemoveContainer(c *gin.Context) {
	id := c.Param("id")
	cli, err := getClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cli.Close()

	if err := cli.ContainerRemove(context.Background(), id, container.RemoveOptions{Force: true}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func RestartContainer(c *gin.Context) {
	id := c.Param("id")
	cli, err := getClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cli.Close()

	if err := cli.ContainerRestart(context.Background(), id, container.StopOptions{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func ListImages(c *gin.Context) {
	cli, err := getClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cli.Close()

	images, err := cli.ImageList(context.Background(), image.ListOptions{All: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, images)
}

func RegisterDockerRoutes(router *gin.Engine) {
	docker := router.Group("/docker", auth.AuthMiddleware())
	{
		docker.GET("/containers", ListContainers)
		docker.POST("/containers/:id/start", StartContainer)
		docker.POST("/containers/:id/stop", StopContainer)
		docker.POST("/containers/:id/restart", RestartContainer)
		docker.DELETE("/containers/:id", RemoveContainer)
		docker.GET("/images", ListImages)
	}
}
