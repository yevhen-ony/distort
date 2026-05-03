package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"google.golang.org/grpc"

	mpb "dos/gen/proto/master/v1"
	"dos/internal/common/config"
	"dos/internal/common/listener"
	"dos/internal/common/logger"
	"dos/internal/services/master/api"
	"dos/internal/services/master/domain"
	"dos/internal/services/master/repo"
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

	objectRepo := repo.NewInMemObjectRepo()
	chunkRepo := repo.NewInMemChunkRepo()
	nodeReg := repo.NewInMemNodeRegistry()

	svc := domain.NewMasterService(chunkRepo, objectRepo, nodeReg, &cfg.Service)

	storageSrv := api.NewStorageServer(svc)
	clientSrv := api.NewClientServer(svc)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
    defer stop()

	err = listener.RunGRPCServer(ctx, &cfg.Listen, func(s *grpc.Server) {
		mpb.RegisterMasterStorageServiceServer(s, storageSrv)
		mpb.RegisterMasterClientServiceServer(s, clientSrv)
	})

	if err != nil {
		panic(fmt.Errorf("run grpc server: %w", err))
	}
}

