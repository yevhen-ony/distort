package inventory

import (
	"dos/internal/common/digest"
	"dos/internal/common/metrics"
	t "dos/internal/common/types"
	s "dos/internal/services/storage"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChunkInventory_Add(tt *testing.T) {
	f := newInventoryFixture(tt)
	inventory, err := NewChunkInventory(f.deps())
	require.NoError(tt, err)

	meta := testChunkMeta("chunk-1", 123)
	
	// add chunk
	err = inventory.Add(&meta)
	require.NoError(tt, err)
	require.True(tt, inventory.Has("chunk-1"))

	// give expected stats
	expectedStats := t.NodeStats{
		FreeBytes:  877,
		UsedBytes:  123,
		ChunkCount: 1,
	}
	require.Equal(tt, expectedStats, inventory.GetStats())

	// reject duplicates
	err = inventory.Add(&meta)
	require.ErrorIs(tt, err, s.ErrChunkConflict)
}

func TestChunkInventory_StageActivate(tt *testing.T) {
	f := newInventoryFixture(tt)
	inventory, err := NewChunkInventory(f.deps())
	require.NoError(tt, err)

	fst := testChunkMeta("chunk-1", 123)
	snd := testChunkMeta("chunk-2", 456)

	require.NoError(tt, inventory.Add(&fst))
	require.NoError(tt, inventory.Add(&snd))

	// promote fst to 'active'
	meta, err := inventory.Activate("chunk-1")
	require.NoError(tt, err)
	require.Equal(tt, fst, meta)

	state, err := inventory.GetState("chunk-1")
	require.NoError(tt, err)
	require.Equal(tt, s.ChunkStateActive, state)

	require.ElementsMatch(tt, []t.ChunkMeta{snd}, inventory.ListStaged())
	
	// demote fst back to 'staged'
	meta, err = inventory.Stage("chunk-1")
	require.NoError(tt, err)
	require.Equal(tt, fst, meta)

	require.ElementsMatch(tt, []t.ChunkMeta{fst, snd}, inventory.ListStaged())
}

func TestChunkInventory_Remove(tt *testing.T) {
	f := newInventoryFixture(tt)
	inventory, err := NewChunkInventory(f.deps())
	require.NoError(tt, err)

	fst := testChunkMeta("chunk-1", 123)
	snd := testChunkMeta("chunk-2", 456)

	require.NoError(tt, inventory.Add(&fst))
	require.NoError(tt, inventory.Add(&snd))

	require.True(tt, inventory.Remove("chunk-1"))
	require.False(tt, inventory.Remove("chunk-1"))

	require.False(tt, inventory.Has("chunk-1"))
	require.True(tt, inventory.Has("chunk-2"))

	// only scd chunk left
	require.Equal(tt, t.NodeStats{
		FreeBytes:  544,
		UsedBytes:  456,
		ChunkCount: 1,
	}, inventory.GetStats())
}

func TestChunkInventory_ListIDs(tt *testing.T) {
	f := newInventoryFixture(tt)
	inventory, err := NewChunkInventory(f.deps())
	require.NoError(tt, err)

	fst := testChunkMeta("chunk-1", 123)
	snd := testChunkMeta("chunk-2", 456)

	require.NoError(tt, inventory.Add(&fst))
	require.NoError(tt, inventory.Add(&snd))

	expectedIDs := []t.ChunkID{"chunk-1", "chunk-2"}
	require.ElementsMatch(tt, expectedIDs, inventory.ListIDs())

	require.True(tt, inventory.Remove("chunk-1"))
	require.ElementsMatch(tt, []t.ChunkID{"chunk-2"}, inventory.ListIDs())
}

// fixture

type inventoryFixture struct {
	config  fakeChunkCatalogConfig
	metrics *ChunkInventoryMetrics
}

func newInventoryFixture(tt *testing.T) *inventoryFixture {
	tt.Helper()

	return &inventoryFixture{
		config:  fakeChunkCatalogConfig{maxStorageBytes: 1000},
		metrics: NewChunkInventoryMetrics(metrics.NopProvider{}),
	}
}

func (f *inventoryFixture) deps() ChunkInventoryDeps {
	return ChunkInventoryDeps{
		Config:  f.config,
		Metrics: f.metrics,
	}
}

// fakes config

type fakeChunkCatalogConfig struct{ maxStorageBytes int64 }

func (c fakeChunkCatalogConfig) MaxStorageBytes() int64 { return c.maxStorageBytes }

// test data gen

func testChunkMeta(id t.ChunkID, size int64) t.ChunkMeta {
	return t.ChunkMeta{
		ID: id,
		Digest: digest.Digest{
			Checksum: digest.Checksum("checksum-" + string(id)),
			Size:     size,
		},
	}
}
