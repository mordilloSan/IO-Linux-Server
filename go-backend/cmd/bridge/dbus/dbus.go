package dbus

import (
	"strings"
	"time"
)

// --- Retry Wrapper ---
func RetryOnceIfClosed(initialErr error, do func() error) error {
	if initialErr == nil {
		err := do()
		if err != nil && strings.Contains(err.Error(), "use of closed network connection") {
			time.Sleep(150 * time.Millisecond)
			return do()
		}
		return err
	}
	if strings.Contains(initialErr.Error(), "use of closed network connection") {
		time.Sleep(150 * time.Millisecond)
		return do()
	}
	return initialErr
}
