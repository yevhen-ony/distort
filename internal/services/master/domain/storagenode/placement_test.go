package storagenode

import (
	"context"
	"testing"

	t "dos/internal/common/types"
	m "dos/internal/services/master"
	"dos/internal/services/master/repo"

	"github.com/stretchr/testify/require"
)

func TestPlacementService_GetCandidates(tt *testing.T) {
	ctx := context.Background()
	f := newPlacementFixture()

	node1, err := f.nodes.Register(ctx, "node-1:10001")
	require.NoError(tt, err)
	node2, err := f.nodes.Register(ctx, "node-2:10001")
	require.NoError(tt, err)
	node3, err := f.nodes.Register(ctx, "node-3:10001")
	require.NoError(tt, err)

	require.NoError(tt, f.nodes.UpdateStats(ctx, node1.ID, t.NodeStats{FreeBytes: 100}))
	require.NoError(tt, f.nodes.UpdateStats(ctx, node2.ID, t.NodeStats{FreeBytes: 100}))
	require.NoError(tt, f.nodes.UpdateStats(ctx, node3.ID, t.NodeStats{FreeBytes: 40}))

	require.True(tt, f.index.AttachChunk(ctx, node2.ID, "chunk-1"))

	service, err := NewPlacementService(f.deps())
	require.NoError(tt, err)

	tt.Run("no_conditions", func(tt *testing.T) {
		got, err := service.GetCandidates(ctx, m.CandidateNodesQuery{})
		require.NoError(tt, err)
		require.Len(tt, got, 3)
	})

	tt.Run("min_free_bytes", func(tt *testing.T) {
		got, err := service.GetCandidates(ctx, m.CandidateNodesQuery{
			MinFreeBytes: 50,
		})
		require.NoError(tt, err)
		require.Len(tt, got, 2)
		require.NotContains(tt, got, node3)
	})

	tt.Run("exclude_node", func(tt *testing.T) {
		got, err := service.GetCandidates(ctx, m.CandidateNodesQuery{
			ExcludeNodes: []t.NodeRef{node1},
		})
		require.NoError(tt, err)
		require.Len(tt, got, 2)
		require.NotContains(tt, got, node1)

	})

	tt.Run("exclude_chunk", func(tt *testing.T) {
		got, err := service.GetCandidates(ctx, m.CandidateNodesQuery{
			ExcludeChunk: t.ChunkID("chunk-1"),
		})
		require.NoError(tt, err)
		require.Len(tt, got, 2)
		require.NotContains(tt, got, node2)
	})

	tt.Run("combined_conditions", func(tt *testing.T) {
		got, err := service.GetCandidates(ctx, m.CandidateNodesQuery{
			MinFreeBytes: 50,
			ExcludeChunk: t.ChunkID("chunk-1"),
			ExcludeNodes: []t.NodeRef{node1},
		})
		require.NoError(tt, err)
		require.Empty(tt, got)
	})
}

func TestPlacementService_GetChunkNodes(tt *testing.T) {
	ctx := context.Background()
	f := newPlacementFixture()

	node1, err := f.nodes.Register(ctx, "node-1:10001")
	require.NoError(tt, err)
	node2, err := f.nodes.Register(ctx, "node-2:10001")
	require.NoError(tt, err)

	require.True(tt, f.index.AttachChunk(ctx, node1.ID, "chunk-1"))
	require.True(tt, f.index.AttachChunk(ctx, node2.ID, "chunk-1"))

	service, err := NewPlacementService(f.deps())
	require.NoError(tt, err)

	tt.Run("missing_chunk", func(tt *testing.T) {
		got, err := service.GetChunkNodes(ctx, "missing")
		require.NoError(tt, err)
		require.Nil(tt, got)
	})

	tt.Run("success", func(tt *testing.T) {
		got, err := service.GetChunkNodes(ctx, "chunk-1")
		require.NoError(tt, err)
		require.ElementsMatch(tt, []t.NodeRef{node1, node2}, got)
	})
}

// fixture

type placementFixture struct {
	nodes *repo.InMemNodeRegistry
	index *repo.InMemChunkNodeIndex
}

func newPlacementFixture() *placementFixture {
	return &placementFixture{
		nodes: repo.NewInMemNodeRegistry(),
		index: repo.NewInMemChunkNodeIndex(),
	}
}

func (f *placementFixture) deps() PlacementDeps {
	return PlacementDeps{
		ChunkNodeIndex: f.index,
		NodeRegistry:   f.nodes,
		Config:         fakePlacementConfig{margin: 10},
	}
}

// fake config

type fakePlacementConfig struct {
	margin int64
}

func (c fakePlacementConfig) ChunkAllocationMarginBytes() int64 {
	return c.margin
}
