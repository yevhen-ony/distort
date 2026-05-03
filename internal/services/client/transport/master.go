package transport

import (
	"context"
	"errors"
	"fmt"

	
	pb "dos/gen/proto/master/v1"
	"dos/internal/common/connect"
	"dos/internal/common/convert"
	t "dos/internal/common/types"
	c "dos/internal/services/client"
)

type MasterTransport struct {
	conn   *connect.ConnCache
	config *MasterTransportConfig
}

func NewMasterTransport(conn *connect.ConnCache, config *MasterTransportConfig) (*MasterTransport, error) {
	if conn == nil {
		return nil, errors.New("missing conn")
	}
	if config == nil {
		return nil, errors.New("missing config")
	}
	return &MasterTransport{conn: conn, config: config}, nil
}

func (mt *MasterTransport) CreateObject(ctx context.Context, oid t.ObjectID) error {
	conn, err := mt.conn.Get(mt.config.Addr)
	if err != nil {
		return fmt.Errorf("get conn: %w", err)
	}
	client := pb.NewMasterClientServiceClient(conn)
	
	req := &pb.CreateObjectRequest{ObjectId: string(oid)}

	_, err = client.CreateObject(ctx, req)
	if err != nil {
		return fmt.Errorf("transport: %w", err)
	}

	return nil
}

func (mt *MasterTransport) AllocateChunk(
	ctx context.Context, query *c.AllocateChunkQuery,
) (t.ChunkPlacement, error)  {
	conn, err := mt.conn.Get(mt.config.Addr)
	if err != nil {
		return t.ChunkPlacement{}, fmt.Errorf("get conn: %w", err) 
	}
	client := pb.NewMasterClientServiceClient(conn)

	req := &pb.AllocateChunkRequest{
		ObjectId: string(query.ObjectID),
		ChunkKey: string(query.ChunkKey),
		ChunkSize: query.ChunkSize,
	}
	rsp, err := client.AllocateChunk(ctx, req)
	if err != nil {
		return t.ChunkPlacement{}, fmt.Errorf("transport: %w", err) 
	}
	chunks := *convert.ChunkPlacementFromPB(rsp.GetChunk())
	return chunks, nil
}

func (mt *MasterTransport) GetObjectAccess(
	ctx context.Context, oid t.ObjectID,
) (t.ObjectAccess, error)  {
	conn, err := mt.conn.Get(mt.config.Addr)
	if err != nil {
		return t.ObjectAccess{}, fmt.Errorf("get conn: %w", err)
	}
	client :=  pb.NewMasterClientServiceClient(conn)

	req := &pb.GetObjectAccessRequest{ObjectId: string(oid)}
	rsp, err := client.GetObjectAccess(ctx, req)
	if err != nil {
		return t.ObjectAccess{}, fmt.Errorf("transport: %w", err)
	}
	objAccess := *convert.ObjectAccessFromPB(rsp)
	return objAccess, nil  
}
