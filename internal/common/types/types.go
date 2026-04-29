package types

import (
	"dos/internal/common/digest"
	"time"
)

type ObjectID string
type ChunkID string
type ChunkKey string
type NodeID string

type NodeRef struct {
	ID   NodeID
	Addr string
}

type ChunkPlacement struct {
	ID    ChunkID
	Key   ChunkKey
	Nodes []NodeRef
}

type ObjectAccess struct {
	ID        ObjectID
	TotalSize int64
	Chunks    []ChunkPlacement
}

type NodeStats struct {
	FreeBytes  int64
	UsedBytes  int64
	ChunkCount int
}

type ChunkDesc struct {
	ID     ChunkID
	Digest digest.Digest
}

type ChunkMeta struct {
	ChunkDesc
	ModifiedAt time.Time
}

type ChunkStorageReject struct {
	ChunkID ChunkID
	Reason  string
}
