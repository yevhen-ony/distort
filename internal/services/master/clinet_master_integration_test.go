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
	"dos/internal/common/digest"
	t "dos/internal/common/types"
	"dos/internal/services/master/api"
	"dos/internal/services/master/domain"
	"dos/internal/services/master/repo"
)

type masterTestDeps struct {
	chunkRepo *repo.InMemChunkRepo
	nodeReg   *repo.InMemNodeRegistry
}

func startMasterClientServer(test *testing.T) (pb.MasterClientServiceClient, *masterTestDeps, func()) {
	test.Helper()

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
	require.NoError(test, err)

	gs := grpc.NewServer()
	pb.RegisterMasterClientServiceServer(gs, server)

	go func() { _ = gs.Serve(lis) }()

	conn, err := grpc.NewClient(
		lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(test, err)

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

func TestClientServer_CreateAllocateGetObjectAccess_HappyPath(test *testing.T) {
	client, deps, cleanup := startMasterClientServer(test)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const objectID = "obj-1"
	const chunkKey = "0" 
	const chunkSize = int64(123)

	_, err := client.CreateObject(ctx, &pb.CreateObjectRequest{
		ObjectId: objectID,
	})
	require.NoError(test, err)

	nid, err := deps.nodeReg.Register(ctx, &t.NodeStats{
		Addr:      "127.0.0.1:9001",
		FreeBytes: 1024,
	})
	require.NoError(test, err)

	alloc, err := client.AllocateChunk(ctx, &pb.AllocateChunkRequest{
		ObjectId:  objectID,
		ChunkKey:  chunkKey,
		ChunkSize: chunkSize,
	})
	require.NoError(test, err)
	require.NotEmpty(test, alloc.GetChunkId())
	require.Len(test, alloc.GetNodes(), 1)

	// Simulate post-write metadata sync so GetObjectAccess can report size and nodes.
	require.NoError(test, deps.chunkRepo.SetDigest(
		ctx,
		t.ChunkID(alloc.GetChunkId()),
		digest.Digest{Size: chunkSize, Checksum: "checksum-1"},
	))
	require.NoError(test, deps.nodeReg.AttachChunk(
		ctx,
		nid,
		t.ChunkID(alloc.GetChunkId()),
	))

	got, err := client.GetObjectAccess(ctx, &pb.GetObjectAccessRequest{
		ObjectId: objectID,
	})
	require.NoError(test, err)

	assert.Equal(test, objectID, got.GetObjectId())
	assert.Equal(test, chunkSize, got.GetTotalSize())
	require.Len(test, got.GetChunks(), 1)
	assert.Equal(test, alloc.GetChunkId(), got.GetChunks()[0].GetChunkId())
	assert.Equal(test, chunkKey, got.GetChunks()[0].GetChunkKey())
	require.Len(test, got.GetChunks()[0].GetNodes(), 1)
	assert.Equal(test, string(nid), got.GetChunks()[0].GetNodes()[0].GetNodeId())
	assert.Equal(test, "127.0.0.1:9001", got.GetChunks()[0].GetNodes()[0].GetAddr())
}

func TestClientServer_CreateObject_InvalidArgument(test *testing.T) {
	client, _, cleanup := startMasterClientServer(test)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.CreateObject(ctx, &pb.CreateObjectRequest{})
	require.Error(test, err)
	assert.Equal(test, codes.InvalidArgument, status.Code(err))
}

func TestClientServer_AllocateChunk_NoCandidateNodes(test *testing.T) {
	client, _, cleanup := startMasterClientServer(test)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.CreateObject(ctx, &pb.CreateObjectRequest{
		ObjectId: "obj-1",
	})
	require.NoError(test, err)

	_, err = client.AllocateChunk(ctx, &pb.AllocateChunkRequest{
		ObjectId:  "obj-1",
		ChunkKey:  "0",
		ChunkSize: 1,
	})
	require.Error(test, err)
	assert.Equal(test, codes.Internal, status.Code(err))
	assert.Contains(test, err.Error(), "no suitable nodes found")
}

func TestClientServer_GetObjectAccess_NotFound(test *testing.T) {
	client, _, cleanup := startMasterClientServer(test)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.GetObjectAccess(ctx, &pb.GetObjectAccessRequest{
		ObjectId: "missing",
	})
	require.Error(test, err)
	assert.Equal(test, codes.Internal, status.Code(err))
	assert.Contains(test, err.Error(), "object not found")
}
