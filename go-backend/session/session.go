// session/session.go
package session

import (
	"go-backend/utils"
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
				for id, s := range Sessions {
					if s.ExpiresAt.Before(now) {
						delete(Sessions, id)
					}
				}
			}
		}
	}()
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
	if !exists || session.ExpiresAt.Before(time.Now()) {
		return utils.User{}, "", false
	}
	return session.User, cookie.Value, true
}
