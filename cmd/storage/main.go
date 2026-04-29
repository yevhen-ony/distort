package main

import (
	"context"
	"flag"
	"fmt"
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
	configPath := flag.String("config", "config.yml", "path to config file")
	flag.Parse()

	cfg := Config{}
	err := config.LoadConfig(*configPath, &cfg)
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}
	
	logger.Init(&cfg.Logger)


	storage, err := store.New(&cfg.Store)
	if err != nil {
		panic(fmt.Errorf("construct storage: %w", err))
	}

	conn := connect.NewConnCache()
	defer conn.Close()

	master, err := transport.NewMasterTransport(conn, &cfg.Master)
	if err != nil {
		panic(fmt.Errorf("construct master transport: %w", err))
	}

	app, err := core.New(storage, master, cfg.Service)
	if err != nil {
		panic(fmt.Errorf("construct service: %w", err))
	}

	srv := api.New(app, &cfg.API)
    
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
    defer stop()

	err = listener.RunGRPCServer(ctx, &cfg.Listen, func(s *grpc.Server) {
		spb.RegisterChunkServiceServer(s, srv)
	})
	if err != nil {
		panic(fmt.Errorf("run grpc server: %w", err))
	}
}

