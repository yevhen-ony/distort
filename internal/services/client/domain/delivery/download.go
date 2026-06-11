package delivery

import (
	"context"
	"dos/internal/common/transport/chunkrpc"
	t "dos/internal/common/types"
	"dos/internal/services/client/io"
	"fmt"

	"golang.org/x/sync/errgroup"
)



func (d *ObjectDelivery) Download(ctx context.Context, asm io.ObjectAssembler) error {

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

func (d *ObjectDelivery) downloadChunk(ctx context.Context, placement t.ChunkPlacement, sink io.ObjectSink) error {
	opt := chunkrpc.WithProgress(func(prog chunkrpc.Progress) {
		d.progress.UpdateChunk(placement.Slot.ChunkKey, prog)
		d.emitProgress()
	})
	session := d.chunkT.NewDownloadSession(placement.Sources, opt)
	chunk, err := session.Download(ctx, placement.Meta.ID)
	if err != nil {
		return fmt.Errorf("download chunk %s: %w", placement.Meta.ID, err)
	}
	if err := sink.WriteChunk(placement.Slot.ChunkKey, chunk.Data); err != nil {
		return fmt.Errorf("write chunk %s: %w", placement.Meta.ID, err)
	}
	return nil
}
