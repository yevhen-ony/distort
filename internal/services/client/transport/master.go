package transport

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "dos/gen/proto/master/v1"
	"dos/internal/common/convert"
	"dos/internal/common/master/route"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
)

type MasterTransportConfig interface {
	RPCTimeout() time.Duration
}

type MasterTransportDeps struct {
	Config MasterTransportConfig
	Router *route.MasterRouter
}

type MasterTransport struct {
	mrouter *route.MasterRouter
	config  MasterTransportConfig
}

func NewMasterTransport(deps MasterTransportDeps) (*MasterTransport, error) {
	if deps.Router == nil {
		return nil, errors.New("missing master router")
	}
	if deps.Config == nil {
		return nil, errors.New("missing config")
	}

	mt := &MasterTransport{
		mrouter: deps.Router,
		config:  deps.Config,
	}
	return mt, nil
}

func (mt *MasterTransport) client(ctx context.Context) (pb.MasterClientServiceClient, error) {
	conn, err := mt.mrouter.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("get master conn: %w", err)
	}
	return pb.NewMasterClientServiceClient(conn), nil
}

func (mt *MasterTransport) admin(ctx context.Context) (pb.AdminServiceClient, error) {
	conn, err := mt.mrouter.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("get master conn: %w", err)
	}
	return pb.NewAdminServiceClient(conn), nil
}

func (mt *MasterTransport) CreateObject(ctx context.Context, oid t.ObjectID) error {

	ctx, cancel := context.WithTimeout(ctx, mt.config.RPCTimeout())
	defer cancel()

	req := &pb.CreateObjectRequest{ObjectId: string(oid)}

	client, err := mt.client(ctx)
	if err != nil {
		return err
	}

	if _, err = client.CreateObject(ctx, req); err != nil {
		return fmt.Errorf("transport: %w", err)
	}

	return nil
}

type AllocateChunkCommand struct {
	Slot      t.ObjectSlot
	ChunkSize int64
}

func (mt *MasterTransport) AllocateChunk(
	ctx context.Context,
	query *AllocateChunkCommand,
) (*t.ChunkAllocation, error) {

	ctx, cancel := context.WithTimeout(ctx, mt.config.RPCTimeout())
	defer cancel()

	req := &pb.AllocateChunkRequest{
		ObjectSlot: convert.ObjectSlotToPB(query.Slot),
		ChunkSize:  query.ChunkSize,
	}

	client, err := mt.client(ctx)
	if err != nil {
		return nil, err
	}

	rsp, err := client.AllocateChunk(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("transport: %w", err)
	}

	alloc := &t.ChunkAllocation{
		ID:      t.ChunkID(rsp.GetChunkId()),
		Slot:    convert.ObjectSlotFromPB(rsp.GetObjectSlot()),
		Targets: utils.Map(rsp.GetTargets(), convert.NodeRefFromPB),
	}
	return alloc, nil
}

func (mt *MasterTransport) SetReplication(ctx context.Context, objectID t.ObjectID, count int) error {
	req := &pb.SetReplicationRequest{
		ObjectId: string(objectID),
		Count:    int64(count),
	}

	ctx, cancel := context.WithTimeout(ctx, mt.config.RPCTimeout())
	defer cancel()

	client, err := mt.client(ctx)
	if err != nil {
		return err
	}

	_, err = client.SetReplication(ctx, req)
	if err != nil {
		return fmt.Errorf("transport: %w", err)
	}
	return nil
}

func (mt *MasterTransport) DescribeChunk(
	ctx context.Context,
	chunkID t.ChunkID,
) (*t.ChunkDesc, error) {

	req := &pb.DescribeChunkRequest{
		ChunkId: string(chunkID),
	}

	ctx, cancel := context.WithTimeout(ctx, mt.config.RPCTimeout())
	defer cancel()

	client, err := mt.client(ctx)
	if err != nil {
		return nil, err
	}

	rsp, err := client.DescribeChunk(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("transport: %w", err)
	}

	description := convert.ChunkDesc1FromPB(rsp.GetDescription())
	return &description, nil
}

func (mt *MasterTransport) DescribeObject(
	ctx context.Context,
	objectID t.ObjectID,
) (*t.ObjectDesc, error) {

	ctx, cancel := context.WithTimeout(ctx, mt.config.RPCTimeout())
	defer cancel()

	req := &pb.DescribeObjectRequest{
		ObjectId: string(objectID),
	}

	client, err := mt.client(ctx)
	if err != nil {
		return nil, err
	}

	rsp, err := client.DescribeObject(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("transport: %w", err)
	}

	description := convert.ObjectDesc1FromPB(rsp.GetDescription())
	return &description, nil
}
