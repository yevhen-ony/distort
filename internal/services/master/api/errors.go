package api

import (
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	m "dos/internal/services/master"
)

var (
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
	case errors.Is(err, m.ErrNodeNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
	

}

