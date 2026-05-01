package types

import (
	"dos/internal/common/digest"
)

type ObjectID string
type ChunkID string
type ChunkKey string
type NodeID string

type NodeRef struct {
	ID   NodeID
	Addr string
}

type ChunkLocation struct {
	ChunkID    ChunkID
	ChunkKey   ChunkKey
	Nodes []NodeRef
}

type ObjectAccess struct {
	ID        ObjectID
	TotalSize int64
	Chunks    []ChunkLocation
}

type NodeStats struct {
	FreeBytes  int64
	UsedBytes  int64
	ChunkCount int
}

type ChunkMeta struct {
	ID     ChunkID
	Digest digest.Digest
}

type ChunkStorageReject struct {
	ChunkID ChunkID
	Reason  string
}
