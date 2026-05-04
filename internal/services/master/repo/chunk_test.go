package repo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dos/internal/common/digest"
	m "dos/internal/services/master"
	t "dos/internal/common/types"
)

func TestInMemChunkRepo_Create(test *testing.T) {
	r := NewInMemChunkRepo()
	ctx := context.Background()

	test.Run("Success", func(test *testing.T) {
		id := r.NewChunkID()
		err := r.Create(ctx, id)
		require.NoError(test, err)
		require.NotEmpty(test, id)
	})
}

func TestInMemChunkRepo_Get(test *testing.T) {
	r := NewInMemChunkRepo()
	ctx := context.Background()
	id := r.NewChunkID()
	err := r.Create(ctx, id)
	require.NoError(test, err)
	
	test.Run("NotFound", func(test *testing.T) {
		_, err := r.Get(ctx, t.ChunkID("missing"))
		require.ErrorIs(test, err, m.ErrChunkNotFound)
	})

	test.Run("Found", func(test *testing.T) {
		_, err := r.Get(ctx, id)
		require.NoError(test, err)
	})

	test.Run("WithDigest", func(test *testing.T) {
		r.SetDigest(ctx, id, &digest.Digest{Size: 1, Checksum: "abc"})
		
		ch, err := r.Get(ctx, id)
		require.NoError(test, err)
		require.NotNil(test, ch.Digest)
		assert.Equal(test, int64(1), ch.Digest.Size)
		assert.Equal(test, "abc", string(ch.Digest.Checksum))
	})
}

func TestInMemChunkRepo_SetDigest(test *testing.T) {
	r := NewInMemChunkRepo()
	ctx := context.Background()

	id := r.NewChunkID()
	err := r.Create(ctx, id)
	require.NoError(test, err)

	dgt := &digest.Digest{Size: 5, Checksum: "abc"}
	test.Run("NewDigest", func(test *testing.T) {
		require.NoError(test, r.SetDigest(ctx, id, dgt))
	})

	test.Run("SameDigest", func(test *testing.T) {
		require.NoError(test, r.SetDigest(ctx, id, dgt))
	})

	test.Run("ConflictDigest", func(test *testing.T) {
		err = r.SetDigest(ctx, id, &digest.Digest{
			Size: 5,
			Checksum: "xyz",
		})
		require.ErrorIs(test, err, digest.ErrDigestMismatch)
	})
	
	test.Run("ChunkNotFound", func(test *testing.T) {
		err := r.SetDigest(ctx, t.ChunkID("missing"), dgt)
		require.ErrorIs(test, err, m.ErrChunkNotFound)
	})
}
