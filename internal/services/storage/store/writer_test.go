package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dos/internal/common/digest"
	t "dos/internal/common/types"
)

func newTestWriter(test *testing.T) *FSChunkWriter {
	test.Helper()

	root := test.TempDir()
	commitDir := filepath.Join(root, "chunks")
	tempDir := filepath.Join(root, "temp")

	require.NoError(test, os.MkdirAll(commitDir, 0o755))
	require.NoError(test, os.MkdirAll(tempDir, 0o755))

	fd, err := os.CreateTemp(tempDir, "chunk-*")
	require.NoError(test, err)

	return &FSChunkWriter{fd: fd, commitDir: commitDir, dg: digest.New()}
}

func TestFSChunkWriter_Cleanup(test *testing.T) {
  	test.Run("RemovesTempFile", func(test *testing.T) {
		w := newTestWriter(test)

		_, err := w.Write([]byte("hello"))
		require.NoError(test, err, "write to temp")

		require.NoError(test, w.Close(), "cleanup")

		_, err = os.Stat(w.fd.Name())
		assert.ErrorIs(test, err, os.ErrNotExist)
	})

  	test.Run("Idempotent", func(test *testing.T) {
		w := newTestWriter(test)

		assert.NoError(test, w.Close())
		assert.NoError(test, w.Close())
	})
}

func TestFSChunkWriter_Commit(test *testing.T) {
	test.Run("CreateCommitedFile", func(test *testing.T) {
		w := newTestWriter(test)

		chunkID := t.ChunkID("chunk-001")
		want := []byte("payload")

		_, err := w.Write(want)
		require.NoError(test, err, "write to temp")

		_, err = w.Commit(chunkID)
		require.NoError(test, err, "commit")

		commitPath := filepath.Join(w.commitDir, string(chunkID))
		got, err := os.ReadFile(commitPath)

		assert.NoError(test, err, "read commited")
		assert.Equal(test, want, got, "compare content")
  	})

  	test.Run("TargetAlreadyExists", func (test *testing.T) {
		w := newTestWriter(test)

		chunkID := t.ChunkID("chunk-precreated")
		commitPath := filepath.Join(w.commitDir, string(chunkID))

		origContent := []byte("dummy data")
		err := os.WriteFile(commitPath, origContent, 0o600)
		require.NoError(test, err, "create commited")
		
		_, err = w.Write([]byte("hello, world"))
		assert.NoError(test, err, "write to temp")

		_,  err = w.Commit(chunkID)
		assert.Error(test, err, "try commit with commited id")

		got, err := os.ReadFile(commitPath)
		assert.NoError(test, err, "read commited")
		assert.Equal(test, origContent, got, "compare content")
	})
}

func TestFSChunkWriter_Write(test *testing.T) {
	test.Run("WriteAfterClose", func(test *testing.T) {
		w := newTestWriter(test)

		require.NoError(test, w.Close(), "run cleanup")

		n, err := w.Write([]byte("x"))
		assert.Error(test, err, "write after cleanup")
		assert.Zero(test, n, "nothing was written")
	})
}
