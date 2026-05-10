package master

import (
	"context"
	"dos/internal/common/digest"
	t "dos/internal/common/types"
	"time"
)


type Service interface {
	CreateObject(context.Context, t.ObjectID) error
	AllocateChunk(context.Context, *AllocateChunkCommand) (t.ChunkPlacement, error)
	GetObjectAccess(context.Context, t.ObjectID) (t.ObjectAccess, error)
	ListObjects(context.Context) ([]t.ObjectItem, error)

	RegisterStorageNode(context.Context, string) (t.NodeRef, error)
	ReportReplication(context.Context, t.NodeID, []t.ReplicaReport) (t.ReportResult, error)
	Heartbeat(context.Context, t.NodeID, t.NodeStats) error
	EvictStorageNode(ctx context.Context, nodeID t.NodeID) error
}

type ObjectRepo interface {
	Create(context.Context, t.ObjectID, int) error
	Get(context.Context, t.ObjectID) (Object, error)
	GetReplication(context.Context, t.ObjectID) (int, error)
	List(context.Context) []t.ObjectItem
	AddChunk(context.Context, t.ObjectID, t.ChunkKey, t.ChunkID) error
}

type ChunkRepo interface {
	NewChunkID() t.ChunkID
	Create(context.Context, t.ChunkID, t.ObjectID) error
	Get(context.Context, t.ChunkID) (Chunk, error)
	SetDigest(context.Context, t.ChunkID, *digest.Digest) error
	IncReplication(context.Context, t.ChunkID) error
	DecReplication(context.Context, t.ChunkID) error
}

type NodeQuery struct {
	MinFreeBytes int64
	ExcludeIDs []t.NodeID
}

type NodeRegistry interface {
	Register(context.Context, string) (t.NodeRef, error)
	Unregister(context.Context, t.NodeID) error 

	Get(context.Context, t.NodeID) (Node, error)
	GetMany(context.Context, ...t.NodeID) []Node
	Find(context.Context, NodeQuery) ([]Node, error)
	UpdateStats(context.Context, t.NodeID, t.NodeStats) error
	GetInactive(context.Context, time.Time) []t.NodeID
}

type ChunkNodeIndex interface {
	AttachChunk(ctx context.Context, nodeID t.NodeID, chunkID t.ChunkID) bool
	DetachNode(ctx context.Context, nodeID t.NodeID)

	GetChunkNodes(ctx context.Context, chunkID t.ChunkID) []t.NodeID
	GetNodeChunks(ctx context.Context, nodeID t.NodeID) []t.ChunkID
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
