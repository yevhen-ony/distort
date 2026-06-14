package storage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"dos/internal/common/dosctx"
	t "dos/internal/common/types"
	s "dos/internal/services/storage"
)

func (cs *StorageService) StartUpload(
	ctx context.Context,
	meta *t.ChunkMeta,
) (s.UploadSession, error) {
	if cs.inventory.Has(meta.ID) {
		return nil, s.ErrChunkConflict
	}

	release, err := cs.AcquireOpSlot(ctx, defaultOpSlotAcquireTimeout)
	if err != nil {
		return nil, err
	}

	start := time.Now()

	session := &UploadSession{
		id:   meta.ID,
		data: make([]byte, meta.Digest.Size),
		onCommit: func(ctx context.Context, chunk t.Chunk) error {
			defer release()
			err := cs.commitUpload(ctx, chunk, meta)
			if err != nil {
				cs.metrics.UploadsFailedDuration.Observe(time.Since(start).Seconds())
			} else {
				cs.metrics.UploadsSuccessDuration.Observe(time.Since(start).Seconds())
			}
			return err
		},
		onAbort: func() error {
			defer release()
			cs.metrics.UploadsFailedDuration.Observe(time.Since(start).Seconds())
			return nil
		},
	}
	return session, nil
}

func (cs *StorageService) commitUpload(
	ctx context.Context,
	chunk t.Chunk,
	meta *t.ChunkMeta,
) error {

	ctx = dosctx.WithOperation(ctx, "commit_upload")

	if err := meta.Digest.Match(&chunk.Meta.Digest); err != nil {
		return err
	}

	if err := cs.storageBE.Store(chunk); err != nil {
		return fmt.Errorf("store chunk: %w", err)
	}

	if err := cs.inventory.Add(meta); err != nil {
		if err := cs.storageBE.Delete(meta.ID); err != nil {
			slog.ErrorContext(ctx, "rollback failed", "error", err)
		}
		return err
	}

	cs.StageAndReportOne(ctx, meta.ID)
	return nil
}
