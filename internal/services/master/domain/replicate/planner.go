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

type ReplicationPlannerConfig interface {
	ChunkStaleAfter() time.Duration
	ReplicationPlannerInterval() time.Duration
}

type ReplicationPlanner struct {
	objectRepo m.ObjectRepo
	chunkRepo m.ChunkRepo
	replicate ReplicationScheduler

	config ReplicationPlannerConfig

	looper *loop.Looper
}

func NewReplicationPlanner(
	objectRepo m.ObjectRepo,
	chunkRepo m.ChunkRepo,
	replicate ReplicationScheduler,
	config ReplicationPlannerConfig,	
) (*ReplicationPlanner, error) {

	if objectRepo == nil {
		return nil, errors.New("missing object repo")
	}
	if chunkRepo == nil {
		return nil, errors.New("missing chunk repo")
	}
	if replicate == nil {
		return nil, errors.New("missing replication scheduler")
	}
	if config == nil {
		return nil, errors.New("missing config")
	}
	planner := &ReplicationPlanner{
		objectRepo: objectRepo,
		chunkRepo: chunkRepo,
		replicate: replicate,
		config: config,
		looper: loop.NewLooper(config.ReplicationPlannerInterval()),
	}
	return planner, nil
}

func (p *ReplicationPlanner) ScheduleStaleChunks(ctx context.Context) {
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

func (p *ReplicationPlanner) RunLoop(ctx context.Context) {
	ctx = dosctx.WithService(ctx, "replication_planner")
	p.looper.Run(ctx, p.ScheduleStaleChunks)
}

func (p *ReplicationPlanner) Flush() {
	p.looper.Flush()
}
