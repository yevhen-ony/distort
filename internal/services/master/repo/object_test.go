package repo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	m "dos/internal/services/master"
)

func TestInMemObjectRepo_Create(t *testing.T) {
	r := NewInMemObjectRepo()
	ctx := context.Background()

	t.Run("CreateNew", func(t *testing.T) {
		err := r.Create(ctx, m.ObjectID("obj-1"))
		require.NoError(t, err)

		got, err := r.Get(ctx, m.ObjectID("obj-1"))
		require.NoError(t, err)
		assert.Equal(t, m.ObjectID("obj-1"), got.ID)
		assert.Empty(t, got.Chunks)
	})

	t.Run("Conflict", func(t *testing.T) {
		err := r.Create(ctx, m.ObjectID("obj-1"))
		require.ErrorIs(t, err, m.ErrObjectExists)
	})
}

func TestInMemObjectRepo_Get(t *testing.T) {
	r := NewInMemObjectRepo()
	ctx := context.Background()
	oid := m.ObjectID("obj-1")

	require.NoError(t, r.Create(ctx, oid))
	require.NoError(t, r.AddChunk(ctx, oid, m.ChunkKey("1"), m.ChunkID("chunk-a")))

	t.Run("NotFound", func(t *testing.T) {
		ctx := context.Background()
		_, err := r.Get(ctx, m.ObjectID("missing"))
		require.ErrorIs(t, err, m.ErrObjectNotFound)
	})

	t.Run("ReturnsClone", func(t *testing.T) {
		obj, err := r.Get(ctx, oid)
		require.NoError(t, err)

		// Mutate returned object.
		obj.Chunks[m.ChunkKey("1")] = m.ChunkID("tampered")
		obj.Chunks[m.ChunkKey("2")] = m.ChunkID("new")

		// Repo state must stay unchanged.
		again, err := r.Get(ctx, oid)
		require.NoError(t, err)
		assert.Equal(t, m.ChunkID("chunk-a"), again.Chunks[m.ChunkKey("1")])
		_, ok := again.Chunks[m.ChunkKey("2")]
		assert.False(t, ok)
	})
}

func TestInMemObjectRepo_AddChunk(t *testing.T) {
	r := NewInMemObjectRepo()
	ctx := context.Background()
	oid := m.ObjectID("obj-1")

	require.NoError(t, r.Create(ctx, oid))

	t.Run("UniqueKey", func(t *testing.T) {
		err := r.AddChunk(ctx, oid, m.ChunkKey("0"), m.ChunkID("chunk-a"))
		require.NoError(t, err)

		obj, err := r.Get(ctx, oid)
		require.NoError(t, err)
		require.Equal(t, m.ChunkID("chunk-a"), obj.Chunks[m.ChunkKey("0")])
	})

	t.Run("DuplicateKey", func(t *testing.T) {
		err := r.AddChunk(ctx, oid, m.ChunkKey("0"), m.ChunkID("chunk-b"))
		require.ErrorIs(t, err, m.ErrChunkKeyExists)
	})

	t.Run("NonExistingObject", func(t *testing.T) {
		err := r.AddChunk(ctx, m.ObjectID("missing"), m.ChunkKey("1"), m.ChunkID("chunk-c"))
		require.ErrorIs(t, err, m.ErrObjectNotFound)
	})
}
