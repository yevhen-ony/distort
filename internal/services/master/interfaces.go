package master

import (
	"context"
	"dos/internal/common/digest"
	t "dos/internal/common/types"
	"time"
)

type StorageNodeLifecycle interface {
	Register(context.Context, string) (t.NodeRef, error)
	UpdateStats(context.Context, t.NodeID, t.NodeStats) error
	Remove(context.Context, t.NodeID) ([]t.ChunkID, error)
}

type StorageNodePlacement interface {
	GetCandidates(context.Context, CandidateNodesQuery) ([]t.NodeRef, error)
	GetChunkNodes(context.Context, t.ChunkID) ([]t.NodeRef, error)
}

type StorageNodeReport interface {
	Report(context.Context, t.NodeID, []t.StorageNodeReport) (t.ReportResult, error)
}

type ClientFacade interface {
	CreateObject(context.Context, t.ObjectID) error
	AllocateChunk(context.Context, AllocateChunkCommand) (t.ChunkPlacement, error)
	GetObjectAccess(context.Context, t.ObjectID) (t.ObjectAccess, error)

	ListObjects(context.Context) []t.ObjectInfo
	ListChunks(context.Context) []t.ChunkInfo
	ListNodes(context.Context) []t.NodeInfo
	SetReplication(context.Context, t.ObjectID, int) error
}

type ObjectCatalog interface {
	Create(context.Context, t.ObjectID, int) error
	GetReplicaCount(context.Context, t.ObjectID) (int, error)
	AllocateChunk(context.Context, t.ObjectID, t.ChunkKey, int64) (t.ChunkDesc, error)
	GetChunks(ctx context.Context, objectID t.ObjectID) ([]t.ChunkDesc, error)
}


type ObjectRepo interface {
	Create(context.Context, t.ObjectID, int) error
	Get(context.Context, t.ObjectID) (Object, error)
	GetReplication(context.Context, t.ObjectID) (int, error)
	SetReplication(context.Context, t.ObjectID, int) error
	List(context.Context) []Object
	AddChunk(context.Context, t.ObjectID, t.ChunkKey, t.ChunkID) error
	RemoveChunk(context.Context, t.ObjectID, t.ChunkKey)
	DeleteObject(context.Context, t.ObjectID) error
}

type ChunkRepo interface {
	NewChunkID() t.ChunkID
	Create(context.Context, t.ChunkID, t.ObjectID, t.ChunkKey) error
	Get(context.Context, t.ChunkID) (Chunk, error)
	SetDigest(context.Context, t.ChunkID, *digest.Digest) error
	IncReplication(context.Context, t.ChunkID)
	DecReplication(context.Context, t.ChunkID)
	List(context.Context) []Chunk
	DeleteWithNoReplicas(context.Context, t.ChunkID) bool
	Touch(context.Context, t.ChunkID)

	ForEach(context.Context, func(Chunk))
}

type NodeQuery struct {
	MinFreeBytes int64
	ExcludeIDs   []t.NodeID
}

type NodeRegistry interface {
	Register(context.Context, string) (t.NodeRef, error)
	Unregister(context.Context, t.NodeID)

	Get(context.Context, t.NodeID) (Node, error)
	GetMany(context.Context, ...t.NodeID) []Node
	Find(context.Context, NodeQuery) []Node
	UpdateStats(context.Context, t.NodeID, t.NodeStats) error

	Count(context.Context) int 
	GetInactive(context.Context, time.Time) []t.NodeID
}

type ChunkNodeIndex interface {
	AttachChunk(ctx context.Context, nodeID t.NodeID, chunkID t.ChunkID) bool
	DetachChunk(ctx context.Context, nodeID t.NodeID, chunkID t.ChunkID) bool
	DetachNode(ctx context.Context, nodeID t.NodeID)

	GetChunkNodes(ctx context.Context, chunkID t.ChunkID) []t.NodeID
	GetNodeChunks(ctx context.Context, nodeID t.NodeID) []t.ChunkID
}

type CandidateNodesQuery struct {
	MinFreeBytes int64
	ExcludeChunk      t.ChunkID
	MaxCount          int
}

type PlacementPolicy interface {
	Select(nodes []Node, n int) []Node
}

type AllocateChunkCommand struct {
	ObjectID  t.ObjectID
	ChunkKey  t.ChunkKey
	ChunkSize int64
}

type ReplicaScheduler interface {
	Schedule(context.Context, t.ChunkID)
}
