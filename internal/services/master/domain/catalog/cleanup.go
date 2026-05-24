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

type CleanupConfig interface {
	CatalogCleanupInterval() time.Duration
}

type CleanupDeps struct {
	ObjectRepo m.ObjectRepo
	ChunkRepo  m.ChunkRepo
	Config     CleanupConfig
	Metrics    *CatalogMetrics
}

type CleanupService struct {
	objectRepo m.ObjectRepo
	chunkRepo  m.ChunkRepo
	metrics    *CatalogMetrics

	config CleanupConfig

	looper *loop.Looper
}

func NewCleanupService(deps CleanupDeps) (*CleanupService, error) {

	if deps.ObjectRepo == nil {
		return nil, errors.New("missing object repository")
	}
	if deps.ChunkRepo == nil {
		return nil, errors.New("missing chunk repository")
	}
	if deps.Config == nil {
		return nil, errors.New("missing config")
	}
	if deps.Metrics == nil {
		return nil, errors.New("missing metrics")
	}

	looper := loop.NewLooper(deps.Config.CatalogCleanupInterval())

	cleanup := &CleanupService{
		objectRepo: deps.ObjectRepo,
		chunkRepo:  deps.ChunkRepo,
		config:     deps.Config,
		metrics:    deps.Metrics,
		looper:     looper,
	}
	return cleanup, nil
}

func (cc *CleanupService) DeleteUnwanted(ctx context.Context) []t.ObjectID {
	removedObjectIDs := []t.ObjectID{}

	objects := cc.objectRepo.List(ctx)
	for _, object := range objects {

		if object.Replication == 0 { // unwanted

			cleaned := true
			for chunkKey, chunkID := range object.Chunks {
				if ok, _ := cc.chunkRepo.Delete(ctx, chunkID); ok {
					cc.objectRepo.DeleteChunk(ctx, t.ObjectSlot{
						ObjectID: object.ID,
						ChunkKey: chunkKey,
					})
					cc.metrics.ChunkCount.Add(-1)
				} else {
					cleaned = false
				}
			}
			if !cleaned {
				continue
			}
			if err := cc.objectRepo.Delete(ctx, object.ID); err != nil {
				slog.ErrorContext(ctx, "delete object failed", "object_id", object.ID, "error", err)
			} else {
				removedObjectIDs = append(removedObjectIDs, object.ID)
				cc.metrics.ObjectCount.Add(-1)
			}
		}
	}
	return removedObjectIDs
}

func (cc *CleanupService) RunLoop(ctx context.Context) {
	cc.looper.Run(ctx, func(ctx context.Context) {
		removed := cc.DeleteUnwanted(ctx)
		if len(removed) > 0 {
			slog.DebugContext(ctx, "removed objects", "count", len(removed), "object_ids", removed)
		}
	})
}
