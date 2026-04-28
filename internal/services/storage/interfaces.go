package storage

import (
	"context"
	"io"
	"time"

	"dos/internal/common/digest"
	t "dos/internal/common/types"
)

type ChunkWriter interface {
	io.WriteCloser
	Digest() digest.Digest
	Commit(t.ChunkID) (time.Time, error)
}

type ChunkStorage interface {
	Get(chunkID t.ChunkID) (io.ReadCloser, error)
	GetMeta(chunkID t.ChunkID) (*t.ChunkMeta, error)
	NewWriter() (ChunkWriter, error)
	GetAllIDs() ([]t.ChunkID, error)
}

type HeartbeatResult struct {
	NodeUnknown bool
}

type MasterTransport interface {
	Heartbeat(context.Context,t.NodeID,t.NodeStats) (HeartbeatResult, error)
	ReportChunkStorage(context.Context, t.NodeID, []t.ChunkDesc) ([]t.ChunkStorageReject, error)
	RegisterStorageNode(ctx context.Context, addr string) (t.NodeID, error)
}
