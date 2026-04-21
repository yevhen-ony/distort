package storage

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cs "dos/internal/chunkserver"
)

func TestFSChunkStorage_New(t *testing.T) {
	t.Run("DirsCreated", func(t *testing.T) {
		rootDir := t.TempDir() 
		cfg := &ChunkStorageConfig{RootDir: rootDir}

		s, err := New(cfg)
		require.NoError(t, err)

		assert.NotNil(t, s)
		assert.NotEqual(t, s.commitDir, s.tempDir, "commit and temp dirs are different")
		assert.DirExists(t, s.commitDir, "commit dir exists")
		assert.DirExists(t, s.tempDir, "temp dir exists")
	})
}

func TestFSChunkStorage_Get(t *testing.T) {
	t.Run("HappyPath", func(t *testing.T) {
		rootDir := t.TempDir()
		cfg := &ChunkStorageConfig{RootDir: rootDir}

		s, err := New(cfg)
		require.NoError(t, err)
		
		chunkID := cs.ChunkID("chunk-1")
		content := []byte("1234567")
		storeChunk(t, s.commitDir, chunkID, content)

		r, err := s.Get(chunkID)
		require.NoError(t, err)
		defer r.Close()

		got, err := io.ReadAll(r)
		require.NoError(t, err)
		assert.Equal(t, content, got)
	})
}

func TestFSChunkStorage_GetMeta(t *testing.T) {
	cfg := &ChunkStorageConfig{RootDir: t.TempDir()}
	s, err := New(cfg)
	require.NoError(t, err)

	chunkID := cs.ChunkID("chunk-2")
	content := []byte("1234567")
	storeChunk(t, s.commitDir, chunkID, content)

	t.Run("ChunkExists", func(t *testing.T){
		meta, err := s.GetMeta(chunkID)
		require.NoError(t, err)
		require.NotNil(t, meta)
		require.Equal(t, int64(len(content)), meta.Size)
	})

	t.Run("ChunkNotExists", func(t *testing.T) {
		_, err := s.GetMeta(cs.ChunkID("notexist"))
		require.Error(t, err, "access nonexisting chunk") 
	})
}

func TestFSChunkStorage_GetAllIDs(t *testing.T) {
	t.Run("EmptyStorage", func(t *testing.T) {
		cfg := &ChunkStorageConfig{RootDir: t.TempDir()}
		s, err := New(cfg)
		require.NoError(t, err)

		ids, err := s.GetAllIDs()
		require.NoError(t, err, "get all ids")
		require.Empty(t, ids, "no ids returned")

	})

	t.Run("WithTwoChunks", func(t *testing.T) {
		cfg := &ChunkStorageConfig{RootDir: t.TempDir()}
		s, err := New(cfg)
		require.NoError(t, err)

		storeChunk(t, s.commitDir, cs.ChunkID("ch-1"), []byte("hello"))
		storeChunk(t, s.commitDir, cs.ChunkID("ch-2"), []byte("world"))

		ids, err := s.GetAllIDs()
		require.NoError(t, err)
		assert.ElementsMatch(t, []cs.ChunkID{"ch-1", "ch-2"}, ids)
	})
}

func TestFSChunkStorage_NewWriter(t *testing.T) {
	cfg := &ChunkStorageConfig{RootDir: t.TempDir()}
  	s, err := New(cfg)
  	require.NoError(t, err)

  	w, err := s.NewWriter()
  	require.NoError(t, err)
  	require.NotNil(t, w)
  	defer w.Close()

  	fsw, ok := w.(*FSChunkWriter)
  	require.True(t, ok)
  	assert.Equal(t, s.commitDir, fsw.commitDir)
	assert.Equal(t, s.tempDir, filepath.Dir(fsw.fd.Name()))
}


func storeChunk(t *testing.T, dir string, id cs.ChunkID, content []byte) {
	path := filepath.Join(dir, string(id))
	err := os.WriteFile(path, content, 0o600)
	require.NoError(t, err, "store chunk")
}

