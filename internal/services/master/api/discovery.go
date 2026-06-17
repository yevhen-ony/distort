package api

import (
	"context"
	"errors"
	"log/slog"

	mpb "dos/gen/proto/master/v1"
	"dos/internal/common/convert"
	t "dos/internal/common/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MasterDiscovery interface {
	GetActiveMaster(_ context.Context) (t.MasterRef, error)
}

type MasterDiscoveryServer struct {
	mpb.UnimplementedMasterDiscoveryServiceServer

	discovery MasterDiscovery
}

func NewMasterDiscoveryServer(discovery MasterDiscovery) (*MasterDiscoveryServer, error) {
	if discovery == nil {
		return nil, errors.New("missing discovery service")
	}
	s := &MasterDiscoveryServer{
		discovery: discovery,
	}
	return s, nil
}

func (s *MasterDiscoveryServer) GetActiveMaster(
	ctx context.Context,
	req *mpb.GetActiveMasterRequest,
) (*mpb.GetActiveMasterResponse, error) {

	slog.DebugContext(ctx, "get active master requested")

	ref, err := s.discovery.GetActiveMaster(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	rsp := &mpb.GetActiveMasterResponse{
		Active: convert.MasterRefToPB(ref),
	}
	return rsp, nil
}
