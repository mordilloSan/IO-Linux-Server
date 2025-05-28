package main

import (
	"encoding/json"
	"fmt"
	"go-backend/cmd/bridge/cleanup"
	"go-backend/cmd/bridge/dbus"
	"go-backend/cmd/bridge/system"
	"go-backend/internal/bridge"
	"go-backend/internal/logger"
	"net"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"
)

type Request struct {
	Type    string   `json:"type"`
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

type Response struct {
	Status string      `json:"status"`
	Output interface{} `json:"output,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// Handler function signature for all commands
type HandlerFunc func(args []string) (interface{}, error)

// -- D-Bus Handlers --
var dbusHandlers = map[string]HandlerFunc{
	"Reboot":   func(args []string) (interface{}, error) { return nil, dbus.CallLogin1Action("Reboot") },
	"PowerOff": func(args []string) (interface{}, error) { return nil, dbus.CallLogin1Action("PowerOff") },
	"GetUpdates": func(args []string) (interface{}, error) {
		result, err := dbus.GetUpdatesWithDetails()
		if result == nil {
			result = []dbus.UpdateDetail{}
		}
		return result, err
	},
	"InstallPackage": func(args []string) (interface{}, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("missing package ID")
		}
		return nil, dbus.InstallPackage(args[0])
	},
	"ListServices": func(args []string) (interface{}, error) {
		return dbus.ListServices()
	},
	"GetServiceInfo": func(args []string) (interface{}, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("missing service name")
		}
		return dbus.GetServiceInfo(args[0])
	},
	"StartService":   func(args []string) (interface{}, error) { return nil, dbus.StartService(args[0]) },
	"StopService":    func(args []string) (interface{}, error) { return nil, dbus.StopService(args[0]) },
	"RestartService": func(args []string) (interface{}, error) { return nil, dbus.RestartService(args[0]) },
	"ReloadService":  func(args []string) (interface{}, error) { return nil, dbus.ReloadService(args[0]) },
	"EnableService":  func(args []string) (interface{}, error) { return nil, dbus.EnableService(args[0]) },
	"DisableService": func(args []string) (interface{}, error) { return nil, dbus.DisableService(args[0]) },
	"MaskService":    func(args []string) (interface{}, error) { return nil, dbus.MaskService(args[0]) },
	"UnmaskService":  func(args []string) (interface{}, error) { return nil, dbus.UnmaskService(args[0]) },
	"GetNetworkInfo": func(args []string) (interface{}, error) { return dbus.GetNetworkInfo() },
	"SetDNS": func(args []string) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("SetDNS requires interface and at least one DNS server")
		}
		return nil, dbus.SetDNS(args[0], args[1:])
	},
	"SetGateway": func(args []string) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("SetGateway requires interface and gateway address")
		}
		return nil, dbus.SetGateway(args[0], args[1])
	},
	"SetIPv4": func(args []string) (interface{}, error) {
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
	"SetIPv6": func(args []string) (interface{}, error) {
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
	"SetMTU": func(args []string) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("SetMTU requires interface and mtu value")
		}
		return nil, dbus.SetMTU(args[0], args[1])
	},
}

// -- Control Handlers (example, you can extend) --
var controlHandlers = map[string]HandlerFunc{
	"shutdown": func(args []string) (interface{}, error) {
		logger.Info.Println("Received shutdown command, exiting bridge")
		go func() {
			time.Sleep(200 * time.Millisecond)
			os.Exit(0)
		}()
		return "Bridge shutting down", nil
	},
	// Add more as needed...
}

var systemHandlers = map[string]HandlerFunc{
	"get_drive_info": func(args []string) (interface{}, error) {
		return system.FetchDriveInfo()
	},
	"get_smart_info": func(args []string) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("missing device argument")
		}
		return system.FetchSmartInfo(args[0])
	},
	"get_nvme_power": func(args []string) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("missing device argument")
		}
		return system.GetNVMePowerState(args[0])
	},
}

// -- Handler groups by type --
var handlersByType = map[string]map[string]HandlerFunc{
	"dbus":    dbusHandlers,
	"control": controlHandlers,
	"system":  systemHandlers,
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

	logger.Info.Printf("➡️ Received request: type=%s, command=%s, args=%v", req.Type, req.Command, req.Args)

	group, found := handlersByType[req.Type]
	if !found {
		logger.Warning.Printf("❌ Unknown request type: %s", req.Type)
		_ = encoder.Encode(Response{Status: "error", Error: fmt.Sprintf("invalid type: %s", req.Type)})
		return
	}

	handler, found := group[req.Command]
	if !found {
		logger.Warning.Printf("❌ Unknown command for type %s: %s", req.Type, req.Command)
		_ = encoder.Encode(Response{Status: "error", Error: fmt.Sprintf("unknown command: %s", req.Command)})
		return
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Error.Printf("🔥 Panic in %s command handler: %v", req.Type, r)
			_ = encoder.Encode(Response{Status: "error", Error: fmt.Sprintf("panic: %v", r)})
		}
	}()

	out, err := handler(req.Args)
	if err == nil {
		_ = encoder.Encode(Response{Status: "ok", Output: out})
		return
	}
	logger.Error.Printf("❌ %s %s failed: %v", req.Type, req.Command, err)
	_ = encoder.Encode(Response{Status: "error", Error: err.Error()})
}
