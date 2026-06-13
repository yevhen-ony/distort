package storage

import (
	"context"
	"testing"

	t "dos/internal/common/types"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestStorageService_DeleteChunk_RemovesAndReports(tt *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(tt)

	inventory := NewMockInventory(ctrl)
	storageBE := NewMockChunkStorage(ctrl)
	reporter := NewMockReporter(ctrl)

	chunkID := t.ChunkID("chunk-1")
	service := &StorageService{
		inventory: inventory,
		storageBE: storageBE,
		reporter:  reporter,
	}

	inventory.EXPECT().Has(chunkID).Return(true)
	storageBE.EXPECT().Delete(chunkID).Return(nil)
	inventory.EXPECT().Remove(chunkID).Return(true)
	reporter.EXPECT().
		Report(gomock.Any(), t.NewReplicaDeleted(chunkID).ToRecord())

	err := service.DeleteChunk(ctx, chunkID)

	require.NoError(tt, err)
}

func TestStorageService_DeleteChunk_IgnoresMissingChunk(tt *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(tt)

	inventory := NewMockInventory(ctrl)
	storageBE := NewMockChunkStorage(ctrl)
	reporter := NewMockReporter(ctrl)

	chunkID := t.ChunkID("missing-chunk")
	service := &StorageService{
		inventory: inventory,
		storageBE: storageBE,
		reporter:  reporter,
	}

	inventory.EXPECT().Has(chunkID).Return(false)

	err := service.DeleteChunk(ctx, chunkID)
	require.NoError(tt, err)
}

// this case is plausible if multiple threads trying to delete the same chunk
func TestStorageService_DeleteChunk_NoReportWhenInventoryRejects(tt *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(tt)

	inventory := NewMockInventory(ctrl)
	storageBE := NewMockChunkStorage(ctrl)
	reporter := NewMockReporter(ctrl)

	chunkID := t.ChunkID("chunk-1")
	service := &StorageService{
		inventory: inventory,
		storageBE: storageBE,
		reporter:  reporter,
	}

	inventory.EXPECT().Has(chunkID).Return(true)
	storageBE.EXPECT().Delete(chunkID).Return(nil)
	inventory.EXPECT().Remove(chunkID).Return(false)
	// no reporter calls

	err := service.DeleteChunk(ctx, chunkID)
	require.NoError(tt, err)
}
