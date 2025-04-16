package main

import (
	"time"
)

type Session struct {
	User      User
	ExpiresAt time.Time
}

var (
	sessions   = make(map[string]Session)
	sessionMux = make(chan func())
)

func init() {
	go func() {
		for f := range sessionMux {
			f()
		}
	}()
}

func startSessionGC() {
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			now := time.Now()
			sessionMux <- func() {
				for id, s := range sessions {
					if s.ExpiresAt.Before(now) {
						delete(sessions, id)
					}
				}
			}
		}
	}()
}
