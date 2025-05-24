package main

import (
	"encoding/json"
	"fmt"
	"go-backend/cmd/bridge/dbus"
	"go-backend/internal/logger"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strconv"
	"syscall"
)

type Request struct {
	Type    string   `json:"type"`    // "dbus", "command", or "control"
	Command string   `json:"command"` // e.g., "reboot", "poweroff", "pkcon", "shutdown"
	Args    []string `json:"args,omitempty"`
}

type Response struct {
	Status string `json:"status"`           // "ok" or "error"
	Output string `json:"output,omitempty"` // stdout/stderr
	Error  string `json:"error,omitempty"`
}

func main() {
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "development"
	}
	verbose := os.Getenv("VERBOSE") == "true"
	logger.Init(env, verbose)

	sessionID := os.Getenv("LINUXIO_SESSION_ID")
	username := os.Getenv("LINUXIO_SESSION_USER")
	backendURL := os.Getenv("LINUXIO_BACKEND_URL")

	if sessionID == "" || username == "" {
		logger.Error.Fatalf("❌ LINUXIO_SESSION_ID and LINUXIO_SESSION_USER env vars required")
	}

	u, err := user.Lookup(username)
	if err != nil {
		logger.Error.Fatalf("❌ Failed to lookup user %s: %v", username, err)
	}
	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(u.Gid)
	socketPath := fmt.Sprintf("/run/user/%d/linuxio-bridge-%s.sock", uid, sessionID)

	_ = os.Remove(socketPath)
	defer func() {
		logger.Info.Println("🔐 linuxio-bridge shut down.")
		_ = os.Remove(socketPath)
	}()

	logger.Info.Printf("linuxio-bridge: starting up for session %s user %s", sessionID, username)
	logger.Info.Printf("sessionID=%s username=%s backendURL=%s", sessionID, username, backendURL)

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

	_ = os.Chmod(socketPath, 0600)
	if err := os.Chown(socketPath, uid, gid); err != nil {
		logger.Error.Printf("❌ Failed to chown socket to %s: %v", username, err)
	} else {
		logger.Info.Printf("🔑 Socket ownership set to %s (%d:%d)", username, uid, gid)
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

	if req.Type == "control" && req.Command == "shutdown" {
		logger.Info.Println("Received shutdown command, exiting bridge")
		_ = encoder.Encode(Response{Status: "ok", Output: "Bridge shutting down"})
		os.Exit(0)
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
	case "GetUpdates":
		var jsonOut string
		jsonOut, err = dbus.GetUpdatesWithDetails()
		if err == nil {
			_ = enc.Encode(Response{Status: "ok", Output: jsonOut})
			return
		}
	case "InstallPackage":
		if len(req.Args) == 0 {
			_ = enc.Encode(Response{Status: "error", Error: "missing package ID"})
			return
		}
		err = dbus.InstallPackage(req.Args[0])
	default:
		err = fmt.Errorf("unknown dbus command: %s", req.Command)
	}

	if err != nil {
		logger.Error.Printf("❌ D-Bus %s failed: %v", req.Command, err)
		_ = enc.Encode(Response{Status: "error", Error: err.Error()})
		return
	}

	logger.Info.Printf("✅ D-Bus %s succeeded\n", req.Command)
	_ = enc.Encode(Response{Status: "ok"})
}

func handleShellCommand(req Request, enc *json.Encoder) {
	logger.Info.Printf("🔧 Handling shell command: %s %v", req.Command, req.Args)
	if req.Command == "" {
		_ = enc.Encode(Response{Status: "error", Error: "missing command"})
		return
	}
	cmd := exec.Command(req.Command, req.Args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		_ = enc.Encode(Response{Status: "error", Output: string(out), Error: err.Error()})
	} else {
		_ = enc.Encode(Response{Status: "ok", Output: string(out)})
	}
}
