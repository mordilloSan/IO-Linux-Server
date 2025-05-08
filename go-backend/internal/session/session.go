// session/session.go
package session

import (
	"go-backend/internal/logger"
	"go-backend/internal/utils"
	"net/http"
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
					logger.Info.Printf("[session] Garbage collected %d expired sessions", count)
				}
			}
		}
	}()
}

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
