package main

import (
	"encoding/json"
	"fmt"
	"go-backend/internal/dbus"
	"go-backend/internal/logger"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strconv"
	"syscall"
	"time"
)

const socketPath = "/run/linuxio-bridge.sock"

type Request struct {
	Type    string   `json:"type"`    // "dbus" or "command"
	Command string   `json:"command"` // e.g., "reboot", "poweroff", "pkcon"
	Args    []string `json:"args,omitempty"`
}

type Response struct {
	Status string `json:"status"`           // "ok" or "error"
	Output string `json:"output,omitempty"` // stdout/stderr
	Error  string `json:"error,omitempty"`
}

func main() {
	// Setup logging: auto stdout (dev) or journal (prod)
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "development"
	}
	verbose := os.Getenv("VERBOSE") == "true"
	logger.Init(env, verbose)

	defer func() {
		logger.Info.Println("🔐 linuxio-bridge shut down.")
		_ = os.Remove(socketPath)
	}()

	logger.Info.Println("linuxio-bridge: starting up")
	_ = os.RemoveAll(socketPath)

	sessionID := os.Getenv("LINUXIO_SESSION_ID")
	username := os.Getenv("LINUXIO_SESSION_USER")
	backendURL := os.Getenv("LINUXIO_BACKEND_URL")

	logger.Info.Printf("sessionID=%s username=%s backendURL=%s", sessionID, username, backendURL)

	// Start session watcher
	if sessionID != "" && backendURL != "" {
		logger.Info.Println("Starting sessionWatcher goroutine")
		go sessionWatcher(backendURL, sessionID)
	}

	// Trap SIGINT/SIGTERM for graceful exit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		logger.Info.Println("Received shutdown signal")
		os.Exit(0)
	}()

	logger.Info.Printf("Trying to listen on socket: %s", socketPath)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		logger.Error.Printf("❌ Failed to listen on socket: %v", err)
		os.Exit(1)
	}
	defer listener.Close()
	logger.Info.Println("Listening succeeded.")

	// Set permissions and ownership
	_ = os.Chmod(socketPath, 0660) // rw for user/group

	// Try to chown to user
	if username != "" {
		logger.Info.Printf("Chowning socket to user: %s", username)
		u, err := user.Lookup(username)
		if err != nil {
			logger.Error.Printf("❌ Failed to lookup user %s: %v", username, err)
		} else {
			uid, _ := strconv.Atoi(u.Uid)
			gid, _ := strconv.Atoi(u.Gid)
			if err := os.Chown(socketPath, uid, gid); err != nil {
				logger.Error.Printf("❌ Failed to chown socket to %s: %v", username, err)
			} else {
				logger.Info.Printf("🔑 Socket ownership set to %s (%d:%d)", username, uid, gid)
			}
		}
	}

	logger.Info.Printf("🔐 linuxio-bridge listening: %s", socketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Warning.Printf("⚠️ Accept failed: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	var req Request
	if err := decoder.Decode(&req); err != nil {
		logger.Warning.Printf("❌ Invalid JSON from client: %v", err)
		_ = encoder.Encode(Response{Status: "error", Error: "invalid JSON"})
		return
	}

	logger.Info.Printf("➡️ Received request: type=%s, command=%s, args=%v", req.Type, req.Command, req.Args)

	switch req.Type {
	case "dbus":
		handleDbusCommand(req, encoder)
	case "command":
		handleShellCommand(req, encoder)
	default:
		logger.Warning.Printf("❌ Unknown request type: %s", req.Type)
		_ = encoder.Encode(Response{Status: "error", Error: "invalid type"})
	}
}

func handleDbusCommand(req Request, enc *json.Encoder) {
	logger.Info.Printf("🔒 Handling D-Bus command: %s\n", req.Command)
	var err error

	switch req.Command {
	case "Reboot", "PowerOff":
		err = dbus.CallLogin1Action(req.Command)
	// ...add other  methods as needed
	default:
		err = fmt.Errorf("unknown dbus command: %s", req.Command)
	}

	if err != nil {
		logger.Error.Printf("❌ D-Bus %s failed: %v\n", req.Command, err)
		_ = enc.Encode(Response{Status: "error", Error: err.Error()})
		return
	}

	logger.Info.Printf("✅ D-Bus %s succeeded\n", req.Command)

}

func handleShellCommand(req Request, enc *json.Encoder) {
	logger.Info.Printf("🔧 Handling shell command: %s %v", req.Command, req.Args)
	if req.Command == "" {
		logger.Error.Println("❌ Missing shell command")
		_ = enc.Encode(Response{Status: "error", Error: "missing command"})
		return
	}
	cmd := exec.Command(req.Command, req.Args...)
	out, err := cmd.CombinedOutput()

	// Detect "pkcon get-updates" exit 5 (no updates)
	if req.Command == "pkcon" && len(req.Args) > 0 && req.Args[0] == "get-updates" {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 5 {
			logger.Info.Printf("✅ pkcon get-updates: no updates available (exit 5)")
			_ = enc.Encode(Response{Status: "ok", Output: string(out)})
			return
		}
	}

	if err != nil {
		logger.Error.Printf("❌ Command failed: %s %v - %v", req.Command, req.Args, err)
		_ = enc.Encode(Response{Status: "error", Output: string(out), Error: err.Error()})
		return
	}
	logger.Info.Printf("✅ Command succeeded: %s %v", req.Command, req.Args)
	_ = enc.Encode(Response{Status: "ok", Output: string(out)})
}

func sessionWatcher(backendURL, sessionID string) {
	client := &http.Client{Timeout: 5 * time.Second}
	url := backendURL + "/auth/me"
	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			logger.Error.Printf("Session watcher: failed to create request: %v", err)
			os.Exit(0)
		}
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			logger.Warning.Printf("Session watcher: session invalid or backend unreachable (status %v, err %v)", resp.StatusCode, err)
			os.Exit(0)
		}
		resp.Body.Close()
		time.Sleep(30 * time.Second) // Check every 30 seconds
	}
}
