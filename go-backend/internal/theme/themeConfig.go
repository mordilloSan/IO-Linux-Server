package theme

import (
	"encoding/json"
	"errors"
	embed "go-backend"
	"go-backend/internal/auth"
	"go-backend/internal/logger"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type ThemeSettings struct {
	Theme           string `json:"theme"`
	PrimaryColor    string `json:"primaryColor"`
	SidebarColapsed bool   `json:"sidebarColapsed"`
}

const themeFilePath = "/etc/linuxio/themeConfig.json"

func InitTheme() error {
	if _, err := os.Stat(themeFilePath); os.IsNotExist(err) {
		logger.Infof("No theme file found, creating from embedded default...")
		if err := os.WriteFile(themeFilePath, embed.DefaultThemeConfig, 0660); err != nil {
			logger.Errorf("Failed to write embedded theme config: %v", err)
			return err
		}
		return nil
	}
	return nil
}

func LoadTheme() (ThemeSettings, error) {
	var settings ThemeSettings

	data, err := os.ReadFile(themeFilePath)
	if err != nil {
		logger.Errorf("Failed to read theme file: %v", err)
		return settings, err
	}

	if err := json.Unmarshal(data, &settings); err != nil {
		logger.Errorf("Failed to parse theme file: %v", err)
		return settings, err
	}

	return settings, nil
}

func SaveThemeToFile(settings ThemeSettings) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		logger.Errorf("Failed to encode theme settings: %v", err)
		return err
	}

	// Open file with group and user write/read permissions
	file, err := os.OpenFile(themeFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0660)
	if err != nil {
		logger.Errorf("Failed to open theme file: %v", err)
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		logger.Errorf("Failed to write theme file: %v", err)
		return err
	}

	logger.Infof("Theme settings saved to %s", themeFilePath)
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
