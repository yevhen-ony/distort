package master

import (
	"context"
	"dos/internal/libraries/digest"
)

type Service interface {
	CreateObject(context.Context, ObjectID) error
	AllocateChunk(context.Context, *AllocateChunkCommand) (ChunkPlacement, error)
	GetObjectAccess(context.Context, ObjectID) (ObjectAccess, error)
}

type ObjectRepo interface {
	Create(context.Context, ObjectID) error
	Get(context.Context, ObjectID) (Object, error)
	AddChunk(context.Context, ObjectID, ChunkKey, ChunkID) error
}

type ChunkRepo interface {
	Create(context.Context) (ChunkID, error)
	Get(context.Context, ChunkID) (Chunk, error)
	SetDigest(context.Context, ChunkID, digest.Digest) error
}

type NodeRegistry interface {
	Register(context.Context, string) (NodeID, error)
	Unregister(context.Context, NodeID) error

	GetNode(context.Context, NodeID) (Node, error)

	AttachChunk(context.Context, NodeID, ChunkID) error
	GetNodeChunks(context.Context, NodeID) ([]ChunkID, error)
	GetChunkNodes(context.Context, ChunkID) ([]Node, error)

	GetCandidateNodes(context.Context, *CandidateNodesQuery) ([]Node, error)
}

type CandidateNodesQuery struct {
	MinFreeBytes int64
	ExcludeChunk ChunkID
}

type PlacementPolicy interface {
	Select(nodes []Node, n int) []Node
}

type AllocateChunkCommand struct {
	ObjectID  ObjectID
	ChunkKey  ChunkKey
	ChunkSize int64
}
