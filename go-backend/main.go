package main

import (
	"fmt"
	"go-backend/auth"
	"go-backend/config"
	"go-backend/docker"
	"go-backend/services"
	"go-backend/session"
	"go-backend/update"
	"go-backend/utils"
	"go-backend/websocket"

	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var (
	env       = "development" // default
	version   = "dev"
	buildTime = "unknown"
)

func main() {

	// Load configuration
	err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Ensure docker apps folder exists
	err = config.EnsureDockerAppsDirExists()
	if err != nil {
		log.Fatalf("Failed to create docker apps directory: %v", err)
	}

	// Load .env variables into os.Environ()
	godotenv.Load("../.env")

	// Override env from GO_ENV if set
	if goEnv := os.Getenv("GO_ENV"); goEnv != "" {
		env = goEnv
	}

	log.Printf("🌱 Starting server in %s mode...\n", env)

	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	if env == "development" {
		router.SetTrustedProxies(nil)
		router.Use(auth.CorsMiddleware())
		router.Use(gin.Logger())
	}

	router.Use(gin.Recovery())
	auth.RegisterAuthRoutes(router)
	registerSystemRoutes(router)
	websocket.RegisterWebSocketRoutes(router)
	update.RegisterUpdateRoutes(router)
	services.RegisterServiceRoutes(router)
	docker.RegisterDockerRoutes(router)
	docker.RegisterDockerComposeRoutes(router)
	config.RegisterThemeRoutes(router)

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
					output = append(output, gin.H{
						"endpoint": r.Endpoint,
						"error":    r.Error.Error(),
					})
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

	// Serve frontend in production
	if env == "production" {
		router.Static("/assets", "./frontend/assets")
		router.NoRoute(func(c *gin.Context) {
			c.File("./frontend/index.html")
		})
	}

	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"env":       env,
			"version":   version,
			"buildTime": buildTime,
		})
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
		log.Println("⚠️  SERVER_PORT not set, defaulting to 8080")
	}

	log.Printf("🚀 Server running in %s mode on http://localhost:%s", env, port)
	log.Fatal(router.Run(":" + port))
}
