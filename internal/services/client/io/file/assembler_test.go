package file

import (
	"os"
	"path/filepath"
	"testing"

	"dos/internal/common/digest"
	t "dos/internal/common/types"

	"github.com/stretchr/testify/require"
)

func TestObjectAssembler_FollowsLayout(tt *testing.T) {
	path := filepath.Join(tt.TempDir(), "out", "object")

	asm, err := NewObjectAssembler(path)
	require.NoError(tt, err)

	sink, err := asm.NewSink([]t.ChunkPlacement{
		placementWithSize("000001", 5),
		placementWithSize("000000", 5),
	})
	require.NoError(tt, err)

	require.NoError(tt, sink.WriteChunk("000001", []byte("world")))
	require.NoError(tt, sink.WriteChunk("000000", []byte("hello")))
	require.NoError(tt, sink.Close())

	data, err := os.ReadFile(path)
	require.NoError(tt, err)
	require.Equal(tt, []byte("helloworld"), data)
}

func placementWithSize(key t.ChunkKey, size int64) t.ChunkPlacement {
	return t.ChunkPlacement{
		Meta: t.ChunkMeta{Digest: digest.Digest{Size: size}},
		Slot: t.ObjectSlot{ChunkKey: key},
	}
}
