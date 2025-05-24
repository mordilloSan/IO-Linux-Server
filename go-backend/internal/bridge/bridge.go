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
	"os/user"
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

// Utility: Per-session/per-user socket path
func BridgeSocketPath(sessionID, username string) string {
	u, err := user.Lookup(username)
	if err != nil {
		panic(fmt.Sprintf("could not find user %s: %v", username, err))
	}
	return fmt.Sprintf("/run/user/%s/linuxio-bridge-%s.sock", u.Uid, sessionID)
}

// Public: Use this everywhere from backend for bridge actions
func CallWithSession(sessionID, username, reqType, command string, args []string) (string, error) {
	socketPath := BridgeSocketPath(sessionID, username)
	return CallViaSocket(socketPath, reqType, command, args)
}

// Low-level: Direct call by socket path
func CallViaSocket(socketPath, reqType, command string, args []string) (string, error) {
	req := map[string]interface{}{
		"type":    reqType,
		"command": command,
	}
	if args != nil {
		req["args"] = args
	}
	resp, err := sendBridgeRequest(socketPath, req)
	if err != nil {
		return "", err
	}
	if resp["status"] != "ok" {
		output := ""
		if o, ok := resp["output"].(string); ok {
			output = o
		}
		return output, fmt.Errorf("bridge error: %v", resp["error"])
	}
	output, _ := resp["output"].(string)
	return output, nil
}

// sendBridgeRequest sends a request to the bridge at socketPath and returns the response.
func sendBridgeRequest(socketPath string, req map[string]interface{}) (map[string]interface{}, error) {
	conn, err := net.DialTimeout("unix", socketPath, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to bridge: %w", err)
	}
	defer conn.Close()
	logger.Debug.Printf("[bridge] Sending request to %s: %+v", socketPath, req)
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

// StopBridge stops the bridge for a session (SIGTERM, then SIGKILL if needed)
func StopBridge(sessionID string) {
	processesMu.Lock()
	proc, exists := processes[sessionID]
	processesMu.Unlock()

	if !exists || proc.Cmd.Process == nil {
		return
	}

	pgid, err := syscall.Getpgid(proc.Cmd.Process.Pid)
	if err == nil {
		// Kill the entire process group (including sudo, env, bridge)
		syscall.Kill(-pgid, syscall.SIGTERM)
	} else {
		_ = proc.Cmd.Process.Signal(syscall.SIGTERM)
	}

	done := make(chan error, 1)
	go func() {
		done <- proc.Cmd.Wait()
	}()

	select {
	case <-done:
		logger.Info.Printf("[bridge] Bridge for session %s stopped gracefully", sessionID)
	case <-time.After(5 * time.Second):
		logger.Warning.Printf("[bridge] Bridge for session %s did not stop, killing...", sessionID)
		if err == nil {
			syscall.Kill(-pgid, syscall.SIGKILL)
		} else {
			_ = proc.Cmd.Process.Kill()
		}
	}

	processesMu.Lock()
	delete(processes, sessionID)
	processesMu.Unlock()
}
