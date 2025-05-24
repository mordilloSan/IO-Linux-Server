// session/session.go
package session

import (
	"go-backend/internal/logger"
	"go-backend/internal/utils"
	"net/http"
	"time"
)

type Session struct {
	User       utils.User
	ExpiresAt  time.Time
	Privileged bool
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

// Starts a goroutine that periodically checks for expired sessions
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
func CreateSession(id string, user utils.User, duration time.Duration, privileged bool) {
	sess := Session{
		User:       user,
		ExpiresAt:  time.Now().Add(duration),
		Privileged: privileged,
	}
	SessionMux <- func() {
		Sessions[id] = sess
	}
	logger.Info.Printf("[session] Created session for user '%s'", user.ID)
}

// Deletes a session
func DeleteSession(id string) {
	SessionMux <- func() {
		sess, exists := Sessions[id]
		if exists {
			delete(Sessions, id)
			logger.Info.Printf("[session] Deleted session for user '%s'", sess.User.ID)
		}
	}
}

// Checks if a session is privileged
func IsPrivileged(sessionID string) bool {
	done := make(chan bool)
	var privileged bool
	SessionMux <- func() {
		sess, exists := Sessions[sessionID]
		privileged = exists && sess.Privileged
		done <- true
	}
	<-done
	return privileged
}

// Utility to validate session cookie and return user/ID/status
func ValidateFromRequest(r *http.Request) (utils.User, string, bool, bool) {
	cookie, err := r.Cookie("session_id")
	if err != nil || cookie.Value == "" {
		return utils.User{}, "", false, false
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
		return utils.User{}, "", false, false
	}

	if session.ExpiresAt.Before(time.Now()) {
		logger.Warning.Printf("[session] Expired session access attempt by user '%s'", session.User.ID)
		return utils.User{}, "", false, false
	}

	return session.User, cookie.Value, true, session.Privileged
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

// Changes the privileged status of a session
func SetPrivileged(sessionID string, privileged bool) {
	SessionMux <- func() {
		sess, exists := Sessions[sessionID]
		if exists {
			sess.Privileged = privileged
			Sessions[sessionID] = sess
		}
	}
}

// Returns a list of all currently valid (non-expired) session IDs
func GetActiveSessionIDs() []string {
	done := make(chan []string)

	SessionMux <- func() {
		now := time.Now()
		var active []string
		for id, s := range Sessions {
			if s.ExpiresAt.After(now) {
				active = append(active, id)
			}
		}
		done <- active
	}

	return <-done
}
