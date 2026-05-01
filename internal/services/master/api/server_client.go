package api

import (
	"context"
	"fmt"
	"log/slog"

	pb "dos/gen/proto/master/v1"
	"dos/internal/common/convert"
	t "dos/internal/common/types"
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

	err = s.service.CreateObject(ctx, t.ObjectID(req.GetObjectId()))
	if err != nil {
		return nil, fmt.Errorf("create object %s: %w", req.GetObjectId(), err)
	}
	return &pb.CreateObjectResponse{}, nil
}

func (s *ClientServer) AllocateChunk(
	ctx context.Context, req *pb.AllocateChunkRequest,
) (rsp *pb.AllocateChunkResponse, err error) {

	defer func() {
		if err != nil {
			slog.ErrorContext(ctx, "allocate chunk failed",
				"object_id", req.GetObjectId(),
				"chunk_key", req.GetChunkKey(),
				"chunk_size", req.GetChunkSize(),
				"error", err,
			)
			err = toStatus(err)
		}
	}()
	slog.Info("allocate chunk requested", "object_id", req.GetObjectId())

	if err = validateAllocateChunkRequest(req); err != nil {
		return nil, err
	}

	cmd := &m.AllocateChunkCommand{
		ObjectID:  t.ObjectID(req.GetObjectId()),
		ChunkKey:  t.ChunkKey(req.GetChunkKey()),
		ChunkSize: req.GetChunkSize(),
	}
	chunk, err := s.service.AllocateChunk(ctx, cmd)
	if err != nil {
		return nil, err
	}

	rsp = &pb.AllocateChunkResponse{
		Chunk: &pb.ChunkPlacement{
			ChunkId:   string(chunk.ChunkID),
			ChunkKey:  string(chunk.ChunkKey),
			ChunkSize: chunk.ChunkSize,
			Nodes:     convert.NodeRefToPB(chunk.Nodes...),
		},
	}

	return rsp, nil
}

func (s *ClientServer) GetObjectAccess(
	ctx context.Context, req *pb.GetObjectAccessRequest,
) (rsp *pb.GetObjectAccessResponse, err error) {

	defer func() {
		if err != nil {
			slog.ErrorContext(ctx, "object access failed",
				"object_id", req.GetObjectId(), "error", err,
			)
			err = toStatus(err)
		}
	}()
	slog.DebugContext(ctx, "object access requested", "object_id", req.GetObjectId())

	if err = validateGetObjectAccessRequest(req); err != nil {
		return nil, err
	}

	object, err := s.service.GetObjectAccess(ctx, t.ObjectID(req.GetObjectId()))
	if err != nil {
		return nil, err
	}

	rsp = &pb.GetObjectAccessResponse{
		ObjectId:  string(object.ID),
		TotalSize: object.TotalSize,
		Chunks:    convert.ChunkPlacementToPB(object.Chunks...),
	}
	return rsp, nil
}
