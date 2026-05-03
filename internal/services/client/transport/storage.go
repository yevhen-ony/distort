package transport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	pb "dos/gen/proto/common/v1"
	spb "dos/gen/proto/storage/v1"
	"dos/internal/common/connect"
	"dos/internal/common/convert"
	"dos/internal/common/digest"
	t "dos/internal/common/types"
	c "dos/internal/services/client"
)

type StorageTransport struct {
	conn *connect.ConnCache
	config *StorageTransportConfig
} 

func NewStorageTransport(conn *connect.ConnCache, config *StorageTransportConfig) (*StorageTransport, error) {
	if conn == nil {
		return nil, errors.New("missing connection pool")
	}
	if config == nil {
		return nil, errors.New("missing config")
	}
	return &StorageTransport{conn: conn, config: config}, nil
}

func (ct *StorageTransport) PushChunk(ctx context.Context, nodes []t.NodeRef, chunk *c.Chunk) error {
	var errs []error
	for _, node := range nodes {
		err := ct.pushChunkToNode(ctx, node, chunk) 
		if err == nil {
			return nil	
		}
		slog.WarnContext(ctx, "send chunk failed", "addr", node.Addr, "chunk", chunk.Meta.ID, "error", err)
		errs = append(errs, fmt.Errorf("send chunk %s to %s failed: %w", chunk.Meta.ID, node.Addr, err))
	}
	return fmt.Errorf("all candidate nodes failed: %w", errors.Join(errs...))	
}

func (ct *StorageTransport) pushChunkToNode(ctx context.Context, node t.NodeRef, chunk *c.Chunk) error {
	conn, err := ct.conn.Get(node.Addr)
	if err != nil {
		return fmt.Errorf("get conn: %w", err) 
	}

	client := spb.NewChunkServiceClient(conn)

	stream, err := client.PutChunk(ctx)
	if err != nil {
		return fmt.Errorf("open put stream: %w", err)
	}
	header := &spb.PutChunkHeader{
		NodeId: string(node.ID),
		ChunkId: string(chunk.Meta.ID),
		Digest: &pb.Digest{
			Size: int64(chunk.Meta.Digest.Size),
			Checksum: string(chunk.Meta.Digest.Checksum),
		},
	}

	err = stream.Send(&spb.PutChunkRequest{Header: header})
	if err != nil {
		return fmt.Errorf("send header: %w", err)
	}

	if err = ct.sendData(stream, chunk.Data); err != nil {
		return fmt.Errorf("send data: %w", err) 
	}

	if _, err := stream.CloseAndRecv(); err != nil {
		return fmt.Errorf("close stream: %w", err)
	}

	return nil
}

func (ct *StorageTransport) PullChunk(
	ctx context.Context, nodes []t.NodeRef, chunkID t.ChunkID,
) (c.Chunk, error) {

	var errs []error
	for _, node := range nodes {
		chunk, err := ct.pullChunkFromNode(ctx, node, chunkID)
		if err == nil {
			return chunk, nil
		}
		slog.WarnContext(ctx, "pull chunk failed", "addr", node.Addr, "chunk", chunkID, "error", err)
		errs = append(errs, fmt.Errorf("send chunk %s to %s: %w", chunkID, node.Addr, err))
	}
	return c.Chunk{} , fmt.Errorf("all candidate nodes failed: %w", errors.Join(errs...))
}

func (ct *StorageTransport) pullChunkFromNode(
	ctx context.Context, node t.NodeRef, chunkID t.ChunkID,
) (c.Chunk, error) {

	if err := ReceiveChunkValidate(node, chunkID); err != nil {
		return c.Chunk{}, err
	}

	conn, err := ct.conn.Get(node.Addr)
	if err != nil {
		return c.Chunk{}, fmt.Errorf("get conn: %w", err)
	}
	
	client := spb.NewChunkServiceClient(conn)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	stream, err := client.GetChunk(ctx, &spb.GetChunkRequest{
		NodeId: string(node.ID),
		ChunkId: string(chunkID), 
	})
	if err != nil {
		return c.Chunk{}, fmt.Errorf("send request: %w", err)
	}

	rsp, err := stream.Recv()
	if err != nil {
		return c.Chunk{}, fmt.Errorf("recv header: %w", err)
	}

	header := rsp.GetHeader()
	if header == nil {
		return c.Chunk{}, ErrHeaderInvalid 
	}

	data, digest, err := ct.recvData(stream)
	if err != nil {
		return c.Chunk{}, fmt.Errorf("recv data: %w", err)
	}

	meta := t.ChunkMeta{
		ID: chunkID,
		Digest: digest,
	}

	err = matchChunkMeta(convert.ChunkDescFromPB(header), meta)
	if err != nil {
		return c.Chunk{}, err 
	}

	chunk :=  c.Chunk{
		Meta: meta,
		Data: data,
	}
	return chunk, nil 
}

func (ct *StorageTransport) sendData(stream spb.ChunkService_PutChunkClient, data []byte) error {
	for len(data) > 0 {
		n := min(int64(ct.config.FrameSize), int64(len(data)))
		err := stream.Send(&spb.PutChunkRequest{Data: data[:n]})
		if err != nil {
			return err 
		}
		data = data[n:]
	}
	return nil
}

func (ct *StorageTransport) recvData(stream spb.ChunkService_GetChunkClient) ([]byte, digest.Digest, error) {
	var buf bytes.Buffer
	dg := digest.New()

	for {
		rsp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, digest.Digest{}, err
		}

		data := rsp.GetData()
		if data == nil {
			return nil, digest.Digest{}, ErrDataInvalid 
		}

		buf.Write(rsp.Data)
		dg.Write(rsp.Data)
	}
	return buf.Bytes(), dg.Digest(), nil
}

func matchChunkMeta(want, got t.ChunkMeta) error {
	if err := got.Digest.Match(want.Digest); err != nil {
		return err
	}

	if want.ID != got.ID {
		return fmt.Errorf("id mismatch: %w", ErrChunkMetaMismatch)
	}
	return nil
}
