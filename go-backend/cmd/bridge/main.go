package main

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"os/exec"
)

const socketPath = "/run/linuxio-bridge.sock"

type Request struct {
	Command string   `json:"command"` // e.g., "pkcon"
	Args    []string `json:"args"`    // e.g., ["update", "--noninteractive"]
}

type Response struct {
	Status string `json:"status"`           // "ok" or "error"
	Output string `json:"output,omitempty"` // stdout/stderr
	Error  string `json:"error,omitempty"`  // exec error
}

func main() {
	_ = os.RemoveAll(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("❌ Failed to listen on socket: %v", err)
	}
	defer listener.Close()

	if err := os.Chmod(socketPath, 0660); err != nil {
		log.Fatalf("❌ Failed to chmod socket: %v", err)
	}

	log.Println("🔐 Privileged bridge is listening:", socketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("⚠️ Accept failed: %v", err)
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
		log.Printf("❌ Invalid JSON: %v", err)
		_ = encoder.Encode(Response{Status: "error", Error: "invalid JSON"})
		return
	}

	if req.Command == "" {
		_ = encoder.Encode(Response{Status: "error", Error: "missing command"})
		return
	}

	cmd := exec.Command(req.Command, req.Args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("❌ Command failed: %s %v - %v", req.Command, req.Args, err)
		_ = encoder.Encode(Response{Status: "error", Error: err.Error(), Output: string(out)})
		return
	}

	_ = encoder.Encode(Response{Status: "ok", Output: string(out)})
}
