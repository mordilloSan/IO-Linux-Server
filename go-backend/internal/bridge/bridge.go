package bridge

import (
	"bytes"
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

type BridgeHealthRequest struct {
	Type    string `json:"type"`    // e.g., "healthcheck" or "validate"
	Session string `json:"session"` // sessionID
}
type BridgeHealthResponse struct {
	Status  string `json:"status"` // "ok" or "invalid"
	Message string `json:"message,omitempty"`
}

// Response struct to parse bridge response
type BridgeResponse struct {
	Status string          `json:"status"`
	Output json.RawMessage `json:"output"`
	Error  string          `json:"error"`
}

var (
	processes   = make(map[string]*BridgeProcess)
	processesMu sync.Mutex
)

var (
	mainSocketListeners   = make(map[string]net.Listener) // sessionID → Listener
	mainSocketListenersMu sync.Mutex
)

// MainSocketPath returns the per-session main (healthcheck) socket path for the user.
func MainSocketPath(sess *session.Session) string {
	u, err := user.Lookup(sess.User.ID)
	if err != nil {
		panic(fmt.Sprintf("could not find user %s: %v", sess.User.ID, err))
	}
	return fmt.Sprintf("/run/user/%s/linuxio-main-%s.sock", u.Uid, sess.SessionID)
}

// BridgeSocketPath returns the per-session bridge command socket path for the user.
func BridgeSocketPath(sess *session.Session) string {
	u, err := user.Lookup(sess.User.ID)
	if err != nil {
		panic(fmt.Sprintf("could not find user %s: %v", sess.User.ID, err))
	}
	return fmt.Sprintf("/run/user/%s/linuxio-bridge-%s.sock", u.Uid, sess.SessionID)
}

// Use everywhere for bridge actions: returns *raw* JSON response string (for HTTP handler to decode output as needed)
func CallWithSession(sess *session.Session, reqType, command string, args []string) (string, error) {
	socketPath := BridgeSocketPath(sess)
	return CallViaSocket(socketPath, reqType, command, args)
}

// Now returns the full JSON bridge response as string, not just output!
func CallViaSocket(socketPath, reqType, command string, args []string) (string, error) {
	req := map[string]any{
		"type":    reqType,
		"command": command,
	}
	if args != nil {
		req["args"] = args
	}
	conn, err := net.DialTimeout("unix", socketPath, 2*time.Second)
	if err != nil {
		return "", fmt.Errorf("failed to connect to bridge: %w", err)
	}
	defer conn.Close()
	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)
	if err := enc.Encode(req); err != nil {
		return "", fmt.Errorf("failed to send request to bridge: %w", err)
	}
	var resp BridgeResponse
	if err := dec.Decode(&resp); err != nil {
		return "", fmt.Errorf("failed to decode response from bridge: %w", err)
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return "", fmt.Errorf("failed to marshal bridge response: %w", err)
	}
	return string(b), nil
}

// StartBridge starts the bridge process for a given session.
func StartBridge(sess *session.Session, sudoPassword string) error {
	processesMu.Lock()
	defer processesMu.Unlock()

	if _, exists := processes[sess.SessionID]; exists {
		return errors.New("bridge already running for this session")
	}

	var cmd *exec.Cmd
	if sess.Privileged {
		cmd = exec.Command("sudo", "-S", "env",
			"LINUXIO_SESSION_ID="+sess.SessionID,
			"LINUXIO_SESSION_USER="+sess.User.ID,
			"GO_ENV="+os.Getenv("GO_ENV"),
			"VERBOSE="+os.Getenv("VERBOSE"),
			bridgeBinary,
		)
	} else {
		cmd = exec.Command(bridgeBinary)
		cmd.Env = append(os.Environ(),
			"LINUXIO_SESSION_ID="+sess.SessionID,
			"LINUXIO_SESSION_USER="+sess.User.ID,
			"GO_ENV="+os.Getenv("GO_ENV"),
			"VERBOSE="+os.Getenv("VERBOSE"),
		)
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	if sess.Privileged && sudoPassword != "" {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			logger.Errorf("Failed to get stdin pipe: %v", err)
			return err
		}
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, sudoPassword+"\n")
		}()
	}

	if err := cmd.Start(); err != nil {
		logger.Errorf("Failed to start bridge for session %s: %v", sess.SessionID, err)
		return err
	}

	logger.Infof("Started %sbridge for session %s (pid=%d)",
		func() string {
			if sess.Privileged {
				return "privileged "
			}
			return ""
		}(), sess.SessionID, cmd.Process.Pid)

	processes[sess.SessionID] = &BridgeProcess{
		Cmd:       cmd,
		SessionID: sess.SessionID,
		StartedAt: time.Now(),
	}

	// Panic guard for process cleanup goroutine
	go func(sessID string, cmd *exec.Cmd, stdoutBuf, stderrBuf *bytes.Buffer) {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("Panic in process cleanup goroutine for session %s: %v", sessID, r)
			}
		}()
		logger.Infof("Captured output buffers for session %s: STDOUT=%d bytes, STDERR=%d bytes", sessID, stdoutBuf.Len(), stderrBuf.Len())

		err := cmd.Wait()
		processesMu.Lock()
		defer processesMu.Unlock()
		delete(processes, sessID)

		stdout := strings.TrimSpace(stdoutBuf.String())
		stderr := strings.TrimSpace(stderrBuf.String())

		if stdout != "" {
			logger.Infof("STDOUT for session %s:\n%s", sessID, stdout)
		}
		if stderr != "" {
			logger.Warnf("STDERR for session %s:\n%s", sessID, stderr)
		}

		if err != nil {
			logger.Warnf("Bridge for session %s exited with error: %v", sessID, err)
		} else {
			logger.Infof("Bridge for session %s exited", sessID)
		}
	}(sess.SessionID, cmd, &stdoutBuf, &stderrBuf)

	return nil
}

