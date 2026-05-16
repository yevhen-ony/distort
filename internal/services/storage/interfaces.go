package storage

import (
	"context"
	"io"

	t "dos/internal/common/types"
)

type ChunkStorage interface {
	Get(t.ChunkID) (io.ReadCloser, error)
	GetMeta(t.ChunkID) (t.ChunkMeta, error)
	List() ([]t.ChunkID, error)
	Delete(t.ChunkID) error
	Store(chunk t.Chunk) error
}

type HeartbeatResult struct {
	NodeUnknown bool
}

type MasterTransport interface {
	Heartbeat(context.Context,t.NodeID,t.NodeStats) (HeartbeatResult, error)
	ReportChunks(context.Context, t.NodeID, []t.StorageNodeReport) (t.ReportResult, error)
	RegisterNode(ctx context.Context, addr string) (t.NodeID, error)
}
