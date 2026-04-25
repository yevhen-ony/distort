package repo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	t "dos/internal/common/types"
	m "dos/internal/services/master"
)

func TestInMemObjectRepo_Create(test *testing.T) {
	r := NewInMemObjectRepo()
	ctx := context.Background()

	test.Run("CreateNew", func(test *testing.T) {
		err := r.Create(ctx, t.ObjectID("obj-1"))
		require.NoError(test, err)

		got, err := r.Get(ctx, t.ObjectID("obj-1"))
		require.NoError(test, err)
		assert.Equal(test, t.ObjectID("obj-1"), got.ID)
		assert.Empty(test, got.Chunks)
	})

	test.Run("Conflict", func(test *testing.T) {
		err := r.Create(ctx, t.ObjectID("obj-1"))
		require.ErrorIs(test, err, m.ErrObjectExists)
	})
}

func TestInMemObjectRepo_Get(test *testing.T) {
	r := NewInMemObjectRepo()
	ctx := context.Background()
	oid := t.ObjectID("obj-1")

	require.NoError(test, r.Create(ctx, oid))
	require.NoError(test, r.AddChunk(ctx, oid, t.ChunkKey("1"), t.ChunkID("chunk-a")))

	test.Run("NotFound", func(test *testing.T) {
		ctx := context.Background()
		_, err := r.Get(ctx, t.ObjectID("missing"))
		require.ErrorIs(test, err, m.ErrObjectNotFound)
	})

	test.Run("ReturnsClone", func(test *testing.T) {
		obj, err := r.Get(ctx, oid)
		require.NoError(test, err)

		// Mutate returned object.
		obj.Chunks[t.ChunkKey("1")] = t.ChunkID("tampered")
		obj.Chunks[t.ChunkKey("2")] = t.ChunkID("new")

		// Repo state must stay unchanged.
		again, err := r.Get(ctx, oid)
		require.NoError(test, err)
		assert.Equal(test, t.ChunkID("chunk-a"), again.Chunks[t.ChunkKey("1")])
		_, ok := again.Chunks[t.ChunkKey("2")]
		assert.False(test, ok)
	})
}

func TestInMemObjectRepo_AddChunk(test *testing.T) {
	r := NewInMemObjectRepo()
	ctx := context.Background()
	oid := t.ObjectID("obj-1")

	require.NoError(test, r.Create(ctx, oid))

	test.Run("UniqueKey", func(test *testing.T) {
		err := r.AddChunk(ctx, oid, t.ChunkKey("0"), t.ChunkID("chunk-a"))
		require.NoError(test, err)

		obj, err := r.Get(ctx, oid)
		require.NoError(test, err)
		require.Equal(test, t.ChunkID("chunk-a"), obj.Chunks[t.ChunkKey("0")])
	})

	test.Run("DuplicateKey", func(test *testing.T) {
		err := r.AddChunk(ctx, oid, t.ChunkKey("0"), t.ChunkID("chunk-b"))
		require.ErrorIs(test, err, m.ErrChunkKeyExists)
	})

	test.Run("NonExistingObject", func(test *testing.T) {
		err := r.AddChunk(ctx, t.ObjectID("missing"), t.ChunkKey("1"), t.ChunkID("chunk-c"))
		require.ErrorIs(test, err, m.ErrObjectNotFound)
	})
}
