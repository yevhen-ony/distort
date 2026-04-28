package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"

	pb "dos/gen/proto/storage/v1"
	"dos/internal/common/connect"
	"dos/internal/services/storage/api"
	"dos/internal/services/storage/core"
	"dos/internal/services/storage/store"
	"dos/internal/services/storage/transport"

	"google.golang.org/grpc"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	config, err := loadConfig(*configPath)
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}

	storage, err := store.New(&config.Store)
	if err != nil {
		panic(fmt.Errorf("construct storage: %w", err))
	}

	conn := connect.NewConnCache()
	defer conn.Close()

	master, err := transport.NewMasterTransport(conn, &config.Master)
	if err != nil {
		panic(fmt.Errorf("construct master transport: %w", err))
	}

	app, err := core.New(storage, master, config.Service)
	if err != nil {
		panic(fmt.Errorf("construct service: %w", err))
	}

	srv := api.New(app, &config.API)

	if err := runGRPCServer(srv, &config.Listen); err != nil {
		panic(err)
	}
}

func runGRPCServer(server *api.Server, cfg *ListenerConfig) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	lis, err := net.Listen("tcp", cfg.Addr())
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer lis.Close()

	gs := grpc.NewServer()

	pb.RegisterChunkServiceServer(gs, server)

	errCh := make(chan error, 1)

	slog.Info("listening ...", "addr", cfg.Addr())
	go func() { errCh <- gs.Serve(lis) }()

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("grpc serve: %w", err)
		}
		return nil
	case <-ctx.Done():
		slog.Info("shutting down", "addr", cfg.Addr())
		gs.GracefulStop()
		return nil
	}
}
