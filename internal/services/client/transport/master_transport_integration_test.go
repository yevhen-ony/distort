package transport_test

import (
	"context"
	"net"
	"testing"
	"time"

	pb "dos/gen/proto/master/v1"
	"dos/internal/common/digest"
	t "dos/internal/common/types"
	tr "dos/internal/services/client/transport"
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

func startMasterServer(test *testing.T) (string, *MasterDependencies, func()) {
	test.Helper()

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
	require.NoError(test, err)

	gs := grpc.NewServer()
	pb.RegisterMasterClientServiceServer(gs, server)

	go func() { _ = gs.Serve(lis) }()

	cleanup := func() {
		gs.Stop()
		lis.Close()
	}

	return lis.Addr().String(), deps, cleanup
}

func TestMasterTransport_HappyPath_AgainstMasterServer(test *testing.T) {
	addr, deps, stopServer := startMasterServer(test)
	defer stopServer()

	cp := tr.NewConnectionPool()
	defer func() { _ = cp.Close() }()

	mt, err := tr.NewMasterTransport(cp, &tr.MasterTransportConfig{Addr: addr})
	require.NoError(test, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	nodeID, err := deps.nodeReg.Register(ctx, &t.NodeStats{
		Addr:      "127.0.0.1:9001",
		FreeBytes: 1024,
	})
	require.NoError(test, err)

	const (
		objectID  = t.ObjectID("obj-1")
		chunkKey  = t.ChunkKey("0")
		chunkSize = int64(128)
	)

	require.NoError(test, mt.CreateObject(ctx, objectID))

	placement, err := mt.AllocateChunk(ctx, &tr.AllocateChunkQuery{
		ObjectID:  objectID,
		ChunkKey:  chunkKey,
		ChunkSize: chunkSize,
	})
	require.NoError(test, err)
	require.NotEmpty(test, placement.ID)
	require.Len(test, placement.Nodes, 1)

	assert.Equal(test, nodeID, placement.Nodes[0].ID)
	assert.Equal(test, "127.0.0.1:9001", placement.Nodes[0].Addr)

	// Seed metadata needed by GetObjectAccess.
	require.NoError(test, deps.chunkRepo.SetDigest(
		ctx,
		t.ChunkID(placement.ID),
		digest.Digest{Size: chunkSize, Checksum: "checksum-1"},
	))
	require.NoError(test, deps.nodeReg.AttachChunk(ctx, nodeID, t.ChunkID(placement.ID)))

	obj, err := mt.GetObjectAccess(ctx, objectID)
	require.NoError(test, err)

	assert.Equal(test, objectID, obj.ID)
	assert.Equal(test, chunkSize, obj.TotalSize)
	require.Len(test, obj.Chunks, 1)
	assert.Equal(test, placement.ID, obj.Chunks[0].ID)
	assert.Equal(test, chunkKey, obj.Chunks[0].Key)
	require.Len(test, obj.Chunks[0].Nodes, 1)
	assert.Equal(test, nodeID, obj.Chunks[0].Nodes[0].ID)
}
