package storagenode

import (
	"context"
	"testing"
	"time"

	t "dos/internal/common/types"

	"github.com/stretchr/testify/require"
)

func TestCleanupWorker_RemoveInactive(tt *testing.T) {
	ctx := context.Background()

	lifecycle := &fakeNodeLifecycle{
		inactive: []t.NodeID{"node-1", "node-2"},
		chunks: map[t.NodeID][]t.ChunkID{
			"node-1": {"chunk-1", "chunk-2"},
			"node-2": {"chunk-3"},
		},
	}
	replication := &fakeReplicaScheduler{}

	worker, err := NewCleanupWorker(CleanupDeps{
		Lifecycle:   lifecycle,
		Replication: replication,
		Config:      fakeCleanupConfig{},
	})
	require.NoError(tt, err)

	got := worker.RemoveInactive(ctx)

	require.Equal(tt, 2, got)
	require.Equal(tt, []t.NodeID{"node-1", "node-2"}, lifecycle.removed)
	require.Equal(tt, []t.ChunkID{"chunk-1", "chunk-2", "chunk-3"}, replication.chunkIDs)
}

// fake lifecycle

type fakeNodeLifecycle struct {
	inactive []t.NodeID
	removed  []t.NodeID
	chunks   map[t.NodeID][]t.ChunkID
}

func (l *fakeNodeLifecycle) GetInactive(context.Context, time.Time) []t.NodeID {
	return l.inactive
}

func (l *fakeNodeLifecycle) Remove(_ context.Context, nodeID t.NodeID) ([]t.ChunkID, error) {
	l.removed = append(l.removed, nodeID)
	return l.chunks[nodeID], nil
}

// fake scheduler

type fakeReplicaScheduler struct {
	chunkIDs []t.ChunkID
}

func (s *fakeReplicaScheduler) Schedule(_ context.Context, chunkID t.ChunkID) {
	s.chunkIDs = append(s.chunkIDs, chunkID)
}

// fake config

type fakeCleanupConfig struct{}

func (fakeCleanupConfig) NodeInactivityTimeout() time.Duration {
	return time.Hour
}

func (fakeCleanupConfig) NodeCleanupInterval() time.Duration {
	return time.Hour
}
