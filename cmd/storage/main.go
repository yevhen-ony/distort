package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	spb "dos/gen/proto/storage/v1"
	"dos/internal/common/config"
	"dos/internal/common/connect"
	"dos/internal/common/listener"
	"dos/internal/common/logger"
	"dos/internal/services/storage/api"
	"dos/internal/services/storage/core"
	"dos/internal/services/storage/store"
	"dos/internal/services/storage/transport"

	"google.golang.org/grpc"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", "config.yml", "path to config file")
	flag.Parse()

	cfg := Config{}
	err := config.LoadConfig(*configPath, &cfg)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	
	logger.Init(&cfg.Logger)
	
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
    defer stop()

	storage, err := store.New(&cfg.Store)
	if err != nil {
		return fmt.Errorf("construct storage: %w", err)
	}

	conn := connect.NewConnCache()
	defer conn.Close()

	master, err := transport.NewMasterTransport(conn, &cfg.Master)
	if err != nil {
		return fmt.Errorf("construct master transport: %w", err)
	}

	svc, err := core.New(storage, master, cfg.Service)
	if err != nil {
		return fmt.Errorf("construct service: %w", err)
	}
	
	if err := svc.Start(ctx); err != nil {
		return fmt.Errorf("start service: %w", err)
	}

	srv := api.New(svc, &cfg.API)
	err = listener.RunGRPCServer(ctx, &cfg.Listen, func(s *grpc.Server) {
		spb.RegisterChunkServiceServer(s, srv)
	})
	if err != nil {
		return fmt.Errorf("run grpc server: %w", err)
	}
	return nil
}

