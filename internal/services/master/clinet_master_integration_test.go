package master_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "dos/gen/proto/master/v1"
	"dos/internal/libraries/digest"
	m "dos/internal/services/master"
	"dos/internal/services/master/api"
	"dos/internal/services/master/domain"
	"dos/internal/services/master/repo"
)

type masterTestDeps struct {
	chunkRepo *repo.InMemChunkRepo
	nodeReg   *repo.InMemNodeRegistry
}

func startMasterClientServer(t *testing.T) (pb.MasterClientServiceClient, *masterTestDeps, func()) {
	t.Helper()

	chunkRepo := repo.MakeInMemChunkRepo()
	objectRepo := repo.NewInMemObjectRepo()
	nodeReg := repo.NewInMemNodeRegistry()

	svc := domain.NewMasterService(
		chunkRepo,
		objectRepo,
		nodeReg,
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

	conn, err := grpc.NewClient(
		lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	client := pb.NewMasterClientServiceClient(conn)

	cleanup := func() {
		conn.Close()
		gs.Stop()
		lis.Close()
	}

	deps := &masterTestDeps{
		chunkRepo: chunkRepo,
		nodeReg:   nodeReg,
	}

	return client, deps, cleanup
}

func TestClientServer_CreateAllocateGetObjectAccess_HappyPath(t *testing.T) {
	client, deps, cleanup := startMasterClientServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const objectID = "obj-1"
	const chunkKey = int64(0)
	const chunkSize = int64(123)

	_, err := client.CreateObject(ctx, &pb.CreateObjectRequest{
		ObjectId: objectID,
	})
	require.NoError(t, err)

	nid, err := deps.nodeReg.Register(ctx, &m.NodeReport{
		Addr:      "127.0.0.1:9001",
		FreeBytes: 1024,
	})
	require.NoError(t, err)

	alloc, err := client.AllocateChunk(ctx, &pb.AllocateChunkRequest{
		ObjectId:  objectID,
		ChunkKey:  chunkKey,
		ChunkSize: chunkSize,
	})
	require.NoError(t, err)
	require.NotEmpty(t, alloc.GetChunkId())
	require.Len(t, alloc.GetNodes(), 1)

	// Simulate post-write metadata sync so GetObjectAccess can report size and nodes.
	require.NoError(t, deps.chunkRepo.SetDigest(
		ctx,
		m.ChunkID(alloc.GetChunkId()),
		&digest.Digest{Size: chunkSize, Checksum: "checksum-1"},
	))
	require.NoError(t, deps.nodeReg.AttachChunk(
		ctx,
		nid,
		m.ChunkID(alloc.GetChunkId()),
	))

	got, err := client.GetObjectAccess(ctx, &pb.GetObjectAccessRequest{
		ObjectId: objectID,
	})
	require.NoError(t, err)

	assert.Equal(t, objectID, got.GetObjectId())
	assert.Equal(t, chunkSize, got.GetObjectSize())
	require.Len(t, got.GetChunks(), 1)
	assert.Equal(t, alloc.GetChunkId(), got.GetChunks()[0].GetChunkId())
	assert.Equal(t, chunkKey, got.GetChunks()[0].GetChunkKey())
	require.Len(t, got.GetChunks()[0].GetNodes(), 1)
	assert.Equal(t, string(nid), got.GetChunks()[0].GetNodes()[0].GetNodeId())
	assert.Equal(t, "127.0.0.1:9001", got.GetChunks()[0].GetNodes()[0].GetAddress())
}

func TestClientServer_CreateObject_InvalidArgument(t *testing.T) {
	client, _, cleanup := startMasterClientServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.CreateObject(ctx, &pb.CreateObjectRequest{})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestClientServer_AllocateChunk_NoCandidateNodes(t *testing.T) {
	client, _, cleanup := startMasterClientServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.CreateObject(ctx, &pb.CreateObjectRequest{
		ObjectId: "obj-1",
	})
	require.NoError(t, err)

	_, err = client.AllocateChunk(ctx, &pb.AllocateChunkRequest{
		ObjectId:  "obj-1",
		ChunkKey:  0,
		ChunkSize: 1,
	})
	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.Contains(t, err.Error(), "no suitable nodes found")
}

func TestClientServer_GetObjectAccess_NotFound(t *testing.T) {
	client, _, cleanup := startMasterClientServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.GetObjectAccess(ctx, &pb.GetObjectAccessRequest{
		ObjectId: "missing",
	})
	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.Contains(t, err.Error(), "object not found")
}
