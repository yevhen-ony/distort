package api

import (
	"context"
	"fmt"
	"log/slog"

	mpb "dos/gen/proto/master/v1"
	pb "dos/gen/proto/master/v1"
	"dos/internal/common/convert"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
	m "dos/internal/services/master"
)

type ClientServer struct {
pb.UnimplementedMasterClientServiceServer
	facade m.ClientFacade
}

func NewClientServer(facade m.ClientFacade) *ClientServer {

	return &ClientServer{facade: facade}
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

	chunk, err := s.facade.AllocateChunk(ctx, m.AllocateChunkCommand{
		ObjectID:  t.ObjectID(req.GetObjectId()),
		ChunkKey:  t.ChunkKey(req.GetChunkKey()),
		ChunkSize: req.GetChunkSize(),
		ExcludeNodes: utils.Map(req.GetExcludeNodes(), convert.NodeRefFromPB),
	})
	if err != nil {
		return nil, err
	}

	rsp = &pb.AllocateChunkResponse{
		Chunk: convert.ChunkPlacementToPB(*chunk),
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
	object, err := s.facade.GetObjectAccess(ctx, t.ObjectID(req.GetObjectId()))
	if err != nil {
		return nil, err
	}

	rsp = &pb.GetObjectAccessResponse{
		ObjectId:  string(object.ID),
		TotalSize: object.TotalSize,
		Chunks:    utils.Map(object.Chunks, func(cp t.ChunkPlacement) *pb.ChunkPlacement { 
			return convert.ChunkPlacementToPB(cp) 
		}),
		
	}
	return rsp, nil
}

func (s *ClientServer) ListObjects(
	ctx context.Context, req *pb.ListObjectsRequest,
) ( *pb.ListObjectsResponse, error) {

	slog.DebugContext(ctx, "list objects requested")

	objects := s.facade.ListObjects(ctx)
	rsp := &mpb.ListObjectsResponse{
		Objects: utils.Map(objects, convert.ObjectInfoToPB),
	}
	return rsp, nil	
}

func (s *ClientServer) ListChunks(
	ctx context.Context, req *pb.ListChunksRequest,
) (*pb.ListChunksResponse, error) {

	slog.DebugContext(ctx, "list chunks requested")

	chunks := s.facade.ListChunks(ctx)
	rsp := &mpb.ListChunksResponse{
		Chunks: utils.Map(chunks, convert.ChunkInfoToPB),
	}
	return rsp, nil
}

func (s *ClientServer) ListNodes(
	ctx context.Context, req *pb.ListNodesRequest,
) (*pb.ListNodesResponse, error) {

	slog.DebugContext(ctx, "list objects requested")

	nodes := s.facade.ListNodes(ctx)

	rsp := &mpb.ListNodesResponse{
		Nodes: utils.Map(nodes, convert.NodeInfoToPB),
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


