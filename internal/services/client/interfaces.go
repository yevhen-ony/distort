package client

import (
	"context"

	t "dos/internal/common/types"
)
type AllocateChunkQuery struct {
	ObjectID t.ObjectID
	ChunkKey t.ChunkKey
	ChunkSize int64
}

type MasterTransport interface {
	CreateObject(context.Context, t.ObjectID) error
	AllocateChunk(context.Context, *AllocateChunkQuery) (t.ChunkPlacement, error)
	GetObjectAccess(context.Context, t.ObjectID) (t.ObjectAccess, error)
}
 
type StorageTransport interface {
	PushChunk(context.Context, []t.NodeRef, *Chunk) error
	PullChunk(context.Context, []t.NodeRef, t.ChunkID) (Chunk, error)
}

type ObjectInfo struct {
	ID        t.ObjectID
	TotalSize int64
	Chunks    []t.ChunkDesc
}

type ObjectAssembler interface {
	NewWriter(t.ObjectDesc, []t.ChunkDesc) (ObjectWriter, error)
}

type ObjectWriter interface {
	WriteChunk(t.ChunkID, []byte) error
	Close() error
}

type ChunkSource interface {
	Next() (t.ChunkKey, []byte, error)
}
