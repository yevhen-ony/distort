package main

import (
	"context"
	"fmt"

	spb "dos/gen/proto/storage/v1"
	"dos/internal/common/connect"
	"dos/internal/common/listener"
	"dos/internal/common/metrics/prom"
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
	conn   *connect.ConnCache

	chunkT    *chunkrpc.Transport
	masterT   *transport.Master
	storageBE s.ChunkStorage

	metricsS   *prom.Service
	catalogS   *storage.ChunkCatalogService
	heartbeatS *storage.HeartbeatService
	identityS  *identity.IdentityService
	reportS    *report.ReportService
	storageS   *storage.StorageService

	apiServer *api.Server
}

func NewApp(config *Config) (*App, error) {

	conn := connect.NewConnCache()

	chunkT, err := chunkrpc.NewTransport(conn, config)
	if err != nil {
		return nil, fmt.Errorf("chunk transport init: %w", err)
	}

	masterT, err := transport.NewMaster(conn, config)
	if err != nil {
		return nil, fmt.Errorf("master transport init: %w", err)
	}

	storageBE, err := store.NewChunkStorage(config)
	if err != nil {
		return nil, fmt.Errorf("chunk store init: %w", err)
	}

	metricsS := prom.NewService(config.Metrics)

	identityS, err := identity.NewIdentityService(identity.IdentityDeps{
		MasterT: masterT,
		Config:  config,
	})
	if err != nil {
		return nil, fmt.Errorf("identity service init: %w", err)
	}

	reportS, err := report.NewReportService(report.ReportDeps{
		Identity: identityS,
		MasterT:  masterT,
		Config:   config,
		Metrics:  report.NewReportMetrics(metricsS.Provider()),
	})
	if err != nil {
		return nil, fmt.Errorf("report service init: %w", err)
	}

	catalogS, err := storage.NewChunkCatalogService(storage.ChunkCatalogDeps{
		Config:  config,
		Metrics: storage.NewChunkCatalogMetrics(metricsS.Provider()),
	})

	heartbeatS, err := storage.NewHeartbeatService(storage.HeartbeatDeps{
		Catalog:  catalogS,
		Identity: identityS,
		MasterT:  masterT,
		Config:   config,
		Metrics:  storage.NewHeartbeatMetrics(metricsS.Provider()),
	})
	if err != nil {
		return nil, fmt.Errorf("heartbeat service init: %w", err)
	}

	storageS, err := storage.NewStorageService(storage.StorageDeps{
		Catalog:   catalogS,
		Identity:  identityS,
		Reporter:  reportS,
		StorageBE: storageBE,
		MasterT:   masterT,
		ChunkT:    chunkT,
		Config:    config,
		Metrics:   storage.NewStorageMetrics(metricsS.Provider()),
	})
	if err != nil {
		return nil, fmt.Errorf("storage service init: %w", err)
	}

	apiServer := api.New(identityS, storageS, config)

	app := &App{
		config: config,
		conn:   conn,

		storageBE: storageBE,
		masterT:   masterT,
		chunkT:    chunkT,

		metricsS:   metricsS,
		catalogS:   catalogS,
		identityS:  identityS,
		reportS:    reportS,
		heartbeatS: heartbeatS,

		storageS: storageS,

		apiServer: apiServer,
	}
	return app, nil
}

func (app *App) Start(ctx context.Context) error {
	if err := app.identityS.RequestNewID(ctx); err != nil {
		return fmt.Errorf("request id: %w", err)
	}

	go app.metricsS.Serve(ctx)
	go app.reportS.RunLoop(ctx)

	if err := app.storageS.Start(ctx); err != nil {
		return fmt.Errorf("start storage service: %w", err)
	}

	go app.heartbeatS.RunLoop(ctx)
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
