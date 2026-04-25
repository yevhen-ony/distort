package api

import (
	"context"
	"fmt"
	"log/slog"

	pb "dos/gen/proto/master/v1"
	t "dos/internal/common/types"
	"dos/internal/common/convert"
	m "dos/internal/services/master"

)

type ClientServer struct {
	pb.UnimplementedMasterClientServiceServer
	service m.Service
}

func NewClientServer(service m.Service) *ClientServer {
	return &ClientServer{service: service}
}

func (s *ClientServer) CreateObject(
	ctx context.Context, req *pb.CreateObjectRequest,
) (rsp *pb.CreateObjectResponse, err error) {

	defer func() { err = toStatus(err) }()
	slog.Info("create object requested", "object_id", req.GetObjectId())

	if err = validateCreateObjectRequest(req); err != nil {
		return nil, err
	}

	err = s.service.CreateObject(ctx, t.ObjectID(req.GetObjectId()))
	if err != nil {
		return nil, fmt.Errorf("create object %s: %w", req.GetObjectId(), err)
	}
	return &pb.CreateObjectResponse{}, nil
}

func (s *ClientServer) AllocateChunk(
	ctx context.Context, req *pb.AllocateChunkRequest,
) (rsp *pb.AllocateChunkResponse, err error) {

	defer func() { err = toStatus(err) }()
	slog.Info("allocate chunk requested", "object_id", req.GetObjectId())

	if err = validateAllocateChunkRequest(req); err != nil {
		return nil, err
	}

	cmd := &m.AllocateChunkCommand{
		ObjectID:  t.ObjectID(req.GetObjectId()),
		ChunkKey:  t.ChunkKey(req.GetChunkKey()),
		ChunkSize: req.GetChunkSize(),
	}
	chunks, err := s.service.AllocateChunk(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("allocate chunk for object %s: %w", req.GetObjectId(), err)
	}

	rsp = &pb.AllocateChunkResponse{
		ChunkId: string(chunks.ChunkID),
		Nodes:   convert.NodeAccessToPB(chunks.Nodes),
	}

	return rsp, nil
}

func (s *ClientServer) GetObjectAccess(
	ctx context.Context, req *pb.GetObjectAccessRequest,
) (rsp *pb.GetObjectAccessResponse, err error) {

	defer func() { err = toStatus(err) }()
	slog.Info("object access requested", "object_id", req.GetObjectId())

	if err = validateGetObjectAccessRequest(req); err != nil {
		return nil, err
	}

	object, err := s.service.GetObjectAccess(ctx, t.ObjectID(req.GetObjectId()))
	if err != nil {
		return nil, fmt.Errorf("get object access %s: %w", req.GetObjectId(), err)
	}

	rsp = &pb.GetObjectAccessResponse{
		ObjectId:  string(object.ObjectID),
		TotalSize: object.TotalSize,
		Chunks:    convert.ChunkPlacementToPB(object.Chunks),
	}
	return rsp, nil
}
