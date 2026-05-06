package main

import (
	"context"
	mpb "dos/gen/proto/master/v1"
	"dos/internal/common/listener"
	"dos/internal/services/master/api"
	"dos/internal/services/master/domain"
	"dos/internal/services/master/repo"
	"fmt"

	"google.golang.org/grpc"
)

type App struct {
	objectRepo *repo.InMemObjectRepo
	chunkRepo  *repo.InMemChunkRepo
	nodeReg    *repo.InMemNodeRegistry

	Service *domain.MasterService
	Config  *Config
}

func NewApp(cfg *Config) (*App, error) {
	objectRepo := repo.NewInMemObjectRepo()
	chunkRepo := repo.NewInMemChunkRepo()
	nodeReg := repo.NewInMemNodeRegistry()

	service, err := domain.NewMasterService(chunkRepo, objectRepo, nodeReg, &cfg.Service)
	if err != nil {
		return nil, fmt.Errorf("init service: %w", err)
	}

	app := &App{
		objectRepo: objectRepo,
		chunkRepo:  chunkRepo,
		nodeReg:    nodeReg,
		Service:    service,
		Config:     cfg,
	}
	return app, nil
}

func (app *App) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go app.Service.RunNodeCleanupLoop(ctx)

	storageSrv := api.NewStorageServer(app.Service)
	clientSrv := api.NewClientServer(app.Service)
	
	err := listener.RunGRPCServer(ctx, &app.Config.Listen, func(s *grpc.Server) {
		mpb.RegisterMasterStorageServiceServer(s, storageSrv)
		mpb.RegisterMasterClientServiceServer(s, clientSrv)
	})
	return err
}
