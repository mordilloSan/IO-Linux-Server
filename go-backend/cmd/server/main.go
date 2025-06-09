package main

import (
	embed "github.com/mordilloSan/LinuxIO/go-backend"

	"github.com/mordilloSan/LinuxIO/go-backend/internal/auth"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/benchmark"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/config"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/docker"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/logger"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/network"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/power"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/services"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/session"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/system"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/templates"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/updates"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/utils"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/websocket"

	"crypto/tls"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var env = "production"

func main() {
	_ = godotenv.Load("../.env")

	if goEnv := os.Getenv("GO_ENV"); goEnv != "" {
		env = goEnv
	}

	verbose := os.Getenv("VERBOSE") == "true"
	logger.Init("env", verbose)

	logger.Infof("📦 Loading configuration...")
	if err := config.LoadConfig(); err != nil {
		logger.Error.Fatalf("❌ Failed to load config: %v", err)
	}
	if err := config.EnsureDockerAppsDirExists(); err != nil {
		logger.Error.Fatalf("❌ Failed to create docker apps directory: %v", err)
	}
	if err := config.InitTheme(); err != nil {
		logger.Error.Fatalf("❌ Failed to initialize theme file: %v", err)
	}

	logger.Infof("🌱 Starting server in %s mode...", env)

	if !verbose {
		gin.SetMode(gin.ReleaseMode)
	}

	// Start the session garbage collector
	session.StartSessionGC()
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
	updates.RegisterUpdateRoutes(router)
	services.RegisterServiceRoutes(router)
	network.RegisterNetworkRoutes(router)
	docker.RegisterDockerRoutes(router)
	docker.RegisterDockerComposeRoutes(router)
	config.RegisterThemeRoutes(router)
	power.RegisterPowerRoutes(router)
	// API Benchmark route
	if env != "production" {
		benchmark.RegisterDebugRoutes(router, env)
	}

	// Static files (only needed in production if files exist on disk)
	if env == "production" {
		templates.RegisterStaticRoutes(router, embed.StaticFS, embed.PWAManifest)
	}

	// WebSocket route
	router.GET("/ws", websocket.WebSocketHandler)

	// ✅ Serve frontend on "/" and fallback routes
	router.GET("/", func(c *gin.Context) {
		templates.ServeIndex(c, env, embed.ViteManifest)
	})
	router.NoRoute(func(c *gin.Context) {
		templates.ServeIndex(c, env, embed.ViteManifest)
	})

	// Port config
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
		logger.Warnf("⚠️  SERVER_PORT not set, defaulting to 8080")
	}
	os.Setenv("SERVER_PORT", port)

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
		logger.Infof("🚀 Server running at https://localhost:%s", port)
		logger.Error.Fatal(srv.ListenAndServeTLS("", "")) // Empty filenames = use TLSConfig.Certificates
	} else {
		logger.Infof("🚀 Server running at http://localhost:%s", port)
		logger.Error.Fatal(router.Run(addr))
	}

}
