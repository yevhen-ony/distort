package catalog

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"dos/internal/common/dosctx"
	"dos/internal/common/loop"
	t "dos/internal/common/types"
	m "dos/internal/services/master"
)

type CleanupConfig interface {
	CatalogCleanupInterval() time.Duration
}

type CleanupDeps struct {
	ObjectAuthority m.ObjectRW 
	ChunkRepository m.ChunkRepo
	Config          CleanupConfig
	Metrics         *CatalogMetrics
}

type CleanupService struct {
	objects m.ObjectRW
	chunks  m.ChunkRepo
	metrics *CatalogMetrics

	config CleanupConfig

	looper *loop.Looper
}

func NewCleanupService(deps CleanupDeps) (*CleanupService, error) {

	if deps.ObjectAuthority == nil {
		return nil, errors.New("missing object repository")
	}
	if deps.ChunkRepository == nil {
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
		objects: deps.ObjectAuthority,
		chunks:  deps.ChunkRepository,
		config:  deps.Config,
		metrics: deps.Metrics,
		looper:  looper,
	}
	return cleanup, nil
}

func (cs *CleanupService) ReconcileChunks(ctx context.Context) error {
	ctx = dosctx.WithOperation(ctx, "reconcile_chunks")

	for _, chunk := range cs.chunks.List(ctx) {
		slot := chunk.Slot
		chunkID, err := cs.objects.GetChunk(ctx, chunk.Slot)
		if errors.Is(err, m.ErrObjectNotFound) || errors.Is(err, m.ErrChunkKeyNotFound) { 
			slog.ErrorContext(ctx, "missing slot",
				"chunk_id", chunk.Meta.ID,
				"object_id", slot.ObjectID,
				"chunk_key", slot.ChunkKey,
			)
			cs.chunks.Drop(ctx, chunk.Meta.ID)
			continue
		}
		if err != nil {
			slog.ErrorContext(ctx, "access chunk", "chunk_id", chunk.Meta.ID, "error", err)
			continue
		}
		if chunkID != chunk.Meta.ID {
			slog.ErrorContext(ctx, "wrong chunk",
				"wanted_chunk_id", chunk.Meta.ID,
				"object_id", slot.ObjectID,
				"chunk_key", slot.ChunkKey,
				"actual_chunk_id", chunkID,
			)
			cs.chunks.Drop(ctx, chunk.Meta.ID)
			continue
		}
	}

	var reconcileErr error
	for _, object := range cs.objects.List(ctx) {
		for chunkKey, chunkID := range object.Chunks {
			err := cs.chunks.Create(ctx, chunkID, t.ObjectSlot{ 
				ChunkKey: chunkKey,
				ObjectID: object.ID,
			})
			if err != nil && !errors.Is(err, m.ErrChunkExists) {
				slog.ErrorContext(ctx, "create chunk",
					"chunk_id", chunkID,
					"chunk_key", chunkKey,
					"object_id", object.ID,
					"error", err)
				reconcileErr = errors.New("reconcile incomplete")
			}
		}
	}
	return reconcileErr
}

func (cs *CleanupService) DeleteUnwanted(ctx context.Context) []t.ObjectID {
	removedObjectIDs := []t.ObjectID{}

	objects := cs.objects.List(ctx)
	for _, object := range objects {

		if object.Replication == 0 { // unwanted

			cleaned := true
			for chunkKey, chunkID := range object.Chunks {
				if ok, _ := cs.chunks.Delete(ctx, chunkID); ok {
					cs.objects.DeleteChunk(ctx, t.ObjectSlot{
						ObjectID: object.ID,
						ChunkKey: chunkKey,
					})
					cs.metrics.ChunkCount.Add(-1)
				} else {
					cleaned = false
				}
			}
			if !cleaned {
				continue
			}
			if err := cs.objects.Delete(ctx, object.ID); err != nil {
				slog.ErrorContext(ctx, "delete object failed", "object_id", object.ID, "error", err)
			} else {
				removedObjectIDs = append(removedObjectIDs, object.ID)
				cs.metrics.ObjectCount.Add(-1)
			}
		}
	}
	return removedObjectIDs
}

func (cs *CleanupService) RunLoop(ctx context.Context) {
	
	_ = cs.ReconcileChunks(ctx)

	go cs.looper.Run(ctx, func(ctx context.Context) {
		removed := cs.DeleteUnwanted(ctx)
		if len(removed) > 0 {
			slog.DebugContext(ctx, "removed objects", "count", len(removed), "object_ids", removed)
		}
	})
}
