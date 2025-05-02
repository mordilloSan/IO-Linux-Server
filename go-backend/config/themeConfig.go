package config

import (
	"encoding/json"
	"errors"
	"go-backend/auth"
	"go-backend/logger"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type ThemeSettings struct {
	Theme           string `json:"theme"`
	PrimaryColor    string `json:"primaryColor"`
	SidebarColapsed bool   `json:"sidebarColapsed"`
}

const themeFilePath = "./theme.json"

var defaultTheme = ThemeSettings{
	Theme:           "DARK",
	PrimaryColor:    "#1976d2",
	SidebarColapsed: false,
}

func InitTheme() error {
	if _, err := os.Stat(themeFilePath); os.IsNotExist(err) {
		logger.Info.Println("[theme] No theme file found, creating default...")
		return SaveThemeToFile(defaultTheme)
	}
	return nil
}

func LoadTheme() (ThemeSettings, error) {
	var settings ThemeSettings

	data, err := os.ReadFile(themeFilePath)
	if err != nil {
		logger.Error.Printf("[theme] Failed to read theme file: %v", err)
		return settings, err
	}

	if err := json.Unmarshal(data, &settings); err != nil {
		logger.Error.Printf("[theme] Failed to parse theme file: %v", err)
		return settings, err
	}

	return settings, nil
}

func SaveThemeToFile(settings ThemeSettings) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		logger.Error.Printf("[theme] Failed to encode theme settings: %v", err)
		return err
	}

	err = os.WriteFile(themeFilePath, data, 0644)
	if err != nil {
		logger.Error.Printf("[theme] Failed to write theme file: %v", err)
		return err
	}

	logger.Info.Printf("[theme] Theme settings saved to %s", themeFilePath)
	return nil
}

// --- Gin Routes ---

func RegisterThemeRoutes(router *gin.Engine) {
	router.GET("/theme/get", func(c *gin.Context) {
		settings, err := LoadTheme()
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				if initErr := InitTheme(); initErr != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize theme"})
					return
				}
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
