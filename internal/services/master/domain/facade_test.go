package domain

import (
	"context"
	"testing"

	t "dos/internal/common/types"
	m "dos/internal/services/master"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestClientFacadeService_AllocateChunk(tt *testing.T) {
	ctx := context.Background()
	f := newFacadeFixture(tt)

	slot := t.ObjectSlot{ObjectID: "object-1", ChunkKey: "chunk-key-1"}
	targets := []t.NodeRef{{ID: "node-1", Addr: "node-1:10001"}}

	f.catalog.EXPECT().ExistsChunk(ctx, slot).Return(false, nil)
	f.catalog.EXPECT().GetReplication(ctx, slot.ObjectID).Return(1, nil)
	f.placement.EXPECT().GetCandidates(ctx, m.CandidateNodesQuery{
		MinFreeBytes: 123,
		MaxCount:     1,
	}).Return(targets, nil)
	f.catalog.EXPECT().AddChunk(ctx, slot, int64(123)).Return(t.ChunkID("chunk-1"), nil)

	s, err := NewClientFacadeService(f.deps())
	require.NoError(tt, err)

	got, err := s.AllocateChunk(ctx, m.AllocateChunkCommand{
		Slot: slot,
		Size: 123,
	})

	require.NoError(tt, err)
	require.Equal(tt, t.ChunkID("chunk-1"), got.ID)
	require.Equal(tt, slot, got.Slot)
	require.Equal(tt, targets, got.Targets)
}

func TestClientFacadeService_AllocateChunk_NoCandidates(tt *testing.T) {
	ctx := context.Background()
	f := newFacadeFixture(tt)

	slot := t.ObjectSlot{ObjectID: "object-1", ChunkKey: "chunk-key-1"}

	f.catalog.EXPECT().ExistsChunk(ctx, slot).Return(false, nil)
	f.catalog.EXPECT().GetReplication(ctx, slot.ObjectID).Return(1, nil)
	f.placement.EXPECT().GetCandidates(ctx, m.CandidateNodesQuery{
		MinFreeBytes: 123,
		MaxCount:     1,
	}).Return(nil, nil)

	s, err := NewClientFacadeService(f.deps())
	require.NoError(tt, err)

	_, err = s.AllocateChunk(ctx, m.AllocateChunkCommand{
		Slot: slot,
		Size: 123,
	})

	require.ErrorIs(tt, err, m.ErrNoCandidateNodes)
}

func TestClientFacadeService_AllocateChunk_Occupied(tt *testing.T) {
	ctx := context.Background()
	f := newFacadeFixture(tt)

	slot := t.ObjectSlot{ObjectID: "object-1", ChunkKey: "chunk-key-1"}
	chunkID := t.ChunkID("chunk-1")

	f.catalog.EXPECT().ExistsChunk(ctx, slot).Return(true, nil)
	f.catalog.EXPECT().GetReplication(ctx, slot.ObjectID).Return(1, nil)
	f.placement.EXPECT().GetCandidates(ctx, m.CandidateNodesQuery{
		MinFreeBytes: 123,
		MaxCount:     1,
	}).Return([]t.NodeRef{{ID: "node-2"}}, nil)
	f.catalog.EXPECT().GetChunkID(ctx, slot).Return(chunkID, nil)
	f.placement.EXPECT().GetChunkNodes(ctx, chunkID).
		Return([]t.NodeRef{{ID: "node-1"}}, nil)

	s, err := NewClientFacadeService(f.deps())
	require.NoError(tt, err)

	_, err = s.AllocateChunk(ctx, m.AllocateChunkCommand{
		Slot: slot,
		Size: 123,
	})
	require.ErrorIs(tt, err, m.ErrChunkKeyOccupied)
}

func TestClientFacadeService_SetReplication(tt *testing.T) {
	ctx := context.Background()
	f := newFacadeFixture(tt)

	f.lifecycle.EXPECT().GetNodeCount(ctx).Return(3)
	f.catalog.EXPECT().SetReplication(ctx, t.ObjectID("object-1"), 2).Return(nil)
	f.catalog.EXPECT().
		GetObjectChunks(ctx, t.ObjectID("object-1")).
		Return([]t.ChunkID{"chunk-1", "chunk-2"}, nil)
	f.replication.EXPECT().Schedule(ctx, t.ChunkID("chunk-1"))
	f.replication.EXPECT().Schedule(ctx, t.ChunkID("chunk-2"))

	s, err := NewClientFacadeService(f.deps())
	require.NoError(tt, err)

	err = s.SetReplication(ctx, "object-1", 2)

	require.NoError(tt, err)
}

type facadeFixture struct {
	catalog     *MockCatalog
	placement   *MockPlacement
	lifecycle   *MockLifecycle
	replication *MockReplicationScheduler
	config      *MockClientFacadeConfig
}

func newFacadeFixture(tt *testing.T) *facadeFixture {
	ctrl := gomock.NewController(tt)
	return &facadeFixture{
		catalog:     NewMockCatalog(ctrl),
		placement:   NewMockPlacement(ctrl),
		lifecycle:   NewMockLifecycle(ctrl),
		replication: NewMockReplicationScheduler(ctrl),
		config:      NewMockClientFacadeConfig(ctrl),
	}
}

func (f *facadeFixture) deps() ClientFacadeDeps {
	return ClientFacadeDeps{
		Catalog:     f.catalog,
		Placement:   f.placement,
		Lifecycle:   f.lifecycle,
		Replication: f.replication,
		Config:      f.config,
	}
}
