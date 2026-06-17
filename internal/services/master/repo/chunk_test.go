package repo

import (
	"context"
	"testing"

	"dos/internal/common/digest"
	t "dos/internal/common/types"
	m "dos/internal/services/master"

	"github.com/stretchr/testify/require"
)

func TestInMemChunkRepo_CreateGet(tt *testing.T) {
	ctx := context.Background()
	repo := NewInMemChunkRepo()
	chunkID := t.ChunkID("chunk-1")
	slot1 := t.ObjectSlot{ObjectID: "object-1", ChunkKey: "chunk-key-1"}
	slot2 := t.ObjectSlot{ObjectID: "object-1", ChunkKey: "chunk-key-2"}

	// create new chunk
	err := repo.Create(ctx, chunkID, slot1)
	require.NoError(tt, err, "new chunk")

	// try create with the same id
	err = repo.Create(ctx, chunkID, slot2)
	require.ErrorIs(tt, err, m.ErrChunkExists, "duplicated")

	// get chunk
	chunk, err := repo.Get(ctx, chunkID)
	require.NoError(tt, err, "accessable")

	require.Equal(tt, chunkID, chunk.Meta.ID)
	require.Equal(tt, int64(0), chunk.Meta.Digest.Size)
	require.Empty(tt, chunk.Meta.Digest.Checksum)
	require.Equal(tt, slot1, chunk.Slot)

	// access non existing
	_, err = repo.Get(ctx, t.ChunkID("missing"))
	require.ErrorIs(tt, err, m.ErrChunkNotFound, "missing")
}

func TestInMemChunkRepo_SetDigest(tt *testing.T) {
	ctx := context.Background()
	repo := NewInMemChunkRepo()
	chunkID := t.ChunkID("chunk-1")
	slot := t.ObjectSlot{ObjectID: "object-1", ChunkKey: "chunk-key-1"}

	err := repo.Create(ctx, chunkID, slot)
	require.NoError(tt, err)

	// set digest
	first := digest.Digest{Checksum: "checksum-1", Size: 123}
	err = repo.SetDigest(ctx, chunkID, first)
	require.NoError(tt, err)

	// increment replica count
	err = repo.IncReplicaCount(ctx, chunkID)
	require.NoError(tt, err)

	// noop if digest correct
	err = repo.SetDigest(ctx, chunkID, first)
	require.NoError(tt, err)

	// try set conflicting digest
	conflicting := digest.Digest{Checksum: "checksum-2", Size: 123}
	err = repo.SetDigest(ctx, chunkID, conflicting)
	require.ErrorIs(tt, err, digest.ErrDigestMismatch)

	// digest untouched
	got, err := repo.GetDigest(ctx, chunkID)
	require.NoError(tt, err)
	require.Equal(tt, first, got)
}

func TestInMemChunkRepo_DecReplicaCount(tt *testing.T) {
	ctx := context.Background()
	repo := NewInMemChunkRepo()
	chunkID := t.ChunkID("chunk-1")
	slot := t.ObjectSlot{ObjectID: "object-1", ChunkKey: "chunk-key-1"}

	err := repo.Create(ctx, chunkID, slot)
	require.NoError(tt, err)

	err = repo.DecReplicaCount(ctx, chunkID)
	require.ErrorIs(tt, err, m.ErrChunkReplicaUnderflow)

	err = repo.IncReplicaCount(ctx, chunkID)
	require.NoError(tt, err)

	err = repo.DecReplicaCount(ctx, chunkID)
	require.NoError(tt, err)

	got, err := repo.Get(ctx, chunkID)
	require.NoError(tt, err)
	require.Equal(tt, 0, got.ReplicaCount)
}

func TestInMemChunkRepo_Delete(tt *testing.T) {
	ctx := context.Background()
	repo := NewInMemChunkRepo()
	chunkID := t.ChunkID("chunk-1")
	slot := t.ObjectSlot{ObjectID: "object-1", ChunkKey: "chunk-key-1"}

	err := repo.Create(ctx, chunkID, slot)
	require.NoError(tt, err)

	// inc replica count
	err = repo.IncReplicaCount(ctx, chunkID)
	require.NoError(tt, err)

	// faile to delete
	deleted, err := repo.Delete(ctx, chunkID)
	require.Error(tt, err)
	require.False(tt, deleted)

	// still accessable
	got, err := repo.Get(ctx, chunkID)
	require.NoError(tt, err)
	require.Equal(tt, 1, got.ReplicaCount)

	// dec replica count
	err = repo.DecReplicaCount(ctx, chunkID)
	require.NoError(tt, err)

	// dec is applied
	got, err = repo.Get(ctx, chunkID)
	require.NoError(tt, err)
	require.Zero(tt, got.ReplicaCount)

	// delete
	deleted, err = repo.Delete(ctx, chunkID)
	require.NoError(tt, err)
	require.True(tt, deleted)

	// delete once again
	deleted, err = repo.Delete(ctx, chunkID)
	require.NoError(tt, err)
	require.False(tt, deleted)
}
