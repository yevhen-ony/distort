package file

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestObjectWriter_WriteRegion_OrderIrrelevant(tt *testing.T) {
	path := filepath.Join(tt.TempDir(), "object")

	writer, err := NewObjectWriter(path, 10)
	require.NoError(tt, err)

	r1 := ChunkRegion{Offset: 5, Size: 5}
	require.NoError(tt, writer.WriteRegion(r1, []byte("world")))

	r0 := ChunkRegion{Offset: 0, Size: 5}
	require.NoError(tt, writer.WriteRegion(r0, []byte("hello")))

	require.NoError(tt, writer.Close())

	data, err := os.ReadFile(path)
	require.NoError(tt, err)
	require.Equal(tt, []byte("helloworld"), data)
}

func TestObjectWriter_WriteRegion_SizeMismatch(tt *testing.T) {
	path := filepath.Join(tt.TempDir(), "object")

	writer, err := NewObjectWriter(path, 10)
	require.NoError(tt, err)
	defer writer.Close()
	
	r := ChunkRegion{Offset: 0, Size: 5}
	err = writer.WriteRegion(r, []byte("hi"))

	require.ErrorIs(tt, err, ErrChunkSizeMismatch)
}

func TestObjectWriter_FileAlreadyExists(tt *testing.T) {
	path := filepath.Join(tt.TempDir(), "object")
	require.NoError(tt, os.WriteFile(path, []byte("exists"), 0o600))

	_, err := NewObjectWriter(path, 10)

	require.Error(tt, err)
	require.True(tt, errors.Is(err, os.ErrExist), "err = %v", err)
}
