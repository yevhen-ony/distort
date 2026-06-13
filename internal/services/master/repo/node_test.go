package repo

import (
	"context"
	"testing"

	t "dos/internal/common/types"
	"dos/internal/common/utils"
	m "dos/internal/services/master"

	"github.com/stretchr/testify/require"
)


func TestInMemNodeRegistry_RegisterUnregister(tt *testing.T) {
  	ctx := context.Background()
  	registry := NewInMemNodeRegistry()

  	first, err := registry.Register(ctx, "node-1:10001")
  	require.NoError(tt, err)

	// try register with the same addr
	_, err = registry.Register(ctx, "node-1:10001")
	require.ErrorIs(tt, err, m.ErrNodeAddrInUse)

  	registry.Unregister(ctx, first.ID)

  	_, err = registry.Get(ctx, first.ID)
  	require.ErrorIs(tt, err, m.ErrNodeNotFound)

	// register after unregister
  	second, err := registry.Register(ctx, "node-1:10001")
  	require.NoError(tt, err)
  	require.Equal(tt, "node-1:10001", second.Addr)
}

func TestInMemNodeRegistry_Find(tt *testing.T) {
  	ctx := context.Background()
  	registry := NewInMemNodeRegistry()

  	n1, err := registry.Register(ctx, "node-1:10001")
  	require.NoError(tt, err)
  	n2, err := registry.Register(ctx, "node-2:10001")
  	require.NoError(tt, err)
  	n3, err := registry.Register(ctx, "node-3:10001")
  	require.NoError(tt, err)

	err = registry.UpdateStats(ctx, n1.ID, t.NodeStats{FreeBytes: 100})
  	require.NoError(tt, err)
	err = registry.UpdateStats(ctx, n2.ID, t.NodeStats{FreeBytes: 200})
  	require.NoError(tt, err)
	err = registry.UpdateStats(ctx, n3.ID, t.NodeStats{FreeBytes: 300})
  	require.NoError(tt, err)

	tt.Run("with_min_free_bytes", func(tt *testing.T) {
  		nodes := registry.Find(ctx, m.NodeQuery{
			MinFreeBytes: 200,
		})
		got := utils.Map(nodes, func(n m.Node) t.NodeID {return n.ID})
		require.ElementsMatch(tt, []t.NodeID{n2.ID, n3.ID}, got)
	})

	tt.Run("with_excluded_ids",  func(tt *testing.T)  {
		nodes := registry.Find(ctx, m.NodeQuery{
			ExcludeIDs: []t.NodeID{n3.ID},
		})
		got := utils.Map(nodes, func(n m.Node) t.NodeID {return n.ID})
		require.ElementsMatch(tt, []t.NodeID{n1.ID, n2.ID}, got)
	})

	tt.Run("combined_constraints", func (tt *testing.T) {
  		nodes := registry.Find(ctx, m.NodeQuery{
			MinFreeBytes: 200,
			ExcludeIDs: []t.NodeID{n3.ID},
		})
		require.Len(tt, nodes, 1)
		require.Equal(tt, n2.ID, nodes[0].ID)
	})
}
