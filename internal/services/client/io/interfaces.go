package io

import (
	t "dos/internal/common/types"
)

type ObjectAssembler interface {
	NewSink([]t.ChunkPlacement) (ObjectSink, error)
}

type ObjectSink interface {
	WriteChunk(t.ChunkKey, []byte) error
	Close() error
}
