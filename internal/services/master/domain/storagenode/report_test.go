package storagenode

import (
	"context"
	"testing"

	"dos/internal/common/digest"
	"dos/internal/common/metrics"
	t "dos/internal/common/types"
	"dos/internal/services/master/repo"

	"github.com/stretchr/testify/require"
)

func TestReportService_Report(tt *testing.T) {
	ctx := context.Background()
	f := newReportFixture(tt)

	node1, err := f.nodes.Register(ctx, "node-1:10001")
	require.NoError(tt, err)
	node2, err := f.nodes.Register(ctx, "node-2:10001")
	require.NoError(tt, err)

	slot := t.ObjectSlot{ObjectID: "object-1", ChunkKey: "chunk-key-1"}
	err = f.chunks.Create(ctx, "chunk-1", slot)
	require.NoError(tt, err)

	s, err := NewReportService(f.deps())
	require.NoError(tt, err)

	dgst := testDigest("checksum-1", 123)

	tt.Run("staged", func(tt *testing.T) {
		report := t.NewReplicaStaged(t.ChunkMeta{
			ID:     "chunk-1",
			Digest: dgst},
		).ToRecord()

		got, err := s.Report(ctx, node1.ID, []t.StorageNodeReport{report})
		require.NoError(tt, err)

		want := t.ReportResult{Accepted: []t.ChunkID{"chunk-1"}}
		require.Equal(tt, want, got)

		chunk, err := f.chunks.Get(ctx, "chunk-1")
		require.NoError(tt, err)
		require.Equal(tt, 1, chunk.ReplicaCount)
		require.Equal(tt, dgst, chunk.Meta.Digest)
		require.Equal(tt, []t.NodeID{node1.ID}, f.index.GetChunkNodes(ctx, "chunk-1"))
	})

	tt.Run("repeated_staged_same_node", func(tt *testing.T) {
		report := t.NewReplicaStaged(t.ChunkMeta{
			ID:     "chunk-1",
			Digest: dgst},
		).ToRecord()

		got, err := s.Report(ctx, node1.ID, []t.StorageNodeReport{report})
		require.NoError(tt, err)

		want := t.ReportResult{Accepted: []t.ChunkID{"chunk-1"}}
		require.Equal(tt, want, got)

		chunk, err := f.chunks.Get(ctx, "chunk-1")
		require.NoError(tt, err)
		require.Equal(tt, 1, chunk.ReplicaCount)
	})

	tt.Run("repeated_staged_other_node", func(tt *testing.T) {
		report := t.NewReplicaStaged(t.ChunkMeta{
			ID:     "chunk-1",
			Digest: dgst},
		).ToRecord()

		got, err := s.Report(ctx, node2.ID, []t.StorageNodeReport{report})
		require.NoError(tt, err)

		want := t.ReportResult{Accepted: []t.ChunkID{"chunk-1"}}
		require.Equal(tt, want, got)

		chunk, err := f.chunks.Get(ctx, "chunk-1")
		require.NoError(tt, err)
		require.Equal(tt, 2, chunk.ReplicaCount)

		nodes := f.index.GetChunkNodes(ctx, "chunk-1")
		require.ElementsMatch(tt, []t.NodeID{node1.ID, node2.ID}, nodes)
	})

	tt.Run("digest_mismatch", func(tt *testing.T) {
		report := t.NewReplicaStaged(t.ChunkMeta{
			ID:     "chunk-1",
			Digest: testDigest("checksum-2", 123),
		}).ToRecord()

		got, err := s.Report(ctx, node1.ID, []t.StorageNodeReport{report})
		require.NoError(tt, err)

		want := t.ReportResult{Rejected: []t.ChunkID{"chunk-1"}}
		require.Equal(tt, want, got)
	})

	tt.Run("unknown_chunk", func(tt *testing.T) {
		report := t.NewReplicaStaged(t.ChunkMeta{
			ID:     "missing-chunk",
			Digest: dgst,
		}).ToRecord()

		got, err := s.Report(ctx, node1.ID, []t.StorageNodeReport{report})
		require.NoError(tt, err)

		want := t.ReportResult{Rejected: []t.ChunkID{"missing-chunk"}}
		require.Equal(tt, want, got)
	})

	tt.Run("deleted", func(tt *testing.T) {
		report := t.NewReplicaDeleted("chunk-1").ToRecord()

		got, err := s.Report(ctx, node1.ID, []t.StorageNodeReport{report})
		require.NoError(tt, err)

		require.Empty(tt, got.Accepted)
		require.Empty(tt, got.Rejected)

		chunk, err := f.chunks.Get(ctx, "chunk-1")
		require.NoError(tt, err)
		require.Equal(tt, 1, chunk.ReplicaCount)

		nodes := f.index.GetChunkNodes(ctx, "chunk-1")
		require.Equal(tt, []t.NodeID{node2.ID}, nodes)
	})

	tt.Run("chain_failed", func(tt *testing.T) {
		report := t.NewReplicaChainFailed("chunk-1", nil).ToRecord()
		got, err := s.Report(ctx, node1.ID, []t.StorageNodeReport{report})

		require.NoError(tt, err)
		require.Empty(tt, got.Accepted)
		require.Empty(tt, got.Rejected)

		//  replication rescheduled
		require.Equal(tt, []t.ChunkID{"chunk-1"}, f.replication.chunkIDs)
	})
}

// fixture

type reportFixture struct {
	chunks      *repo.InMemChunkRepo
	nodes       *repo.InMemNodeRegistry
	index       *repo.InMemChunkNodeIndex
	replication *fakeReportReplicaScheduler
}

func newReportFixture(tt *testing.T) *reportFixture {
	tt.Helper()

	return &reportFixture{
		chunks:      repo.NewInMemChunkRepo(),
		nodes:       repo.NewInMemNodeRegistry(),
		index:       repo.NewInMemChunkNodeIndex(),
		replication: &fakeReportReplicaScheduler{},
	}
}

func (f *reportFixture) deps() ReportDeps {
	return ReportDeps{
		ChunkNodeIndex: f.index,
		ChunkRepo:      f.chunks,
		NodeRegistry:   f.nodes,
		Replication:    f.replication,
		Metrics:        NewReportMetrics(metrics.NopProvider{}),
	}
}

type fakeReportReplicaScheduler struct {
	chunkIDs []t.ChunkID
}

func (s *fakeReportReplicaScheduler) Schedule(_ context.Context, chunkID t.ChunkID) {
	s.chunkIDs = append(s.chunkIDs, chunkID)
}

func testDigest(checksum string, size int64) digest.Digest {
	return digest.Digest{
		Checksum: digest.Checksum(checksum),
		Size:     size,
	}
}
