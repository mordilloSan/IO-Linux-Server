package dockers

import (
	"go-backend/internal/auth"
	"go-backend/internal/bridge"
	"go-backend/internal/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ListContainers(c *gin.Context) {
	sess := auth.GetSessionOrAbort(c)
	if sess == nil {
		return
	}
	data, err := bridge.CallWithSession(sess, "docker", "list_containers", nil)
	if err != nil {
		logger.Errorf("Bridge ListContainers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}

func StartContainer(c *gin.Context) {
	sess := auth.GetSessionOrAbort(c)
	if sess == nil {
		return
	}
	id := c.Param("id")
	data, err := bridge.CallWithSession(sess, "docker", "start_container", []string{id})
	if err != nil {
		logger.Errorf("Bridge StartContainer: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}

func StopContainer(c *gin.Context) {
	sess := auth.GetSessionOrAbort(c)
	if sess == nil {
		return
	}
	id := c.Param("id")
	data, err := bridge.CallWithSession(sess, "docker", "stop_container", []string{id})
	if err != nil {
		logger.Errorf("Bridge StopContainer: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}

func RemoveContainer(c *gin.Context) {
	sess := auth.GetSessionOrAbort(c)
	if sess == nil {
		return
	}
	id := c.Param("id")
	data, err := bridge.CallWithSession(sess, "docker", "remove_container", []string{id})
	if err != nil {
		logger.Errorf("Bridge RemoveContainer: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}

func RestartContainer(c *gin.Context) {
	sess := auth.GetSessionOrAbort(c)
	if sess == nil {
		return
	}
	id := c.Param("id")
	data, err := bridge.CallWithSession(sess, "docker", "restart_container", []string{id})
	if err != nil {
		logger.Errorf("Bridge RestartContainer: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}

func ListImages(c *gin.Context) {
	sess := auth.GetSessionOrAbort(c)
	if sess == nil {
		return
	}
	data, err := bridge.CallWithSession(sess, "docker", "list_images", nil)
	if err != nil {
		logger.Errorf("Bridge ListImages: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/json", data)
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
