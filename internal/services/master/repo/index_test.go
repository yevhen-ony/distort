package repo

import (
	"context"
	"testing"

	t "dos/internal/common/types"

	"github.com/stretchr/testify/require"
)

func TestInMemChunkNodeIndex_AttachChunk(tt *testing.T) {
	ctx := context.Background()
	index := NewInMemChunkNodeIndex()

	attached := index.AttachChunk(ctx, "node-1", "chunk-1")
	require.True(tt, attached)

	require.ElementsMatch(tt, []t.ChunkID{"chunk-1"}, index.GetNodeChunks(ctx, "node-1"))
	require.ElementsMatch(tt, []t.NodeID{"node-1"}, index.GetChunkNodes(ctx, "chunk-1"))

	// attaching the same chunk rejected
	attached = index.AttachChunk(ctx, "node-1", "chunk-1")
	require.False(tt, attached)

	require.ElementsMatch(tt, []t.ChunkID{"chunk-1"}, index.GetNodeChunks(ctx, "node-1"))
	require.ElementsMatch(tt, []t.NodeID{"node-1"}, index.GetChunkNodes(ctx, "chunk-1"))
}

func TestInMemChunkNodeIndex_DetachChunk(tt *testing.T) {
	ctx := context.Background()
	index := NewInMemChunkNodeIndex()

	require.True(tt, index.AttachChunk(ctx, "node-1", "chunk-1"))
	require.True(tt, index.AttachChunk(ctx, "node-1", "chunk-2"))
	require.True(tt, index.AttachChunk(ctx, "node-2", "chunk-1"))

	// detaches chunk from first node only
	detached := index.DetachChunk(ctx, "node-1", "chunk-1")
	require.True(tt, detached)
	
	require.ElementsMatch(tt, []t.ChunkID{"chunk-2"}, index.GetNodeChunks(ctx, "node-1"))
	require.ElementsMatch(tt, []t.NodeID{"node-2"}, index.GetChunkNodes(ctx, "chunk-1"))

	// detaches missing chunk from first node
	detached = index.DetachChunk(ctx, "node-1", "chunk-1")
	require.False(tt, detached)
}

func TestInMemChunkNodeIndex_DetachNode(tt *testing.T) {
	ctx := context.Background()
	index := NewInMemChunkNodeIndex()

	require.True(tt, index.AttachChunk(ctx, "node-1", "chunk-1"))
	require.True(tt, index.AttachChunk(ctx, "node-1", "chunk-2"))
	require.True(tt, index.AttachChunk(ctx, "node-2", "chunk-1"))

	index.DetachNode(ctx, "node-1")

	require.Empty(tt, index.GetNodeChunks(ctx, "node-1"))
	require.ElementsMatch(tt, []t.NodeID{"node-2"}, index.GetChunkNodes(ctx, "chunk-1"))
	require.Empty(tt, index.GetChunkNodes(ctx, "chunk-2"))
}
