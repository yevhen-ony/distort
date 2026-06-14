package storage

import (
	"context"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
	s "dos/internal/services/storage"
	"fmt"
	"log/slog"
	"time"
)

func (cs *StorageService) ForwardChunk(
	ctx context.Context,
	chunkID t.ChunkID,
	targets []t.NodeRef,
) error {

	slog.DebugContext(ctx, "forward chunk")
	nodeID, err := cs.identity.GetID()
	if err != nil {
		return fmt.Errorf("access node id: %w", err)
	}

	targets = utils.Select(targets, func(r t.NodeRef) bool {
		return r.ID != nodeID
	})
	if len(targets) == 0 {
		return s.ErrNoValidTargets
	}

	slog.DebugContext(ctx, "load chunk")
	chunk, err := cs.LoadChunk(chunkID)
	if err != nil {
		return fmt.Errorf("get chunk: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, cs.config.ReplicationTimeout())
	defer cancel()

	slog.DebugContext(ctx, "send chunk")
	chosen, err := cs.SendChunk(ctx, chunk, targets)
	if err != nil {
		return fmt.Errorf("send chunk: %w", err)
	}

	targets = utils.Select(targets, func(r t.NodeRef) bool {
		return r.ID != chosen.ID
	})

	if len(targets) == 0 {
		return nil
	}

	slog.DebugContext(ctx, "handoff replicate chunk", "source", chosen.ID)
	err = cs.ReplicateChunk(ctx, chunkID, chosen, targets)
	if err != nil {
		return fmt.Errorf("replicate chunk: %w", err)
	}

	return nil
}

func (cs *StorageService) ScheduleForwardChunk(
	ctx context.Context,
	chunkID t.ChunkID,
	targets []t.NodeRef,
) error {

	if !cs.inventory.Has(chunkID) {
		return s.ErrChunkNotFound
	}

	release, err := cs.AcquireOpSlot(ctx, defaultOpSlotAcquireTimeout)
	if err != nil {
		return err
	}

	fwdCtx := context.WithoutCancel(ctx)
	go func() {
		defer release()
		_ = cs.ForwardChunk(fwdCtx, chunkID, targets)
	}()

	return nil
}

func (cs *StorageService) ReplicateChunk(
	ctx context.Context,
	chunkID t.ChunkID,
	source t.NodeRef,
	targets []t.NodeRef,
) error {

	start := time.Now()

	err := cs.chunkT.ReplicateChunk(ctx, chunkID, source, targets)
	if err != nil {
		cs.metrics.ReplicateFailedDuration.Observe(time.Since(start).Seconds())

		slog.ErrorContext(ctx, "replicate chunk failed",
			"source", source, "targets", targets, "error", err)

		cs.reporter.Report(ctx, t.NewReplicaChainFailed(chunkID, targets).ToRecord())
		cs.reporter.Flush(ctx)

		return fmt.Errorf("replicate chunk: %w", err)
	}

	cs.metrics.ReplicateSuccessDuration.Observe(time.Since(start).Seconds())
	return nil
}

func (cs *StorageService) SendChunk(
	ctx context.Context,
	chunk t.Chunk,
	targets []t.NodeRef,
) (t.NodeRef, error) {

	start := time.Now()

	session := cs.chunkT.NewUploadSession(targets)
	chosen, err := session.Upload(ctx, &chunk)
	if err != nil {
		cs.metrics.SendsFailedDuration.Observe(time.Since(start).Seconds())
		slog.ErrorContext(ctx, "push chunk failed", "targets", targets, "error", err)

		cs.reporter.Report(ctx, t.NewReplicaChainFailed(chunk.Meta.ID, targets).ToRecord())
		cs.reporter.Flush(ctx)

		return t.NodeRef{}, fmt.Errorf("upload replica: %w", err)
	}

	cs.metrics.SendsSuccessDuration.Observe(time.Since(start).Seconds())
	return chosen, nil
}
