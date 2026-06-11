package delivery

import (
	"context"
	"errors"
	"iter"

	"dos/internal/common/transport/chunkrpc"
	t "dos/internal/common/types"
	"dos/internal/services/client/domain/progress"
	"dos/internal/services/client/transport"
)

type ObjectDeliveryConfig interface {
	TransferConcurrency() int
}

type ChunkSource interface {
	Chunks() iter.Seq2[t.ChunkKey, []byte]
	Err() error
}

type MasterTransport interface {
	CreateObject(context.Context, t.ObjectID) error
	AllocateChunk(context.Context, *transport.AllocateChunkCommand) (*t.ChunkAllocation, error)
	DescribeObject(context.Context, t.ObjectID) (*t.ObjectDesc, error)
}

type ChunkTransport interface {
	NewDownloadSession([]t.NodeRef, ...chunkrpc.SessionOption) chunkrpc.DownloadSession
	NewUploadSession([]t.NodeRef, ...chunkrpc.SessionOption) chunkrpc.UploadSession
}

type ObjectDeliveryDeps struct {
	MasterT MasterTransport
	ChunkT  ChunkTransport
	Config  ObjectDeliveryConfig

	ObjectID t.ObjectID
}

type ObjectDelivery struct {
	objectID   t.ObjectID
	progress   *progress.ObjectProgress
	onProgress func(*progress.ObjectProgress)

	masterT MasterTransport
	chunkT  ChunkTransport

	config ObjectDeliveryConfig
}

func NewObjectDelivery(deps ObjectDeliveryDeps) (*ObjectDelivery, error) {
	if deps.ObjectID == "" {
		return nil, errors.New("missing object id")
	}
	if deps.MasterT == nil {
		return nil, errors.New("missing master transport")
	}
	if deps.ChunkT == nil {
		return nil, errors.New("missing chunk transport")
	}

	if deps.Config == nil {
		return nil, errors.New("missing config")
	}

	if deps.Config.TransferConcurrency() < 1 {
		return nil, errors.New("invalid concurrency")
	}

	uploader := &ObjectDelivery{
		objectID: deps.ObjectID,
		masterT:  deps.MasterT,
		chunkT:   deps.ChunkT,
		config:   deps.Config,

		progress:   progress.NewObjectProgress(deps.ObjectID),
		onProgress: func(*progress.ObjectProgress) {}, // nop
	}
	return uploader, nil
}

func (d *ObjectDelivery) WithProgress(h progress.ProgressHandler) {
	d.onProgress = h
}

func (d *ObjectDelivery) emitProgress() {
	if d.onProgress != nil {
		d.onProgress(d.progress)
	}
}
