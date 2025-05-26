package dbus

// In cmd/bridge/dbus/dbus.go

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/godbus/dbus/v5"
)

type ServiceStatus struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	LoadState   string `json:"load_state"`
	ActiveState string `json:"active_state"`
	SubState    string `json:"sub_state"`
}

func ListServices() (string, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return "", fmt.Errorf("failed to connect to system bus: %w", err)
	}
	defer conn.Close()

	systemd := conn.Object("org.freedesktop.systemd1", "/org/freedesktop/systemd1")

	var units [][]interface{}
	err = systemd.Call("org.freedesktop.systemd1.Manager.ListUnits", 0).Store(&units)
	if err != nil {
		return "", fmt.Errorf("ListUnits failed: %w", err)
	}

	var services []ServiceStatus
	for _, u := range units {
		name := u[0].(string)
		if !strings.HasSuffix(name, ".service") {
			continue
		}
		svc := ServiceStatus{
			Name:        name,
			Description: u[1].(string),
			LoadState:   u[2].(string),
			ActiveState: u[3].(string),
			SubState:    u[4].(string),
		}
		services = append(services, svc)
	}
	jsonBytes, err := json.MarshalIndent(services, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