// StartBridgeSocket starts a Unix socket server for the main process.
func StartBridgeSocket(sess *session.Session) error {
	socketPath := MainSocketPath(sess)
	_ = os.Remove(socketPath)
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		logger.Errorf("Failed to listen on main socket for session %s: %v", sess.SessionID, err)
		return err
	}

	// Set permissions strictly to 0600 (owner read/write only)
	if err := os.Chmod(socketPath, 0600); err != nil {
		_ = ln.Close()
		_ = os.Remove(socketPath)
		logger.Errorf("Failed to chmod main socket %s: %v", socketPath, err)
		return fmt.Errorf("failed to chmod socket: %w", err)
	}

	// Store the listener so we can close and remove it on logout
	mainSocketListenersMu.Lock()
	mainSocketListeners[sess.SessionID] = ln
	mainSocketListenersMu.Unlock()

	logger.Infof("Main socket for session %s is now listening on %s", sess.SessionID, socketPath)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				logger.Warnf("Accept failed on main socket for session %s: %v", sess.SessionID, err)
				// Exit the goroutine if the listener is closed
				break
			}
			logger.Infof("Main socket for session %s accepted a connection", sess.SessionID)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Errorf("Panic in main socket handler: %v", r)
					}
				}()
				handleBridgeRequest(conn)
			}()
		}
	}()

	return nil
}

func handleBridgeRequest(conn net.Conn) {
	defer conn.Close()
	logger.Infof("Main socket accepted a connection")
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	var req BridgeHealthRequest
	if err := decoder.Decode(&req); err != nil {
		logger.Warnf("Invalid JSON on main socket: %v", err)
		_ = encoder.Encode(BridgeHealthResponse{Status: "error", Message: "invalid json"})
		return
	}

	if req.Type == "validate" {
		logger.Infof("Healthcheck received for session %s", req.Session)
		if session.IsValid(req.Session) {
			_ = encoder.Encode(BridgeHealthResponse{Status: "ok"})
		} else {
			_ = encoder.Encode(BridgeHealthResponse{Status: "invalid", Message: "session expired"})
		}
		return
	}
	logger.Warnf("Unknown healthcheck request type: %s (session %s)", req.Type, req.Session)
	_ = encoder.Encode(BridgeHealthResponse{Status: "error", Message: "unknown request type"})
}

func CleanupBridgeSocket(sess *session.Session) {
	mainSocketListenersMu.Lock()
	ln, ok := mainSocketListeners[sess.SessionID]
	if ok {
		err := ln.Close()
		if err != nil {
			logger.Warnf("Error closing main socket listener for session %s: %v", sess.SessionID, err)
		} else {
			logger.Infof("Closed main socket listener for session %s", sess.SessionID)
		}
		delete(mainSocketListeners, sess.SessionID)
	}
	mainSocketListenersMu.Unlock()
	//remove main socket
	mainSock := MainSocketPath(sess)
	if err := os.Remove(mainSock); err == nil {
		logger.Infof("Removed main socket file %s for session %s", mainSock, sess.SessionID)
	} else if !os.IsNotExist(err) {
		logger.Warnf("Failed to remove main socket file %s: %v", mainSock, err)
	}
	//remove bridge socket
	bridgeSock := BridgeSocketPath(sess)
	if err := os.Remove(bridgeSock); err == nil {
		logger.Infof("Removed bridge socket file %s for session %s", bridgeSock, sess.SessionID)
	} else if !os.IsNotExist(err) {
		logger.Warnf("Failed to remove bridge socket file %s: %v", bridgeSock, err)
	}
}
