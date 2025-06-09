package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-backend/cmd/bridge/cleanup"
	"go-backend/cmd/bridge/dbus"
	"go-backend/cmd/bridge/system"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/mordilloSan/LinuxIO/go-backend/internal/bridge"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/logger"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/session"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/utils"

	"github.com/google/uuid"
)

// Request represents the standard JSON request format sent to both built-in handlers and external helpers.
type Request struct {
	Type    string   `json:"type"`
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

// Response represents the standard JSON response returned by both built-in handlers and helpers.
type Response struct {
	Status string `json:"status"`
	Output any    `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

// HandlerFunc is the function signature for all built-in command handlers.
type HandlerFunc func(args []string) (any, error)

// ---- Built-in Handler Registration ----
// -- D-Bus Handlers --
var dbusHandlers = map[string]HandlerFunc{
	"Reboot":         func(args []string) (any, error) { return nil, dbus.CallLogin1Action("Reboot") },
	"PowerOff":       func(args []string) (any, error) { return nil, dbus.CallLogin1Action("PowerOff") },
	"GetUpdates":     func(args []string) (any, error) { return dbus.GetUpdatesWithDetails() },
	"InstallPackage": func(args []string) (any, error) { return nil, dbus.InstallPackage(args[0]) },
	"ListServices":   func(args []string) (any, error) { return dbus.ListServices() },
	"GetServiceInfo": func(args []string) (any, error) { return dbus.GetServiceInfo(args[0]) },
	"StartService":   func(args []string) (any, error) { return nil, dbus.StartService(args[0]) },
	"StopService":    func(args []string) (any, error) { return nil, dbus.StopService(args[0]) },
	"RestartService": func(args []string) (any, error) { return nil, dbus.RestartService(args[0]) },
	"ReloadService":  func(args []string) (any, error) { return nil, dbus.ReloadService(args[0]) },
	"EnableService":  func(args []string) (any, error) { return nil, dbus.EnableService(args[0]) },
	"DisableService": func(args []string) (any, error) { return nil, dbus.DisableService(args[0]) },
	"MaskService":    func(args []string) (any, error) { return nil, dbus.MaskService(args[0]) },
	"UnmaskService":  func(args []string) (any, error) { return nil, dbus.UnmaskService(args[0]) },
	"GetNetworkInfo": func(args []string) (any, error) { return dbus.GetNetworkInfo() },
	"SetDNS":         func(args []string) (any, error) { return nil, dbus.SetDNS(args[0], args[1:]) },
	"SetGateway":     func(args []string) (any, error) { return nil, dbus.SetGateway(args[0], args[1]) },
	"SetMTU":         func(args []string) (any, error) { return nil, dbus.SetMTU(args[0], args[1]) },
	"SetIPv4": func(args []string) (any, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("SetIPv4 requires interface and method (dhcp/static)")
		}
		iface, method := args[0], strings.ToLower(args[1])
		switch method {
		case "dhcp":
			return nil, dbus.SetIPv4DHCP(iface)
		case "static":
			if len(args) != 3 {
				return nil, fmt.Errorf("SetIPv4 static requires addressCIDR")
			}
			return nil, dbus.SetIPv4Static(iface, args[2])
		default:
			return nil, fmt.Errorf("SetIPv4 method must be 'dhcp' or 'static'")
		}
	},
	"SetIPv6": func(args []string) (any, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("SetIPv6 requires interface and method (dhcp/static)")
		}
		iface, method := args[0], strings.ToLower(args[1])
		switch method {
		case "dhcp":
			return nil, dbus.SetIPv6DHCP(iface)
		case "static":
			if len(args) != 3 {
				return nil, fmt.Errorf("SetIPv6 static requires addressCIDR")
			}
			return nil, dbus.SetIPv6Static(iface, args[2])
		default:
			return nil, fmt.Errorf("SetIPv6 method must be 'dhcp' or 'static'")
		}
	},
}

// -- Control Handlers --
var controlHandlers = map[string]HandlerFunc{
	"shutdown": func(args []string) (any, error) {
		logger.Infof("Received shutdown command, exiting bridge")
		go func() {
			time.Sleep(200 * time.Millisecond)
			os.Exit(0)
		}()
		return "Bridge shutting down", nil
	},
}

// -- System Handlers --
var systemHandlers = map[string]HandlerFunc{
	"get_drive_info": func(args []string) (any, error) {
		return system.FetchDriveInfo()
	},
	"get_smart_info": func(args []string) (any, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("missing device argument")
		}
		return system.FetchSmartInfo(args[0])
	},
	"get_nvme_power": func(args []string) (any, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("missing device argument")
		}
		return system.GetNVMePowerState(args[0])
	},
}

// -- Handler groups by type (built-in, for backwards compatibility) --
var handlersByType = map[string]map[string]HandlerFunc{
	"dbus":    dbusHandlers,
	"control": controlHandlers,
	"system":  systemHandlers,
	"modules": {}, // Placeholder for external helpers
}

// modulesDir returns the modules directory, overridable via env for flexibility.
func modulesDir() string {
	if val := os.Getenv("LINUXIO_MODULES_DIR"); val != "" {
		return val
	}
	return "/usr/lib/linuxio/modules"
}

func main() {
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "production"
	}
	verbose := os.Getenv("VERBOSE") == "true"
	logger.Init(env, verbose)

	sessionID := os.Getenv("LINUXIO_SESSION_ID")
	username := os.Getenv("LINUXIO_SESSION_USER")
	// Build minimal session object
	sess := &session.Session{
		SessionID: sessionID,
		User:      utils.User{ID: username, Name: username},
		// If you want, also read and set .Privileged from another env var
	}
	socketPath := bridge.BridgeSocketPath(sess)
	listener, uid, gid, err := createAndOwnSocket(socketPath, sess.User.ID)
	if err != nil {
		logger.Error.Fatalf("❌ %v", err)
	}

	// Trap SIGTERM and SIGINT for clean shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
		sig := <-sigChan
		logger.Infof("🛑 Caught signal: %s — shutting down bridge", sig)
		listener.Close()
		_ = os.Remove(socketPath)
		os.Exit(0)
	}()

	defer listener.Close()
	defer func() {
		logger.Infof("🔐 linuxio-bridge shut down.")
		_ = os.Remove(socketPath)
	}()
	logger.Debugf("🔑 Socket ownership set to %s (%d:%d)", username, uid, gid)
	logger.Infof("🔐 linuxio-bridge listening: %s", socketPath)
	runSelfTestIfDev(env)

	go func() {
		logger.Infof("Starting periodic health check (session: %s)", sess.SessionID)
		for {
			logger.Debugf("Healthcheck: pinging main process for session %s", sess.SessionID)
			ok := cleanup.CheckMainProcessHealth(sess)
			if !ok {
				logger.Warnf("❌ Main process unreachable or session invalid, bridge exiting...")
				bridge.CleanupBridgeSocket(sess)
				os.Exit(1)
			}
			time.Sleep(time.Minute)
		}
	}()

	// When in production, clean up any lingering bridge startup processes.
	if env == "production" {
		cleanup.KillLingeringBridgeStartupProcesses()
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Warnf("⚠️ Accept failed: %v", err)
			continue
		}
		id := uuid.NewString()
		logger.Debugf("MAIN: spawning handler %s", id)
		go handleConnection(conn, id)
	}
}

// createAndOwnSocket creates a unix socket at socketPath, ensures only the target user can access it.
func createAndOwnSocket(socketPath, username string) (net.Listener, int, int, error) {
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

// handleConnection processes incoming bridge requests.
// If the command is not built-in, dispatches to an external helper binary in the modules directory.
func handleConnection(conn net.Conn, id string) {
	logger.Debugf("HANDLECONNECTION: [%s] called!", id)
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	var req Request
	if err := decoder.Decode(&req); err != nil {
		if err == io.EOF {
			logger.Debugf("🔁 [%s] connection closed without data (likely healthcheck probe)", id)
		} else {
			logger.Warnf("❌ [%s] invalid JSON from client: %v", id, err)
		}
		_ = encoder.Encode(Response{Status: "error", Error: "invalid JSON"})
		return
	}

	// (1) DEFENSE-IN-DEPTH: Validate handler name for fallback
	if strings.ContainsAny(req.Type, "./\\") || strings.ContainsAny(req.Command, "./\\") {
		logger.Warnf("❌ [%s] Invalid characters in type/command: type=%q, command=%q", id, req.Type, req.Command)
		_ = encoder.Encode(Response{Status: "error", Error: "invalid characters in command/type"})
		return
	}

	logger.Infof("➡️ Received request: type=%s, command=%s, args=%v", req.Type, req.Command, req.Args)

	// (2) Avoid nil map panic and clarify intent
	group, found := handlersByType[req.Type]
	if found && group != nil {
		if handler, ok := group[req.Command]; ok {
			defer func() {
				if r := recover(); r != nil {
					logger.Errorf("🔥 Panic in %s command handler: %v", req.Type, r)
					_ = encoder.Encode(Response{Status: "error", Error: fmt.Sprintf("panic: %v", r)})
				}
			}()
			out, err := handler(req.Args)
			if err == nil {
				_ = encoder.Encode(Response{Status: "ok", Output: out})
				return
			}
			logger.Errorf("❌ %s %s failed: %v", req.Type, req.Command, err)
			_ = encoder.Encode(Response{Status: "error", Error: err.Error()})
			return
		}
	}

	// Fallback: try running external helper in modulesDir (modular extension point)
	helperPath := filepath.Join(modulesDir(), fmt.Sprintf("%s_%s", req.Type, req.Command))
	info, err := os.Stat(helperPath)
	if err == nil && !info.IsDir() && info.Mode()&0111 != 0 {
		logger.Infof("🔎 Dispatching to helper: %s", helperPath)
		runHelper(helperPath, req, encoder)
		return
	}

	logger.Warnf("❌ Unknown command for type %s: %s", req.Type, req.Command)
	_ = encoder.Encode(Response{Status: "error", Error: fmt.Sprintf("unknown command: %s", req.Command)})
}

// runHelper executes an external helper script or binary, passing the entire Request as JSON on stdin,
// and expects a JSON Response on stdout.
// If the helper fails, its stderr output is included in the error response for diagnostics.
// Malformed output is logged for troubleshooting.
func runHelper(path string, req Request, encoder *json.Encoder) {
	logger.Debugf("RUNHELPER: called for %s", path)

	inputBytes, _ := json.Marshal(req)

	cmd := exec.Command(path)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		logger.Errorf("Helper %s: failed to open stdin: %v", path, err)
		_ = encoder.Encode(Response{Status: "error", Error: "failed to open stdin for helper"})
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Errorf("Helper %s: failed to open stdout: %v", path, err)
		_ = encoder.Encode(Response{Status: "error", Error: "failed to open stdout for helper"})
		return
	}

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	// (3) Helper timeout: configurable
	timeout := 10 * time.Second
	if val := os.Getenv("BRIDGE_HELPER_TIMEOUT"); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil {
			timeout = parsed
		}
	}

	if err := cmd.Start(); err != nil {
		logger.Errorf("Helper %s: failed to start: %v", path, err)
		_ = encoder.Encode(Response{Status: "error", Error: fmt.Sprintf("failed to start helper: %v", err)})
		return
	}

	// Write input
	go func() {
		defer stdin.Close()
		_, _ = stdin.Write(inputBytes)
	}()

	// Read stdout concurrently
	var outBytes []byte
	stdoutDone := make(chan error, 1)
	go func() {
		var readErr error
		outBytes, readErr = io.ReadAll(stdout)
		stdoutDone <- readErr
	}()

	// Wait for process and output
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case <-time.After(timeout): // <--- use timeout variable
		_ = cmd.Process.Kill()
		logger.Errorf("Helper %s timed out.\n  STDOUT: %s\n  STDERR: %s", path, string(outBytes), stderrBuf.String())
		_ = encoder.Encode(Response{Status: "error", Error: "helper timed out"})
		return

	case err := <-done:
		<-stdoutDone // ensure output is fully read
		logger.Debugf("Helper %s finished. Exit error: %v\n  STDOUT: %s\n  STDERR: %s",
			path, err, string(outBytes), stderrBuf.String())
		if err != nil {
			_ = encoder.Encode(Response{Status: "error", Error: fmt.Sprintf("helper exited with error: %s", stderrBuf.String())})
			return
		}
	}

	// DEBUG: decode response
	logger.Debugf("DEBUG: About to decode helper output:\n=====\n%s\n=====", string(outBytes))

	var resp Response
	if err := json.Unmarshal(outBytes, &resp); err != nil {
		logger.Infof("Helper %s output (malformed JSON):\n  STDOUT: %s\n  STDERR: %s", path, string(outBytes), stderrBuf.String())
		_ = encoder.Encode(Response{Status: "error", Error: "invalid JSON from helper"})
		return
	}

	_ = encoder.Encode(resp)
}

func runSelfTestIfDev(env string) {
	if env != "development" {
		return
	}
	go func() {
		logger.Infof("🔍 Running bridge self-test for system_teste helper...")

		req := Request{Type: "system", Command: "teste"}
		var buf bytes.Buffer
		encoder := json.NewEncoder(&buf)

		runHelper(filepath.Join(modulesDir(), "system_teste"), req, encoder)

		rawJSON := buf.String()
		logger.Debugf("🧪 Self-test raw JSON output:\n%s", rawJSON)

		var resp Response
		if err := json.Unmarshal([]byte(rawJSON), &resp); err != nil {
			logger.Warnf("⚠️  Self-test: failed to decode response from system_teste helper: %v", err)
			return
		}

		if resp.Status != "ok" {
			logger.Warnf("⚠️  Self-test: system_teste returned error: %s", resp.Error)
		} else {
			// (4) Pretty-print Output
			pretty, err := json.MarshalIndent(resp.Output, "", "  ")
			if err != nil {
				logger.Infof("✅ Self-test succeeded, but failed to pretty-print output:")
				logger.Infof("%v", resp.Output)

			} else {
				logger.Infof("✅ Self-test succeeded. Output:")
				logger.Infof("%s", pretty)

			}
		}
	}()
}
func MainSocketPath(sess *session.Session) (string, error) {
	u, err := user.Lookup(sess.User.ID)
	if err != nil {
		logger.Errorf("could not find user %s: %v", sess.User.ID, err)
		return "", err
	}
	return fmt.Sprintf("/run/user/%s/linuxio-main-%s.sock", u.Uid, sess.SessionID), nil
}

func BridgeSocketPath(sess *session.Session) (string, error) {
	u, err := user.Lookup(sess.User.ID)
	if err != nil {
		logger.Errorf("could not find user %s: %v", sess.User.ID, err)
		return "", err
	}
	return fmt.Sprintf("/run/user/%s/linuxio-bridge-%s.sock", u.Uid, sess.SessionID), nil
}

/*
Auto-docs & Contributor note:

- To add a new bridge extension, drop an executable into the modules directory (default /usr/lib/linuxio/modules or override with LINUXIO_MODULES_DIR env).
- Name helpers as <type>_<command> (e.g., system_myfeature), chmod +x.
- Helper receives full Request JSON on stdin, must return a Response JSON on stdout.
- Example helper input: {"type":"system","command":"myfeature","args":["foo"]}
- Example helper output: {"status":"ok", "output":{...}}, or {"status":"error","error":"explanation"}
- stderr from helpers is logged and returned on error for troubleshooting.
*/
