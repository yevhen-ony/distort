package catalog

import (
	"context"
	"testing"
	"time"

	"dos/internal/common/metrics"
	t "dos/internal/common/types"
	"dos/internal/services/master/repo"

	"github.com/stretchr/testify/require"
)

type testCleanupConfig struct{}

func (testCleanupConfig) CatalogCleanupInterval() time.Duration {
	return time.Hour
}

func TestCleanupService_ReconcileChunks(tt *testing.T) {
	ctx := context.Background()
	fixture := newCleanupFixture(tt)

	slot := t.ObjectSlot{ObjectID: "object-1", ChunkKey: "000001"}
	chunkID := t.ChunkID("chunk-1")

	// modify objects only
	require.NoError(tt, fixture.objects.Create(ctx, slot.ObjectID, 1))
	require.NoError(tt, fixture.objects.AddChunk(ctx, slot, chunkID))

	// apply reconcile
	err := fixture.cleanup.ReconcileChunks(ctx)
	require.NoError(tt, err)

	// changes reflected in chunks
	chunk, err := fixture.chunks.Get(ctx, chunkID)
	require.NoError(tt, err)
	require.Equal(tt, slot, chunk.Slot)
}

func TestCleanupService_DeleteUnwanted(tt *testing.T) {
	ctx := context.Background()
	fixture := newCleanupFixture(tt)

	objectID := t.ObjectID("object-1")
	slot := t.ObjectSlot{ObjectID: objectID, ChunkKey: "000001"}
	chunkID := t.ChunkID("chunk-1")

	require.NoError(tt, fixture.objects.Create(ctx, objectID, 0))
	require.NoError(tt, fixture.objects.AddChunk(ctx, slot, chunkID))
	require.NoError(tt, fixture.chunks.Create(ctx, chunkID, slot))

	removed := fixture.cleanup.DeleteUnwanted(ctx)
	require.Equal(tt, []t.ObjectID{objectID}, removed)

	_, err := fixture.objects.Get(ctx, objectID)
	require.Error(tt, err)

	_, err = fixture.chunks.Get(ctx, chunkID)
	require.Error(tt, err)
}

type cleanupFixture struct {
	objects *repo.InMemObjectRepo
	chunks  *repo.InMemChunkRepo
	cleanup *CleanupService
}

func newCleanupFixture(tt *testing.T) cleanupFixture {
	tt.Helper()

	objects := repo.NewInMemObjectRepo()
	chunks := repo.NewInMemChunkRepo()

	cleanup, err := NewCleanupService(CleanupDeps{
		ObjectAuthority: objects,
		ChunkRepository: chunks,
		Config:          testCleanupConfig{},
		Metrics:         NewCatalogMetrics(metrics.NopProvider{}),
	})
	require.NoError(tt, err)

	return cleanupFixture{
		objects: objects,
		chunks:  chunks,
		cleanup: cleanup,
	}
}
