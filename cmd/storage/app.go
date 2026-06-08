package main

import (
	"context"
	"fmt"

	cpb "dos/gen/proto/common/v1"
	spb "dos/gen/proto/storage/v1"
	"dos/internal/common/connect"
	"dos/internal/common/listener"
	"dos/internal/common/master/resolve"
	"dos/internal/common/master/route"
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
	master    *MasterHolder
	storageBE s.ChunkStorage

	metricsS   *prom.Service
	inventoryS *storage.ChunkInventory
	heartbeatS *storage.HeartbeatService
	identityS  *identity.IdentityService
	reportS    *report.ReportService
	storageS   *storage.StorageService

	apiServer *api.Server
	apiHealth *api.HealthServer
	apiAdmin  *api.AdminServer
}

func NewApp(config *Config) (*App, error) {

	conn := connect.NewConnCache()

	chunkT, err := chunkrpc.NewTransport(conn, config)
	if err != nil {
		return nil, fmt.Errorf("chunk transport init: %w", err)
	}

	master, err := initMasterTransport(config)
	if err != nil {
		return nil, err
	}

	storageBE, err := store.NewChunkStorage(config)
	if err != nil {
		return nil, fmt.Errorf("chunk store init: %w", err)
	}

	metricsS := prom.NewService(config.Metrics)

	identityS, err := identity.NewIdentityService(identity.IdentityDeps{
		MasterT: master.transport,
		Config:  config,
	})
	if err != nil {
		return nil, fmt.Errorf("identity service init: %w", err)
	}

	reportS, err := report.NewReportService(report.ReportDeps{
		Identity: identityS,
		MasterT:  master.transport,
		Config:   config,
		Metrics:  report.NewReportMetrics(metricsS.Provider()),
	})
	if err != nil {
		return nil, fmt.Errorf("report service init: %w", err)
	}

	inventoryS, err := storage.NewChunkInventory(storage.ChunkInventoryDeps{
		Config:  config,
		Metrics: storage.NewChunkCatalogMetrics(metricsS.Provider()),
	})

	storageS, err := storage.NewStorageService(storage.StorageDeps{
		Inventory: inventoryS,
		Identity:  identityS,
		StorageBE: storageBE,
		MasterT:   master.transport,
		ChunkT:    chunkT,
		Config:    config,
		Metrics:   storage.NewStorageMetrics(metricsS.Provider()),
	})
	if err != nil {
		return nil, fmt.Errorf("storage service init: %w", err)
	}

	heartbeatS, err := storage.NewHeartbeatService(storage.HeartbeatDeps{
		Inventory: inventoryS,
		Identity:  identityS,
		Storage:   storageS,
		MasterT:   master.transport,
		Config:    config,
		Metrics:   storage.NewHeartbeatMetrics(metricsS.Provider()),
	})
	if err != nil {
		return nil, fmt.Errorf("heartbeat service init: %w", err)
	}

	apiServer, err := api.NewServer(api.ServerDeps{
		Identity: identityS,
		Storage:  storageS,
		Config:   config,
	})
	if err != nil {
		return nil, fmt.Errorf("api server init: %w", err)
	}

	apiHealth, err := api.NewHealthServer(api.HealthDeps{
		Identity: identityS,
	})
	if err != nil {
		return nil, fmt.Errorf("api health init: %w", err)
	}

	apiAdmin, err := api.NewAdminServer(api.AdminDeps{
		Inventory: inventoryS,
		Storage:   storageS,
		Heartbeat: heartbeatS,
	})
	if err != nil {
		return nil, fmt.Errorf("api admin init: %w", err)
	}

	storageS.SetReporter(reportS)
	reportS.SetReportProcessor(storageS)
	master.router.SetOnMasterChange(func(context.Context) {
		heartbeatS.Flush()
	})

	app := &App{
		config: config,
		conn:   conn,

		storageBE: storageBE,
		master:    master,
		chunkT:    chunkT,

		metricsS:   metricsS,
		inventoryS: inventoryS,
		identityS:  identityS,
		reportS:    reportS,
		heartbeatS: heartbeatS,

		storageS: storageS,

		apiServer: apiServer,
		apiHealth: apiHealth,
		apiAdmin:  apiAdmin,
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
		cpb.RegisterHealthServiceServer(s, app.apiHealth)
		spb.RegisterAdminServiceServer(s, app.apiAdmin)
	})
}

func (app *App) Close() error {
	if app == nil || app.conn == nil {
		return nil
	}
	return app.conn.Close()
}

type MasterHolder struct {
	resolver  *resolve.Resolver
	router    *route.MasterRouter
	transport *transport.Master
}

func initMasterTransport(config *Config) (*MasterHolder, error) {
	mresolver, err := resolve.New(&config.Master)
	if err != nil {
		return nil, fmt.Errorf("master resolver init: %w", err)

	}
	mrouter, err := route.New(mresolver)
	if err != nil {
		return nil, fmt.Errorf("master router init: %w", err)
	}

	mtransport, err := transport.NewMaster(transport.MasterTransportDeps{
		Router: mrouter,	
		Config: config,
	})
	if err != nil {
		return nil, fmt.Errorf("master transport init: %w", err)
	}

	holder := &MasterHolder{
		resolver:  mresolver,
		router:    mrouter,
		transport: mtransport,
	}
	return holder, nil
}
