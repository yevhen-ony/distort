package storage

import (
	"bytes"
	"errors"
	"io"
	"testing"

	t "dos/internal/common/types"
	s "dos/internal/services/storage"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestStorageService_LoadChunk_ReturnsStoredChunk(tt *testing.T) {
	ctrl := gomock.NewController(tt)

	inventory := NewMockInventory(ctrl)
	storageBE := NewMockChunkStorage(ctrl)

	chunk := t.NewChunk("chunk-1", []byte("hello"))
	service := &StorageService{
		inventory: inventory,
		storageBE: storageBE,
	}

	inventory.EXPECT().
		GetRecord(chunk.Meta.ID).
		Return(s.NewChunkRecord(chunk.Meta), nil)

	storageBE.EXPECT().
		Get(chunk.Meta.ID).
		Return(io.NopCloser(bytes.NewReader(chunk.Data)), nil)

	got, err := service.LoadChunk(chunk.Meta.ID)

	require.NoError(tt, err)
	require.Equal(tt, chunk, got)
}

func TestStorageService_LoadChunk_ReturnsInventoryError(tt *testing.T) {
	ctrl := gomock.NewController(tt)

	inventory := NewMockInventory(ctrl)
	storageBE := NewMockChunkStorage(ctrl)

	chunkID := t.ChunkID("missing-chunk")
	service := &StorageService{
		inventory: inventory,
		storageBE: storageBE,
	}

	inventory.EXPECT().
		GetRecord(chunkID).
		Return(nil, s.ErrChunkNotFound)

	got, err := service.LoadChunk(chunkID)

	require.ErrorIs(tt, err, s.ErrChunkNotFound)
	require.Empty(tt, got)
}

func TestStorageService_LoadChunk_ReturnsStorageGetError(tt *testing.T) {
	ctrl := gomock.NewController(tt)

	inventory := NewMockInventory(ctrl)
	storageBE := NewMockChunkStorage(ctrl)

	service := &StorageService{
		inventory: inventory,
		storageBE: storageBE,
	}

	chunk := t.NewChunk("chunk-1", []byte("hello"))
	getErr := errors.New("get failed")

	inventory.EXPECT().
		GetRecord(chunk.Meta.ID).
		Return(s.NewChunkRecord(chunk.Meta), nil)

	storageBE.EXPECT().
		Get(chunk.Meta.ID).
		Return(nil, getErr)

	got, err := service.LoadChunk(chunk.Meta.ID)

	require.ErrorIs(tt, err, getErr)
	require.Contains(tt, err.Error(), "get from store")
	require.Empty(tt, got)
}
