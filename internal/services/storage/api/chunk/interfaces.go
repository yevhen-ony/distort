package chunk

import (
	"context"
	"time"

	t "dos/internal/common/types"
	"dos/internal/services/storage/core/storage"
)

//go:generate mockgen -source=$GOFILE -destination=mocks_test.go -package=chunk

type ChunkConfig interface {
	FrameSize() int64
}

type ChunkStorage interface {
	StartUpload(context.Context, *t.ChunkMeta) (*storage.UploadSession, error)
	AcquireOpSlot(context.Context, time.Duration) (func(), error)
	LoadChunk(t.ChunkID) (t.Chunk, error)
	ScheduleForwardChunk(context.Context, t.ChunkID, []t.NodeRef) error
	DeleteChunk(context.Context, t.ChunkID) error
}

type NodeIdentity interface {
	Validate(t.NodeID) error
}
