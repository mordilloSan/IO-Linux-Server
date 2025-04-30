package config

import (
	"go-backend/auth"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func RegisterThemeRoutes(router *gin.Engine) {
	// Route for fetching the theme (does not require authentication)
	router.GET("/theme/get", func(c *gin.Context) {
		currentTheme := LoadTheme()
		c.JSON(http.StatusOK, gin.H{"theme": currentTheme})
	})

	// Authenticated route for setting the theme (requires authentication)
	theme := router.Group("/theme", auth.AuthMiddleware())
	theme.POST("/set", SaveTheme)
}

func SaveTheme(c *gin.Context) {
	var body struct {
		Theme string `json:"theme"`
	}

	if err := c.BindJSON(&body); err != nil || (body.Theme != "LIGHT" && body.Theme != "DARK") {
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
		return "DARK" // fallback default
	}
	theme := string(data)
	if theme != "LIGHT" && theme != "DARK" {
		return "DARK"
	}
	return theme
}
