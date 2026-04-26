package domain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	m "dos/internal/services/master"
	t "dos/internal/common/types"
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

func TestMasterService_CreateObject(test *testing.T) {
	svc := newTestService(test)
	ctx := context.Background()

	require.NoError(test, svc.CreateObject(ctx, t.ObjectID("obj-1")))
	require.ErrorIs(test, svc.CreateObject(ctx, t.ObjectID("obj-1")), m.ErrObjectExists)
}

func TestMasterService_AllocateChunk_HappyPath(test *testing.T) {
	svc := newTestService(test)
	ctx := context.Background()

	// Prepare object.
	require.NoError(test, svc.CreateObject(ctx, t.ObjectID("obj-1")))

	// Prepare at least one candidate node.
	_, err := svc.nodeReg.Register(ctx, &t.NodeStats{
		Addr:      "127.0.0.1:9001",
		FreeBytes: 1024,
	})
	require.NoError(test, err)

	placement, err := svc.AllocateChunk(ctx, &m.AllocateChunkCommand{
		ObjectID:  t.ObjectID("obj-1"),
		ChunkKey:  t.ChunkKey("0"),
		ChunkSize: 100,
	})
	require.NoError(test, err)

	assert.NotEmpty(test, placement.ID)
	require.Len(test, placement.Nodes, 1)
	assert.NotEmpty(test, placement.Nodes[0].ID)
	assert.NotEmpty(test, placement.Nodes[0].Addr)
}
