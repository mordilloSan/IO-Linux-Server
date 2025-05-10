package system

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jaypipes/ghw/pkg/gpu"
)

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
