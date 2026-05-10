package main

import (
	"context"
	spb "dos/gen/proto/storage/v1"
	"dos/internal/common/connect"
	"dos/internal/common/listener"
	"dos/internal/common/transport/chunkrpc"
	s "dos/internal/services/storage"
	"dos/internal/services/storage/api"
	"dos/internal/services/storage/core"
	"dos/internal/services/storage/store"
	"dos/internal/services/storage/transport"

	"fmt"

	"google.golang.org/grpc"
)

type App struct {
	config *Config	

	storageInfra s.ChunkStorage 
	
	identityService *core.IdentityService
	reportService *core.ReportService
	storageService *core.StorageService

	chunkTransport *chunkrpc.Transport
	masterTransport *transport.Master

	apiServer *api.Server
}

func NewApp(cfg *Config) (*App, error) {
	conn := connect.NewConnCache()

	chunkTransport, err := chunkrpc.NewTransport(conn, cfg)
	if err != nil {
		return nil, fmt.Errorf("chunk transport init: %w", err)
	}

	masterTransport, err := transport.NewMaster(conn, cfg)
	if err != nil {
		return nil, fmt.Errorf("master transport init: %w", err)
	}

	identityService := core.NewIdentityService(masterTransport, cfg)

	reportService := core.NewReportService(identityService, masterTransport, cfg)

	storageInfra, err := store.NewChunkStorage(cfg)
	if err != nil {
		return nil, fmt.Errorf("chunk store init: %w", err)
	}

	storageService, err := core.NewStorageService(
		storageInfra, masterTransport, chunkTransport, reportService, cfg)
	if err != nil {
		return nil, fmt.Errorf("storage service init: %w", err)
	}

	apiServer := api.New(identityService, storageService, cfg)

	app := &App{
		config: cfg,
		
		storageInfra: storageInfra,
		masterTransport: masterTransport,
		chunkTransport: chunkTransport,

		identityService: identityService,
		reportService: reportService,
		storageService: storageService,

		apiServer: apiServer,
	}
	return app, nil
}

func (app *App) Start(ctx context.Context) error {
	if err := app.identityService.RequestNewID(ctx); err != nil {
		return err
	}

	go app.reportService.RunReportLoop(ctx)
	go app.storageService.RunHearbeatLoop(ctx)

	go app.runGrpcServer(ctx)

	return nil
}

func (app *App) runGrpcServer(ctx context.Context) {
	_ = listener.RunGRPCServer(ctx, &app.config.Listen, func(s *grpc.Server) {
		spb.RegisterChunkServiceServer(s, app.apiServer)
	})
}
