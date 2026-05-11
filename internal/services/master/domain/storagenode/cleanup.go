package storagenode

import (
	"context"
	"log/slog"
	"time"
)

type CleanupConfig interface {
	NodeInactivityTimeout() time.Duration
	NodeCleanupInterval() time.Duration
}

type CleanupWorker struct {
	lifecycle *LifecycleService
	config    CleanupConfig
}

func NewCleanupWorker(
	lifecycle *LifecycleService,
	config CleanupConfig,
) *CleanupWorker {

	return &CleanupWorker{
		lifecycle: lifecycle,
		config:    config,
	}
}

func (s *CleanupWorker) RunLoop(ctx context.Context) {

	timer := time.NewTimer(s.config.NodeCleanupInterval())
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}

		cutoff := time.Now().Add(-s.config.NodeInactivityTimeout())
		count, err := s.lifecycle.RemoveInactive(ctx, cutoff)

		slog.DebugContext(ctx, "removed inactive nodes", "count", count)
		if err != nil {
			slog.ErrorContext(ctx, "cleanup inactive nodes", "error", err)
		}

		timer.Reset(s.config.NodeCleanupInterval())
	}
}
