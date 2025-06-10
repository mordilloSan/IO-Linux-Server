package docker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type Metrics struct {
	CPUPercent float64 `json:"cpu_percent"`
	MemUsage   uint64  `json:"mem_usage"`
	NetInput   uint64  `json:"net_input"`
	NetOutput  uint64  `json:"net_output"`
	BlockRead  uint64  `json:"block_read"`
	BlockWrite uint64  `json:"block_write"`
}

type ContainerWithMetrics struct {
	types.Container
	Metrics *Metrics `json:"metrics,omitempty"`
}

// Helper to get a docker client
func getClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv)
}

// List all containers with metrics
func ListContainers() (any, error) {
	cli, err := getClient()
	if err != nil {
		return nil, fmt.Errorf("docker client error: %w", err)
	}
	defer cli.Close()

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
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
				cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage)
				systemDelta := float64(stats.CPUStats.SystemCPUUsage)
				if systemDelta > 0 && len(stats.CPUStats.CPUUsage.PercpuUsage) > 0 {
					metrics.CPUPercent = (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
				}

				metrics.MemUsage = stats.MemoryStats.Usage

				for _, net := range stats.Networks {
					metrics.NetInput += net.RxBytes
					metrics.NetOutput += net.TxBytes
				}

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

	return enriched, nil
}

// Start a container by ID
func StartContainer(id string) (any, error) {
	cli, err := getClient()
	if err != nil {
		return nil, fmt.Errorf("docker client error: %w", err)
	}
	defer cli.Close()

	if err := cli.ContainerStart(context.Background(), id, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	return "started", nil
}

// Stop a container by ID
func StopContainer(id string) (any, error) {
	cli, err := getClient()
	if err != nil {
		return nil, fmt.Errorf("docker client error: %w", err)
	}
	defer cli.Close()

	if err := cli.ContainerStop(context.Background(), id, container.StopOptions{}); err != nil {
		return nil, fmt.Errorf("failed to stop container: %w", err)
	}

	return "stopped", nil
}

// Remove a container by ID
func RemoveContainer(id string) (any, error) {
	cli, err := getClient()
	if err != nil {
		return nil, fmt.Errorf("docker client error: %w", err)
	}
	defer cli.Close()

	if err := cli.ContainerRemove(context.Background(), id, container.RemoveOptions{Force: true}); err != nil {
		return nil, fmt.Errorf("failed to remove container: %w", err)
	}

	return "removed", nil
}

// Restart a container by ID
func RestartContainer(id string) (any, error) {
	cli, err := getClient()
	if err != nil {
		return nil, fmt.Errorf("docker client error: %w", err)
	}
	defer cli.Close()

	if err := cli.ContainerRestart(context.Background(), id, container.StopOptions{}); err != nil {
		return nil, fmt.Errorf("failed to restart container: %w", err)
	}

	return "restarted", nil
}

// List all images
func ListImages() (any, error) {
	cli, err := getClient()
	if err != nil {
		return nil, fmt.Errorf("docker client error: %w", err)
	}
	defer cli.Close()

	images, err := cli.ImageList(context.Background(), image.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	return images, nil
}
