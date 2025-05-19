package main

import (
	"crypto/tls"
	"fmt"
	"go-backend/internal/auth"
	"go-backend/internal/config"
	"go-backend/internal/docker"
	"go-backend/internal/logger"
	"go-backend/internal/power"
	"go-backend/internal/services"
	"go-backend/internal/session"
	"go-backend/internal/system"
	"go-backend/internal/templates"
	"go-backend/internal/update"
	"go-backend/internal/utils"
	"go-backend/internal/wireguard"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var env = "development"

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

	// Initialize cache functions

	system.InitGPUInfo()

	router := gin.New()
	router.Use(gin.Recovery())

	if env == "development" {
		router.SetTrustedProxies(nil)
		router.Use(auth.CorsMiddleware())
		router.Use(gin.Logger())
	}

	// Backend routes
	auth.RegisterAuthRoutes(router)
	system.RegisterSystemRoutes(router)
	update.RegisterUpdateRoutes(router)
	services.RegisterServiceRoutes(router)
	docker.RegisterDockerRoutes(router)
	docker.RegisterDockerComposeRoutes(router)
	config.RegisterThemeRoutes(router)
	wireguard.RegisterWireguardRoutes(router)
	power.RegisterPowerRoutes(router)

	session.StartSessionGC()

	// Debug route
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

	// Static files (only needed in production if files exist on disk)
	if env == "production" {
		router.Static("/assets", "../../frontend/assets")
		router.StaticFile("/manifest.json", "../../frontend/manifest.json")
		router.StaticFile("/favicon.ico", "../../frontend/favicon-6.png")
		for i := 1; i <= 6; i++ {
			router.StaticFile(fmt.Sprintf("/favicon-%d.png", i), fmt.Sprintf("../../frontend/favicon-%d.png", i))
		}
	}

	// ✅ Serve frontend on "/" and fallback routes
	router.GET("/", ServeIndex)
	router.NoRoute(ServeIndex)

	// Port config
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
		logger.Warning.Println("⚠️  SERVER_PORT not set, defaulting to 8080")
	}

	// Start the server
	addr := ":" + port

	if env == "production" {
		cert, err := utils.GenerateSelfSignedCert()
		if err != nil {
			logger.Error.Fatalf("❌ Failed to generate self-signed certificate: %v", err)
		}

		srv := &http.Server{
			Addr:      addr,
			Handler:   router,
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}},
		}

		logger.Info.Printf("🚀 Server running at https://localhost:%s", port)
		fmt.Printf("🚀 Server running at https://localhost:%s\n", port)
		logger.Error.Fatal(srv.ListenAndServeTLS("", "")) // Empty filenames = use TLSConfig.Certificates
	} else {
		logger.Info.Printf("🚀 Server running at http://localhost:%s", port)
		fmt.Printf("🚀 Server running at http://localhost:%s\n", port)
		logger.Error.Fatal(router.Run(addr))
	}

}

func ServeIndex(c *gin.Context) {
	logger.Info.Println("📄 ServeIndex called for:", c.Request.URL.Path)

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
		js, css, err = utils.ParseViteManifest("../../frontend/.vite/manifest.json")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load bundle info")
			return
		}
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
	if err := templates.IndexTemplate.Execute(c.Writer, data); err != nil {
		logger.Error.Printf("❌ Failed to execute index template: %v", err)
	}
}
