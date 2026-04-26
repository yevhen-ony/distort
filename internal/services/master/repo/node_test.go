package repo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	m "dos/internal/services/master"
	t "dos/internal/common/types"
)

func TestInMemNodeRegistry_Register(test *testing.T) {
	r := NewInMemNodeRegistry()
	ctx := context.Background()
	addr := "127.0.0.1:9001"
	report := t.NodeStats{Addr: addr, FreeBytes: 100}

	test.Run("Success", func(test *testing.T) {
		nid, err := r.Register(ctx, &report)
		require.NoError(test, err)
		require.NotEmpty(test, nid)
	})

	test.Run("DuplicateAddr", func(test *testing.T) {
		_, err := r.Register(ctx, &t.NodeStats{Addr: addr})
		require.ErrorIs(test, err, m.ErrNodeAddrInUse)
	})
}

func TestInMemNodeRegistry_GetChunkNodes(test *testing.T) {
	r := NewInMemNodeRegistry()
	ctx := context.Background()
	addr := "127.0.0.1:9001"
	report := t.NodeStats{Addr: addr, FreeBytes: 100}

	nid, err := r.Register(ctx, &report)
	require.NoError(test, err)
	
	cid := t.ChunkID("chunk-x")
	require.NoError(test, r.AttachChunk(ctx, nid, cid))
	
	test.Run("Success", func(test *testing.T) {
		nodes, err := r.GetChunkNodes(ctx, cid)
		require.NoError(test, err)
		require.Len(test, nodes, 1)

		node := nodes[0]
		require.Equal(test, nid, node.ID)
	})
}

func TestInMemNodeRegistry_Unregister(test *testing.T) {
	r := NewInMemNodeRegistry()
	ctx := context.Background()
	addr := "127.0.0.1:9001"
	report := t.NodeStats{Addr: addr, FreeBytes: 100}

	nid, err := r.Register(ctx, &report)
	require.NoError(test, err)
	
	cid := t.ChunkID("chunk-x")
	require.NoError(test, r.AttachChunk(ctx, nid, cid))

	test.Run("Success", func(test *testing.T) {
		require.NoError(test, r.Unregister(ctx, nid))

		err := r.AttachChunk(ctx, nid, t.ChunkID("chunk-y"))
		require.ErrorIs(test, err, m.ErrNodeNotFound)

		_, err = r.GetNodeChunks(ctx, nid)
		require.ErrorIs(test, err, m.ErrNodeNotFound) 

		nodes, err := r.GetChunkNodes(ctx, cid)
		assert.Empty(test, nodes)
		assert.NoError(test, err)
	})
}

func TestInMemNodeRegistry_GetCandidateNodes(test *testing.T) {
	r := NewInMemNodeRegistry()
	ctx := context.Background()

	n1, err := r.Register(ctx, &t.NodeStats{Addr: "127.0.0.1:9001", FreeBytes: 100})
	require.NoError(test, err)
	_, err = r.Register(ctx, &t.NodeStats{Addr: "127.0.0.1:9002", FreeBytes: 50})
	require.NoError(test, err)
	_, err = r.Register(ctx, &t.NodeStats{Addr: "127.0.0.1:9003", FreeBytes: 10})
	require.NoError(test, err)

	// to test 'ExcludeChunk'
	cid := t.ChunkID("chunk-x")
	require.NoError(test, r.AttachChunk(ctx, n1, cid))

	nodes, err := r.GetCandidateNodes(ctx, &m.CandidateNodesQuery{
		MinFreeBytes: 40,
		ExcludeChunk: cid,
	})
	require.NoError(test, err)

	assert.Len(test, nodes, 1)
	assert.Equal(test, int64(50), nodes[0].Stats.FreeBytes)
}

