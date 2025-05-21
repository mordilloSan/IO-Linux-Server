package main

import (
	"encoding/json"
	"go-backend/internal/dbus"
	"go-backend/internal/logger"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
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
	// Clean up socket file on exit
	defer func() {
		_ = os.Remove(socketPath)
		logger.Info.Println("🔐 linuxio-bridge shut down.")
	}()

	_ = os.RemoveAll(socketPath)

	// Trap SIGINT/SIGTERM for graceful exit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		os.Exit(0)
	}()

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		logger.Error.Fatalf("❌ Failed to listen on socket: %v", err)
	}
	defer listener.Close()
	_ = os.Chmod(socketPath, 0600) // Only root
	logger.Info.Println("🔐 linuxio-bridge listening:", socketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error.Printf("⚠️ Accept failed: %v", err)
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
		logger.Error.Println("❌ Invalid JSON from client:", err)
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
		logger.Error.Printf("❌ Unknown request type: %s", req.Type)
		_ = encoder.Encode(Response{Status: "error", Error: "invalid type"})
	}
}

func handleDbusCommand(req Request, enc *json.Encoder) {
	logger.Info.Printf("🔒 Handling D-Bus command: %s", req.Command)
	switch req.Command {
	case "reboot":
		if err := dbus.RebootSystem(); err != nil {
			logger.Error.Println("❌ D-Bus reboot failed:", err)
			_ = enc.Encode(Response{Status: "error", Error: err.Error()})
			return
		}
		logger.Info.Println("✅ D-Bus reboot succeeded")
		_ = enc.Encode(Response{Status: "ok"})
	case "poweroff":
		if err := dbus.PowerOffSystem(); err != nil {
			logger.Error.Println("❌ D-Bus poweroff failed:", err)
			_ = enc.Encode(Response{Status: "error", Error: err.Error()})
			return
		}
		logger.Info.Println("✅ D-Bus poweroff succeeded")
		_ = enc.Encode(Response{Status: "ok"})
	default:
		logger.Error.Printf("❌ Unknown D-Bus command: %s", req.Command)
		_ = enc.Encode(Response{Status: "error", Error: "unknown dbus command"})
	}
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
	if err != nil {
		logger.Error.Printf("❌ Command failed: %s %v - %v", req.Command, req.Args, err)
		_ = enc.Encode(Response{Status: "error", Output: string(out), Error: err.Error()})
		return
	}
	logger.Info.Printf("✅ Command succeeded: %s %v", req.Command, req.Args)
	_ = enc.Encode(Response{Status: "ok", Output: string(out)})
}
