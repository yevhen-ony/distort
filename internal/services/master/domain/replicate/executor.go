package replicate

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"dos/internal/common/dosctx"
	"dos/internal/common/loop"
	"dos/internal/common/queue"
	"dos/internal/common/transport/chunkrpc"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
	m "dos/internal/services/master"
)

var (
	ErrReplicationAttemptsExhausted = errors.New("all replication attempts exhausted")
)

type ReplicationConfig interface {
	ReplicationQueueLength() int
	ReplicationExecInterval() time.Duration
}

type ReplicationExecutor struct {
	chunkRepo  m.ChunkRepo
	objectRepo m.ObjectRepo
	placement  m.StorageNodePlacement
	transport  *chunkrpc.Transport
	config     ReplicationConfig

	queue  *queue.DedupQueue[t.ChunkID]
	looper *loop.Looper
}

func NewReplicationExecutor(
	chunkRepo m.ChunkRepo,
	objectRepo m.ObjectRepo,
	placement m.StorageNodePlacement,
	transport *chunkrpc.Transport,
	config ReplicationConfig,
) *ReplicationExecutor {

	queue := queue.NewDedupQueue[t.ChunkID](config.ReplicationQueueLength())
	looper := loop.NewLooper(config.ReplicationExecInterval())
	return &ReplicationExecutor{
		chunkRepo:  chunkRepo,
		objectRepo: objectRepo,
		placement:  placement,
		transport:  transport,
		config:     config,

		queue:  queue,
		looper: looper,
	}
}

func (r *ReplicationExecutor) ReplicateChunk(ctx context.Context, chunkID t.ChunkID) error {

	ctx = dosctx.WithChunkID(ctx, chunkID)

	slog.DebugContext(ctx, "do replication")

	chunk, err := r.chunkRepo.Get(ctx, chunkID)
	if err != nil {
		return fmt.Errorf("read chunk %s: %w", chunkID, err)
	}

	wantedReplicaCount, err := r.objectRepo.GetReplication(ctx, chunk.ObjectID)
	if err != nil {
		return fmt.Errorf("read object %s: %w", chunk.ObjectID, err)
	}

	count := wantedReplicaCount - chunk.ReplicaCount
	if count == 0 {
		return nil
	}
	slog.DebugContext(ctx, "replication decision",
		"wanted", wantedReplicaCount,
		"actual", chunk.ReplicaCount,
		"delta", count,
	)

	if chunk.ReplicaCount == 0 {
		return nil
	}

	r.chunkRepo.Touch(ctx, chunk.Meta.ID)

	if count > 0 {
		_, err = r.AddReplica(ctx, chunk.Meta, count)
		if err != nil {

			return fmt.Errorf("replicate chunk %s: %w", chunkID, err)
		}
		return nil
	}

	// count < 0
	err = r.DeleteReplica(ctx, chunk.Meta, -count)
	if err != nil {
		return fmt.Errorf("delete chunk %s: %w", chunkID, err)
	}
	return nil
}

func (r *ReplicationExecutor) AddReplica(ctx context.Context, meta t.ChunkMeta, count int) (t.NodeID, error) {

	ctx = dosctx.WithOperation(ctx, "add")

	slog.DebugContext(ctx, "add replica")

	sources, err := r.placement.GetChunkNodes(ctx, meta.ID)
	if err != nil {
		slog.ErrorContext(ctx, "list chunk's nodes failed")
		return "", fmt.Errorf("list chunk's nodes: %w", err)
	}
	if len(sources) == 0 {
		slog.ErrorContext(ctx, "no replication sources found")
		return "", errors.New("no replication sources found")
	}

	targets, err := r.placement.GetCandidates(ctx, m.CandidateNodesQuery{
		MinFreeBytes: meta.Digest.Size,
		ExcludeChunk: meta.ID,
		MaxCount:     count,
	})
	if err != nil {
		slog.ErrorContext(ctx, "find candidate nodes failed", "error", err)
		return "", fmt.Errorf("find candidate nodes: %w", err)
	}
	if len(targets) == 0 {
		slog.ErrorContext(ctx, "no candidate nodes found")
		return "", fmt.Errorf("no candidate nodes found for %s", meta.ID)
	}

	for _, source := range utils.RandomSelect(sources, len(sources)) {

		err = r.transport.ReplicateChunk(ctx, meta.ID, source, targets)
		if err != nil {
			slog.ErrorContext(ctx, "replicate chunk failed", "source", source.ID, "error", err)
			continue
		}
		return source.ID, nil
	}
	return "", ErrReplicationAttemptsExhausted
}

func (r *ReplicationExecutor) DeleteReplica(ctx context.Context, meta t.ChunkMeta, count int) error {

	ctx = dosctx.WithOperation(ctx, "delete")

	slog.DebugContext(ctx, "delete replica")

	nodeRefs, err := r.placement.GetChunkNodes(ctx, meta.ID)
	if err != nil {
		slog.ErrorContext(ctx, "get chunk nodes while deleting replica", "error", err)
		return fmt.Errorf("get chunk nodes %s: %w", meta.ID, err)
	}

	var errs []error
	for _, nodeRef := range utils.RandomSelect(nodeRefs, count) {
		err = r.transport.DeleteChunk(ctx, meta.ID, nodeRef)
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

func (r *ReplicationExecutor) RunReplicationIteration(ctx context.Context) {
	chunkIDs := r.queue.Drain()
	if len(chunkIDs) == 0 {
		return
	}
	slog.DebugContext(ctx, "replicate chunks", "count", len(chunkIDs))
	for _, chunkID := range chunkIDs {
		r.ReplicateChunk(ctx, chunkID)
	}
}

func (r *ReplicationExecutor) RunLoop(ctx context.Context) {
	ctx = dosctx.WithService(ctx, "replication_executor")
	r.looper.Run(ctx, r.RunReplicationIteration)
}

func (r *ReplicationExecutor) Schedule(ctx context.Context, chunkID t.ChunkID) {
	r.queue.Enq(ctx, chunkID)
}

func (r *ReplicationExecutor) Flush(_ context.Context) {
	r.looper.Flush()
}
