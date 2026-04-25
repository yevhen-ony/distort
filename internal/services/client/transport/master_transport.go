package transport

import (
	"context"
	"errors"
	"fmt"

	pb "dos/gen/proto/master/v1"
	c "dos/internal/services/client"
)

type MasterTransport struct {
	conn   *ConnectionPool
	config *MasterTransportConfig
}

func NewMasterTransport(conn *ConnectionPool, config *MasterTransportConfig) (*MasterTransport, error) {
	if conn == nil {
		return nil, errors.New("missing connection pool")
	}
	if config == nil {
		return nil, errors.New("missing config")
	}
	return &MasterTransport{conn: conn, config: config}, nil
}

func (mt *MasterTransport) CreateObject(ctx context.Context, oid c.ObjectID) error {
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
	ObjectID c.ObjectID
	ChunkKey c.ChunkKey
	ChunkSize int64
}

func (mt *MasterTransport) AllocateChunk(
	ctx context.Context, query *AllocateChunkQuery,
) (c.ChunkPlacement, error)  {
	conn, err := mt.conn.Get(mt.config.Addr)
	if err != nil {
		return c.ChunkPlacement{}, fmt.Errorf("get conn: %w", err) 
	}
	client := pb.NewMasterClientServiceClient(conn)

	req := &pb.AllocateChunkRequest{
		ObjectId: string(query.ObjectID),
		ChunkKey: string(query.ChunkKey),
		ChunkSize: query.ChunkSize,
	}
	rsp, err := client.AllocateChunk(ctx, req)
	if err != nil {
		return c.ChunkPlacement{}, fmt.Errorf("allocate chunk: %w", err) 
	}
	chunks := *ChunkPlacementFromPB(rsp)
	return chunks, nil
}

func (mt *MasterTransport) GetObjectAccess(
	ctx context.Context, oid c.ObjectID,
) (c.ObjectAccess, error)  {
	conn, err := mt.conn.Get(mt.config.Addr)
	if err != nil {
		return c.ObjectAccess{}, fmt.Errorf("get conn: %w", err)
	}
	client :=  pb.NewMasterClientServiceClient(conn)

	req := &pb.GetObjectAccessRequest{ObjectId: string(oid)}
	rsp, err := client.GetObjectAccess(ctx, req)
	if err != nil {
		return c.ObjectAccess{}, fmt.Errorf("get object access: %w", err)
	}
	objAccess := *ObjectAccessFromPB(rsp)
	return objAccess, nil  
}
