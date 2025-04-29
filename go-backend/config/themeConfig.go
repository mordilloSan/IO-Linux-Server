package config

import (
	"go-backend/auth"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func SaveTheme(c *gin.Context) {
	var body struct {
		Theme string `json:"theme"` // expects "light" or "dark"
	}

	if err := c.BindJSON(&body); err != nil || (body.Theme != "light" && body.Theme != "dark") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid theme value"})
		return
	}

	err := os.WriteFile("./theme.txt", []byte(body.Theme), 0644)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save theme"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Theme saved"})
}

func LoadTheme() string {
	data, err := os.ReadFile("./theme.txt")
	if err != nil {
		return "dark" // fallback default
	}
	theme := string(data)
	if theme != "light" && theme != "dark" {
		return "dark"
	}
	return theme
}

func RegisterThemeRoutes(router *gin.Engine) {
	theme := router.Group("/theme", auth.AuthMiddleware())

	theme.GET("/", func(c *gin.Context) {
		currentTheme := LoadTheme()
		c.JSON(http.StatusOK, gin.H{"theme": currentTheme})
	})

	theme.POST("/", SaveTheme)
}
