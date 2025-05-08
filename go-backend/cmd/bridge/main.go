package main

import (
	"go-backend/internal/bridge"
	"log"
	"net"
	"os"
)

const socketPath = "/run/linuxio-bridge.sock"

func main() {
	// Clean up existing socket
	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatalf("Failed to remove existing socket: %v", err)
	}

	l, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to listen on socket: %v", err)
	}
	defer l.Close()

	// Make sure socket is only accessible by root or appropriate group
	if err := os.Chmod(socketPath, 0660); err != nil {
		log.Fatalf("Failed to chmod socket: %v", err)
	}

	log.Printf("🔐 linuxio-bridge listening on %s", socketPath)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("⚠️ Accept error: %v", err)
			continue
		}
		go bridge.HandleConnection(conn)
	}
}
