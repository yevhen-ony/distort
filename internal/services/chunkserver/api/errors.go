package api 

import (
	svc "dos/internal/services/chunkserver/core"
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrHeaderInvalid = errors.New("invalid header")
	ErrDataInvalid = errors.New("invalid data")
)	

func toStatus(err error) error {
	if err == nil {
		return nil
	}

	slog.Error(err.Error())

	if _, ok := status.FromError(err); ok {
		return err
	}

	switch {
	case errors.Is(err, svc.ErrInvalid):
		return status.Error(codes.InvalidArgument, "validation failed")
	case errors.Is(err, svc.ErrNotFound):
		return status.Error(codes.NotFound, "not found")
	case errors.Is(err, svc.ErrConflict):
		return status.Error(codes.AlreadyExists, "already exists")
	case errors.Is(err, ErrHeaderInvalid):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, ErrDataInvalid):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
