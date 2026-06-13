package storage

import (
	"context"
	"dos/internal/common/transport/chunkrpc"
	t "dos/internal/common/types"
	s "dos/internal/services/storage"
	"io"
	"time"
)

//go:generate mockgen -source=$GOFILE -destination=mocks_test.go -package=storage

type StorageConfig interface {
	AdvertiseAddr() string
	ReplicationTimeout() time.Duration
	MaxParallelHeavyOps() int
}

type Reporter interface {
	Report(context.Context, t.StorageNodeReport)
	Flush(context.Context)
}

type Identity interface {
	GetID() (t.NodeID, error)
}

type ChunkUploadSession = chunkrpc.UploadSession

type ChunkTransport interface {
	NewUploadSession([]t.NodeRef, ...chunkrpc.SessionOption) ChunkUploadSession
	ReplicateChunk(context.Context, t.ChunkID, t.NodeRef, []t.NodeRef) error
}

type Inventory interface {
	BuildCatalog(context.Context, s.CatalogSource) error
	Has(t.ChunkID) bool
	Add(*t.ChunkMeta) error
	GetRecord(t.ChunkID) (*s.ChunkRecord, error)
	Remove(t.ChunkID) bool
	ListIDs() []t.ChunkID
	Stage(t.ChunkID) (t.ChunkMeta, error)
	Activate(t.ChunkID) (t.ChunkMeta, error)
}

type ChunkStorage interface {
	Get(t.ChunkID) (io.ReadCloser, error)
	GetMeta(t.ChunkID) (t.ChunkMeta, error)
	List() ([]t.ChunkID, error)
	Delete(t.ChunkID) error
	Store(chunk t.Chunk) error
}
