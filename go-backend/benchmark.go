// benchmark.go
package main

import (
	"io"
	"net/http"
	"sync"
	"time"
)

type BenchmarkResult struct {
	Endpoint string
	Status   int
	Latency  time.Duration
	Error    error
}

var benchmarkEndpoints = []string{
	"/system/info",
	"/system/cpu",
	"/system/mem",
	"/system/disk",
	"/system/load",
	"/system/uptime",
	"/system/network",
	"/system/processes",
}

// RunBenchmark performs parallel benchmarking of all endpoints
func RunBenchmark(baseURL string, sessionCookie string, concurrency int) []BenchmarkResult {
	client := &http.Client{Timeout: 5 * time.Second}
	var wg sync.WaitGroup
	results := make([]BenchmarkResult, len(benchmarkEndpoints))
	resultChan := make(chan BenchmarkResult, len(benchmarkEndpoints))

	for _, endpoint := range benchmarkEndpoints {
		wg.Add(1)
		go func(endpoint string) {
			defer wg.Done()

			req, err := http.NewRequest("GET", baseURL+endpoint, nil)
			if err != nil {
				resultChan <- BenchmarkResult{Endpoint: endpoint, Error: err}
				return
			}
			req.Header.Set("Cookie", sessionCookie)

			start := time.Now()
			resp, err := client.Do(req)
			latency := time.Since(start)

			if err != nil {
				resultChan <- BenchmarkResult{Endpoint: endpoint, Latency: latency, Error: err}
				return
			}
			defer resp.Body.Close()
			io.Copy(io.Discard, resp.Body)

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

	return results
}
