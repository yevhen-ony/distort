package io

import (
	t "dos/internal/common/types"
)

//go:generate mockgen -source=$GOFILE -destination=mock/mocks.go -package=mock

type ObjectAssembler interface {
	NewSink([]t.ChunkPlacement) (ObjectSink, error)
}

type ObjectSink interface {
	WriteChunk(t.ChunkKey, []byte) error
	Close() error
}
