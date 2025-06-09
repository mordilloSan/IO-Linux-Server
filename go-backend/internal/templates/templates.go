package templates

import (
	"embed"
	"fmt"
	"go-backend/internal/config"
	"go-backend/internal/logger"
	"go-backend/internal/utils"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

//go:embed index.tmpl
var tmplFS embed.FS

var IndexTemplate *template.Template

func init() {
	var err error
	IndexTemplate, err = template.ParseFS(tmplFS, "index.tmpl")
	if err != nil {
		log.Fatalf("❌ Failed to parse embedded template: %v", err)
	}
}

func ServeIndex(c *gin.Context, env string, viteManifest []byte) {
	logger.Infof("📄 ServeIndex called for: %s", c.Request.URL.Path)

	var js, css string

	if env == "development" {
		// Use Vite dev server directly
		vitePort := os.Getenv("VITE_DEV_PORT")
		if vitePort == "" {
			vitePort = "5173"
		}
		js = fmt.Sprintf("http://localhost:%s/src/main.tsx", vitePort)
		css = "" // Vite injects CSS in dev mode
	} else {
		// Load from manifest (production build)
		var err error
		js, css, err = utils.ParseViteManifestBytes(viteManifest)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load bundle info")
			return
		}
	}

	theme, err := config.LoadTheme()
	if err != nil {
		logger.Warnf("⚠️ Failed to load theme, using defaults: %v", err)
		theme = config.ThemeSettings{
			Theme:           "DARK",
			PrimaryColor:    "#1976d2",
			SidebarColapsed: false,
		}
	}

	background := "#ffffff"
	shimmer := "#eeeeee"
	if theme.Theme == "DARK" {
		background = "#1B2635"
		shimmer = "#233044"
	}

	data := map[string]string{
		"JSBundle":          js,
		"CSSBundle":         css,
		"PrimaryColor":      theme.PrimaryColor,
		"ThemeColor":        theme.PrimaryColor,
		"Background":        background,
		"ShimmerBackground": shimmer,
	}

	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := IndexTemplate.Execute(c.Writer, data); err != nil {
		logger.Errorf("❌ Failed to execute index template: %v", err)
	}
}
