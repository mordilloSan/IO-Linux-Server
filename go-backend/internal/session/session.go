// session/session.go
package session

import (
	"go-backend/internal/logger"
	"go-backend/internal/utils"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Session struct {
	User      utils.User
	ExpiresAt time.Time
}

var (
	Sessions   = make(map[string]Session)
	SessionMux = make(chan func())
)

func init() {
	go func() {
		for f := range SessionMux {
			f()
		}
	}()
}

func StartSessionGC() {
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			now := time.Now()
			SessionMux <- func() {
				count := 0
				for id, s := range Sessions {
					if s.ExpiresAt.Before(now) {
						delete(Sessions, id)
						count++
					}
				}
				if count > 0 {
					logger.Info.Printf("[session] Garbage collected %d expired sessions (and stopped bridges)", count)
				}
			}
		}
	}()
}

// Creates a new session
func CreateSession(id string, user utils.User, duration time.Duration) {
	sess := Session{
		User:      user,
		ExpiresAt: time.Now().Add(duration),
	}
	SessionMux <- func() {
		Sessions[id] = sess
	}
	logger.Info.Printf("[session] Created session for user '%s'", user.ID)
}

// Use this to delete a session and stop its bridge process
func DeleteSession(id string) {
	SessionMux <- func() {
		sess, exists := Sessions[id]
		if exists {
			delete(Sessions, id)
			logger.Info.Printf("[session] Deleted session for user '%s'", sess.User.ID)
		}
	}
}

// Utility to validate session cookie and return user/ID/status
func ValidateFromRequest(r *http.Request) (utils.User, string, bool) {
	cookie, err := r.Cookie("session_id")
	if err != nil || cookie.Value == "" {
		return utils.User{}, "", false
	}

	var session Session
	var exists bool
	done := make(chan bool)

	SessionMux <- func() {
		session, exists = Sessions[cookie.Value]
		done <- true
	}
	<-done

	if !exists {
		logger.Warning.Printf("[session] Access attempt with unknown session_id: %s", cookie.Value)
		return utils.User{}, "", false
	}

	if session.ExpiresAt.Before(time.Now()) {
		logger.Warning.Printf("[session] Expired session access attempt by user '%s'", session.User.ID)
		return utils.User{}, "", false
	}

	return session.User, cookie.Value, true
}

// Checks if a session is valid
func IsValid(id string) bool {
	done := make(chan bool)
	var valid bool
	SessionMux <- func() {
		session, exists := Sessions[id]
		valid = exists && session.ExpiresAt.After(time.Now())
		done <- true
	}
	<-done
	return valid
}

// --- Bridge process handling --- //
func stopBridge(pid int) {
	if pid <= 0 {
		return
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		logger.Warning.Printf("[session] Could not find bridge process (PID %d): %v", pid, err)
		return
	}
	_ = proc.Signal(syscall.SIGTERM)
	// Wait a short time, then force kill if needed
	for i := 0; i < 10; i++ {
		if !processExists(pid) {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	_ = proc.Kill()
	logger.Info.Printf("[session] Forced kill bridge process (PID %d)", pid)
}

func processExists(pid int) bool {
	if pid <= 0 {
		return false
	}
	// signal 0 is used to check for existence
	err := syscall.Kill(pid, 0)
	return err == nil
}

// Optionally, add a helper to clean up bridges at startup
func CleanupLeftoverBridges() {
	out, _ := exec.Command("pgrep", "-x", "linuxio-bridge").Output()
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		pid, _ := strconv.Atoi(strings.TrimSpace(line))
		stopBridge(pid)
	}
}
