package master

import (
	"context"
	"dos/internal/common/digest"
	t "dos/internal/common/types"
	"time"
)

//go:generate mockgen -source=$GOFILE -destination=mock/mocks.go -package=mock

type StorageNodePlacement interface {
	GetCandidates(context.Context, CandidateNodesQuery) ([]t.NodeRef, error)
	GetChunkNodes(context.Context, t.ChunkID) ([]t.NodeRef, error)
}

type ChunkRepo interface {
	NewChunkID() t.ChunkID

	Create(context.Context, t.ChunkID, t.ObjectSlot) error
	Delete(context.Context, t.ChunkID) (bool, error)
	Touch(context.Context, t.ChunkID) error
	Drop(context.Context, t.ChunkID)

	Exists(context.Context, t.ChunkID) (bool, error)
	Get(context.Context, t.ChunkID) (Chunk, error)
	List(context.Context) []Chunk

	SetDigest(context.Context, t.ChunkID, digest.Digest) error
	GetDigest(context.Context, t.ChunkID) (digest.Digest, error)

	IncReplicaCount(context.Context, t.ChunkID) error
	DecReplicaCount(context.Context, t.ChunkID) error
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
	TransferLeadership(context.Context) error
}

type ObjectReader interface {
	List(ctx context.Context) []Object
	Get(ctx context.Context, objectID t.ObjectID) (Object, error)
	Exists(ctx context.Context, objectID t.ObjectID) (bool, error)

	GetReplication(ctx context.Context, objectID t.ObjectID) (int, error)
	ExistsChunk(ctx context.Context, slot t.ObjectSlot) (bool, error)
	GetChunk(ctx context.Context, slot t.ObjectSlot) (t.ChunkID, error)
}

type ObjectWriter interface {
	Create(context.Context, t.ObjectID, int) error
	Delete(context.Context, t.ObjectID) error
	SetReplication(context.Context, t.ObjectID, int) error

	AddChunk(context.Context, t.ObjectSlot, t.ChunkID) error
	DeleteChunk(context.Context, t.ObjectSlot) error
}

type ObjectRW interface {
	ObjectWriter
	ObjectReader
}
