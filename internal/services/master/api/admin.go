package api

import (
	"context"
	mpb "dos/gen/proto/master/v1"
	"dos/internal/common/convert"
	"dos/internal/common/utils"
	m "dos/internal/services/master"
	"log/slog"
)

var _ mpb.AdminServiceServer = (*AdminServer)(nil)

type AdminServer struct {
	mpb.UnimplementedAdminServiceServer

	facade m.ClientFacade
}

func NewAdminServer(facade m.ClientFacade) *AdminServer {
	return &AdminServer{facade: facade}
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

	slog.DebugContext(ctx, "list objects requested")

	nodes := s.facade.ListNodes(ctx)

	rsp := &mpb.ListNodesResponse{
		Nodes: utils.Map(nodes, convert.NodeInfoToPB),
	}
	return rsp, nil
}
