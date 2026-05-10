package store

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	t "dos/internal/common/types"
)


func TestFSChunkStorage_Get(test *testing.T) {
	test.Run("HappyPath", func(test *testing.T) {
		cfg := NewTestConfig(test) 

		store, err := NewChunkStorage(cfg)
		require.NoError(test, err)
		
		chunkID := t.ChunkID("chunk-1")
		content := []byte("1234567")
		storeChunk(test, store.commitDir, chunkID, content)

		r, err := store.Get(chunkID)
		require.NoError(test, err)
		defer r.Close()

		got, err := io.ReadAll(r)
		require.NoError(test, err)
		assert.Equal(test, content, got)
	})
}

func TestFSChunkStorage_GetMeta(test *testing.T) {
	cfg := NewTestConfig(test) 
	store, err := NewChunkStorage(cfg)
	require.NoError(test, err)

	chunkID := t.ChunkID("chunk-2")
	content := []byte("1234567")
	storeChunk(test, store.commitDir, chunkID, content)

	test.Run("ChunkExists", func(test *testing.T){
		meta, err := store.GetMeta(chunkID)
		require.NoError(test, err)
		require.NotNil(test, meta)
		require.Equal(test, int64(len(content)), meta.Digest.Size)
	})

	test.Run("ChunkNotExists", func(test *testing.T) {
		_, err := store.GetMeta(t.ChunkID("notexist"))
		require.Error(test, err, "access nonexisting chunk") 
	})
}

func TestFSChunkStorage_GetAllIDs(test *testing.T) {
	test.Run("EmptyStorage", func(test *testing.T) {
		cfg := NewTestConfig(test) 
		store, err := NewChunkStorage(cfg)
		require.NoError(test, err)

		ids, err := store.List()
		require.NoError(test, err, "get all ids")
		require.Empty(test, ids, "no ids returned")

	})

	test.Run("WithTwoChunks", func(test *testing.T) {
		cfg := NewTestConfig(test) 
		store, err := NewChunkStorage(cfg)
		require.NoError(test, err)

		storeChunk(test, store.commitDir, t.ChunkID("ch-1"), []byte("hello"))
		storeChunk(test, store.commitDir, t.ChunkID("ch-2"), []byte("world"))

		ids, err := store.List()
		require.NoError(test, err)
		assert.ElementsMatch(test, []t.ChunkID{"ch-1", "ch-2"}, ids)
	})
}

func TestFSChunkStorage_NewWriter(test *testing.T) {
	cfg := NewTestConfig(test)
  	s, err := NewChunkStorage(cfg)
  	require.NoError(test, err)

  	w, err := s.NewWriter()
  	require.NoError(test, err)
  	require.NotNil(test, w)
  	defer w.Close()

  	fsw, ok := w.(*FSChunkWriter)
  	require.True(test, ok)
  	assert.Equal(test, s.commitDir, fsw.commitDir)
	assert.Equal(test, s.tempDir, filepath.Dir(fsw.fd.Name()))
}


func storeChunk(test *testing.T, dir string, id t.ChunkID, content []byte) {
	path := filepath.Join(dir, string(id))
	err := os.WriteFile(path, content, 0o600)
	require.NoError(test, err, "store chunk")
}

type TestConfig struct {
	test *testing.T
}

func NewTestConfig(test *testing.T) *TestConfig {
	return  &TestConfig{test: test} 
}

func (c *TestConfig) StorageRootDir() string {
	return c.test.TempDir()
}

