package main

import (
	"encoding/json"
	"fmt"
	"go-backend/cmd/bridge/dbus"
	"go-backend/internal/bridge"
	"go-backend/internal/logger"
	"net"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"
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
			ok := checkMainProcessHealth(sessionID, username)
			if !ok {
				logger.Warning.Printf("❌ Main process unreachable or session invalid, bridge exiting...")
				bridge.CleanupBridgeSocket(sessionID, username)
				os.Exit(1)
			}
			time.Sleep(time.Minute)
		}
	}()

	killLingeringBridgeStartupProcesses()

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
	case "control":
		handleInternalCommand(req, encoder)
	default:
		logger.Warning.Printf("❌ Unknown request type: %s", req.Type)
		_ = encoder.Encode(Response{Status: "error", Error: "invalid type"})
	}
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

func handleDbusCommand(req Request, enc *json.Encoder) {
	logger.Info.Printf("🔒 Handling D-Bus command: %s\n", req.Command)
	var err error
	var jsonOut string

	switch req.Command {
	case "Reboot", "PowerOff":
		err = dbus.CallLogin1Action(req.Command)
	case "GetUpdates":
		jsonOut, err = dbus.GetUpdatesWithDetails()
	case "InstallPackage":
		if len(req.Args) == 0 {
			_ = enc.Encode(Response{Status: "error", Error: "missing package ID"})
			return
		}
		err = dbus.InstallPackage(req.Args[0])
	case "ListServices":
		jsonOut, err = dbus.ListServices()
	default:
		err = fmt.Errorf("unknown dbus command: %s", req.Command)
	}

	// Success with output
	if err == nil && jsonOut != "" {
		_ = enc.Encode(Response{Status: "ok", Output: jsonOut})
		return
	}
	// Success but no output
	if err == nil {
		logger.Info.Printf("✅ D-Bus %s succeeded\n", req.Command)
		_ = enc.Encode(Response{Status: "ok"})
		return
	}
	// Error
	logger.Error.Printf("❌ D-Bus %s failed: %v", req.Command, err)
	_ = enc.Encode(Response{Status: "error", Error: err.Error()})
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

func killLingeringBridgeStartupProcesses() {
	procEntries, err := os.ReadDir("/proc")
	if err != nil {
		logger.Error.Printf("❌ Failed to read /proc: %v", err)
		return
	}

	for _, entry := range procEntries {
		if !entry.IsDir() || !IsNumeric(entry.Name()) {
			continue
		}

		pid := entry.Name()
		cmdlineBytes, err := os.ReadFile(fmt.Sprintf("/proc/%s/cmdline", pid))
		if err != nil || len(cmdlineBytes) == 0 {
			continue
		}

		cmdline := strings.ReplaceAll(string(cmdlineBytes), "\x00", " ")

		if strings.Contains(cmdline, "linuxio-bridge") &&
			strings.Contains(cmdline, "sudo -S env") &&
			strings.Contains(cmdline, "LINUXIO_SESSION_USER="+os.Getenv("LINUXIO_SESSION_USER")) {
			pidInt, _ := strconv.Atoi(pid)
			logger.Debug.Printf("⚠️ Found lingering bridge process (pid=%d): %s", pidInt, cmdline)
			killParentTree(pidInt)
		}
	}
}

func killParentTree(pid int) {
	for {
		stat, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
		if err != nil {
			logger.Debug.Printf("killParentTree: could not read stat for pid %d: %v", pid, err)
			break
		}
		fields := strings.Fields(string(stat))
		if len(fields) < 4 {
			logger.Debug.Printf("killParentTree: stat fields < 4 for pid %d", pid)
			break
		}

		ppid, _ := strconv.Atoi(fields[3])
		if ppid <= 1 || ppid == pid {
			logger.Debug.Printf("killParentTree: hit root or self for pid %d (ppid %d)", pid, ppid)
			break
		}

		commBytes, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", ppid))
		if err != nil {
			logger.Debug.Printf("killParentTree: could not read comm for ppid %d: %v", ppid, err)
			break
		}

		comm := strings.TrimSpace(string(commBytes))
		logger.Debug.Printf("killParentTree: pid=%d, ppid=%d, comm='%s'", pid, ppid, comm)
		if comm == "sudo" || comm == "env" {
			logger.Debug.Printf("🛑 Killing sudo/env process (pid=%d, ppid=%d, comm=%s)", pid, ppid, comm)
			_ = syscall.Kill(ppid, syscall.SIGTERM)
			_ = syscall.Kill(pid, syscall.SIGTERM)
			time.Sleep(250 * time.Millisecond) // Give time for defer/logs to flush
			_ = syscall.Kill(ppid, syscall.SIGKILL)
			_ = syscall.Kill(pid, syscall.SIGKILL)
			break
		}
		pid = ppid
	}
}

func checkMainProcessHealth(sessionID string, username string) bool {
	sock := bridge.MainSocketPath(sessionID, username)
	conn, err := net.DialTimeout("unix", sock, 2*time.Second)
	if err != nil {
		logger.Warning.Printf("⚠️ Could not connect to main socket: %v", err)
		return false
	}
	defer conn.Close()

	req := BridgeHealthRequest{
		Type:    "validate",
		Session: sessionID,
	}
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		logger.Warning.Printf("⚠️ Failed to send health request: %v", err)
		return false
	}

	var resp BridgeHealthResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		logger.Warning.Printf("⚠️ Failed to decode health response: %v", err)
		return false
	}
	logger.Debug.Printf("Healthcheck: got %+v", resp)
	return resp.Status == "ok"
}

func IsNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
