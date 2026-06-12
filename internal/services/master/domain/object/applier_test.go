package object

import (
	"context"
	"testing"

	t "dos/internal/common/types"
	m "dos/internal/services/master"
	"dos/internal/services/master/repo"

	"github.com/stretchr/testify/require"
)

func TestLocalCommandApplier_Apply(tt *testing.T) {
	ctx := context.Background()
	objectRepo := repo.NewInMemObjectRepo()

	applier, err := NewLocalCommandApplier(objectRepo)
	require.NoError(tt, err)

	objectID := t.ObjectID("object-1")
	slot := t.ObjectSlot{
		ObjectID: objectID,
		ChunkKey: "000001",
	}

	tt.Run("create_object", func(tt *testing.T) {
		err := applier.Apply(ctx, (&CreateObjectCommand{
			ObjectID:    objectID,
			Replication: 2,
		}).ToCommand())
		require.NoError(tt, err)

		got, err := objectRepo.Get(ctx, objectID)
		require.NoError(tt, err)
		require.Equal(tt, objectID, got.ID)
		require.Equal(tt, 2, got.Replication)
	})

	tt.Run("set_replication", func(tt *testing.T) {
		err := applier.Apply(ctx, (&SetReplicationCommand{
			ObjectID:    objectID,
			Replication: 3,
		}).ToCommand())
		require.NoError(tt, err)

		got, err := objectRepo.GetReplication(ctx, objectID)
		require.NoError(tt, err)
		require.Equal(tt, 3, got)
	})

	tt.Run("add_chunk", func(tt *testing.T) {
		err := applier.Apply(ctx, (&AddChunkCommand{
			ObjectID: objectID,
			ChunkKey: slot.ChunkKey,
			ChunkID:  "chunk-1",
		}).ToCommand())
		require.NoError(tt, err)

		got, err := objectRepo.GetChunk(ctx, slot)
		require.NoError(tt, err)
		require.Equal(tt, t.ChunkID("chunk-1"), got)
	})

	tt.Run("delete_chunk", func(tt *testing.T) {
		err := applier.Apply(ctx, (&DeleteChunkCommand{
			ObjectID: objectID,
			ChunkKey: slot.ChunkKey,
		}).ToCommand())
		require.NoError(tt, err)

		_, err = objectRepo.GetChunk(ctx, slot)
		require.ErrorIs(tt, err, m.ErrChunkKeyNotFound)
	})

	tt.Run("delete_object", func(tt *testing.T) {
		err := applier.Apply(ctx, (&DeleteObjectCommand{
			ObjectID: objectID,
		}).ToCommand())
		require.NoError(tt, err)

		_, err = objectRepo.Get(ctx, objectID)
		require.ErrorIs(tt, err, m.ErrObjectNotFound)
	})

	tt.Run("unknown", func(tt *testing.T) {
		err = applier.Apply(context.Background(), ObjectCommand{})
		require.ErrorIs(tt, err, ErrUnknownObjectCommand)
	})
}
