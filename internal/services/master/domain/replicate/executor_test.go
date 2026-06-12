package replicate

import (
	"context"
	"dos/internal/common/metrics"
	t "dos/internal/common/types"
	m "dos/internal/services/master"
	"dos/internal/services/master/repo"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestExecutorService_ReplicateChunk(tt *testing.T) {
	tests := map[string]struct {
		wanted        int
		actual        int
		wantAction    string
		wantCallCount int
	}{
		"replicate_when_actual_lt_wanted": {
			wanted:        2,
			actual:        1,
			wantAction:    "replicate",
			wantCallCount: 1,
		},
		"delete_when_actual_gt_wanted": {
			wanted:        1,
			actual:        2,
			wantAction:    "delete",
			wantCallCount: 1,
		},
		"do_nothing_when_actual_eq_wanted": {
			wanted:        2,
			actual:        2,
			wantAction:    "",
			wantCallCount: 0,
		},
		"do_nothing_when_actual_eq_0": {
			wanted:        2,
			actual:        0,
			wantAction:    "",
			wantCallCount: 0,
		},
	}

	for name, test := range tests {
		tt.Run(name, func(tt *testing.T) {
			ctx := context.Background()
			f := newExecutorFixture(tt)

			slot := t.ObjectSlot{ObjectID: "object-1", ChunkKey: "chunk-key-1"}
			meta := t.NewChunk("chunk-1", []byte("hello world!")).Meta

			// populate repos
			require.NoError(tt, f.objects.Create(ctx, slot.ObjectID, test.wanted))
			require.NoError(tt, f.chunks.Create(ctx, meta.ID, slot))
			for range test.actual {
				require.NoError(tt, f.chunks.SetDigest(ctx, meta.ID, meta.Digest))
				require.NoError(tt, f.chunks.IncReplicaCount(ctx, meta.ID))
			}

			executor, err := NewExecutorService(f.deps())
			require.NoError(tt, err)

			err = executor.ReplicateChunk(ctx, meta.ID)

			require.NoError(tt, err)
			require.Equal(tt, test.wantCallCount, f.chunkT.count)
			if test.actual > 0 && test.wanted != test.actual {
				require.Equal(tt, test.wantAction, f.chunkT.action)
				require.Equal(tt, meta.ID, f.chunkT.chunkID)
			} else {
				require.Empty(tt, f.chunkT.action)
				require.Empty(tt, f.chunkT.chunkID)
			}
		})
	}
}

type executorFixture struct {
	chunks    *repo.InMemChunkRepo
	objects   *repo.InMemObjectRepo
	placement *fakePlacement
	chunkT    *fakeChunkTransport
}

func newExecutorFixture(tt *testing.T) *executorFixture {
	tt.Helper()

	return &executorFixture{
		chunks:  repo.NewInMemChunkRepo(),
		objects: repo.NewInMemObjectRepo(),
		placement: &fakePlacement{
			sources: []t.NodeRef{{ID: "source-1", Addr: "source:1"}},
			targets: []t.NodeRef{{ID: "target-1", Addr: "target:1"}},
		},
		chunkT: &fakeChunkTransport{},
	}
}

func (f *executorFixture) deps() ExecutorDeps {
	return ExecutorDeps{
		ChunkRepository: f.chunks,
		ObjectReader:    f.objects,
		Placement:       f.placement,
		ChunkTransport:  f.chunkT,
		Config:          fakeExecutorConfig{},
		Metrics:         NewExecutorMetrics(metrics.NopProvider{}),
	}
}

// fake placement

type fakePlacement struct {
	sources []t.NodeRef
	targets []t.NodeRef
}

func (p *fakePlacement) GetChunkNodes(context.Context, t.ChunkID) ([]t.NodeRef, error) {
	return p.sources, nil
}

func (p *fakePlacement) GetCandidates(context.Context, m.CandidateNodesQuery) ([]t.NodeRef, error) {
	return p.targets, nil
}

// fake chunk transport

type fakeChunkTransport struct {
	action  string
	chunkID t.ChunkID
	count   int
}

func (tr *fakeChunkTransport) ReplicateChunk(
	_ context.Context,
	chunkID t.ChunkID,
	_ t.NodeRef,
	targets []t.NodeRef,
) error {
	tr.action = "replicate"
	tr.chunkID = chunkID
	tr.count = len(targets)
	return nil
}

func (tr *fakeChunkTransport) DeleteChunk(
	_ context.Context,
	chunkID t.ChunkID,
	_ t.NodeRef,
) error {
	tr.action = "delete"
	tr.chunkID = chunkID
	tr.count++
	return nil
}

// fake config

type fakeExecutorConfig struct{}

func (fakeExecutorConfig) ReplicationQueueLength() int            { return 10 }
func (fakeExecutorConfig) ReplicationExecInterval() time.Duration { return time.Hour }
