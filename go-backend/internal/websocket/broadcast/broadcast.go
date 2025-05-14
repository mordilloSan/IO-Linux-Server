package broadcast

import (
	"context"
	"go-backend/internal/logger"
	"go-backend/internal/system"
	"go-backend/internal/websocket/utils"
	"time"
)

// RunBroadcasterMultiWithIntervalsWithCtx starts multiple goroutines
// to periodically broadcast messages on a channel, cancelable via context.
func RunBroadcasterMultiWithIntervalsWithCtx(ctx context.Context, channel string, funcs map[string]struct {
	Fn       func() (any, error)
	Interval time.Duration
}) {
	for dataType, cfg := range funcs {
		go func(dataType string, cfg struct {
			Fn       func() (any, error)
			Interval time.Duration
		}) {
			ticker := time.NewTicker(cfg.Interval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					logger.Info.Printf("[broadcast] Stopped broadcasting %s updates", dataType)
					return
				case <-ticker.C:
					payload, err := cfg.Fn()
					if err != nil {
						logger.Error.Printf("[broadcast] Error fetching %s: %v", dataType, err)
						continue
					}
					utils.BroadcastToChannel(channel, map[string]any{
						"type":    dataType,
						"channel": channel,
						"payload": payload,
					})
				}
			}
		}(dataType, cfg)
	}
}

func StartDashboardBroadcaster(ctx context.Context) {
	RunBroadcasterMultiWithIntervalsWithCtx(ctx, "dashboard", map[string]struct {
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
