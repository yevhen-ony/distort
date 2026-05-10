package storage

import (
	"context"
	"io"

	"dos/internal/common/digest"
	t "dos/internal/common/types"
)

type ChunkWriter interface {
	io.WriteCloser
	Digest() *digest.Digest
	Commit(t.ChunkID) error
}

type ChunkStorage interface {
	Get(t.ChunkID) (io.ReadCloser, error)
	GetMeta(t.ChunkID) (t.ChunkMeta, error)
	NewWriter() (ChunkWriter, error)
	List() ([]t.ChunkID, error)
	Delete(t.ChunkID) error
}

type HeartbeatResult struct {
	NodeUnknown bool
}

type MasterTransport interface {
	Heartbeat(context.Context,t.NodeID,t.NodeStats) (HeartbeatResult, error)
	ReportChunks(context.Context, t.NodeID, []t.ReplicaReport) (t.ReportResult, error)
	RegisterNode(ctx context.Context, addr string) (t.NodeID, error)
}
