package storage

import (
	"context"
	"testing"

	t "dos/internal/common/types"
	s "dos/internal/services/storage"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestStorageService_StageAndReportOne_StagesAndReports(tt *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(tt)

	inventory := NewMockInventory(ctrl)
	reporter := NewMockReporter(ctrl)
	service := &StorageService{
		inventory: inventory,
		reporter:  reporter,
	}

	chunk := t.NewChunk("chunk-1", []byte("hello"))

	inventory.EXPECT().
		Stage(chunk.Meta.ID).
		Return(chunk.Meta, nil)

	reporter.EXPECT().
		Report(ctx, t.NewReplicaStaged(chunk.Meta).ToRecord())

	err := service.StageAndReportOne(ctx, chunk.Meta.ID)
	require.NoError(tt, err)
}

func TestStorageService_StageAndReportMany(tt *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(tt)

	inventory := NewMockInventory(ctrl)
	reporter := NewMockReporter(ctrl)
	service := &StorageService{
		inventory: inventory,
		reporter:  reporter,
	}

	chunk1 := t.NewChunk("chunk-1", []byte("hello"))
	chunk2 := t.NewChunk("chunk-2", []byte("world"))

	inventory.EXPECT().
		Stage(chunk1.Meta.ID).
		Return(chunk1.Meta, nil)

	reporter.EXPECT().
		Report(ctx, t.NewReplicaStaged(chunk1.Meta).ToRecord())

	inventory.EXPECT().
		Stage(chunk2.Meta.ID).
		Return(t.ChunkMeta{}, s.ErrChunkNotFound)

	reporter.EXPECT().
		Flush(ctx)

	got := service.StageAndReportMany(ctx, []t.ChunkID{
		chunk1.Meta.ID,
		chunk2.Meta.ID,
	})

	require.Equal(tt, []t.ChunkID{chunk1.Meta.ID}, got.Scheduled)
	require.Equal(tt, []t.ChunkID{chunk2.Meta.ID}, got.Failed)
}

func TestStorageService_StageAndReportAll(tt *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(tt)

	inventory := NewMockInventory(ctrl)
	reporter := NewMockReporter(ctrl)
	service := &StorageService{
		inventory: inventory,
		reporter:  reporter,
	}

	chunk := t.NewChunk("chunk-1", []byte("hello"))

	inventory.EXPECT().
		ListIDs().
		Return([]t.ChunkID{chunk.Meta.ID})

	inventory.EXPECT().
		Stage(chunk.Meta.ID).
		Return(chunk.Meta, nil)

	reporter.EXPECT().
		Report(ctx, t.NewReplicaStaged(chunk.Meta).ToRecord())

	reporter.EXPECT().Flush(ctx)

	got := service.StageAndReportAll(ctx)

	require.Equal(tt, []t.ChunkID{chunk.Meta.ID}, got.Scheduled)
	require.Empty(tt, got.Failed)
}

func TestStorageService_ProcessReport(tt *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(tt)

	inventory := NewMockInventory(ctrl)
	reporter := NewMockReporter(ctrl)
	service := &StorageService{
		inventory: inventory,
		reporter:  reporter,
	}

	chunk1 := t.NewChunk("chunk-1", []byte("hello"))
	chunk2 := t.NewChunk("chunk-2", []byte("world"))
	record2 := s.ChunkRecord{Meta: chunk2.Meta, State: s.ChunkStateStaged}
	result := t.ReportResult{
		Accepted: []t.ChunkID{chunk1.Meta.ID},
		Rejected: []t.ChunkID{chunk2.Meta.ID},
	}

	inventory.EXPECT().
		Activate(chunk1.Meta.ID).
		Return(chunk1.Meta, nil)

	inventory.EXPECT().
		GetRecord(chunk2.Meta.ID).
		Return(&record2, nil)

	reporter.EXPECT().
		Report(gomock.Any(), t.NewReplicaStaged(record2.Meta).ToRecord())

	service.ProcessReport(ctx, result)
}
