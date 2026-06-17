package replicate

import (
	"context"
	"testing"
	"time"

	"dos/internal/common/metrics"
	t "dos/internal/common/types"
	"dos/internal/services/master/repo"

	"github.com/stretchr/testify/require"
)

func TestPlannerService_ScheduleStaleChunks(tt *testing.T) {
	tests := map[string]struct {
		wanted        int
		actual        int
		staleAfter    time.Duration
		wantScheduled bool
	}{
		"schedules_when_stale": {
			wanted:        2,
			actual:        1,
			staleAfter:    0, // always stale
			wantScheduled: true,
		},
		"no_schedule_when_replication_matches": {
			wanted:        2,
			actual:        2,
			staleAfter:    0, // always stale
			wantScheduled: false,
		},
		"no_schedule_when_chunk_not_stale": {
			wanted:        2,
			actual:        1,
			staleAfter:    time.Hour, //  never stale
			wantScheduled: false,
		},
	}

	for name, test := range tests {
		tt.Run(name, func(tt *testing.T) {
			ctx := context.Background()
			f := newPlannerFixture(tt)
			f.config.staleAfter = test.staleAfter

			slot := t.ObjectSlot{
				ObjectID: "object-1",
				ChunkKey: "chunk-key-1",
			}
			chunk := t.NewChunk("chunk-1", []byte("hello world!"))

			require.NoError(tt, f.objects.Create(ctx, slot.ObjectID, test.wanted))
			require.NoError(tt, f.chunks.Create(ctx, chunk.Meta.ID, slot))
			for range test.actual {
				require.NoError(tt, f.chunks.SetDigest(ctx, chunk.Meta.ID, chunk.Meta.Digest))
				require.NoError(tt, f.chunks.IncReplicaCount(ctx, chunk.Meta.ID))
			}

			planner, err := NewPlannerService(f.deps())
			require.NoError(tt, err)

			planner.ScheduleStaleChunks(ctx)

			if test.wantScheduled {
				require.Equal(tt, chunk.Meta.ID, f.replication.chunkID)
				require.Equal(tt, 1, f.replication.count)
			} else {
				require.Empty(tt, f.replication.chunkID)
				require.Zero(tt, f.replication.count)
			}
		})
	}
}

// fixuture

type plannerFixture struct {
	chunks      *repo.InMemChunkRepo
	objects     *repo.InMemObjectRepo
	replication *fakeReplicationScheduler
	config      fakePlannerConfig
}

func newPlannerFixture(tt *testing.T) *plannerFixture {
	tt.Helper()

	return &plannerFixture{
		chunks:      repo.NewInMemChunkRepo(),
		objects:     repo.NewInMemObjectRepo(),
		replication: &fakeReplicationScheduler{},
		config: fakePlannerConfig{
			staleAfter: 0,
			interval:   time.Hour,
		},
	}
}

func (f *plannerFixture) deps() PlannerDeps {
	return PlannerDeps{
		ObjectReader:    f.objects,
		ChunkRepository: f.chunks,
		Replication:     f.replication,
		Config:          f.config,
		Metrics:         NewPlannerMetrics(metrics.NopProvider{}),
	}
}

// fake config

type fakePlannerConfig struct {
	staleAfter time.Duration
	interval   time.Duration
}

func (c fakePlannerConfig) ChunkStaleAfter() time.Duration {
	return c.staleAfter
}

func (c fakePlannerConfig) ReplicationPlannerInterval() time.Duration {
	if c.interval == 0 {
		return time.Hour
	}
	return c.interval
}

// fake scheduler

type fakeReplicationScheduler struct {
	chunkID t.ChunkID
	count   int
}

func (s *fakeReplicationScheduler) Schedule(_ context.Context, chunkID t.ChunkID) {
	s.chunkID = chunkID
	s.count++
}
