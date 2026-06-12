package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	mpb "dos/gen/proto/master/v1"
	pb "dos/gen/proto/master/v1"
	"dos/internal/common/convert"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
	m "dos/internal/services/master"
)


var _ mpb.MasterClientServiceServer = (*ClientServer)(nil)

type ClientFacade interface {
	CreateObject(context.Context, t.ObjectID) error
	AllocateChunk(context.Context, m.AllocateChunkCommand) (*t.ChunkAllocation, error)
	SetReplication(context.Context, t.ObjectID, int) error
}

type ResourceView interface {
	DescribeChunk(context.Context, t.ChunkID) (*t.ChunkDesc, error)
	DescribeObject(context.Context, t.ObjectID) (*t.ObjectDesc, error)
}

type ClientServer struct {
	pb.UnimplementedMasterClientServiceServer
	facade ClientFacade
	view ResourceView 
}


func NewClientServer(facade ClientFacade, view ResourceView) (*ClientServer, error) {
	if facade == nil {
		return nil, errors.New("missing facade")
	}

	if view == nil {
		return nil, errors.New("missing view")
	}

	s := &ClientServer{facade: facade, view: view}
	return s, nil
}

func (s *ClientServer) CreateObject(
	ctx context.Context, req *pb.CreateObjectRequest,
) (rsp *pb.CreateObjectResponse, err error) {

	defer func() {
		if err != nil {
			slog.ErrorContext(ctx, "create object failed",
				"object_id", req.GetObjectId(), "error", err,
			)
			err = toStatus(err)
		}
	}()
	slog.DebugContext(ctx, "create object requested", "object_id", req.GetObjectId())

	if err = validateCreateObjectRequest(req); err != nil {
		return nil, err
	}

	err = s.facade.CreateObject(ctx, t.ObjectID(req.GetObjectId()))
	if err != nil {
		return nil, fmt.Errorf("create object %s: %w", req.GetObjectId(), err)
	}
	return &pb.CreateObjectResponse{}, nil
}

func (s *ClientServer) AllocateChunk(
	ctx context.Context, req *pb.AllocateChunkRequest,
) (rsp *pb.AllocateChunkResponse, err error) {
	
	slot := req.GetObjectSlot()
	defer func() {
		if err != nil {
			slog.ErrorContext(ctx, "allocate chunk failed",
				"object_id", slot.GetObjectId(),
				"chunk_key", slot.GetChunkKey(),
				"chunk_size", req.GetChunkSize(),
				"error", err,
			)
			err = toStatus(err)
		}
	}()
	slog.Info("allocate chunk requested", "object_id", slot.GetObjectId())

	if err = validateAllocateChunkRequest(req); err != nil {
		return nil, err
	}

	alloc, err := s.facade.AllocateChunk(ctx, m.AllocateChunkCommand{
		Slot:         convert.ObjectSlotFromPB(req.GetObjectSlot()),
		Size:         req.GetChunkSize(),
		ExcludeNodes: utils.Map(req.GetExcludeNodes(), convert.NodeRefFromPB),
	})
	if err != nil {
		return nil, err
	}

	rsp = &pb.AllocateChunkResponse{
		ChunkId:    string(alloc.ID),
		ObjectSlot: convert.ObjectSlotToPB(alloc.Slot),
		Targets:    utils.Map(alloc.Targets, convert.NodeRefToPB),
	}
	return rsp, nil
}

func (s *ClientServer) SetReplication(
	ctx context.Context, req *pb.SetReplicationRequest,
) (rsp *pb.SetReplicationResponse, err error) {

	defer func() {
		if err != nil {
			slog.ErrorContext(ctx, "set replication failed",
				"object_id", req.GetObjectId(), "error", err,
			)
			err = toStatus(err)
		}
	}()

	slog.DebugContext(ctx, "set replication requested",
		"object_id", req.GetObjectId(),
		"count", req.GetCount(),
	)
	err = s.facade.SetReplication(ctx, t.ObjectID(req.GetObjectId()), int(req.GetCount()))
	if err != nil {
		return nil, err
	}
	rsp = &pb.SetReplicationResponse{}
	return rsp, nil
}

func (s *ClientServer) DescribeChunk(
	ctx context.Context, req *mpb.DescribeChunkRequest,
) (rsp *mpb.DescribeChunkResponse, err error) {

	defer func() {
		if err != nil {
			slog.ErrorContext(ctx, "describe chunk failed", 
				"chunk_id", req.GetChunkId(), "error", err,
			)
			err = toStatus(err)
		}
	}()
	slog.DebugContext(ctx, "describe chunk requested", "chunk_id", req.GetChunkId())

	chunkID := t.ChunkID(req.GetChunkId())
	chunkDesc, err := s.view.DescribeChunk(ctx, chunkID)
	if err != nil {
		return nil, err
	}

	rsp = &mpb.DescribeChunkResponse{
		Description: convert.ChunkDesc1ToPB(*chunkDesc),
	}
	return rsp, nil
}

func (s *ClientServer) DescribeObject(
	ctx context.Context,
	req *mpb.DescribeObjectRequest,
) (rsp *mpb.DescribeObjectResponse, err error) {

	defer func() {
		if err != nil {
			slog.ErrorContext(ctx, "describe objcet failed", 
				"object_id", req.GetObjectId(), "error", err,
			)
			err = toStatus(err)
		}
	}()
	slog.DebugContext(ctx, "describe object requested", "object_id", req.GetObjectId())

	objectID := t.ObjectID(req.GetObjectId())
	objectDesc, err := s.view.DescribeObject(ctx, objectID)
	if err != nil {
		return nil, err
	}
	
	rsp = &mpb.DescribeObjectResponse{
		Description: convert.ObjectDesc1ToPB(*objectDesc),
	}
	return rsp, nil
}

