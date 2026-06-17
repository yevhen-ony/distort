package storage

import (
	"context"
	"testing"

	"dos/internal/common/metrics"
	t "dos/internal/common/types"
	s "dos/internal/services/storage"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestStorageService_CommitUpload_StoresAddsAndReports(tt *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(tt)

	inventory := NewMockInventory(ctrl)
	storageBE := NewMockChunkStorage(ctrl)
	reporter := NewMockReporter(ctrl)
	service := &StorageService{
		inventory: inventory,
		storageBE: storageBE,
		reporter:  reporter,
	}

	chunk := t.NewChunk("chunk-1", []byte("hello"))

	storageBE.EXPECT().Store(chunk).Return(nil)

	inventory.EXPECT().Add(&chunk.Meta).Return(nil)

	inventory.EXPECT().
		Stage(chunk.Meta.ID).
		Return(chunk.Meta, nil)

	reporter.EXPECT().
		Report(gomock.Any(), t.NewReplicaStaged(chunk.Meta).ToRecord())

	err := service.commitUpload(ctx, chunk, &chunk.Meta)
	require.NoError(tt, err)
}

func TestStorageService_CommitUpload_RollbackOnInventoryFailure(tt *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(tt)

	inventory := NewMockInventory(ctrl)
	storageBE := NewMockChunkStorage(ctrl)
	reporter := NewMockReporter(ctrl)
	service := &StorageService{
		inventory: inventory,
		storageBE: storageBE,
		reporter:  reporter,
	}

	chunk := t.NewChunk("chunk-1", []byte("hello"))

	storageBE.EXPECT().
		Store(chunk).
		Return(nil)

	inventory.EXPECT().
		Add(&chunk.Meta).
		Return(s.ErrChunkConflict)

	storageBE.EXPECT().
		Delete(chunk.Meta.ID).
		Return(nil)

	err := service.commitUpload(ctx, chunk, &chunk.Meta)
	require.ErrorIs(tt, err, s.ErrChunkConflict)
}

func TestStorageService_CommitUpload_OnDigestMismatch(tt *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(tt)

	inventory := NewMockInventory(ctrl)
	storageBE := NewMockChunkStorage(ctrl)
	reporter := NewMockReporter(ctrl)
	service := &StorageService{
		inventory: inventory,
		storageBE: storageBE,
		reporter:  reporter,
	}

	chunk := t.NewChunk("chunk-1", []byte("hello"))
	expectedMeta := t.NewChunk("chunk-1", []byte("different")).Meta

	// expect no services called: digest mismatch fails immediately

	err := service.commitUpload(ctx, chunk, &expectedMeta)
	require.Error(tt, err)
}

func TestStorageService_StartUpload_OnChunkConflict(tt *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(tt)

	inventory := NewMockInventory(ctrl)
	service := &StorageService{
		inventory: inventory,
	}

	chunk := t.NewChunk("chunk-1", []byte("hello"))

	inventory.EXPECT().Has(chunk.Meta.ID).Return(true)

	session, err := service.StartUpload(ctx, &chunk.Meta)
	require.ErrorIs(tt, err, s.ErrChunkConflict)
	require.Nil(tt, session)
}

func TestStorageService_StartUpload_OnNewChunk(tt *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(tt)

	inventory := NewMockInventory(ctrl)
	service := &StorageService{
		inventory: inventory,
		metrics:   NewStorageMetrics(metrics.NopProvider{}),
		sem:       make(chan struct{}, 1),
	}

	chunk := t.NewChunk("chunk-1", []byte("hello"))

	inventory.EXPECT().Has(chunk.Meta.ID).Return(false)

	session, err := service.StartUpload(ctx, &chunk.Meta)

	require.NoError(tt, err)
	require.NotNil(tt, session)

	require.NoError(tt, session.Close())
}

func TestStorageService_StartUpload_OnCommit(tt *testing.T) {
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
		metrics:   NewStorageMetrics(metrics.NopProvider{}),
		sem:       make(chan struct{}, 1),
	}

	inventory.EXPECT().Has(chunk.Meta.ID).Return(false)
	storageBE.EXPECT().Store(chunk).Return(nil)
	inventory.EXPECT().Add(&chunk.Meta).Return(nil)
	inventory.EXPECT().
		Stage(chunk.Meta.ID).
		Return(chunk.Meta, nil)
	reporter.EXPECT().
		Report(gomock.Any(), t.NewReplicaStaged(chunk.Meta).ToRecord())

	session, err := service.StartUpload(ctx, &chunk.Meta)
	require.NoError(tt, err)

	n, err := session.Write(chunk.Data)
	require.NoError(tt, err)
	require.Equal(tt, len(chunk.Data), n)

	err = session.Commit(ctx)
	require.NoError(tt, err)
}
