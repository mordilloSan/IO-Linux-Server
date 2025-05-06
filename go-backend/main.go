package main

import (
	"fmt"
	"go-backend/auth"
	"go-backend/config"
	"go-backend/docker"
	"go-backend/logger"
	"go-backend/services"
	"go-backend/session"
	"go-backend/update"
	"go-backend/utils"
	"go-backend/websocket"
	"go-backend/wireguard"
	"html/template"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var env = "development"
var indexTemplate = template.Must(template.ParseFiles("templates/index.tmpl"))

func main() {
	_ = godotenv.Load("../.env")

	if goEnv := os.Getenv("GO_ENV"); goEnv != "" {
		env = goEnv
	}

	verbose := os.Getenv("VERBOSE") == "true"
	logger.Init(env, verbose)

	logger.Info.Println("📦 Loading configuration...")
	if err := config.LoadConfig(); err != nil {
		logger.Error.Fatalf("❌ Failed to load config: %v", err)
	}

	if err := config.EnsureDockerAppsDirExists(); err != nil {
		logger.Error.Fatalf("❌ Failed to create docker apps directory: %v", err)
	}

	if err := config.InitTheme(); err != nil {
		logger.Error.Fatalf("❌ Failed to initialize theme file: %v", err)
	}

	logger.Info.Printf("🌱 Starting server in %s mode...", env)

	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	if env == "development" {
		router.SetTrustedProxies(nil)
		router.Use(auth.CorsMiddleware())
		router.Use(gin.Logger())
	}

	auth.RegisterAuthRoutes(router)
	registerSystemRoutes(router)
	websocket.RegisterWebSocketRoutes(router)
	update.RegisterUpdateRoutes(router)
	services.RegisterServiceRoutes(router)
	docker.RegisterDockerRoutes(router)
	docker.RegisterDockerComposeRoutes(router)
	config.RegisterThemeRoutes(router)
	wireguard.RegisterWireguardRoutes(router)

	session.StartSessionGC()

	if env != "production" {
		router.GET("/debug/benchmark", func(c *gin.Context) {
			cookie, err := c.Cookie("session_id")
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}
			results := utils.RunBenchmark("http://localhost:8080", "session_id="+cookie, router, 8)
			var output []gin.H
			for _, r := range results {
				if r.Error != nil {
					output = append(output, gin.H{"endpoint": r.Endpoint, "error": r.Error.Error()})
				} else {
					output = append(output, gin.H{
						"endpoint": r.Endpoint,
						"status":   r.Status,
						"latency":  fmt.Sprintf("%.2fms", float64(r.Latency.Microseconds())/1000),
					})
				}
			}
			c.JSON(http.StatusOK, output)
		})
	}

	if env == "production" {
		router.Static("/assets", "./frontend/assets")
		router.StaticFile("/manifest.json", "./frontend/manifest.json")
		router.StaticFile("/favicon.ico", "./frontend/favicon-6.png")
		for i := 1; i <= 6; i++ {
			router.StaticFile(fmt.Sprintf("/favicon-%d.png", i), fmt.Sprintf("./frontend/favicon-%d.png", i))
		}
		router.NoRoute(ServeIndex)
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
		logger.Warning.Println("⚠️  SERVER_PORT not set, defaulting to 8080")
	}

	logger.Info.Printf("🚀 Server running at http://localhost:%s", port)
	logger.Error.Fatal(router.Run(":" + port))
}

func ServeIndex(c *gin.Context) {
	js, css, err := utils.ParseViteManifest("./frontend/.vite/manifest.json")
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load bundle info")
		return
	}

	theme, err := config.LoadTheme()
	if err != nil {
		logger.Warning.Println("⚠️ Failed to load theme, using defaults:", err)
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
	if err := indexTemplate.Execute(c.Writer, data); err != nil {
		logger.Error.Printf("❌ Failed to execute index template: %v", err)
	}
}
