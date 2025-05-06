package dbus

import (
	"context"
	"fmt"

	"github.com/godbus/dbus/v5"
)

type Login1Manager struct {
	conn *dbus.Conn
	obj  dbus.BusObject
}

// Connects to system D-Bus and prepares the org.freedesktop.login1.Manager interface.
func NewLogin1Manager(ctx context.Context) (*Login1Manager, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to system bus: %w", err)
	}

	obj := conn.Object("org.freedesktop.login1", "/org/freedesktop/login1")
	return &Login1Manager{
		conn: conn,
		obj:  obj,
	}, nil
}

// Internal generic call handler for login1 methods.
func (m *Login1Manager) call(ctx context.Context, method string) error {
	call := m.obj.CallWithContext(ctx, "org.freedesktop.login1.Manager."+method, 0, false)
	if call.Err != nil {
		return fmt.Errorf("failed to call %s: %w", method, call.Err)
	}
	return nil
}

// Reboot calls the D-Bus method to reboot the system.
func (m *Login1Manager) Reboot(ctx context.Context) error {
	return m.call(ctx, "Reboot")
}

// PowerOff calls the D-Bus method to power off the system.
func (m *Login1Manager) PowerOff(ctx context.Context) error {
	return m.call(ctx, "PowerOff")
}