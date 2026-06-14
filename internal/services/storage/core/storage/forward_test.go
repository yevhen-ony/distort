package storage

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"dos/internal/common/metrics"
	chunkrpcmock "dos/internal/common/transport/chunkrpc/mock"
	t "dos/internal/common/types"
	s "dos/internal/services/storage"
)

func TestStorageService_SendChunk(tt *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(tt)

	chunkT := NewMockChunkTransport(ctrl)
	session := chunkrpcmock.NewMockUploadSession(ctrl)
	service := &StorageService{
		chunkT:  chunkT,
		metrics: NewStorageMetrics(metrics.NopProvider{}),
	}

	chunk := t.NewChunk("chunk-1", []byte("hello"))
	target := t.NodeRef{ID: "node-2", Addr: "addr-2"}

	chunkT.EXPECT().
		NewUploadSession([]t.NodeRef{target}).
		Return(session)

	session.EXPECT().
		Upload(ctx, &chunk).
		Return(target, nil)

	got, err := service.SendChunk(ctx, chunk, []t.NodeRef{target})

	require.NoError(tt, err)
	require.Equal(tt, target, got)
}

func TestStorageService_ReplicateChunk(tt *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(tt)

	chunkT := NewMockChunkTransport(ctrl)

	chunkID := t.ChunkID("chunk-1")
	source := t.NodeRef{ID: "node-2", Addr: "addr-2"}
	targets := []t.NodeRef{
		{ID: "node-3", Addr: "addr-3"},
	}

	service := &StorageService{
		chunkT:  chunkT,
		metrics: NewStorageMetrics(metrics.NopProvider{}),
	}

	chunkT.EXPECT().
		ReplicateChunk(ctx, chunkID, source, targets).
		Return(nil)

	err := service.ReplicateChunk(ctx, chunkID, source, targets)
	require.NoError(tt, err)
}

func TestStorageService_ForwardChunk_SendsToSingleValidTarget(tt *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(tt)

	identity := NewMockIdentity(ctrl)
	inventory := NewMockInventory(ctrl)
	storageBE := NewMockChunkStorage(ctrl)
	chunkT := NewMockChunkTransport(ctrl)
	session := chunkrpcmock.NewMockUploadSession(ctrl)

	self := t.NodeID("node-1")
	chunk := t.NewChunk("chunk-1", []byte("hello"))
	target := t.NodeRef{ID: "node-2", Addr: "addr-2"}

	service := &StorageService{
		identity:  identity,
		inventory: inventory,
		storageBE: storageBE,
		chunkT:    chunkT,
		config:    fixedStorageConfig{replicationTimeout: time.Second},
		metrics:   NewStorageMetrics(metrics.NopProvider{}),
	}

	identity.EXPECT().
		GetID().
		Return(self, nil)

	inventory.EXPECT().
		GetRecord(chunk.Meta.ID).
		Return(s.NewChunkRecord(chunk.Meta), nil)

	storageBE.EXPECT().
		Get(chunk.Meta.ID).
		Return(io.NopCloser(bytes.NewReader(chunk.Data)), nil)

	chunkT.EXPECT().
		NewUploadSession([]t.NodeRef{target}).
		Return(session)

	session.EXPECT().
		Upload(gomock.Any(), &chunk).
		Return(target, nil)

	err := service.ForwardChunk(ctx, chunk.Meta.ID, []t.NodeRef{
		{ID: self, Addr: "addr-1"},
		target,
	})

	require.NoError(tt, err)
}


func TestStorageService_ForwardChunk_ReplicatesRemainingTargetsThroughChosenNode(tt *testing.T) {
  	ctx := context.Background()
  	ctrl := gomock.NewController(tt)

  	identity := NewMockIdentity(ctrl)
  	inventory := NewMockInventory(ctrl)
  	storageBE := NewMockChunkStorage(ctrl)
  	chunkT := NewMockChunkTransport(ctrl)
  	session := chunkrpcmock.NewMockUploadSession(ctrl)
  	service := &StorageService{
  		identity:  identity,
  		inventory: inventory,
  		storageBE: storageBE,
  		chunkT:    chunkT,
  		config:    fixedStorageConfig{replicationTimeout: time.Second},
  		metrics:   NewStorageMetrics(metrics.NopProvider{}),
  	}

  	self := t.NodeID("node-1")
  	chunk := t.NewChunk("chunk-1", []byte("hello"))
  	target1 := t.NodeRef{ID: "node-2", Addr: "addr-2"}
  	target2 := t.NodeRef{ID: "node-3", Addr: "addr-3"}

  	identity.EXPECT().
  		GetID().
  		Return(self, nil)

  	inventory.EXPECT().
  		GetRecord(chunk.Meta.ID).
  		Return(s.NewChunkRecord(chunk.Meta), nil)

  	storageBE.EXPECT().
  		Get(chunk.Meta.ID).
  		Return(io.NopCloser(bytes.NewReader(chunk.Data)), nil)

  	chunkT.EXPECT().
  		NewUploadSession([]t.NodeRef{target1, target2}).
  		Return(session)

  	session.EXPECT().
  		Upload(gomock.Any(), &chunk).
  		Return(target1, nil)

  	chunkT.EXPECT().
  		ReplicateChunk(gomock.Any(), chunk.Meta.ID, target1, []t.NodeRef{target2}).
  		Return(nil)

  	err := service.ForwardChunk(ctx, chunk.Meta.ID, []t.NodeRef{
  		{ID: self, Addr: "addr-1"},
  		target1,
  		target2,
  	})

  	require.NoError(tt, err)
}


// fake config

type fixedStorageConfig struct{ replicationTimeout time.Duration }

func (c fixedStorageConfig) AdvertiseAddr() string             { return "" }
func (c fixedStorageConfig) ReplicationTimeout() time.Duration { return c.replicationTimeout }
func (c fixedStorageConfig) MaxParallelHeavyOps() int          { return 1 }
