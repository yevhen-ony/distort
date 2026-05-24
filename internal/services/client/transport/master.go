package transport

import (
	"context"
	"errors"
	"fmt"

	pb "dos/gen/proto/master/v1"
	"dos/internal/common/connect"
	"dos/internal/common/convert"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
)

type MasterTransportConfig interface {
	MasterAddr() string
}

type MasterTransport struct {
	client pb.MasterClientServiceClient
	config MasterTransportConfig
}

func NewMasterTransport(conn *connect.ConnCache, config MasterTransportConfig) (*MasterTransport, error) {
	if conn == nil {
		return nil, errors.New("missing conn")
	}
	if config == nil {
		return nil, errors.New("missing config")
	}

	c, err := conn.Get(config.MasterAddr())
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

type AllocateChunkCommand struct {
	Slot      t.ObjectSlot
	ChunkSize int64
}

func (mt *MasterTransport) AllocateChunk(
	ctx context.Context, query *AllocateChunkCommand,
) (*t.ChunkAllocation1, error) {

	req := &pb.AllocateChunkRequest{
		ObjectSlot: convert.ObjectSlotToPB(query.Slot),
		ChunkSize:  query.ChunkSize,
	}
	rsp, err := mt.client.AllocateChunk(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("transport: %w", err)
	}

	alloc := &t.ChunkAllocation1{
		ID:       t.ChunkID(rsp.GetChunkId()),
		Slot:     convert.ObjectSlotFromPB(rsp.GetObjectSlot()),
		Targets: utils.Map(rsp.GetTargets(), convert.NodeRefFromPB),
	}
	return alloc, nil
}

func (mt *MasterTransport) GetObjectAccess(
	ctx context.Context, oid t.ObjectID,
) (t.ObjectAccess, error) {

	req := &pb.GetObjectAccessRequest{ObjectId: string(oid)}
	rsp, err := mt.client.GetObjectAccess(ctx, req)
	if err != nil {
		return t.ObjectAccess{}, fmt.Errorf("transport: %w", err)
	}
	objAccess := *convert.ObjectAccessFromPB(rsp)
	return objAccess, nil
}

func (mt *MasterTransport) ListObjects(ctx context.Context) ([]t.ObjectInfo, error) {

	rsp, err := mt.client.ListObjects(ctx, &pb.ListObjectsRequest{})
	if err != nil {
		return nil, fmt.Errorf("transport: %w", err)
	}
	infos := utils.Map(rsp.GetObjects(), convert.ObjectInfoFromPB)
	return infos, nil
}

func (mt *MasterTransport) ListChunks(ctx context.Context) ([]t.ChunkInfo, error) {

	rsp, err := mt.client.ListChunks(ctx, &pb.ListChunksRequest{})
	if err != nil {
		return nil, fmt.Errorf("transport: %w", err)
	}
	infos := utils.Map(rsp.GetChunks(), convert.ChunkInfoFromPB)
	return infos, nil
}

func (mt *MasterTransport) ListNodes(ctx context.Context) ([]t.NodeInfo, error) {

	rsp, err := mt.client.ListNodes(ctx, &pb.ListNodesRequest{})
	if err != nil {
		return nil, fmt.Errorf("transport: %w", err)
	}
	infos := utils.Map(rsp.GetNodes(), convert.NodeInfoFromPB)
	return infos, nil
}

func (mt *MasterTransport) SetReplication(ctx context.Context, objectID t.ObjectID, count int) error {
	req := &pb.SetReplicationRequest{
		ObjectId: string(objectID),
		Count:    int64(count),
	}

	_, err := mt.client.SetReplication(ctx, req)
	if err != nil {
		return fmt.Errorf("transport: %w", err)
	}
	return nil
}
