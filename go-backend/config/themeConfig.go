package config

import (
	"encoding/json"
	"go-backend/auth"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type ThemeSettings struct {
	Theme        string `json:"theme"`
	PrimaryColor string `json:"primaryColor,omitempty"`
}

const themeFilePath = "./theme.txt"

func RegisterThemeRoutes(router *gin.Engine) {
	// Public: Get current theme + primary color
	router.GET("/theme/get", func(c *gin.Context) {
		settings := LoadTheme()
		c.JSON(http.StatusOK, gin.H{
			"theme":        settings.Theme,
			"primaryColor": settings.PrimaryColor,
		})
	})

	// Authenticated: Save both theme and primary color
	theme := router.Group("/theme", auth.AuthMiddleware())
	theme.POST("/set", SaveTheme)
}

func SaveTheme(c *gin.Context) {
	var body ThemeSettings

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if body.Theme != "LIGHT" && body.Theme != "DARK" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid theme value"})
		return
	}

	data, err := json.Marshal(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode theme settings"})
		return
	}

	err = os.WriteFile(themeFilePath, data, 0644)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save theme settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Theme settings saved"})
}

func LoadTheme() ThemeSettings {
	var settings ThemeSettings

	data, err := os.ReadFile(themeFilePath)
	if err != nil {
		// fallback default
		return ThemeSettings{Theme: "DARK"}
	}

	if err := json.Unmarshal(data, &settings); err != nil {
		return ThemeSettings{Theme: "DARK"}
	}

	// Validate theme value
	if settings.Theme != "LIGHT" && settings.Theme != "DARK" {
		settings.Theme = "DARK"
	}

	return settings
}
