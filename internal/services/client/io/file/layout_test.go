package file

import (
	"testing"

	t "dos/internal/common/types"

	"github.com/stretchr/testify/require"
)

func TestObjectLayout_SortsChunks(tt *testing.T) {
	layout, err := NewObjectLayout(&LayoutSpec{
		chunks: []LayoutChunk{
			{Key: "000001", Size: 3},
			{Key: "000000", Size: 2},
		},
	})
	require.NoError(tt, err)

	require.Equal(tt, int64(5), layout.TotalBytes)

	region, err := layout.Region("000000")
	require.NoError(tt, err)
	require.Equal(tt, ChunkRegion{Offset: 0, Size: 2}, region)

	region, err = layout.Region("000001")
	require.NoError(tt, err)
	require.Equal(tt, ChunkRegion{Offset: 2, Size: 3}, region)
}

func TestObjectLayout_UnexpectedKey(tt *testing.T) {
	layout, err := NewObjectLayout(&LayoutSpec{
		chunks: []LayoutChunk{{Key: "000000", Size: 2}},
	})
	require.NoError(tt, err)

	_, err = layout.Region(t.ChunkKey("missing"))

	require.ErrorIs(tt, err, ErrUnexpectedKey)
}
