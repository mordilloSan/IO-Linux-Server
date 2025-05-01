package config

import (
	"encoding/json"
	"go-backend/auth"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type ThemeSettings struct {
	Theme           string `json:"theme"`
	PrimaryColor    string `json:"primaryColor"`
	SidebarColapsed bool   `json:"sidebarColapsed"`
}

const themeFilePath = "./theme.txt"

func InitTheme() error {
	if _, err := os.Stat(themeFilePath); os.IsNotExist(err) {
		defaultSettings := ThemeSettings{
			Theme:           "DARK",
			PrimaryColor:    "#1976d2",
			SidebarColapsed: false,
		}
		return SaveThemeToFile(defaultSettings)
	}
	return nil
}

func LoadTheme() (ThemeSettings, error) {
	var settings ThemeSettings

	data, err := os.ReadFile(themeFilePath)
	if err != nil {
		return settings, err
	}

	if err := json.Unmarshal(data, &settings); err != nil {
		return settings, err
	}

	return settings, nil
}

func SaveThemeToFile(settings ThemeSettings) error {
	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	return os.WriteFile(themeFilePath, data, 0644)
}

// --- Gin routes ---

func RegisterThemeRoutes(router *gin.Engine) {
	router.GET("/theme/get", func(c *gin.Context) {
		// Attempt to load theme
		settings, err := LoadTheme()
		if err != nil {
			// If file is missing, auto-init
			if os.IsNotExist(err) {
				if initErr := InitTheme(); initErr != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize theme"})
					return
				}
				// Try again after init
				settings, err = LoadTheme()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load initialized theme"})
					return
				}
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load theme"})
				return
			}
		}
		c.JSON(http.StatusOK, settings)
	})

	theme := router.Group("/theme", auth.AuthMiddleware())
	theme.POST("/set", func(c *gin.Context) {
		var body ThemeSettings

		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if body.Theme != "LIGHT" && body.Theme != "DARK" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid theme value"})
			return
		}

		if err := SaveThemeToFile(body); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save theme settings"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Theme settings saved"})
	})
}
