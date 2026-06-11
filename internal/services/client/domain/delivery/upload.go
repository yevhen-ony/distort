package delivery

import (
	"context"
	"dos/internal/common/transport/chunkrpc"
	t "dos/internal/common/types"
	"dos/internal/services/client/transport"
	"fmt"

	"golang.org/x/sync/errgroup"
)

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

	session := d.chunkT.NewUploadSession(loc.Targets, opt)
	if _, err := session.Upload(ctx, &chunk); err != nil {
		return err
	}
	return nil
}
