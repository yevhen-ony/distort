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

type ChunkPlacement struct {
	ChunkID c.ChunkID
	
}

func (mt *MasterTransport) AllocateChunk(ctx context.Context, query *AllocateChunkQuery)  {
}
