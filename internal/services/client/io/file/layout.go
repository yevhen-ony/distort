package file

import (
	"errors"
	"slices"

	t "dos/internal/common/types"
)

var (
	ErrTotalSizeMismatch = errors.New("layout total size mismatch sum of chunks")
	ErrUnexpectedKey     = errors.New("chunk key unexpected for layout")
)

type ChunkRegion struct {
	Offset int64
	Size   int64
}

type ObjectLayout struct {
	Layout     map[t.ChunkKey]ChunkRegion
	TotalBytes int64
}

type LayoutChunk struct {
	Key  t.ChunkKey
	Size int64
}

func compareLayoutChunks(lhs, rhs LayoutChunk) int {
	if lhs.Key < rhs.Key {
		return -1
	}
	if lhs.Key > rhs.Key {
		return 1
	}
	return 0
}

type LayoutSpec struct {
	chunks []LayoutChunk
}

func NewObjectLayout(spec *LayoutSpec) (*ObjectLayout, error) {
	slices.SortFunc(spec.chunks, compareLayoutChunks)

	layout := make(map[t.ChunkKey]ChunkRegion, len(spec.chunks))
	offset := int64(0)
	for _, chunk := range spec.chunks {
		layout[chunk.Key] = ChunkRegion{
			Offset: offset,
			Size:   chunk.Size,
		}
		offset += chunk.Size
	}

	res := &ObjectLayout{
		TotalBytes: offset,
		Layout:     layout,
	}

	return res, nil
}

func (ol *ObjectLayout) Region(key t.ChunkKey) (ChunkRegion, error) {
	r, ok := ol.Layout[key]
	if !ok {
		return ChunkRegion{}, ErrUnexpectedKey
	}
	return r, nil
}
