package transport_test

import (
	"context"
	"net"
	"testing"
	"time"

	pb "dos/gen/proto/master/v1"
	"dos/internal/common/digest"
	c "dos/internal/services/client"
	tr "dos/internal/services/client/transport"
	m "dos/internal/services/master"
	"dos/internal/services/master/api"
	"dos/internal/services/master/domain"
	"dos/internal/services/master/repo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type MasterDependencies struct {
	chunkRepo *repo.InMemChunkRepo
	objectRepo *repo.InMemObjectRepo
	nodeReg *repo.InMemNodeRegistry
}

func startMasterServer(t *testing.T) (string, *MasterDependencies, func()) {
	t.Helper()

	deps := &MasterDependencies{
		chunkRepo: repo.MakeInMemChunkRepo(),
		objectRepo: repo.NewInMemObjectRepo(),
		nodeReg: repo.NewInMemNodeRegistry(),
	}

	svc := domain.NewMasterService(
		deps.chunkRepo,
		deps.objectRepo,
		deps.nodeReg,
		&domain.MasterServiceConfig{
			ReplicationCount:           1,
			ChunkAllocationMarginBytes: 0,
		},
	)

	server := api.NewClientServer(svc)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	gs := grpc.NewServer()
	pb.RegisterMasterClientServiceServer(gs, server)

	go func() { _ = gs.Serve(lis) }()

	cleanup := func() {
		gs.Stop()
		lis.Close()
	}

	return lis.Addr().String(), deps, cleanup
}

func TestMasterTransport_HappyPath_AgainstMasterServer(t *testing.T) {
	addr, deps, stopServer := startMasterServer(t)
	defer stopServer()

	cp := tr.NewConnectionPool()
	defer func() { _ = cp.Close() }()

	mt, err := tr.NewMasterTransport(cp, &tr.MasterTransportConfig{Addr: addr})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	nodeID, err := deps.nodeReg.Register(ctx, &m.NodeReport{
		Addr:      "127.0.0.1:9001",
		FreeBytes: 1024,
	})
	require.NoError(t, err)

	const (
		objectID  = c.ObjectID("obj-1")
		chunkKey  = c.ChunkKey("0")
		chunkSize = int64(128)
	)

	require.NoError(t, mt.CreateObject(ctx, objectID))

	placement, err := mt.AllocateChunk(ctx, &tr.AllocateChunkQuery{
		ObjectID:  objectID,
		ChunkKey:  chunkKey,
		ChunkSize: chunkSize,
	})
	require.NoError(t, err)
	require.NotEmpty(t, placement.ChunkID)
	require.Len(t, placement.Nodes, 1)

	assert.Equal(t, string(nodeID), placement.Nodes[0].NodeID)
	assert.Equal(t, "127.0.0.1:9001", placement.Nodes[0].Addr)

	// Seed metadata needed by GetObjectAccess.
	require.NoError(t, deps.chunkRepo.SetDigest(
		ctx,
		m.ChunkID(placement.ChunkID),
		&digest.Digest{Size: chunkSize, Checksum: "checksum-1"},
	))
	require.NoError(t, deps.nodeReg.AttachChunk(ctx, nodeID, m.ChunkID(placement.ChunkID)))

	obj, err := mt.GetObjectAccess(ctx, objectID)
	require.NoError(t, err)

	assert.Equal(t, objectID, obj.ObjectID)
	assert.Equal(t, chunkSize, obj.TotalSize)
	require.Len(t, obj.Chunks, 1)
	assert.Equal(t, placement.ChunkID, obj.Chunks[0].ChunkID)
	assert.Equal(t, chunkKey, obj.Chunks[0].ChunkKey)
	require.Len(t, obj.Chunks[0].Nodes, 1)
	assert.Equal(t, string(nodeID), obj.Chunks[0].Nodes[0].NodeID)
}
