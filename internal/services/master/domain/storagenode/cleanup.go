package storagenode

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"dos/internal/common/loop"
	t "dos/internal/common/types"
)

type CleanupConfig interface {
	NodeInactivityTimeout() time.Duration
	NodeCleanupInterval() time.Duration
}

type ReplicaScheduler interface {
	Schedule(context.Context, t.ChunkID)
}

type NodeLifecycle interface {
	GetInactive(context.Context, time.Time) []t.NodeID
	Remove(context.Context, t.NodeID) ([]t.ChunkID, error)
}

type CleanupDeps struct {
	Lifecycle   NodeLifecycle
	Replication ReplicaScheduler
	Config      CleanupConfig
}

type CleanupWorker struct {
	lifecycle NodeLifecycle
	replicate ReplicaScheduler

	config CleanupConfig
	looper *loop.Looper
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

	looper := loop.NewLooper(deps.Config.NodeCleanupInterval())
	service := &CleanupWorker{
		lifecycle: deps.Lifecycle,
		replicate: deps.Replication,
		config:    deps.Config,

		looper: looper,
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

	go s.looper.Run(ctx, func(ctx context.Context) {
		count := s.RemoveInactive(ctx)
		slog.DebugContext(ctx, "removed inactive nodes", "count", count)
	})
}
