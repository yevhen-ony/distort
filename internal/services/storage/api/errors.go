package api 

import (
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	
	s "dos/internal/services/storage"
)

func ToStatus(err error) error {
	if err == nil {
		return nil
	}

	slog.Error(err.Error())

	if _, ok := status.FromError(err); ok {
		return err
	}

	switch {
	case errors.Is(err, s.ErrServiceBusy):
		return status.Error(codes.ResourceExhausted, err.Error())
	case errors.Is(err, s.ErrInvalidDigest):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, s.ErrChunkNotFound):
		return status.Error(codes.NotFound,  err.Error())
	case errors.Is(err, s.ErrChunkConflict):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, s.ErrInvalidHeader):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, s.ErrInvalidData):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
