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

	metrics *ReplicationMetrics
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

func (r *ReplicationExecutor) ReplicateChunk(ctx context.Context, chunk m.Chunk) error {

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

func (r *ReplicationExecutor) decideReplication(ctx context.Context, chunk m.Chunk) (int, error) {

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

	if actual == 0  {
		slog.WarnContext(ctx, "unreachable chunk detected")
		r.metrics.UnreachableChunkObservationsTotal.Inc()
		return 0, nil
	}

	return delta, nil 
}


func (r *ReplicationExecutor) addReplica(
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

		err = r.transport.ReplicateChunk(ctx, meta.ID, source, targets)
		if err != nil {
			slog.ErrorContext(ctx, "replicate chunk failed", "source", source.ID, "error", err)
			continue
		}
		return source.ID, nil
	}
	return "", ErrReplicationAttemptsExhausted
}

func (r *ReplicationExecutor) deleteReplica(ctx context.Context, meta t.ChunkMeta, count int) (err error) {

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

func (r *ReplicationExecutor) RunLoop(ctx context.Context) {
	ctx = dosctx.WithService(ctx, "replication_executor")
	r.looper.Run(ctx, r.RunReplicationIteration)
}

func (r *ReplicationExecutor) Schedule(ctx context.Context, chunkID t.ChunkID) {
	r.queue.Enq(ctx, chunkID)
	r.metrics.ReplicationScheduledTotal.Inc()
}

func (r *ReplicationExecutor) Flush(_ context.Context) {
	r.looper.Flush()
}

