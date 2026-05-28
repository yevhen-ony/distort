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
	AllocateChunk(context.Context, AllocateChunkCommand) (*t.ChunkAllocation1, error)

	ListObjects(context.Context) []t.ObjectInfo
	ListChunks(context.Context) []t.ChunkInfo
	ListNodes(context.Context) []t.NodeInfo
	SetReplication(context.Context, t.ObjectID, int) error

	DescribeChunk(context.Context, t.ChunkID) (*t.ChunkDesc1, error)
	DescribeObject(context.Context, t.ObjectID) (*t.ObjectDesc1, error)
}

type ObjectCatalog interface {
	CreateObject(context.Context, t.ObjectID, int) error
	GetObject(ctx context.Context, objectID t.ObjectID) (Object, error)

	GetReplicaCount(context.Context, t.ObjectID) (int, error)
	AllocateChunk(context.Context, t.ObjectID, t.ChunkKey, int64) (t.ChunkDesc, error)
	GetChunks(ctx context.Context, objectID t.ObjectID) ([]t.ChunkDesc, error)
}

type ChunkRepo interface {
	NewChunkID() t.ChunkID

	Create(context.Context, t.ChunkID, t.ObjectSlot) error
	Delete(context.Context, t.ChunkID) (bool, error)
	Touch(context.Context, t.ChunkID) error

	Exists(context.Context, t.ChunkID) (bool, error)
	Get(context.Context, t.ChunkID) (Chunk, error)
	List(context.Context) ([]Chunk, error)

	SetDigest(context.Context, t.ChunkID, digest.Digest) error
	GetDigest(context.Context, t.ChunkID) (digest.Digest, error)

	IncReplicaCount(context.Context, t.ChunkID) error
	DecReplicaCount(context.Context, t.ChunkID) error

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
	MaxCount     int
	ExcludeChunk t.ChunkID
	ExcludeNodes []t.NodeRef
}

type PlacementPolicy interface {
	Select(nodes []Node, n int) []Node
}

type AllocateChunkCommand struct {
	Slot         t.ObjectSlot
	Size         int64
	ExcludeNodes []t.NodeRef
}

type ReplicaScheduler interface {
	Schedule(context.Context, t.ChunkID)
}

type MasterState interface {
	GetActiveMaster(context.Context) (t.MasterRef, error)
	IsActiveMaster() bool
	WatchState(context.Context, func(context.Context))
}

