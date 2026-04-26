package transport_test

import (
	"context"
	"net"
	"testing"
	"time"

	pb "dos/gen/proto/chunk/v1"
	t "dos/internal/common/types"
	c "dos/internal/services/client"
	tr "dos/internal/services/client/transport"
	"dos/internal/services/storage/api"
	"dos/internal/services/storage/core"
	"dos/internal/services/storage/store"
	"dos/internal/common/digest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func startChunkServer(test *testing.T) (string, func()) {
	test.Helper()

	storeConfig := &store.ChunkStorageConfig{RootDir: test.TempDir()}
	store, err := store.New(storeConfig)
	require.NoError(test, err)

	svc, err := core.New(store)
	require.NoError(test, err)

	server := api.New(svc, &api.ServerConfig{FrameSize: 4})

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(test, err)

	gs := grpc.NewServer()
	pb.RegisterChunkServiceServer(gs, server)

	go func() { _ = gs.Serve(lis) }()
	
	addr := lis.Addr().String()
	cleanup := func() {
		gs.Stop()
		lis.Close()
	}

	return addr, cleanup 
}

func TestChunkTransport_HappyPath_AgainstChunkServer(test *testing.T) {
	addr, stopServer := startChunkServer(test)
	defer stopServer()

	cp := tr.NewConnectionPool()
	defer func() { _ = cp.Close() }()

	cfg := &tr.StorageTransportConfig{FrameSize: 3}
	tr, err := tr.NewChunkTransport(cp, cfg)
	require.NoError(test, err)

	payload := []byte("hello chunk transport")
	dg := digest.New()
	_, err = dg.Write(payload)
	require.NoError(test, err)

	src := &c.Chunk{
		ID:       t.ChunkID("chunk-1"),
		Checksum: dg.Checksum(),
		Data:     payload,
	}
	target := t.NodeRef{
		ID:   "service-id-123",
		Addr: addr,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(test, tr.SendChunk(ctx, target, src))

	got, err := tr.ReceiveChunk(ctx, target, src.ID)
	require.NoError(test, err)

	assert.Equal(test, src.ID, got.ID)
	assert.Equal(test, src.Checksum, got.Checksum)
	assert.Equal(test, src.Data, got.Data)
}
