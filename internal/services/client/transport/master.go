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
	client pb.MasterClientServiceClient 
	config *MasterTransportConfig
}

func NewMasterTransport(conn *connect.ConnCache, config *MasterTransportConfig) (*MasterTransport, error) {
	if conn == nil {
		return nil, errors.New("missing conn")
	}
	if config == nil {
		return nil, errors.New("missing config")
	}

	c, err := conn.Get(config.Addr)
	if err != nil {
		return nil, fmt.Errorf("get conn: %w", err)
	}
	client := pb.NewMasterClientServiceClient(c)
		
	return &MasterTransport{client: client, config: config}, nil
}

func (mt *MasterTransport) CreateObject(ctx context.Context, oid t.ObjectID) error {
	
	req := &pb.CreateObjectRequest{ObjectId: string(oid)}

	_, err := mt.client.CreateObject(ctx, req)
	if err != nil {
		return fmt.Errorf("transport: %w", err)
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
) (t.ChunkPlacement, error)  {

	req := &pb.AllocateChunkRequest{
		ObjectId: string(query.ObjectID),
		ChunkKey: string(query.ChunkKey),
		ChunkSize: query.ChunkSize,
	}
	rsp, err := mt.client.AllocateChunk(ctx, req)
	if err != nil {
		return t.ChunkPlacement{}, fmt.Errorf("transport: %w", err) 
	}
	chunks := *convert.ChunkPlacementFromPB(rsp.GetChunk())
	return chunks, nil
}

func (mt *MasterTransport) GetObjectAccess(
	ctx context.Context, oid t.ObjectID,
) (t.ObjectAccess, error)  {

	req := &pb.GetObjectAccessRequest{ObjectId: string(oid)}
	rsp, err := mt.client.GetObjectAccess(ctx, req)
	if err != nil {
		return t.ObjectAccess{}, fmt.Errorf("transport: %w", err)
	}
	objAccess := *convert.ObjectAccessFromPB(rsp)
	return objAccess, nil  
}

func (mt *MasterTransport) ListObjects( ctx context.Context) ([]t.ObjectItem, error) {

	rsp, err :=  mt.client.ListObjects(ctx, &pb.ListObjectsRequest{})
	if err != nil {
		return nil, fmt.Errorf("transport: %w", err)
	}

	pbItems := rsp.GetObjects()
	items := make([]t.ObjectItem, len(pbItems))
	for i, pbItem := range pbItems {
		items[i] = convert.ObjectItemFromPB(pbItem)
	}
	return items, nil
}

