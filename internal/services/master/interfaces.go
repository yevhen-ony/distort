package master

import (
	"context"
	"dos/internal/common/digest"
	t "dos/internal/common/types"
)

type Service interface {
	CreateObject(context.Context, t.ObjectID) error
	AllocateChunk(context.Context, *AllocateChunkCommand) (t.ChunkPlacement, error)
	GetObjectAccess(context.Context, t.ObjectID) (ObjectAccess, error)
}

type ObjectRepo interface {
	Create(context.Context, t.ObjectID) error
	Get(context.Context, t.ObjectID) (Object, error)
	AddChunk(context.Context, t.ObjectID, t.ChunkKey, t.ChunkID) error
}

type ChunkRepo interface {
	Create(context.Context) (t.ChunkID, error)
	Get(context.Context, t.ChunkID) (Chunk, error)
	SetDigest(context.Context, t.ChunkID, *digest.Digest) error
}

type NodeRegistry interface {
	Register(context.Context, *t.NodeReport) (t.NodeID, error)
	Unregister(context.Context, t.NodeID) error

	GetNode(context.Context, t.NodeID) (Node, error)

	AttachChunk(context.Context, t.NodeID, t.ChunkID) error
	GetNodeChunks(context.Context, t.NodeID) ([]t.ChunkID, error)
	GetChunkNodes(context.Context, t.ChunkID) ([]Node, error)

	GetCandidateNodes(context.Context, *CandidateNodesQuery) ([]Node, error)
}

type CandidateNodesQuery struct {
	MinFreeBytes int64
	ExcludeChunk t.ChunkID
}

type PlacementPolicy interface {
	Select(nodes []Node, n int) []Node
}

type AllocateChunkCommand struct {
	ObjectID  t.ObjectID
	ChunkKey  t.ChunkKey
	ChunkSize int64
}
