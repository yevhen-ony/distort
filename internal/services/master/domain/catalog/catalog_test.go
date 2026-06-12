package catalog

import (
	"context"
	"testing"

	"dos/internal/common/metrics"
	t "dos/internal/common/types"
	"dos/internal/services/master/domain/object"
	"dos/internal/services/master/repo"

	"github.com/stretchr/testify/require"
)

func TestCatalogService_AddChunk(tt *testing.T) {
	ctx := context.Background()
	catalog := newTestCatalogService(tt)

	objectID := t.ObjectID("object-1")
	slot := t.ObjectSlot{ObjectID: objectID, ChunkKey: "000001"}

	require.NoError(tt, catalog.CreateObject(ctx, objectID, 2))

	chunkID, err := catalog.AddChunk(ctx, slot, 5)
	require.NoError(tt, err)
	require.NotEmpty(tt, chunkID)

	exists, err := catalog.ExistsChunk(ctx, slot)
	require.NoError(tt, err)
	require.True(tt, exists)

	gotChunkID, err := catalog.GetChunkID(ctx, slot)
	require.NoError(tt, err)
	require.Equal(tt, chunkID, gotChunkID)

	chunk, err := catalog.GetChunk(ctx, chunkID)
	require.NoError(tt, err)
	require.Equal(tt, slot, chunk.Slot)

	obj, err := catalog.GetObject(ctx, objectID)
	require.NoError(tt, err)
	require.Equal(tt, chunkID, obj.Chunks[slot.ChunkKey])
}

func newTestCatalogService(tt *testing.T) *CatalogService {
	tt.Helper()

	objectRepo := repo.NewInMemObjectRepo()

	applier, err := object.NewLocalCommandApplier(objectRepo)
	require.NoError(tt, err)

	submitter, err := object.NewLocalCommandSubmitter(applier)
	require.NoError(tt, err)

	writer, err := object.NewObjectWriterImpl(submitter)
	require.NoError(tt, err)

	authority, err := object.NewAuthority(object.AuthorityDeps{
		Reader: objectRepo,
		Writer: writer,
	})
	require.NoError(tt, err)

	service, err := NewCatalogService(CatalogDeps{
		ObjectAuthority: authority,
		ChunkRepository: repo.NewInMemChunkRepo(),
		Metrics:         NewCatalogMetrics(metrics.NopProvider{}),
	})
	require.NoError(tt, err)

	return service
}
