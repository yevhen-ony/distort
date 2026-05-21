package storagenode

import (
	"context"
	t "dos/internal/common/types"
	m "dos/internal/services/master"
	"errors"
	"log/slog"
	"time"
)

type CleanupConfig interface {
	NodeInactivityTimeout() time.Duration
	NodeCleanupInterval() time.Duration
}

type CleanupDeps struct {
	Lifecycle   *LifecycleService
	Replication m.ReplicaScheduler
	Config      CleanupConfig
}

type CleanupWorker struct {
	lifecycle *LifecycleService
	replicate m.ReplicaScheduler

	config CleanupConfig
}

func NewCleanupWorker(deps CleanupDeps) (*CleanupWorker, error) {
	if deps.Lifecycle == nil {
		return nil, errors.New("missing lifecycle service")
	}
	if deps.Replication == nil {
		return nil, errors.New("missing replication scheduler")
	}
	if deps.Config == nil {
		return nil, errors.New("missing config")
	}

	service := &CleanupWorker{
		lifecycle: deps.Lifecycle,
		replicate: deps.Replication,
		config:    deps.Config,
	}
	return service, nil
}

func (s *CleanupWorker) RemoveInactive(ctx context.Context) int {
	cutoff := time.Now().Add(-s.config.NodeInactivityTimeout())
	nodeIDs := s.lifecycle.GetInactive(ctx, cutoff)

	affectedNodeIDs := make([]t.NodeID, 0, len(nodeIDs))
	for _, nodeID := range nodeIDs {
		chunkIDs, err := s.lifecycle.Remove(ctx, nodeID)
		if err != nil {
			slog.ErrorContext(ctx, "remove node failed", "node_id", nodeID, "error", err)
			continue
		}
		for _, chunkID := range chunkIDs {
			s.replicate.Schedule(ctx, chunkID)
		}
		affectedNodeIDs = append(affectedNodeIDs, nodeID)
	}
	return len(affectedNodeIDs)
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

		count := s.RemoveInactive(ctx)
		slog.DebugContext(ctx, "removed inactive nodes", "count", count)

		timer.Reset(s.config.NodeCleanupInterval())
	}
}
