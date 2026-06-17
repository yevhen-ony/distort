package listener

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
)

type ListenerConfig interface {
	ListeningAddr() string
}

func RunGRPCServer(
	ctx context.Context,
	config ListenerConfig,
	register func(*grpc.Server),
	options ...grpc.ServerOption,
) error {

	lis, err := net.Listen("tcp", config.ListeningAddr())
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer lis.Close()

	gs := grpc.NewServer(options...)
	register(gs)

	errCh := make(chan error, 1)

	slog.Info("grpc server listening", "addr", config.ListeningAddr())

	go func() { errCh <- gs.Serve(lis) }()

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("grpc serve: %w", err)
		}
		return nil
	case <-ctx.Done():
		slog.Info("grpc server shutting down", "addr", config.ListeningAddr())
		gs.GracefulStop()
		return nil
	}
}
