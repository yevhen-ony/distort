package storage

import (
	"context"
	t "dos/internal/common/types"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestStorageService_Start_BuildsCatalogAndReportsChunks(tt *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(tt)

	inventory := NewMockInventory(ctrl)
	storageBE := NewMockChunkStorage(ctrl)
	reporter := NewMockReporter(ctrl)

	chunk := t.NewChunk("chunk-1", []byte("hello"))
	service := &StorageService{
		inventory: inventory,
		storageBE: storageBE,
		reporter:  reporter,
	}

	inventory.EXPECT().
		BuildCatalog(ctx, storageBE).
		Return(nil)

	inventory.EXPECT().
		ListIDs().
		Return([]t.ChunkID{chunk.Meta.ID})

	inventory.EXPECT().
		Stage(chunk.Meta.ID).
		Return(chunk.Meta, nil)

	reporter.EXPECT().
		Report(ctx, t.NewReplicaStaged(chunk.Meta).ToRecord())

	reporter.EXPECT().Flush(ctx)

	err := service.Start(ctx)
	require.NoError(tt, err)
}
