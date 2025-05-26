package bridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-backend/internal/logger"
	"go-backend/internal/session"
	"io"
	"net"
	"os"
	"os/exec"
	"os/user"
	"strconv"
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

type BridgeHealthRequest struct {
	Type    string `json:"type"`    // e.g., "healthcheck" or "validate"
	Session string `json:"session"` // sessionID
}
type BridgeHealthResponse struct {
	Status  string `json:"status"` // "ok" or "invalid"
	Message string `json:"message,omitempty"`
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

var (
	mainSocketListeners   = make(map[string]net.Listener) // sessionID → Listener
	mainSocketListenersMu sync.Mutex
)

// StartBridgeSocket starts a Unix socket server for the main process.
// StartBridgeSocket starts a Unix socket server for the main process.
func StartBridgeSocket(sessionID string, username string) error {
	u, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("failed to lookup user %s: %w", username, err)
	}
	uid, _ := strconv.Atoi(u.Uid)
	socketPath := fmt.Sprintf("/run/user/%d/linuxio-main-%s.sock", uid, sessionID)
	_ = os.Remove(socketPath)
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		logger.Error.Printf("[bridge] Failed to listen on main socket for session %s: %v", sessionID, err)
		return err
	}

	// Set permissions strictly to 0600 (owner read/write only)
	if err := os.Chmod(socketPath, 0600); err != nil {
		_ = ln.Close()
		_ = os.Remove(socketPath)
		logger.Error.Printf("[bridge] Failed to chmod main socket %s: %v", socketPath, err)
		return fmt.Errorf("failed to chmod socket: %w", err)
	}

	// Store the listener so we can close and remove it on logout
	mainSocketListenersMu.Lock()
	mainSocketListeners[sessionID] = ln
	mainSocketListenersMu.Unlock()

	logger.Info.Printf("[bridge] Main socket for session %s is now listening on %s", sessionID, socketPath)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				logger.Warning.Printf("[bridge] Accept failed on main socket for session %s: %v", sessionID, err)
				return // Accept fails after Close()
			}
			logger.Info.Printf("[bridge] Main socket for session %s accepted a connection", sessionID)
			go handleBridgeRequest(conn)
		}
	}()
	return nil
}

func handleBridgeRequest(conn net.Conn) {
	defer conn.Close()
	logger.Info.Printf("[bridge] Main socket accepted a connection")
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	var req BridgeHealthRequest
	if err := decoder.Decode(&req); err != nil {
		logger.Warning.Printf("[bridge] Invalid JSON on main socket: %v", err)
		_ = encoder.Encode(BridgeHealthResponse{Status: "error", Message: "invalid json"})
		return
	}

	if req.Type == "validate" {
		logger.Info.Printf("[bridge] Healthcheck received for session %s", req.Session)
		if session.IsValid(req.Session) {
			_ = encoder.Encode(BridgeHealthResponse{Status: "ok"})
		} else {
			_ = encoder.Encode(BridgeHealthResponse{Status: "invalid", Message: "session expired"})
		}
		return
	}
	_ = encoder.Encode(BridgeHealthResponse{Status: "error", Message: "unknown request type"})
}

func CleanupBridgeSocket(sessionID string, username string) {
	mainSocketListenersMu.Lock()
	ln, ok := mainSocketListeners[sessionID]
	if ok {
		err := ln.Close()
		if err != nil {
			logger.Warning.Printf("[bridge] Error closing main socket listener for session %s: %v", sessionID, err)
		} else {
			logger.Info.Printf("[bridge] Closed main socket listener for session %s", sessionID)
		}
		delete(mainSocketListeners, sessionID)
	}
	mainSocketListenersMu.Unlock()

	// Remove the socket file (in case Close() didn't)
	u, err := user.Lookup(username)
	if err == nil {
		uid, _ := strconv.Atoi(u.Uid)
		socketPath := fmt.Sprintf("/run/user/%d/linuxio-main-%s.sock", uid, sessionID)
		if err := os.Remove(socketPath); err == nil {
			logger.Info.Printf("[bridge] Removed socket file %s for session %s", socketPath, sessionID)
		} else if !os.IsNotExist(err) {
			logger.Warning.Printf("[bridge] Failed to remove socket file %s: %v", socketPath, err)
		}
	} else {
		logger.Warning.Printf("[bridge] Could not lookup user %s when cleaning up socket for session %s: %v", username, sessionID, err)
	}
}
