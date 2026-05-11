package reconcile

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"dos/internal/common/dosctx"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
	m "dos/internal/services/master"
	"dos/internal/services/master/transport"
)

var (
	ErrReplicationAttemptsExhausted = errors.New("all replication attempts exhausted")
)

type ReconcileConfig interface {
	ReconcileQueueLength() int
}

type ReconcileWorker struct {
	chunkRepo  m.ChunkRepo
	objectRepo m.ObjectRepo
	placement m.StorageNodePlacement

	transport *transport.Storage

	queue *Queue
}

func NewReconcileWorker(
	chunkRepo m.ChunkRepo,
	objectRepo m.ObjectRepo,
	placement m.StorageNodePlacement,
	transport *transport.Storage,
	config ReconcileConfig,
) *ReconcileWorker {
	return &ReconcileWorker{
		chunkRepo: chunkRepo,
		objectRepo: objectRepo,
		placement: placement,
		transport: transport,
		queue: NewQueue(config.ReconcileQueueLength()),
	}
}

func (s *ReconcileWorker) ReconcileChunk(ctx context.Context, chunkID t.ChunkID) error {

	ctx = dosctx.WithChunkID(ctx, chunkID)
	ctx = dosctx.WithService(ctx, "reconcile")

	chunk, err := s.chunkRepo.Get(ctx, chunkID)
	if err != nil {
		return fmt.Errorf("read chunk %s: %w", chunkID, err)
	}

	wantedReplicaCount, err := s.objectRepo.GetReplication(ctx, chunk.ObjectID)
	if err != nil {
		return fmt.Errorf("read object %s: %w", chunk.ObjectID, err)
	}

	count := wantedReplicaCount - chunk.ReplicaCount
	if count == 0 {
		return nil
	}

	if chunk.ReplicaCount == 0 {
		return nil
	}

	if count > 0 {
		_, err = s.ReplicateChunk(ctx, chunk.ChunkMeta, count)
		if err != nil {
			return fmt.Errorf("replicate chunk %s: %w", chunkID, err)
		}
		return nil
	}

	// count < 0
	err = s.DeleteReplica(ctx, chunk.ChunkMeta, -count)
	if err != nil {
		return fmt.Errorf("delete chunk %s: %w", chunkID, err)
	}
	return nil
}

func (s *ReconcileWorker) ReplicateChunk(ctx context.Context, meta t.ChunkMeta, count int) (t.NodeID, error) {

	ctx = dosctx.WithOperation(ctx, "replicate")

	sources, err := s.placement.GetChunkNodes(ctx, meta.ID)
	if err != nil {
	}
	if len(sources) == 0 {
		return "", fmt.Errorf("no replication sources found for %s", meta.ID)
	}

	targets, err := s.placement.GetCandidates(ctx, m.CandidateNodesQuery{
		MinFreeBytes: meta.Digest.Size,
		ExcludeChunk: meta.ID,
		MaxCount: count,
	})
	if err != nil {
		slog.ErrorContext(ctx, "find candidate nodes failed", "error", err)
		return "", fmt.Errorf("find candidate nodes: %w", err)
	}

	for _, source := range utils.RandomSelect(sources, len(sources)) {

		err = s.transport.ReplicateChunk(ctx, meta.ID, source, targets)
		if err != nil {
			slog.ErrorContext(ctx, "replicate chunk failed", "source", source.ID, "error", err)
			continue
		}
		return source.ID, nil
	}
	return "", ErrReplicationAttemptsExhausted
}

func (s *ReconcileWorker) DeleteReplica(ctx context.Context, meta t.ChunkMeta, count int) error {
	
	ctx = dosctx.WithOperation(ctx, "delete")

	nodeRefs, err := s.placement.GetChunkNodes(ctx, meta.ID)
	if err != nil {
		slog.ErrorContext(ctx, "get chunk nodes while deleting replica", "error", err )
		return fmt.Errorf("get chunk nodes %s: %w", meta.ID, err)
	}

	var errs []error
	for _, nodeRef := range utils.RandomSelect(nodeRefs, count) {
		err = s.transport.DeleteChunk(ctx, meta.ID, nodeRef)
		if err != nil {
			slog.ErrorContext(ctx, "delete replica failed", "source", nodeRef.ID, "error", err)
			errs = append(errs, fmt.Errorf(
				"delete chunk %s from node %s: %w",
				meta.ID, nodeRef.ID, err,
			))
		}
	}

	return errors.Join(errs...)
}

func (s *ReconcileWorker) RunLoop(ctx context.Context) {
	for {
		chunkID, err := s.queue.Pop(ctx)
		if err != nil {
			return
		}
		_ = s.ReconcileChunk(ctx, chunkID)
	}
}

func (s *ReconcileWorker) Enqueue(ctx context.Context, chunkID t.ChunkID) {
	s.queue.Enqueue(ctx, chunkID)
}
