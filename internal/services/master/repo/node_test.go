package repo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	m "dos/internal/services/master"
)

func TestInMemNodeRegistry_Register(t *testing.T) {
	r := NewInMemNodeRegistry()
	ctx := context.Background()
	addr := "127.0.0.1:9001"
	report := m.NodeReport{Addr: addr, FreeBytes: 100}

	t.Run("Success", func(t *testing.T) {
		nid, err := r.Register(ctx, &report)
		require.NoError(t, err)
		require.NotEmpty(t, nid)
	})

	t.Run("DuplicateAddr", func(t *testing.T) {
		_, err := r.Register(ctx, &m.NodeReport{Addr: addr})
		require.ErrorIs(t, err, m.ErrNodeAddrInUse)
	})
}

func TestInMemNodeRegistry_GetChunkNodes(t *testing.T) {
	r := NewInMemNodeRegistry()
	ctx := context.Background()
	addr := "127.0.0.1:9001"
	report := m.NodeReport{Addr: addr, FreeBytes: 100}

	nid, err := r.Register(ctx, &report)
	require.NoError(t, err)
	
	cid := m.ChunkID("chunk-x")
	require.NoError(t, r.AttachChunk(ctx, nid, cid))
	
	t.Run("Success", func(t *testing.T) {
		nodes, err := r.GetChunkNodes(ctx, cid)
		require.NoError(t, err)
		require.Len(t, nodes, 1)

		node := nodes[0]
		require.Equal(t, nid, node.ID)
	})
}

func TestInMemNodeRegistry_Unregister(t *testing.T) {
	r := NewInMemNodeRegistry()
	ctx := context.Background()
	addr := "127.0.0.1:9001"
	report := m.NodeReport{Addr: addr, FreeBytes: 100}

	nid, err := r.Register(ctx, &report)
	require.NoError(t, err)
	
	cid := m.ChunkID("chunk-x")
	require.NoError(t, r.AttachChunk(ctx, nid, cid))

	t.Run("Success", func(t *testing.T) {
		require.NoError(t, r.Unregister(ctx, nid))

		err := r.AttachChunk(ctx, nid, m.ChunkID("chunk-y"))
		require.ErrorIs(t, err, m.ErrNodeNotFound)

		_, err = r.GetNodeChunks(ctx, nid)
		require.ErrorIs(t, err, m.ErrNodeNotFound) 

		nodes, err := r.GetChunkNodes(ctx, cid)
		assert.Empty(t, nodes)
		assert.NoError(t, err)
	})
}

func TestInMemNodeRegistry_GetCandidateNodes(t *testing.T) {
	r := NewInMemNodeRegistry()
	ctx := context.Background()

	n1, err := r.Register(ctx, &m.NodeReport{Addr: "127.0.0.1:9001", FreeBytes: 100})
	require.NoError(t, err)
	_, err = r.Register(ctx, &m.NodeReport{Addr: "127.0.0.1:9002", FreeBytes: 50})
	require.NoError(t, err)
	_, err = r.Register(ctx, &m.NodeReport{Addr: "127.0.0.1:9003", FreeBytes: 10})
	require.NoError(t, err)

	// to test 'ExcludeChunk'
	cid := m.ChunkID("chunk-x")
	require.NoError(t, r.AttachChunk(ctx, n1, cid))

	nodes, err := r.GetCandidateNodes(ctx, &m.CandidateNodesQuery{
		MinFreeBytes: 40,
		ExcludeChunk: cid,
	})
	require.NoError(t, err)

	assert.Len(t, nodes, 1)
	assert.Equal(t, int64(50), nodes[0].Report.FreeBytes)
}

