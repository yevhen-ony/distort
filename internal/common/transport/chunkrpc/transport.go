package chunkrpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	spb "dos/gen/proto/storage/v1"
	"dos/internal/common/connect"
	"dos/internal/common/convert"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
)

//go:generate mockgen -source=$GOFILE -destination=mock/session_mock.go -package=mock

type Config interface {
	FrameSize() int64
	RPCTimeout() time.Duration
}

type Transport struct {
	conn   *connect.ConnCache
	config Config
}

func NewTransport(conn *connect.ConnCache, config Config) (*Transport, error) {
	if conn == nil {
		return nil, errors.New("missing connection pool")
	}
	if config == nil {
		return nil, errors.New("missing config")
	}
	return &Transport{conn: conn, config: config}, nil
}

type UploadSession interface {
	Upload(context.Context, *t.Chunk) (t.NodeRef, error)
}

func (st *Transport) NewUploadSession(nodes []t.NodeRef, opts ...SessionOption) UploadSession {
	features := applySessionOptions(opts)
	return &uploadSession{
		config:  st.config,
		targets: nodes,
		uploader: &ChunkUploader{
			conn:       st.conn,
			config:     st.config,
			onProgress: features.onProgress,
		},
	}
}

type DownloadSession interface {
	Download(context.Context, t.ChunkID) (t.Chunk, error)
}

func (st *Transport) NewDownloadSession(nodes []t.NodeRef, opts ...SessionOption) DownloadSession {
	features := applySessionOptions(opts)
	return &downloadSession{
		config:  st.config,
		targets: nodes,
		downloader: &ChunkDownloader{
			conn:       st.conn,
			onProgress: features.onProgress,
		},
	}
}

func (t *Transport) ReplicateChunk(
	ctx context.Context, chunkID t.ChunkID, source t.NodeRef, targets []t.NodeRef,
) error {

	conn, err := t.conn.Get(source.Addr)
	if err != nil {
		return fmt.Errorf("get conn: %w", err)
	}
	client := spb.NewChunkServiceClient(conn)

	req := &spb.ReplicateChunkRequest{
		NodeId:  string(source.ID),
		ChunkId: string(chunkID),
		Targets: utils.Map(targets, convert.NodeRefToPB),
	}

	ctx, cancel := context.WithTimeout(ctx, t.config.RPCTimeout())
	defer cancel()

	if _, err = client.ReplicateChunk(ctx, req); err != nil {
		return fmt.Errorf("replicate chunk rpc: %w", err)
	}
	return nil
}

func (t *Transport) DeleteChunk(ctx context.Context, chunkID t.ChunkID, node t.NodeRef) error {

	conn, err := t.conn.Get(node.Addr)
	if err != nil {
		return fmt.Errorf("get conn: %w", err)
	}
	client := spb.NewChunkServiceClient(conn)

	req := &spb.DeleteChunkRequest{
		NodeId:  string(node.ID),
		ChunkId: string(chunkID),
	}

	ctx, cancel := context.WithTimeout(ctx, t.config.RPCTimeout())
	defer cancel()

	if _, err = client.DeleteChunk(ctx, req); err != nil {
		return fmt.Errorf("delete chunk rpc: %w", err)
	}
	return nil
}
