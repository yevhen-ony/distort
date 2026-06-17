package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	t "dos/internal/common/types"
)

func TestObjectChunker_Chunks(tt *testing.T) {
	path := filepath.Join(tt.TempDir(), "input")
	require.NoError(tt, os.WriteFile(path, []byte("helloworld!"), 0o600))

	chunker, err := NewObjectChunker(path, 5)
	require.NoError(tt, err)
	defer chunker.Close()

	got := map[t.ChunkKey]string{}
	for key, data := range chunker.Chunks() {
		got[key] = string(data)
	}

	require.NoError(tt, chunker.Err())
	require.Equal(tt, map[t.ChunkKey]string{
		"000000": "hello",
		"000001": "world",
		"000002": "!",
	}, got)
}
