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

type ExecutorConfig interface {
	ReplicationQueueLength() int
	ReplicationExecInterval() time.Duration
}

type ExecutorDeps struct {
	ChunkRepo  m.ChunkRepo
	ObjectRepo m.ObjectRepo
	Placement  m.StorageNodePlacement
	ChunkT     *chunkrpc.Transport
	Config     ExecutorConfig
	Metrics    *ExecutorMetrics
}

type ExecutorService struct {
	chunkRepo  m.ChunkRepo
	objectRepo m.ObjectRepo
	placement  m.StorageNodePlacement
	chunkT     *chunkrpc.Transport

	metrics *ExecutorMetrics
	config  ExecutorConfig

	queue  *queue.DedupQueue[t.ChunkID]
	looper *loop.Looper
}

func NewExecutorService(deps ExecutorDeps) (*ExecutorService, error) {
	if deps.ChunkRepo == nil {
		return nil, errors.New("missing chunk repository")
	}
	if deps.ObjectRepo == nil {
		return nil, errors.New("missing object repository")
	}
	if deps.Placement == nil {
		return nil, errors.New("missing placement service")
	}
	if deps.ChunkT == nil {
		return nil, errors.New("missing chunk placement")
	}
	if deps.Config == nil {
		return nil, errors.New("missing config")
	}
	if deps.Metrics == nil {
		return nil, errors.New("missing metrics")
	}

	config := deps.Config
	queue := queue.NewDedupQueue[t.ChunkID](config.ReplicationQueueLength())
	looper := loop.NewLooper(config.ReplicationExecInterval())
	service := &ExecutorService{
		chunkRepo:  deps.ChunkRepo,
		objectRepo: deps.ObjectRepo,
		placement:  deps.Placement,
		chunkT:     deps.ChunkT,
		metrics:    deps.Metrics,

		config: deps.Config,

		queue:  queue,
		looper: looper,
	}
	return service, nil
}

func (r *ExecutorService) ReplicateChunk(ctx context.Context, chunk m.Chunk) error {

	ctx = dosctx.WithChunkID(ctx, chunk.Meta.ID)

	count, err := r.decideReplication(ctx, chunk)
	if err != nil {
		return fmt.Errorf("decision: %w", err)
	}
	if count == 0 {
		return nil
	}

	if count > 0 {
		_, err = r.addReplica(ctx, chunk.Meta, count)
		if err != nil {
			return fmt.Errorf("add replica %s: %w", chunk.Meta.ID, err)
		}
		return nil
	}

	// count < 0
	err = r.deleteReplica(ctx, chunk.Meta, -count)
	if err != nil {
		return fmt.Errorf("delete replica %s: %w", chunk.Meta.ID, err)
	}
	return nil
}

func (r *ExecutorService) decideReplication(ctx context.Context, chunk m.Chunk) (int, error) {

	actual := chunk.ReplicaCount
	wanted, err := r.objectRepo.GetReplication(ctx, chunk.ObjectID)
	if err != nil {
		slog.ErrorContext(ctx, "access object failed", "object_id", chunk.ObjectID)
		return 0, fmt.Errorf("read object %s: %w", chunk.ObjectID, err)
	}

	delta := wanted - actual
	if delta == 0 {
		return 0, nil
	}

	slog.DebugContext(ctx, "replication decision",
		"wanted", wanted,
		"actual", actual,
		"delta", delta,
	)

	if actual == 0 {
		slog.WarnContext(ctx, "unreachable chunk detected")
		r.metrics.UnreachableChunkObservationsTotal.Inc()
		return 0, nil
	}

	return delta, nil
}

func (r *ExecutorService) addReplica(
	ctx context.Context, meta t.ChunkMeta, count int,
) (chosen t.NodeID, err error) {

	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		if err != nil {
			r.metrics.AddReplicaFailedDuration.Observe(duration)
		} else {
			r.metrics.AddReplicaSuccessDuration.Observe(duration)
		}
	}()

	ctx = dosctx.WithOperation(ctx, "add")
	slog.DebugContext(ctx, "add replica")

	r.chunkRepo.Touch(ctx, meta.ID)

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

		err = r.chunkT.ReplicateChunk(ctx, meta.ID, source, targets)
		if err != nil {
			slog.ErrorContext(ctx, "replicate chunk failed", "source", source.ID, "error", err)
			continue
		}
		return source.ID, nil
	}
	return "", ErrReplicationAttemptsExhausted
}

func (r *ExecutorService) deleteReplica(ctx context.Context, meta t.ChunkMeta, count int) (err error) {

	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		if err != nil {
			r.metrics.DeleteReplicaFailedDuration.Observe(duration)
		} else {
			r.metrics.DeleteReplicaSuccessDuration.Observe(duration)
		}
	}()

	ctx = dosctx.WithOperation(ctx, "delete")
	slog.DebugContext(ctx, "delete replica")

	nodeRefs, err := r.placement.GetChunkNodes(ctx, meta.ID)
	if err != nil {
		slog.ErrorContext(ctx, "get chunk nodes while deleting replica", "error", err)
		return fmt.Errorf("get chunk nodes %s: %w", meta.ID, err)
	}

	var errs []error
	for _, nodeRef := range utils.RandomSelect(nodeRefs, count) {
		err = r.chunkT.DeleteChunk(ctx, meta.ID, nodeRef)
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

func (r *ExecutorService) RunReplicationIteration(ctx context.Context) {
	chunkIDs := r.queue.Drain()
	if len(chunkIDs) == 0 {
		return
	}
	slog.DebugContext(ctx, "replicate chunks", "count", len(chunkIDs))
	for _, chunkID := range chunkIDs {

		chunk, err := r.chunkRepo.Get(ctx, chunkID)
		if err != nil {
			slog.WarnContext(ctx,
				"failed to access scheduled chunk",
				"chunk_id", chunkID,
				"error", err,
			)
		}
		r.ReplicateChunk(ctx, chunk)
	}
}

func (r *ExecutorService) RunLoop(ctx context.Context) {
	ctx = dosctx.WithService(ctx, "replication_executor")
	r.looper.Run(ctx, r.RunReplicationIteration)
}

func (r *ExecutorService) Schedule(ctx context.Context, chunkID t.ChunkID) {
	r.queue.Enq(ctx, chunkID)
	r.metrics.ReplicationScheduledTotal.Inc()
}

func (r *ExecutorService) Flush(_ context.Context) {
	r.looper.Flush()
}
