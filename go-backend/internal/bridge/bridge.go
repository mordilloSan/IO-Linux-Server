package bridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-backend/internal/logger"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var bridgeBinary = os.ExpandEnv("/usr/lib/linuxio/linuxio-bridge")

type BridgeProcess struct {
	Cmd       *exec.Cmd
	SessionID string
	StartedAt time.Time
}

var (
	processes   = make(map[string]*BridgeProcess)
	processesMu sync.Mutex
)

// StartBridge starts the bridge process for a given session ID and username.
func StartBridge(sessionID, username string, privileged bool, sudoPassword string) error {
	processesMu.Lock()
	defer processesMu.Unlock()

	if _, exists := processes[sessionID]; exists {
		return errors.New("bridge already running for this session")
	}

	var cmd *exec.Cmd
	if privileged {
		cmd = exec.Command("sudo", "-S", "env",
			"LINUXIO_SESSION_ID="+sessionID,
			"LINUXIO_SESSION_USER="+username,
			"LINUXIO_BACKEND_URL="+os.Getenv("LINUXIO_BACKEND_URL"),
			"GO_ENV="+os.Getenv("GO_ENV"),
			"VERBOSE="+os.Getenv("VERBOSE"),
			bridgeBinary,
		)
	} else {
		cmd = exec.Command(bridgeBinary)
		cmd.Env = append(os.Environ(),
			"LINUXIO_SESSION_ID="+sessionID,
			"LINUXIO_SESSION_USER="+username,
			"LINUXIO_BACKEND_URL="+os.Getenv("LINUXIO_BACKEND_URL"),
			"GO_ENV="+os.Getenv("GO_ENV"),
			"VERBOSE="+os.Getenv("VERBOSE"),
		)
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Only send the password for privileged launches
	if privileged && sudoPassword != "" {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			logger.Error.Printf("[bridge] Failed to get stdin pipe: %v", err)
			return err
		}
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, sudoPassword+"\n")
		}()
	}

	if err := cmd.Start(); err != nil {
		logger.Error.Printf("[bridge] Failed to start bridge for session %s: %v", sessionID, err)
		return err
	}

	logger.Info.Printf("[bridge] Started %sbridge for session %s (pid=%d)",
		func() string {
			if privileged {
				return "privileged "
			}
			return ""
		}(), sessionID, cmd.Process.Pid)

	processes[sessionID] = &BridgeProcess{
		Cmd:       cmd,
		SessionID: sessionID,
		StartedAt: time.Now(),
	}

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

	if !exists || proc.Cmd.Process == nil {
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

// Finds all running bridge processes and their PIDs
func FindAllBridgeProcesses() ([]int, error) {
	cmd := exec.Command("pgrep", "-f", "linuxio-bridge")
	output, err := cmd.Output()
	if err != nil {
		return nil, nil // If no matches, return empty (not an error)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var pids []int
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		pid, err := strconv.Atoi(l)
		if err == nil {
			pids = append(pids, pid)
		}
	}
	return pids, nil
}

// Kills any real bridge process not tracked by our backend (session).
func CleanupOrphanBridges() {
	pids, err := FindAllBridgeProcesses()
	if err != nil {
		logger.Warning.Printf("[bridge] CleanupOrphanBridges: failed to find bridge processes: %v", err)
		return
	}

	processesMu.Lock()
	validSessionIDs := make(map[string]bool)
	for sid := range processes {
		validSessionIDs[sid] = true
	}
	processesMu.Unlock()

	for _, pid := range pids {
		// Read env to get session id
		envPath := fmt.Sprintf("/proc/%d/environ", pid)
		data, err := os.ReadFile(envPath)
		if err != nil {
			logger.Warning.Printf("[bridge] Could not read env for PID %d: %v", pid, err)
			continue
		}
		envVars := strings.Split(string(data), "\x00")
		var sessionID string
		for _, v := range envVars {
			if strings.HasPrefix(v, "LINUXIO_SESSION_ID=") {
				sessionID = strings.TrimPrefix(v, "LINUXIO_SESSION_ID=")
				break
			}
		}
		if sessionID == "" {
			continue // Not our process or legacy, skip
		}
		if validSessionIDs[sessionID] {
			continue // Active in backend
		}
		// Not valid, kill
		logger.Info.Printf("[bridge] Killing orphan bridge process (PID %d, session '%s')", pid, sessionID)
		_ = syscall.Kill(pid, syscall.SIGTERM)
	}
}

// Generic bridge call
func Call(reqType, command string, args []string) (string, error) {
	req := map[string]interface{}{
		"type":    reqType,
		"command": command,
	}
	if args != nil {
		req["args"] = args
	}
	resp, err := sendBridgeRequest(req)
	if err != nil {
		return "", err
	}
	if resp["status"] != "ok" {
		// output may be present in error cases too
		output := ""
		if o, ok := resp["output"].(string); ok {
			output = o
		}
		return output, fmt.Errorf("bridge error: %v", resp["error"])
	}
	output, _ := resp["output"].(string)
	return output, nil
}

// sendBridgeRequest sends a request to the bridge and returns the response
func sendBridgeRequest(req map[string]interface{}) (map[string]interface{}, error) {
	conn, err := net.DialTimeout("unix", "/run/linuxio-bridge.sock", 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to bridge: %w", err)
	}
	defer conn.Close()
	logger.Debug.Printf("[bridge] Sending request: %+v", req)
	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)

	if err := enc.Encode(req); err != nil {
		return nil, fmt.Errorf("failed to send request to bridge: %w", err)
	}

	var resp map[string]interface{}

	if err := dec.Decode(&resp); err != nil {
		logger.Error.Printf("[bridge] Failed to decode response: %v", err)
		return nil, fmt.Errorf("failed to decode response from bridge: %w", err)
	}
	logger.Debug.Printf("[bridge] Got response: %+v", resp)
	return resp, nil
}
