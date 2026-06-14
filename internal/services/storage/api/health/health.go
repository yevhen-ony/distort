package health

import (
	"context"
	cpb "dos/gen/proto/common/v1"
	"dos/internal/common/dosctx"
	t "dos/internal/common/types"
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IdentityValidator interface {
	Validate(t.NodeID) error
	GetID() (t.NodeID, error)
}

type HealthDeps struct{
	Identity IdentityValidator 
}

type HealthServer struct{
	cpb.UnimplementedHealthServiceServer	

	identity IdentityValidator 
}

func NewHealthServer(deps HealthDeps) (*HealthServer, error) {
	if deps.Identity == nil {
		return nil, errors.New("missing identity")
	}

	server := &HealthServer{
		identity: deps.Identity,
	}
	return server, nil
}

func (h *HealthServer) Ready(
	ctx context.Context,
	req *cpb.ReadyRequest,
) (*cpb.HealthResponse, error) {

	ctx = dosctx.WithService(ctx, "health")
	slog.DebugContext(ctx, "ready requested")

	_, err := h.identity.GetID()
	if err != nil {
		return nil, status.Error(codes.Unavailable, "not ready")
	}

	rsp := &cpb.HealthResponse{Component: cpb.Component_COMPONENT_STORAGE}
	return rsp, nil
}

func (h *HealthServer) VerifyIdentity(
	ctx context.Context,
	req *cpb.VerifyIdentityRequest,
) (*cpb.HealthResponse, error) {

	ctx = dosctx.WithService(ctx, "health")
	slog.DebugContext(ctx, "verify identity requested")

	nodeID := t.NodeID(req.GetExpectedId())
	if err := h.identity.Validate(nodeID); err != nil {
		return nil, status.Error(codes.FailedPrecondition, "invalid identity")
	}
	rsp := &cpb.HealthResponse{Component: cpb.Component_COMPONENT_STORAGE}
	return rsp, nil
}


