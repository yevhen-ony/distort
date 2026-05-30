package replicate

import (
	"context"
	"dos/internal/common/dosctx"
	"dos/internal/common/loop"
	t "dos/internal/common/types"
	m "dos/internal/services/master"
	"dos/internal/services/master/domain/object"
	"errors"
	"log/slog"
	"time"
)

type ReplicationScheduler interface {
	Schedule(context.Context, t.ChunkID)
}

type PlannerConfig interface {
	ChunkStaleAfter() time.Duration
	ReplicationPlannerInterval() time.Duration
}

type PlannerDeps struct {
	ObjectReader    object.ObjectReader
	ChunkRepository m.ChunkRepo
	Replication     ReplicationScheduler
	Config          PlannerConfig
	Metrics         *PlannerMetrics
}

type PlannerService struct {
	objects   object.ObjectReader
	chunks    m.ChunkRepo
	replicate ReplicationScheduler

	config  PlannerConfig
	metrics *PlannerMetrics

	looper *loop.Looper
}

func NewPlannerService(deps PlannerDeps) (*PlannerService, error) {

	if deps.ObjectReader == nil {
		return nil, errors.New("missing object repo")
	}
	if deps.ChunkRepository == nil {
		return nil, errors.New("missing chunk repo")
	}
	if deps.Replication == nil {
		return nil, errors.New("missing replication scheduler")
	}
	if deps.Config == nil {
		return nil, errors.New("missing config")
	}
	if deps.Metrics == nil {
		return nil, errors.New("missing metrics")
	}

	looper := loop.NewLooper(deps.Config.ReplicationPlannerInterval())
	planner := &PlannerService{
		objects:   deps.ObjectReader,
		chunks:    deps.ChunkRepository,
		replicate: deps.Replication,
		config:    deps.Config,
		metrics:   deps.Metrics,

		looper:    looper,
	}
	return planner, nil
}

func (p *PlannerService) ScheduleStaleChunks(ctx context.Context) {

	p.metrics.PlannerIterationsTotal.Inc()

	now := time.Now()
	p.chunks.ForEach(ctx, func(chunk m.Chunk) {
		desired, err := p.objects.GetReplication(ctx, chunk.Slot.ObjectID)
		if err != nil {
			slog.ErrorContext(ctx,
				"read object replication for chunk failed",
				"chunk_id", chunk.Meta.ID,
				"object_id", chunk.Slot.ObjectID,
				"error", err,
			)
			return
		}
		if chunk.ReplicaCount == desired {
			return
		}
		staleAt := chunk.LastTouchedAt.Add(p.config.ChunkStaleAfter())
		if staleAt.Before(now) {
			p.replicate.Schedule(ctx, chunk.Meta.ID)
		}
	})
}

func (p *PlannerService) RunLoop(ctx context.Context) {
	ctx = dosctx.WithService(ctx, "replication_planner")
	go p.looper.Run(ctx, p.ScheduleStaleChunks)
}

func (p *PlannerService) Flush() {
	p.looper.Flush()
}
