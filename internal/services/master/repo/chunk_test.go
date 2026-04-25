package repo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dos/internal/common/digest"
	m "dos/internal/services/master"
)

func TestInMemChunkRepo_Create(t *testing.T) {
	r := MakeInMemChunkRepo()
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		id, err := r.Create(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, id)
	})
}

func TestInMemChunkRepo_Get(t *testing.T) {
	r := MakeInMemChunkRepo()
	ctx := context.Background()
	id, err := r.Create(ctx)
	require.NoError(t, err)
	
	t.Run("NotFound", func(t *testing.T) {
		_, err := r.Get(ctx, m.ChunkID("missing"))
		require.ErrorIs(t, err, m.ErrChunkNotFound)
	})

	t.Run("Found", func(t *testing.T) {
		_, err := r.Get(ctx, id)
		require.NoError(t, err)
	})

	t.Run("WithDigest", func(t *testing.T) {
		r.SetDigest(ctx, id, &digest.Digest{Size: 1, Checksum: "abc"})
		
		ch, err := r.Get(ctx, id)
		require.NoError(t, err)
		require.NotNil(t, ch.Digest)
		assert.Equal(t, int64(1), ch.Digest.Size)
		assert.Equal(t, "abc", ch.Digest.Checksum)
	})
}

func TestInMemChunkRepo_SetDigest(t *testing.T) {
	r := MakeInMemChunkRepo()
	ctx := context.Background()

	id, err := r.Create(ctx)
	require.NoError(t, err)

	dgt := &digest.Digest{Size: 5, Checksum: "abc"}
	t.Run("NewDigest", func(t *testing.T) {
		require.NoError(t, r.SetDigest(ctx, id, dgt))
	})

	t.Run("SameDigest", func(t *testing.T) {
		require.NoError(t, r.SetDigest(ctx, id, dgt))
	})

	t.Run("ConflictDigest", func(t *testing.T) {
		err = r.SetDigest(ctx, id, &digest.Digest{
			Size: 5,
			Checksum: "xyz",
		})
		require.ErrorIs(t, err, m.ErrChunkDigestConflict)
	})
	
	t.Run("ChunkNotFound", func(t *testing.T) {
		err := r.SetDigest(ctx, m.ChunkID("missing"), dgt)
		require.ErrorIs(t, err, m.ErrChunkNotFound)
	})
}
