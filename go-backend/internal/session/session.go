package session

import (
	"go-backend/internal/logger"
	"go-backend/internal/utils"
	"net/http"
	"time"
)

type Session struct {
	SessionID  string
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
			func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Errorf("Panic in session actor: %v", r)
					}
				}()
				f()
			}()
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
					logger.Infof("Garbage collected %d expired sessions", count)
				}
			}
		}
	}()
}

// Creates a new session
func CreateSession(id string, user utils.User, duration time.Duration, privileged bool) {
	sess := Session{
		SessionID:  id,
		User:       user,
		ExpiresAt:  time.Now().Add(duration),
		Privileged: privileged,
	}
	SessionMux <- func() {
		Sessions[id] = sess
	}
	logger.Infof("Created session for user '%s'", user.ID)

}

// Deletes a session
func DeleteSession(id string) {
	SessionMux <- func() {
		sess, exists := Sessions[id]
		if exists {
			delete(Sessions, id)
			logger.Infof("Deleted session for user '%s'", sess.User.ID)
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

// ValidateFromRequest validates the session cookie and returns the session pointer and validity.
func ValidateFromRequest(r *http.Request) (*Session, bool) {
	cookie, err := r.Cookie("session_id")
	if err != nil || cookie.Value == "" {
		return nil, false
	}

	var sess *Session
	var exists bool
	done := make(chan bool)

	SessionMux <- func() {
		s, ok := Sessions[cookie.Value]
		exists = ok
		if ok {
			copy := s // avoid race
			sess = &copy
		}
		done <- true
	}
	<-done

	if !exists {
		logger.Warnf("Access attempt with unknown session_id: %s", cookie.Value)
		return nil, false
	}

	if sess.ExpiresAt.Before(time.Now()) {
		logger.Warnf("Expired session access attempt by user '%s'", sess.User.ID)
		return nil, false
	}

	return sess, true
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

// Get returns a pointer to the Session struct for the given sessionID, or nil if not found.
// WARNING: The returned pointer is to a *copy*; do not modify fields directly!
func Get(id string) *Session {
	done := make(chan *Session)
	SessionMux <- func() {
		sess, exists := Sessions[id]
		if exists {
			s := sess // copy to new var to avoid data race
			done <- &s
		} else {
			done <- nil
		}
	}
	return <-done
}
