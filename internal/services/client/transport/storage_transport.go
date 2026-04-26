package transport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	pb "dos/gen/proto/chunk/v1"
	t "dos/internal/common/types"
	c "dos/internal/services/client"
	"dos/internal/common/digest"
)

type StorageTransport struct {
	conn *ConnectionPool
	config *StorageTransportConfig
} 

func NewChunkTransport(conn *ConnectionPool, config *StorageTransportConfig) (*StorageTransport, error) {
	if conn == nil {
		return nil, errors.New("missing connection pool")
	}
	if config == nil {
		return nil, errors.New("missing config")
	}
	return &StorageTransport{conn: conn, config: config}, nil
}

func (ct *StorageTransport) SendChunk(ctx context.Context, node t.NodeRef, chunk *c.Chunk) error {
	if err := SendChunkValidate(node, chunk); err != nil {
		return err
	}

	conn, err := ct.conn.Get(node.Addr)
	if err != nil {
		return fmt.Errorf("get conn: %w", err) 
	}

	client := pb.NewChunkServiceClient(conn)

	stream, err := client.PutChunk(ctx)
	if err != nil {
		return fmt.Errorf("open put stream: %w", err)
	}
	header := &pb.PutChunkHeader{
		NodeId: string(node.NodeID),
		ChunkId: string(chunk.ID),
		ChunkSize: int64(len(chunk.Data)),
		Checksum: string(chunk.Checksum),
	}

	err = stream.Send(&pb.PutChunkRequest{Header: header})
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

func (ct *StorageTransport) ReceiveChunk(ctx context.Context, node t.NodeRef, chunkID t.ChunkID) (*c.Chunk, error) {
	if err := ReceiveChunkValidate(node, chunkID); err != nil {
		return nil, err
	}

	conn, err := ct.conn.Get(node.Addr)
	if err != nil {
		return nil, fmt.Errorf("get conn: %w", err)
	}
	
	client := pb.NewChunkServiceClient(conn)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	stream, err := client.GetChunk(ctx, &pb.GetChunkRequest{
		NodeId: string(node.NodeID),
		ChunkId: string(chunkID), 
	})
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	rsp, err := stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("recv header: %w", err)
	}

	header := rsp.GetHeader()
	if header == nil {
		return nil, ErrHeaderInvalid 
	}

	data, checksum, err := ct.recvData(stream)
	if err != nil {
		return nil, fmt.Errorf("recv data: %w", err)
	}
	chunk :=  &c.Chunk{
		ID: chunkID,
		Checksum: checksum,
		Data: data,
	}
	if err := ValidateReceivedChunk(chunk, header); err != nil {
		return nil, err 
	}
	return chunk, nil 
}

func (ct *StorageTransport) sendData(stream pb.ChunkService_PutChunkClient, data []byte) error {
	for len(data) > 0 {
		n := ct.config.FrameSize
		if len(data) < n {
			n = len(data)
		}

		err := stream.Send(&pb.PutChunkRequest{Data: data[:n]})
		if err != nil {
			return err 
		}
		data = data[n:]
	}
	return nil
}

func (ct *StorageTransport) recvData(stream pb.ChunkService_GetChunkClient) ([]byte, t.Checksum, error) {
	var buf bytes.Buffer
	dg := digest.New()

	for {
		rsp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, "", err
		}

		data := rsp.GetData()
		if data == nil {
			return nil, "", ErrDataInvalid 
		}

		buf.Write(rsp.Data)
		dg.Write(rsp.Data)
	}
	return buf.Bytes(), dg.Checksum(), nil
}
