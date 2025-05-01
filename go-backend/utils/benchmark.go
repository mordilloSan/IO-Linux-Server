package utils

import (
	"io"
	"net/http"
	"sync"
	"time"

	"go-backend/logger"

	"github.com/gin-gonic/gin"
)

// RunBenchmark performs parallel benchmarking of all GET /system/* endpoints
func RunBenchmark(baseURL string, sessionCookie string, router *gin.Engine, concurrency int) []BenchmarkResult {
	endpoints := getSystemEndpoints(router)
	logger.Info.Printf("📈 Running benchmark for %d /system/ endpoints...", len(endpoints))

	client := &http.Client{Timeout: 5 * time.Second}
	var wg sync.WaitGroup
	results := make([]BenchmarkResult, len(endpoints))
	resultChan := make(chan BenchmarkResult, len(endpoints))

	for _, endpoint := range endpoints {
		wg.Add(1)
		go func(endpoint string) {
			defer wg.Done()

			req, err := http.NewRequest("GET", baseURL+endpoint, nil)
			if err != nil {
				logger.Error.Printf("❌ Failed to create request for %s: %v", endpoint, err)
				resultChan <- BenchmarkResult{Endpoint: endpoint, Error: err}
				return
			}
			req.Header.Set("Cookie", sessionCookie)

			start := time.Now()
			resp, err := client.Do(req)
			latency := time.Since(start)

			if err != nil {
				logger.Warning.Printf("⚠️ Request to %s failed: %v", endpoint, err)
				resultChan <- BenchmarkResult{Endpoint: endpoint, Latency: latency, Error: err}
				return
			}
			defer resp.Body.Close()
			io.Copy(io.Discard, resp.Body)

			logger.Debug.Printf("✅ %s -> %d in %.2fms", endpoint, resp.StatusCode, float64(latency.Microseconds())/1000)

			resultChan <- BenchmarkResult{
				Endpoint: endpoint,
				Status:   resp.StatusCode,
				Latency:  latency,
			}
		}(endpoint)
	}

	wg.Wait()
	close(resultChan)

	i := 0
	for res := range resultChan {
		results[i] = res
		i++
	}

	logger.Info.Println("✅ Benchmark completed.")
	return results
}

// getSystemEndpoints returns all GET routes starting with /system/
func getSystemEndpoints(router *gin.Engine) []string {
	var endpoints []string
	for _, route := range router.Routes() {
		if route.Method == "GET" && len(route.Path) > 7 && route.Path[:8] == "/system/" {
			endpoints = append(endpoints, route.Path)
		}
	}
	logger.Debug.Printf("🔍 Found %d GET /system/* routes to benchmark", len(endpoints))
	return endpoints
}
