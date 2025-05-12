package channels

import (
	"time"

	"go-backend/internal/system"
	"go-backend/internal/websocket"
)

func StartDashboardBroadcaster() {
	websocket.RunBroadcasterMultiWithIntervals("dashboard", map[string]struct {
		Fn       func() (any, error)
		Interval time.Duration
	}{
		"cpu": {
			Fn: func() (any, error) {
				return system.FetchCPUInfo()
			},
			Interval: 1 * time.Second,
		},
		"memory": {
			Fn: func() (any, error) {
				return system.FetchMemoryInfo()
			},
			Interval: 5 * time.Second,
		},
		"filesystem": {
			Fn: func() (any, error) {
				return system.FetchFileSystemInfo()
			},
			Interval: 10 * time.Second,
		},
		"baseboard": {
			Fn: func() (any, error) {
				return system.FetchBaseboardInfo()
			},
			Interval: 5 * time.Second,
		},
		"network": {
			Fn: func() (any, error) {
				return system.FetchNetworkInfo()
			},
			Interval: 1 * time.Second,
		},
		"sensors": {
			Fn: func() (any, error) {
				return system.FetchSensorsInfo(), nil
			},
			Interval: 1 * time.Second,
		},
	})
}
