package repo

import (
	"context"
	"testing"

	t "dos/internal/common/types"
	m "dos/internal/services/master"

	"github.com/stretchr/testify/require"
)

func TestInMemObjectRepo_ChunkLifecycle(tt *testing.T) {
	ctx := context.Background()
	repo := NewInMemObjectRepo()
	objectID := t.ObjectID("object-1")
	slot := t.ObjectSlot{ObjectID: objectID, ChunkKey: "chunk-key-1"}

	require.NoError(tt, repo.Create(ctx, objectID, 2))

	// just created object is empty
	exists, err := repo.ExistsChunk(ctx, slot)
	require.NoError(tt, err)
	require.False(tt, exists)

	// add chunk at slot
	require.NoError(tt, repo.AddChunk(ctx, slot, "chunk-1"))
	exists, err = repo.ExistsChunk(ctx, slot)
	require.NoError(tt, err)
	require.True(tt, exists)

	// chunk at slot is available
	got, err := repo.GetChunk(ctx, slot)
	require.NoError(tt, err)
	require.Equal(tt, t.ChunkID("chunk-1"), got)

	// cannot add another chunk at the same slot
	err = repo.AddChunk(ctx, slot, "chunk-2")
	require.ErrorIs(tt, err, m.ErrChunkKeyExists)

	// delete chunk
	require.NoError(tt, repo.DeleteChunk(ctx, slot))
	_, err = repo.GetChunk(ctx, slot)
	require.ErrorIs(tt, err, m.ErrChunkKeyNotFound)
}

func TestInMemObjectRepo_Delete(tt *testing.T) {
	ctx := context.Background()
	repo := NewInMemObjectRepo()
	objectID := t.ObjectID("object-1")
	slot := t.ObjectSlot{ObjectID: objectID, ChunkKey: "chunk-key-1"}

	require.NoError(tt, repo.Create(ctx, objectID, 2))
	require.NoError(tt, repo.AddChunk(ctx, slot, "chunk-1"))

	// cannot delete non-empty object
	err := repo.Delete(ctx, objectID)
	require.ErrorIs(tt, err, m.ErrObjectNotEmpty)

	// failed deletion doesn't affact object
	got, err := repo.Get(ctx, objectID)
	require.NoError(tt, err)
	require.Equal(tt, objectID, got.ID)

	// delete chunk, delete object
	require.NoError(tt, repo.DeleteChunk(ctx, slot))
	require.NoError(tt, repo.Delete(ctx, objectID))

	// object is not accessable after deleteion
	_, err = repo.Get(ctx, objectID)
	require.ErrorIs(tt, err, m.ErrObjectNotFound)
}
