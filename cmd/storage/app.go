package main

import (
	"context"
	"fmt"

	spb "dos/gen/proto/storage/v1"
	"dos/internal/common/connect"
	"dos/internal/common/listener"
	"dos/internal/common/transport/chunkrpc"
	s "dos/internal/services/storage"
	"dos/internal/services/storage/api"
	"dos/internal/services/storage/core/identity"
	"dos/internal/services/storage/core/report"
	"dos/internal/services/storage/core/storage"
	"dos/internal/services/storage/store"
	"dos/internal/services/storage/transport"

	"google.golang.org/grpc"
)

type App struct {
	config *Config	
	conn *connect.ConnCache

	storageInfra s.ChunkStorage 
	
	identityService *identity.IdentityService
	reportService *report.ReportService
	storageService *storage.StorageService

	chunkTransport *chunkrpc.Transport
	masterTransport *transport.Master

	apiServer *api.Server
}

func NewApp(config *Config) (*App, error) {
	conn := connect.NewConnCache()

	chunkTransport, err := chunkrpc.NewTransport(conn, config)
	if err != nil {
		return nil, fmt.Errorf("chunk transport init: %w", err)
	}

	masterTransport, err := transport.NewMaster(conn, config)
	if err != nil {
		return nil, fmt.Errorf("master transport init: %w", err)
	}

	identityService := identity.NewIdentityService(masterTransport, config)

	reportService := report.NewReportService(identityService, masterTransport, config)

	storageInfra, err := store.NewChunkStorage(config)
	if err != nil {
		return nil, fmt.Errorf("chunk store init: %w", err)
	}

	storageService, err := storage.NewStorageService(
		storageInfra,
		masterTransport,
		chunkTransport,
		identityService,
		config,
	)
	if err != nil {
		return nil, fmt.Errorf("storage service init: %w", err)
	}
	storageService.SetReporter(reportService)

	apiServer := api.New(identityService, storageService, config)

	app := &App{
		config: config,
		conn: conn,	
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
		return fmt.Errorf("request id: %w", err)
	}
	if err := app.storageService.Start(ctx); err != nil {
		return fmt.Errorf("start storage service: %w", err) 
	}

	go app.reportService.RunLoop(ctx)
	go app.storageService.RunHearbeatLoop(ctx)

	go app.runGrpcServer(ctx)

	return nil
}

func (app *App) runGrpcServer(ctx context.Context) {
	_ = listener.RunGRPCServer(ctx, &app.config.Listen, func(s *grpc.Server) {
		spb.RegisterChunkServiceServer(s, app.apiServer)
	})
}

func (app *App) Close() error {
	if app == nil || app.conn == nil {
		return nil
	}
	return app.conn.Close()
}
