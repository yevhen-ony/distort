package api

import (
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	return status.Error(codes.Internal, err.Error())
}

