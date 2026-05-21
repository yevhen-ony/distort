package replicate

import (
	"context"
	"dos/internal/common/dosctx"
	"dos/internal/common/loop"
	t "dos/internal/common/types"
	m "dos/internal/services/master"
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
	ObjectRepo  m.ObjectRepo
	ChunkRepo   m.ChunkRepo
	Replication ReplicationScheduler
	Config      PlannerConfig
}

type PlannerService struct {
	objectRepo m.ObjectRepo
	chunkRepo  m.ChunkRepo
	replicate  ReplicationScheduler

	config PlannerConfig

	looper *loop.Looper
}

func NewPlannerService(deps PlannerDeps) (*PlannerService, error) {

	if deps.ObjectRepo == nil {
		return nil, errors.New("missing object repo")
	}
	if deps.ChunkRepo == nil {
		return nil, errors.New("missing chunk repo")
	}
	if deps.Replication == nil {
		return nil, errors.New("missing replication scheduler")
	}
	if deps.Config == nil {
		return nil, errors.New("missing config")
	}
	looper := loop.NewLooper(deps.Config.ReplicationPlannerInterval())

	planner := &PlannerService{
		objectRepo: deps.ObjectRepo,
		chunkRepo:  deps.ChunkRepo,
		replicate:  deps.Replication,
		config:     deps.Config,
		looper:     looper,
	}
	return planner, nil
}

func (p *PlannerService) ScheduleStaleChunks(ctx context.Context) {
	now := time.Now()
	p.chunkRepo.ForEach(ctx, func(chunk m.Chunk) {
		desired, err := p.objectRepo.GetReplication(ctx, chunk.ObjectID)
		if err != nil {
			slog.ErrorContext(ctx,
				"read object replication for chunk failed",
				"chunk_id", chunk.Meta.ID,
				"object_id", chunk.ObjectID,
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
	p.looper.Run(ctx, p.ScheduleStaleChunks)
}

func (p *PlannerService) Flush() {
	p.looper.Flush()
}
