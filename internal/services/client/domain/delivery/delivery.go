package delivery

import (
	"context"
	"errors"
	"fmt"
	"iter"

	"dos/internal/common/transport/chunkrpc"
	t "dos/internal/common/types"
	"dos/internal/services/client/domain/progress"
	"dos/internal/services/client/io/file"
	"dos/internal/services/client/transport"

	"golang.org/x/sync/errgroup"
)

type ObjectDeliveryConfig interface {
	TransferConcurrency() int
}

type ChunkSource interface {
	Chunks() iter.Seq2[t.ChunkKey, []byte]
	Err() error
}

type ObjectDeliveryDeps struct {
	MasterT *transport.MasterTransport
	ChunkT  *chunkrpc.Transport
	Config  ObjectDeliveryConfig

	ObjectID t.ObjectID
}

type ObjectDelivery struct {
	objectID   t.ObjectID
	progress   *progress.ObjectProgress
	onProgress func(*progress.ObjectProgress)

	masterT *transport.MasterTransport
	chunkT  *chunkrpc.Transport

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

func (d *ObjectDelivery) Upload(ctx context.Context, source ChunkSource) error {

	if err := d.masterT.CreateObject(ctx, d.objectID); err != nil {
		return fmt.Errorf("create object: %w", err)
	}
	d.emitProgress()
	defer d.emitProgress()

	eg := errgroup.Group{}
	eg.SetLimit(d.config.TransferConcurrency())

	for chunkKey, chunkData := range source.Chunks() {  
		eg.Go(func() error {
			if err := d.uploadChunk(ctx, chunkKey, chunkData); err != nil {
				return fmt.Errorf("upload chunk failed %s: %w", chunkKey, err)
			}
			return nil 
		})
	}

	readErr := source.Err()
	uploadErr := eg.Wait()

	switch {
	case readErr != nil:  
		d.progress.Fail("read chunk failed")
		return readErr 
	case uploadErr != nil:
		d.progress.Fail("upload chunk failed")
		return uploadErr 
	default:
		d.progress.Done()
		return nil
	}
}

func (d *ObjectDelivery) uploadChunk(ctx context.Context, chunkKey t.ChunkKey, data []byte) error {

	loc, err := d.masterT.AllocateChunk(ctx, &transport.AllocateChunkCommand{
		Slot: t.ObjectSlot{
			ObjectID: d.objectID,
			ChunkKey: chunkKey,
		},
		ChunkSize: int64(len(data)),
	})
	if err != nil {
		return fmt.Errorf("alloc chunk: %w", err)
	}

	chunk := t.NewChunk(loc.ID, data)

	opt := chunkrpc.WithProgress(func(cp chunkrpc.Progress) {
		d.progress.UpdateChunk(chunkKey, cp)
		d.emitProgress()
	})

	session := d.chunkT.NewTransferSession(loc.Targets, opt)
	if _, err := session.Upload(ctx, &chunk); err != nil {
		return err
	}
	return nil
}

func (d *ObjectDelivery) Download(ctx context.Context, asm *file.ObjectAssembler) error {

	objDesc, err := d.masterT.DescribeObject(ctx, d.objectID)
	if err != nil {
		return fmt.Errorf("describe object: %w", err)
	}

	d.emitProgress()
	defer d.emitProgress()


	sink, err := asm.NewSink(objDesc.Chunks)
	if err != nil {
		return fmt.Errorf("new object writer: %w", err)
	}
	defer sink.Close()
	
	eg := errgroup.Group{}
	eg.SetLimit(d.config.TransferConcurrency())

	for _, placement := range objDesc.Chunks {
		eg.Go(func() error {
			return d.downloadChunk(ctx, placement, sink)
		})
	}
	if err := eg.Wait(); err != nil {
		d.progress.Fail("download chunk failed")
		return err
	} else {
		d.progress.Done()
	}

	return nil
}

type ChunkSink interface {
	WriteChunk(t.ChunkKey, []byte) error
}

func (d *ObjectDelivery) downloadChunk(ctx context.Context, placement t.ChunkPlacement1, sink ChunkSink) error {
	opt := chunkrpc.WithProgress(func(prog chunkrpc.Progress) {
		d.progress.UpdateChunk(placement.Slot.ChunkKey, prog)
		d.emitProgress()
	})
	session := d.chunkT.NewTransferSession(placement.Sources, opt)
	chunk, err := session.Download(ctx, placement.Meta.ID)
	if err != nil {
		return fmt.Errorf("download chunk %s: %w", placement.Meta.ID, err)
	}
	if err := sink.WriteChunk(placement.Slot.ChunkKey, chunk.Data); err != nil {
		return fmt.Errorf("write chunk %s: %w", placement.Meta.ID, err)
	}
	return nil
}

func (d *ObjectDelivery) emitProgress() {
	if d.onProgress != nil {
		d.onProgress(d.progress)
	}
}
