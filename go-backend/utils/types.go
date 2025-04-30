package utils

import "time"

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type BenchmarkResult struct {
	Endpoint string
	Status   int
	Latency  time.Duration
	Error    error
}
