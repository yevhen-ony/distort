package delivery

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"dos/internal/common/transport/chunkrpc"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
	"dos/internal/services/client/domain/progress"
	"dos/internal/services/client/io/file"
	"dos/internal/services/client/transport"
)

type ObjectDeliveryConfig interface {
	TransferConcurrency() int
}

type ChunkSource interface {
	Next() (t.ChunkKey, []byte, error)
}

type ObjectDeliveryDeps struct{
	MasterT *transport.MasterTransport
	ChunkT *chunkrpc.Transport
	Config ObjectDeliveryConfig

	ObjectID t.ObjectID
}

type ObjectDelivery struct {
	objectID t.ObjectID
	progress *progress.ObjectProgress
	onProgress func(*progress.ObjectProgress)

	masterT *transport.MasterTransport
	chunkT *chunkrpc.Transport

	sem chan struct{}
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
	uploader := &ObjectDelivery{
		objectID: deps.ObjectID,
		masterT:  deps.MasterT,
		chunkT:   deps.ChunkT,
		onProgress: func(*progress.ObjectProgress) {}, // nop
		progress: progress.NewObjectProgress(deps.ObjectID),

		sem: make(chan struct{}, deps.Config.TransferConcurrency()),
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
	
	wg := sync.WaitGroup{}
	for {
		chunkKey, chunkData, err := source.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read chunk: %w", err)
		}

		d.sem <- struct{}{}
		wg.Go(func() {
			defer func() { <-d.sem }()
			d.uploadChunk(ctx, chunkKey, chunkData)
		})
	}
	wg.Wait()

	d.progress.Done = true
	d.emitProgress()
	return nil
}

func (d *ObjectDelivery) uploadChunk(ctx context.Context, chunkKey t.ChunkKey, data []byte) error {

	loc, err := d.masterT.AllocateChunk(ctx, &transport.AllocateChunkCommand{
		ObjectID:  d.objectID,
		ChunkKey:  chunkKey,
		ChunkSize: int64(len(data)),
	})
	if err != nil {
		return fmt.Errorf("alloc chunk: %w", err)
	}

	chunk := t.NewChunk(loc.ChunkID, data)

	opt := chunkrpc.WithProgress(func(cp chunkrpc.Progress) {
		d.progress.UpdateChunk(chunkKey, cp)
		d.emitProgress()
	})

	session := d.chunkT.NewTransferSession(loc.Nodes, opt)
	if _, err := session.Upload(ctx, &chunk); err != nil {
		return err
	}
	return nil
}

func (d *ObjectDelivery) Download(ctx context.Context, asm *file.ObjectAssembler) error {

	access, err := d.masterT.GetObjectAccess(ctx, d.objectID)
	if err != nil {
		return fmt.Errorf("get object access: %w", err)
	}

	d.emitProgress()
	defer d.emitProgress()

	chunkDescs := utils.Map(access.Chunks, func(p t.ChunkPlacement) t.ChunkDesc {
		return p.ChunkDesc
	})


	writer, err := asm.NewWriter(access.ObjectDesc, chunkDescs)
	if err != nil {
		return fmt.Errorf("new object writer: %w", err)
	}
	defer writer.Close()

	wg := sync.WaitGroup{}
	for _, placement := range access.Chunks {
		d.sem <- struct{}{}
		wg.Go(func() {
			defer func() { <-d.sem }()
			d.downloadChunk(ctx, placement, writer)
		})
	}
	wg.Wait()

	d.progress.Done = true
	
	return nil
}

type ChunkSink interface {
	WriteChunk(t.ChunkID, []byte) error
}

func (d *ObjectDelivery) downloadChunk(ctx context.Context, placement t.ChunkPlacement, sink ChunkSink) error {
	opt := chunkrpc.WithProgress(func(prog chunkrpc.Progress) {
		d.progress.UpdateChunk(placement.ChunkKey, prog)
		d.emitProgress()
	})
	session := d.chunkT.NewTransferSession(placement.Nodes, opt)
	chunk, err := session.Download(ctx, placement.ChunkID)
	if err != nil {
		return fmt.Errorf("download chunk %s: %w", placement.ChunkID, err)
	}
	if err := sink.WriteChunk(chunk.Meta.ID, chunk.Data); err != nil {
		return fmt.Errorf("write chunk %s: %w", placement.ChunkID, err)
	}
	return nil
}

func (d *ObjectDelivery) emitProgress() {
	if d.onProgress != nil {
		d.onProgress(d.progress)
	}
}
