package worker

import (
	"context"
	"dos/internal/services/storage/core"
	"log/slog"
	"time"
)

type PeriodicConfig struct {
	Delay time.Duration `yaml:"delay"`
}


type HeartbeatWorker struct {
	config *PeriodicConfig
	service *core.Service 
}

func NewHeartbeatWorker(cfg *PeriodicConfig, svc *core.Service) *HeartbeatWorker {
	return &HeartbeatWorker{
		config: cfg,
		service: svc,
	}
}

func (w *HeartbeatWorker) Run(ctx context.Context) {
	go runPeriodicaly(ctx, w.config, func() {
		slog.DebugContext(ctx, "exec heartbeat")
		if err := w.service.Heartbeat(ctx); err != nil {
			slog.ErrorContext(ctx, "heartbeat failed", "error", err)
		}
	})
}


func runPeriodicaly(ctx context.Context, cfg *PeriodicConfig, job func()) {
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}

		job()

		timer.Reset(cfg.Delay)
	}
}


