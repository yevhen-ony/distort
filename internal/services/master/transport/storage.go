package transport

import (
	"context"
	"fmt"

	spb "dos/gen/proto/storage/v1"
	"dos/internal/common/connect"
	"dos/internal/common/convert"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
)

type Storage struct {
	conn *connect.ConnCache
}

func NewStorage(conn *connect.ConnCache) *Storage {
	return &Storage{conn: conn}
}

func (t *Storage) ReplicateChunk(
	ctx context.Context, chunkID t.ChunkID, source t.NodeRef, targets []t.NodeRef,
) error {

	conn, err := t.conn.Get(source.Addr)
	if err != nil {
		return fmt.Errorf("get conn: %w", err)
	}
	client := spb.NewChunkServiceClient(conn)


	req := &spb.ReplicateChunkRequest{
		NodeId: string(source.ID),
		ChunkId: string(chunkID),
		Targets: utils.Map(targets, convert.NodeRefToPB),
	}
	if _, err = client.ReplicateChunk(ctx, req); err != nil {
		return fmt.Errorf("replicate chunk rpc: %w",  err)
	}
	return nil
}

func (t *Storage) DeleteChunk(ctx context.Context, chunkID t.ChunkID, node t.NodeRef) error {

	conn, err := t.conn.Get(node.Addr)
	if err != nil {
		return fmt.Errorf("get conn: %w", err)
	}
	client := spb.NewChunkServiceClient(conn)

	req := &spb.DeleteChunkRequest{
		NodeId: string(node.ID),
		ChunkId: string(chunkID),
	}
	if _, err = client.DeleteChunk(ctx, req); err != nil {
		return fmt.Errorf("delete chunk rpc: %w", err)
	}
	return nil
}
