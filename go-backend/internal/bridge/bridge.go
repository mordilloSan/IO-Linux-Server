package bridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-backend/internal/logger"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// Path to your bridge binary
var bridgeBinary = os.ExpandEnv(" /usr/local/lib/linuxio/linuxio-bridge")

type BridgeProcess struct {
	Cmd       *exec.Cmd
	SessionID string
	StartedAt time.Time
}

// Manages bridge processes per session (sessionID → *BridgeProcess)
var (
	processes   = make(map[string]*BridgeProcess)
	processesMu sync.Mutex
)

// Start a new privileged bridge for the session (returns error if already running)
func StartBridge(sessionID string) error {
	processesMu.Lock()
	defer processesMu.Unlock()

	if _, exists := processes[sessionID]; exists {
		return errors.New("bridge already running for this session")
	}

	cmd := exec.Command("pkexec", bridgeBinary)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Optional: direct stdout/stderr to files or /dev/null
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		logger.Error.Printf("[bridge] Failed to start bridge for session %s: %v", sessionID, err)
		return err
	}

	logger.Info.Printf("[bridge] Started bridge for session %s (pid=%d)", sessionID, cmd.Process.Pid)
	processes[sessionID] = &BridgeProcess{
		Cmd:       cmd,
		SessionID: sessionID,
		StartedAt: time.Now(),
	}

	// Start goroutine to wait and clean up on exit
	go func(sessionID string, cmd *exec.Cmd) {
		err := cmd.Wait()
		processesMu.Lock()
		defer processesMu.Unlock()
		delete(processes, sessionID)
		if err != nil {
			logger.Warning.Printf("[bridge] Bridge for session %s exited with error: %v", sessionID, err)
		} else {
			logger.Info.Printf("[bridge] Bridge for session %s exited", sessionID)
		}
	}(sessionID, cmd)

	return nil
}

// Stop the bridge for a session (SIGTERM, then SIGKILL if needed)
func StopBridge(sessionID string) {
	processesMu.Lock()
	proc, exists := processes[sessionID]
	processesMu.Unlock()

	if !exists {
		return
	}

	if proc.Cmd.Process == nil {
		return
	}

	logger.Info.Printf("[bridge] Stopping bridge for session %s (pid=%d)...", sessionID, proc.Cmd.Process.Pid)
	_ = proc.Cmd.Process.Signal(syscall.SIGTERM)

	done := make(chan error, 1)
	go func() {
		done <- proc.Cmd.Wait()
	}()

	select {
	case <-done:
		logger.Info.Printf("[bridge] Bridge for session %s stopped gracefully", sessionID)
	case <-time.After(5 * time.Second):
		logger.Warning.Printf("[bridge] Bridge for session %s did not stop, killing...", sessionID)
		_ = proc.Cmd.Process.Kill()
	}

	processesMu.Lock()
	delete(processes, sessionID)
	processesMu.Unlock()
}

// Optional: Clean up all orphans at startup
func CleanupOrphanBridges() {
	// Implement if needed: scan for old bridges and kill
}

// Optional: Stop all bridges (e.g., on backend shutdown)
func StopAllBridges() {
	processesMu.Lock()
	sessions := make([]string, 0, len(processes))
	for sid := range processes {
		sessions = append(sessions, sid)
	}
	processesMu.Unlock()

	for _, sid := range sessions {
		StopBridge(sid)
	}
}

const bridgeSocketPath = "/run/linuxio-bridge.sock"

// Sends a request to the bridge and returns the response struct.
func sendBridgeRequest(req map[string]any) (map[string]any, error) {
	conn, err := net.DialTimeout("unix", bridgeSocketPath, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to bridge: %w", err)
	}
	defer conn.Close()

	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)

	if err := enc.Encode(req); err != nil {
		return nil, fmt.Errorf("failed to send request to bridge: %w", err)
	}

	var resp map[string]any
	if err := dec.Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response from bridge: %w", err)
	}

	return resp, nil
}

func RebootSystem() error {
	resp, err := sendBridgeRequest(map[string]any{
		"type":    "dbus",
		"command": "reboot",
	})
	if err != nil {
		return err
	}
	if status, _ := resp["status"].(string); status != "ok" {
		return fmt.Errorf("bridge error: %v, detail: %v", resp["error"], resp["output"])
	}
	return nil
}

func PowerOffSystem() error {
	resp, err := sendBridgeRequest(map[string]any{
		"type":    "dbus",
		"command": "poweroff",
	})
	if err != nil {
		return err
	}
	if status, _ := resp["status"].(string); status != "ok" {
		return fmt.Errorf("bridge error: %v, detail: %v", resp["error"], resp["output"])
	}
	return nil
}
