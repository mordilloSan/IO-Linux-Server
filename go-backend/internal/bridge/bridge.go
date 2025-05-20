// internal/bridge/bridge.go
package bridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"
)

const socketPath = "/run/linuxio-bridge.sock"

type Request struct {
	Type    string   `json:"type"`    // "dbus" or "command"
	Command string   `json:"command"` // "reboot", "poweroff", etc
	Args    []string `json:"args,omitempty"`
}

type Response struct {
	Status string `json:"status"`
	Output string `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

// SendRequest sends a request to the bridge and returns the parsed response.
// timeoutSec: set to 5 for a quick fail, or more for slow commands (updates, etc)
func SendRequest(req Request, timeoutSec int) (*Response, error) {
	conn, err := net.DialTimeout("unix", socketPath, time.Duration(timeoutSec)*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to bridge: %w", err)
	}
	defer conn.Close()

	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)

	if err := enc.Encode(req); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	var resp Response
	if err := dec.Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if resp.Status != "ok" {
		return &resp, errors.New(resp.Error)
	}
	return &resp, nil
}

// Convenience helpers:
func RebootSystem() error {
	_, err := SendRequest(Request{Type: "dbus", Command: "reboot"}, 5)
	return err
}

func PowerOffSystem() error {
	_, err := SendRequest(Request{Type: "dbus", Command: "poweroff"}, 5)
	return err
}

func RunCommand(command string, args ...string) (*Response, error) {
	return SendRequest(Request{Type: "command", Command: command, Args: args}, 20)
}
