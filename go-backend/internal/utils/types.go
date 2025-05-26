package utils

import "time"

type User struct {
	ID   string // Username (unique key)
	Name string // Display name (can be same as ID)
}

type BenchmarkResult struct {
	Endpoint string
	Status   int
	Latency  time.Duration
	Error    error
}

type ManifestEntry struct {
	File string   `json:"file"`
	CSS  []string `json:"css"`
}

type ViteManifest map[string]struct {
	File string   `json:"file"`
	Css  []string `json:"css"`
}

type BridgeProcInfo struct {
	PID       int
	SessionID string // optional (may be blank)
	Username  string // from owner
	Cmdline   string
}
