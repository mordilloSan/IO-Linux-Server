package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var (
	env       = "development" // default for go run
	version   = "dev"         // default version
	buildTime = "unknown"     // default time
)

func getDevPort() string {
	port := os.Getenv("VITE_DEV_PORT")
	if port == "" {
		port = "3000"
	}
	return port
}

func main() {

	// Override env only in development (e.g., when using `go run`)
	if env == "development" && os.Getenv("GO_ENV") != "" {
		env = os.Getenv("GO_ENV")
	}

	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Only in development: relaxed proxy & CORS
	if env == "development" {
		router.SetTrustedProxies(nil) // or a list of known proxy IPs
		router.Use(corsMiddleware())
		router.Use(gin.Logger())
	}

	router.Use(gin.Recovery())
	registerAuthRoutes(router)
	registerSystemRoutes(router)
	registerWebSocketRoutes(router)
	startSessionGC()

	// Dev-only debug route
	if env != "production" {
		router.GET("/debug/benchmark", func(c *gin.Context) {
			cookie, err := c.Cookie("session_id")
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}

			results := RunBenchmark("http://localhost:8080", fmt.Sprintf("session_id=%s", cookie), 8)
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

	// Production-only: serve built frontend
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
