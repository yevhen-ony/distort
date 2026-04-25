package domain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	m "dos/internal/services/master"
	"dos/internal/services/master/repo"
)

func newTestService(t *testing.T) *MasterService {
	t.Helper()

	return NewMasterService(
		repo.MakeInMemChunkRepo(),
		repo.NewInMemObjectRepo(),
		repo.NewInMemNodeRegistry(),
		&MasterServiceConfig{
			ReplicationCount:           1,
			ChunkAllocationMarginBytes: 0,
		},
	)
}

func TestMasterService_CreateObject(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	require.NoError(t, svc.CreateObject(ctx, m.ObjectID("obj-1")))
	require.ErrorIs(t, svc.CreateObject(ctx, m.ObjectID("obj-1")), m.ErrObjectExists)
}

func TestMasterService_AllocateChunk_HappyPath(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	// Prepare object.
	require.NoError(t, svc.CreateObject(ctx, m.ObjectID("obj-1")))

	// Prepare at least one candidate node.
	_, err := svc.nodeReg.Register(ctx, &m.NodeReport{
		Addr:      "127.0.0.1:9001",
		FreeBytes: 1024,
	})
	require.NoError(t, err)

	placement, err := svc.AllocateChunk(ctx, &m.AllocateChunkCommand{
		ObjectID:  m.ObjectID("obj-1"),
		ChunkKey:  m.ChunkKey("0"),
		ChunkSize: 100,
	})
	require.NoError(t, err)

	assert.NotEmpty(t, placement.ChunkID)
	require.Len(t, placement.Nodes, 1)
	assert.NotEmpty(t, placement.Nodes[0].NodeID)
	assert.NotEmpty(t, placement.Nodes[0].Addr)
}
