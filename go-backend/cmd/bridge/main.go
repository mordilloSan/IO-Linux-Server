package main

import (
	"encoding/json"
	"fmt"
	"go-backend/cmd/bridge/cleanup"
	"go-backend/cmd/bridge/dbus"
	"go-backend/internal/bridge"
	"go-backend/internal/logger"
	"net"
	"os"
	"os/user"
	"strconv"
	"time"
)

type Request struct {
	Type    string   `json:"type"`
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

// Output is any Go value, not a string!
type Response struct {
	Status string      `json:"status"`
	Output interface{} `json:"output,omitempty"`
	Error  string      `json:"error,omitempty"`
}

type BridgeHealthRequest struct {
	Type    string `json:"type"`    // e.g., "healthcheck" or "validate"
	Session string `json:"session"` // sessionID
}
type BridgeHealthResponse struct {
	Status  string `json:"status"` // "ok" or "invalid"
	Message string `json:"message,omitempty"`
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

	if sessionID == "" || username == "" {
		logger.Error.Fatalf("❌ LINUXIO_SESSION_ID and LINUXIO_SESSION_USER env vars required")
	}

	socketPath := bridge.BridgeSocketPath(sessionID, username)
	listener, uid, gid, err := createAndOwnSocket(socketPath, username)
	if err != nil {
		logger.Error.Fatalf("❌ %v", err)
	}
	defer listener.Close()
	defer func() {
		logger.Info.Println("🔐 linuxio-bridge shut down.")
		_ = os.Remove(socketPath)
	}()
	logger.Info.Printf("Listening succeeded.")
	logger.Info.Printf("🔑 Socket ownership set to %s (%d:%d)", username, uid, gid)

	logger.Info.Printf("🔐 linuxio-bridge listening: %s", socketPath)

	go func() {
		logger.Info.Printf("[bridge] Starting periodic health check (session: %s)", sessionID)
		for {
			logger.Debug.Printf("[bridge] Healthcheck: pinging main process for session %s", sessionID)
			ok := cleanup.CheckMainProcessHealth(sessionID, username)
			if !ok {
				logger.Warning.Printf("❌ Main process unreachable or session invalid, bridge exiting...")
				bridge.CleanupBridgeSocket(sessionID, username)
				os.Exit(1)
			}
			time.Sleep(time.Minute)
		}
	}()

	cleanup.KillLingeringBridgeStartupProcesses()

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Warning.Printf("⚠️ Accept failed: %v", err)
			continue
		}
		logger.Info.Printf("[bridge] Accepted new connection on bridge socket for session %s", sessionID)
		go handleConnection(conn)
	}

}

func createAndOwnSocket(socketPath, username string) (net.Listener, int, int, error) {
	// Lookup user
	u, err := user.Lookup(username)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to lookup user %s: %w", username, err)
	}
	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(u.Gid)

	_ = os.Remove(socketPath)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to listen on socket: %w", err)
	}

	if err := os.Chmod(socketPath, 0600); err != nil {
		listener.Close()
		os.Remove(socketPath)
		return nil, 0, 0, fmt.Errorf("failed to chmod socket: %w", err)
	}
	if err := os.Chown(socketPath, uid, gid); err != nil {
		listener.Close()
		os.Remove(socketPath)
		return nil, 0, 0, fmt.Errorf("failed to chown socket: %w", err)
	}

	return listener, uid, gid, nil
}

func handleInternalCommand(req Request, enc *json.Encoder) {
	logger.Info.Printf("🔒 Handling internal command: %s\n", req.Command)
	switch req.Command {
	case "shutdown":
		logger.Info.Println("Received shutdown command, exiting bridge")
		_ = enc.Encode(Response{Status: "ok", Output: "Bridge shutting down"})
		os.Exit(0)
	default:
		_ = enc.Encode(Response{Status: "error", Error: "unknown internal command"})
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
	case "control":
		handleInternalCommand(req, encoder)
	default:
		logger.Warning.Printf("❌ Unknown request type: %s", req.Type)
		_ = encoder.Encode(Response{Status: "error", Error: "invalid type"})
	}
}

// Only change here: return Go value directly in Output
func handleDbusCommand(req Request, enc *json.Encoder) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error.Printf("🔥 Panic in D-Bus command handler: %v", r)
			_ = enc.Encode(Response{Status: "error", Error: fmt.Sprintf("panic: %v", r)})
		}
	}()
	logger.Info.Printf("🔒 Handling D-Bus command: %s\n", req.Command)
	var err error
	var out interface{}

	switch req.Command {
	case "Reboot", "PowerOff":
		err = dbus.CallLogin1Action(req.Command)
	case "GetUpdates":
		var result []dbus.UpdateDetail
		result, err = dbus.GetUpdatesWithDetails()
		// Defensive: if result is nil, return empty slice, not null!
		if result == nil {
			result = make([]dbus.UpdateDetail, 0)
		}
		out = result
	case "InstallPackage":
		if len(req.Args) == 0 {
			_ = enc.Encode(Response{Status: "error", Error: "missing package ID"})
			return
		}
		err = dbus.InstallPackage(req.Args[0])
	case "ListServices":
		out, err = dbus.ListServices()
	case "GetServiceInfo":
		if len(req.Args) == 0 {
			_ = enc.Encode(Response{Status: "error", Error: "missing service name"})
			return
		}
		out, err = dbus.GetServiceInfo(req.Args[0])
	case "StartService":
		err = dbus.StartService(req.Args[0])
	case "StopService":
		err = dbus.StopService(req.Args[0])
	case "RestartService":
		err = dbus.RestartService(req.Args[0])
	case "ReloadService":
		err = dbus.ReloadService(req.Args[0])
	case "EnableService":
		err = dbus.EnableService(req.Args[0])
	case "DisableService":
		err = dbus.DisableService(req.Args[0])
	case "MaskService":
		err = dbus.MaskService(req.Args[0])
	case "UnmaskService":
		err = dbus.UnmaskService(req.Args[0])
	case "GetNetworkInfo":
		out, err = dbus.GetNetworkInfo()
	default:
		err = fmt.Errorf("unknown dbus command: %s", req.Command)
	}

	// --- Response logic ---
	if err == nil {
		_ = enc.Encode(Response{Status: "ok", Output: out})
		return
	}
	logger.Error.Printf("❌ D-Bus %s failed: %v", req.Command, err)
	_ = enc.Encode(Response{Status: "error", Error: err.Error()})
}
