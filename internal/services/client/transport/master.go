package transport

import (
	"context"
	"errors"
	"fmt"

	
	pb "dos/gen/proto/master/v1"
	"dos/internal/common/connect"
	"dos/internal/common/convert"
	t "dos/internal/common/types"
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
		return fmt.Errorf("create object: %w", err)
	}

	return nil
}

type AllocateChunkQuery struct {
	ObjectID t.ObjectID
	ChunkKey t.ChunkKey
	ChunkSize int64
}

func (mt *MasterTransport) AllocateChunk(
	ctx context.Context, query *AllocateChunkQuery,
) (t.ChunkLocation, error)  {
	conn, err := mt.conn.Get(mt.config.Addr)
	if err != nil {
		return t.ChunkLocation{}, fmt.Errorf("get conn: %w", err) 
	}
	client := pb.NewMasterClientServiceClient(conn)

	req := &pb.AllocateChunkRequest{
		ObjectId: string(query.ObjectID),
		ChunkKey: string(query.ChunkKey),
		ChunkSize: query.ChunkSize,
	}
	rsp, err := client.AllocateChunk(ctx, req)
	if err != nil {
		return t.ChunkLocation{}, fmt.Errorf("allocate chunk: %w", err) 
	}
	chunks := *convert.ChunkPlacementFromPB(rsp)
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
		return t.ObjectAccess{}, fmt.Errorf("get object access: %w", err)
	}
	objAccess := *convert.ObjectAccessFromPB(rsp)
	return objAccess, nil  
}
