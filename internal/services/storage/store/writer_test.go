package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	s "dos/internal/services/storage"
	"dos/internal/common/digest"
)

func newTestWriter(t *testing.T) *FSChunkWriter {
	t.Helper()

	root := t.TempDir()
	commitDir := filepath.Join(root, "chunks")
	tempDir := filepath.Join(root, "temp")

	require.NoError(t, os.MkdirAll(commitDir, 0o755))
	require.NoError(t, os.MkdirAll(tempDir, 0o755))

	fd, err := os.CreateTemp(tempDir, "chunk-*")
	require.NoError(t, err)

	return &FSChunkWriter{fd: fd, commitDir: commitDir, dg: digest.New()}
}

func TestFSChunkWriter_Cleanup(t *testing.T) {
  	t.Run("RemovesTempFile", func(t *testing.T) {
		w := newTestWriter(t)

		_, err := w.Write([]byte("hello"))
		require.NoError(t, err, "write to temp")

		require.NoError(t, w.Close(), "cleanup")

		_, err = os.Stat(w.fd.Name())
		assert.ErrorIs(t, err, os.ErrNotExist)
	})

  	t.Run("Idempotent", func(t *testing.T) {
		w := newTestWriter(t)

		assert.NoError(t, w.Close())
		assert.NoError(t, w.Close())
	})
}

func TestFSChunkWriter_Commit(t *testing.T) {
	t.Run("CreateCommitedFile", func(t *testing.T) {
		w := newTestWriter(t)

		chunkID := s.ChunkID("chunk-001")
		want := []byte("payload")

		_, err := w.Write(want)
		require.NoError(t, err, "write to temp")

		_, err = w.Commit(chunkID)
		require.NoError(t, err, "commit")

		commitPath := filepath.Join(w.commitDir, string(chunkID))
		got, err := os.ReadFile(commitPath)

		assert.NoError(t, err, "read commited")
		assert.Equal(t, want, got, "compare content")
  	})

  	t.Run("TargetAlreadyExists", func (t *testing.T) {
		w := newTestWriter(t)

		chunkID := s.ChunkID("chunk-precreated")
		commitPath := filepath.Join(w.commitDir, string(chunkID))

		origContent := []byte("dummy data")
		err := os.WriteFile(commitPath, origContent, 0o600)
		require.NoError(t, err, "create commited")
		
		_, err = w.Write([]byte("hello, world"))
		assert.NoError(t, err, "write to temp")

		_,  err = w.Commit(chunkID)
		assert.Error(t, err, "try commit with commited id")

		got, err := os.ReadFile(commitPath)
		assert.NoError(t, err, "read commited")
		assert.Equal(t, origContent, got, "compare content")
	})
}

func TestFSChunkWriter_Write(t *testing.T) {
	t.Run("WriteAfterClose", func(t *testing.T) {
		w := newTestWriter(t)

		require.NoError(t, w.Close(), "run cleanup")

		n, err := w.Write([]byte("x"))
		assert.Error(t, err, "write after cleanup")
		assert.Zero(t, n, "nothing was written")
	})
}
