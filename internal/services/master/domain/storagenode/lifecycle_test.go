package storagenode

import (
	"context"
	"testing"

	"dos/internal/common/metrics"
	t "dos/internal/common/types"
	"dos/internal/services/master/repo"
	m "dos/internal/services/master"

	"github.com/stretchr/testify/require"
)

func TestLifecycleService_HappyPath(tt *testing.T) {
	ctx := context.Background()
	f := newLifecycleFixture(tt)

	node, err := f.nodes.Register(ctx, "node-1:1234")
	require.NoError(tt, err)

	slot := t.ObjectSlot{
		ObjectID: "object-1",
		ChunkKey: "chunk-key-1",
	}

	require.NoError(tt, f.chunks.Create(ctx, "chunk-1", slot))
	require.NoError(tt, f.chunks.IncReplicaCount(ctx, "chunk-1"))
	require.True(tt, f.index.AttachChunk(ctx, node.ID, "chunk-1"))

	got, err := f.service(tt).Remove(ctx, node.ID)

	require.NoError(tt, err)
	require.Equal(tt, []t.ChunkID{"chunk-1"}, got)

	_, err = f.nodes.Get(ctx, node.ID)
	require.ErrorIs(tt, err, m.ErrNodeNotFound)

	require.Empty(tt, f.index.GetNodeChunks(ctx, node.ID))

	chunk, err := f.chunks.Get(ctx, "chunk-1")
	require.NoError(tt, err)
	require.Equal(tt, 0, chunk.ReplicaCount)
}

// fixture 
type lifecycleFixture struct {
	nodes  *repo.InMemNodeRegistry
	chunks *repo.InMemChunkRepo
	index  *repo.InMemChunkNodeIndex
}

func newLifecycleFixture(tt *testing.T) *lifecycleFixture {
	tt.Helper()

	return &lifecycleFixture{
		nodes:  repo.NewInMemNodeRegistry(),
		chunks: repo.NewInMemChunkRepo(),
		index:  repo.NewInMemChunkNodeIndex(),
	}
}

func (f *lifecycleFixture) service(tt *testing.T) *LifecycleService {
	tt.Helper()

	service, err := NewLifecycleService(LifecycleDeps{
		NodeRegistry:    f.nodes,
		ChunkRepository: f.chunks,
		ChunkNodeIndex:  f.index,
		Metrics:         NewLifecycleMetrics(metrics.NopProvider{}),
	})
	require.NoError(tt, err)
	return service
}
