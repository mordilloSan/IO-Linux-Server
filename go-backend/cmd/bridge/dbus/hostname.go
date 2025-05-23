package dbus

import (
	"fmt"
	"log"

	"github.com/godbus/dbus/v5"
)

func GetHostname() {
	// Connect to the system bus (root NOT required for this call)
	conn, err := dbus.SystemBus()
	if err != nil {
		log.Fatalf("Failed to connect to system bus: %v", err)
	}
	defer conn.Close()

	// Get object for hostname method
	obj := conn.Object("org.freedesktop.hostname1", "/org/freedesktop/hostname1")

	// Call the "Hostname" property (not a method, but works the same way)
	var variant dbus.Variant
	err = obj.Call("org.freedesktop.DBus.Properties.Get", 0,
		"org.freedesktop.hostname1", "Hostname").Store(&variant)
	if err != nil {
		log.Fatalf("Failed to get hostname: %v", err)
	}

	// The result comes as a dbus.Variant, extract string value
	hostname, ok := variant.Value().(string)
	if ok {
		fmt.Println("System hostname is:", hostname)
	} else {
		fmt.Println("Failed to extract hostname from variant")
	}
}
