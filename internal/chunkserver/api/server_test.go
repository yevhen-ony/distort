package api_test 

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net"
	"testing"
	"time"

	pb "dos/gen/proto/chunk/v1"
	"dos/internal/chunkserver/api"
	"dos/internal/chunkserver/service"
	"dos/internal/chunkserver/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func checksumHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}



func startChunkServer(t *testing.T) (pb.ChunkServiceClient, func()) {
	t.Helper()

	storeConfig := &storage.ChunkStorageConfig{RootDir: t.TempDir()}
	store, err := storage.New(storeConfig)
	require.NoError(t, err)

	service, err := service.New(store)
	require.NoError(t, err)

	serverConfig := &api.ServerConfig{PartSize: 4}
	server := api.New(service, serverConfig)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	gs := grpc.NewServer()
	pb.RegisterChunkServiceServer(gs, server)

	go gs.Serve(lis)

	conn, err := grpc.NewClient(
		lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	client := pb.NewChunkServiceClient(conn)
	cleanup := func() {
		_ = conn.Close()
		gs.Stop()
		_ = lis.Close()
	}
	return client, cleanup
}

func TestServer_PutAndGetChunk_HappyPath(t *testing.T) {
	client, cleanup := startChunkServer(t)
	defer cleanup()

	payload := []byte("hello world")
	chunkID := "chunk-1"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	put, err := client.PutChunk(ctx)
	require.NoError(t, err)

	require.NoError(t, put.Send(&pb.PutChunkRequest{
		Header: &pb.PutChunkHeader{
			ServerId:  "service-id-123",
			ChunkId:   chunkID,
			ChunkSize: int64(len(payload)),
			Checksum:  checksumHex(payload),
		},
	}))
	require.NoError(t, put.Send(&pb.PutChunkRequest{Data: payload[:5]}))
	require.NoError(t, put.Send(&pb.PutChunkRequest{Data: payload[5:]}))

	_, err = put.CloseAndRecv()
	require.NoError(t, err)

	get, err := client.GetChunk(ctx, &pb.GetChunkRequest{ChunkId: chunkID})
	require.NoError(t, err)

	var header *pb.GetChunkHeader
	var got []byte

	for {
		msg, err := get.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		if msg.GetHeader() != nil {
			header = msg.GetHeader()
			continue
		}
		got = append(got, msg.GetData()...)
	}

	require.NotNil(t, header)
	assert.Equal(t, chunkID, header.GetChunkId())
	assert.Equal(t, int64(len(payload)), header.GetChunkSize())
	assert.Equal(t, checksumHex(payload), header.GetChecksum())
	assert.Equal(t, payload, got)
}

func TestServer_PutChunk_InvalidServerID(t *testing.T) {
	client, cleanup := startChunkServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	put, err := client.PutChunk(ctx)
	require.NoError(t, err)

	require.NoError(t, put.Send(&pb.PutChunkRequest{
		Header: &pb.PutChunkHeader{
			ServerId:  "wrong-id",
			ChunkId:   "chunk-x",
			ChunkSize: 1,
		},
	}))

	_, err = put.CloseAndRecv()
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestServer_GetChunk_NotFound(t *testing.T) {
	client, cleanup := startChunkServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.GetChunk(ctx, &pb.GetChunkRequest{ChunkId: "missing"})
	require.NoError(t, err)

	_, err = stream.Recv()
	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}
