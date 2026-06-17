package api

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	mpb "dos/gen/proto/master/v1"
)

type ActiveGuard interface {
	IsActiveMaster() bool
}

type MasterGuard struct {
	masterState ActiveGuard
}

func NewMasterGuard(state ActiveGuard) *MasterGuard {
	return &MasterGuard{
		masterState: state,
	}
}

func (i *MasterGuard) Intercept(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {

	if isUnguardedMethod(info.FullMethod) {
		return handler(ctx, req)
	}

	if !i.masterState.IsActiveMaster() {
		return nil, status.Error(codes.Unavailable, "not active master")
	}

	return handler(ctx, req)
}

func (i *MasterGuard) AsOption() grpc.ServerOption {
	return grpc.UnaryInterceptor(i.Intercept)
}

func isUnguardedMethod(method string) bool {
	switch method {
	case mpb.MasterDiscoveryService_GetActiveMaster_FullMethodName:
		return true
	default:
		return false
	}
}
