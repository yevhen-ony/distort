package catalog

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"dos/internal/common/loop"
	t "dos/internal/common/types"
	m "dos/internal/services/master"
)

type CatalogCleanupConfig interface {
	CatalogCleanupInterval() time.Duration
}

type CatalogCleanup struct {
	objectRepo m.ObjectRepo
	chunkRepo  m.ChunkRepo

	config CatalogCleanupConfig

	looper *loop.Looper
	metrics *CatalogMetrics
}

var (
	ErrMissingObjectRepo = errors.New("missing object repo")
	ErrMissingChunkRepo  = errors.New("missing chunk repo")
	ErrMissingConfig     = errors.New("missing config")
)

func NewCatalogCleanup(
	objectRepo m.ObjectRepo, chunkRepo m.ChunkRepo, config CatalogCleanupConfig,
) (*CatalogCleanup, error) {

	if objectRepo == nil {
		return nil, ErrMissingObjectRepo
	}
	if chunkRepo == nil {
		return nil, ErrMissingChunkRepo
	}
	if config == nil {
		return nil, ErrMissingConfig
	}

	looper := loop.NewLooper(config.CatalogCleanupInterval())

	cleanup := &CatalogCleanup{
		objectRepo: objectRepo,
		chunkRepo:  chunkRepo,
		config:     config,
		looper:     looper,
	}
	return cleanup, nil
}

func (cc *CatalogCleanup) RemoveUnwanted(ctx context.Context) []t.ObjectID {
	removedObjectIDs := []t.ObjectID{}

	objects := cc.objectRepo.List(ctx)
	for _, object := range objects {

		if object.Replication == 0 { // unwanted

			cleaned := true
			for chunkKey, chunkID := range object.Chunks {
				if cc.chunkRepo.DeleteWithNoReplicas(ctx, chunkID) {
					cc.objectRepo.RemoveChunk(ctx, object.ID, chunkKey)
					cc.metrics.ChunkCount.Add(-1)
				} else {
					cleaned = false
				}
			}
			if !cleaned {
				continue
			}
			if err := cc.objectRepo.DeleteObject(ctx, object.ID); err != nil {
				slog.ErrorContext(ctx, "delete object failed", "object_id", object.ID, "error", err)
			} else {
				removedObjectIDs = append(removedObjectIDs, object.ID)
				cc.metrics.ObjectCount.Add(-1)
			}
		}
	}
	return removedObjectIDs
}

func (cc *CatalogCleanup) RunLoop(ctx context.Context) {
	cc.looper.Run(ctx, func(ctx context.Context) {
		removed := cc.RemoveUnwanted(ctx)
		if len(removed) > 0 {
			slog.DebugContext(ctx, "removed objects", "count", len(removed), "object_ids", removed)
		}
	})
}
