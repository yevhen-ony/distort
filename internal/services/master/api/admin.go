package api

import (
	"context"
	mpb "dos/gen/proto/master/v1"
	"dos/internal/common/convert"
	"dos/internal/common/utils"
	m "dos/internal/services/master"
	"errors"
	"log/slog"
)

var _ mpb.AdminServiceServer = (*AdminServer)(nil)

type AdminServer struct {
	mpb.UnimplementedAdminServiceServer

	facade m.ClientFacade
	state m.MasterState	
}

type AdminDeps struct {
	Facade m.ClientFacade
	State m.MasterState
}

func NewAdminServer(deps AdminDeps) (*AdminServer, error) {
	if deps.Facade == nil {
		return nil, errors.New("missing facade service")
	}
	if deps.State == nil {
		return nil, errors.New("missing master state")
	}

	s := &AdminServer{
		facade: deps.Facade,
		state: deps.State,
	}
	return s, nil
}

func (s *AdminServer) ListObjects(
	ctx context.Context,
	req *mpb.ListObjectsRequest,
) (*mpb.ListObjectsResponse, error) {

	slog.DebugContext(ctx, "list objects requested")

	objects := s.facade.ListObjects(ctx)
	rsp := &mpb.ListObjectsResponse{
		Objects: utils.Map(objects, convert.ObjectInfoToPB),
	}
	return rsp, nil
}

func (s *AdminServer) ListChunks(
	ctx context.Context,
	req *mpb.ListChunksRequest,
) (*mpb.ListChunksResponse, error) {

	slog.DebugContext(ctx, "list chunks requested")

	chunks := s.facade.ListChunks(ctx)
	rsp := &mpb.ListChunksResponse{
		Chunks: utils.Map(chunks, convert.ChunkInfoToPB),
	}
	return rsp, nil
}

func (s *AdminServer) ListNodes(
	ctx context.Context,
	req *mpb.ListNodesRequest,
) (*mpb.ListNodesResponse, error) {

	slog.DebugContext(ctx, "list nodes requested")

	nodes := s.facade.ListNodes(ctx)

	rsp := &mpb.ListNodesResponse{
		Nodes: utils.Map(nodes, convert.NodeInfoToPB),
	}
	return rsp, nil
}

func (s *AdminServer) TransferLeadership(
	ctx context.Context,
	req *mpb.TransferLeadershipRequest,
) (*mpb.TransferLeadershipResponse, error) {

	slog.DebugContext(ctx, "transfer leadership requested")

	if err := s.state.TransferLeadership(ctx); err != nil {
		return nil, err 
	}
	return &mpb.TransferLeadershipResponse{}, nil
}
