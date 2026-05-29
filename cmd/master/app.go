package main

import (
	"context"
	"fmt"

	mpb "dos/gen/proto/master/v1"
	"dos/internal/common/connect"
	"dos/internal/common/listener"
	"dos/internal/common/metrics/prom"
	"dos/internal/common/transport/chunkrpc"
	m "dos/internal/services/master"
	"dos/internal/services/master/api"
	"dos/internal/services/master/domain"
	"dos/internal/services/master/domain/catalog"
	"dos/internal/services/master/domain/replicate"
	"dos/internal/services/master/domain/storagenode"
	"dos/internal/services/master/repo"

	"google.golang.org/grpc"
)

type App struct {
	conn *connect.ConnCache

	masterMode MasterMode

	chunkRepository *repo.InMemChunkRepo
	nodeRegistry    *repo.InMemNodeRegistry
	chunkNodeIndex  *domain.InMemChunkNodeIndex

	chunkTransport *chunkrpc.Transport

	metricsService *prom.Service

	discoveryService m.MasterState

	placement *storagenode.PlacementService

	replicateExecutor *replicate.ExecutorService
	replicatePlanner  *replicate.PlannerService

	nodeLifecycle *storagenode.LifecycleService
	nodeReport    *storagenode.ReportService
	nodeCleanup   *storagenode.CleanupWorker

	catalogService *catalog.CatalogService
	catalogCleanup *catalog.CleanupService

	clientFacade *domain.ClientFacadeService

	clientAPI    *api.ClientServer
	adminAPI     *api.AdminServer
	storageAPI   *api.StorageServer
	discoveryAPI *api.MasterDiscoveryServer

	config *Config
}

func NewApp(config *Config) (app *App, err error) {
	app = &App{config: config}

	app.metricsService = prom.NewService(config.Metrics)

	app.conn = connect.NewConnCache()
	app.chunkRepository = repo.NewInMemChunkRepo()
	app.nodeRegistry = repo.NewInMemNodeRegistry()
	app.chunkNodeIndex = domain.NewInMemChunkNodeIndex()

	app.masterMode, err = InitMasterMode(config)
	if err != nil {
		return nil, err
	}

	app.chunkTransport, err = chunkrpc.NewTransport(app.conn, config)
	if err != nil {
		return nil, fmt.Errorf("chunk transport init: %w", err)
	}

	app.placement, err = storagenode.NewPlacementService(storagenode.PlacementDeps{
		ChunkNodeIndex: app.chunkNodeIndex,
		NodeRegistry:   app.nodeRegistry,
		Config:         config,
	})
	if err != nil {
		return nil, fmt.Errorf("storage node placement service init: %w", err)
	}

	if err := app.initCatalog(config); err != nil {
		return nil, err
	}

	if err := app.initReplication(config); err != nil {
		return nil, err
	}

	if err := app.initStorageNodeServices(config); err != nil {
		return nil, err
	}

	app.clientFacade, err = domain.NewClientFacadeService(domain.ClientFacadeDeps{
		Catalog:     app.catalogService,
		Placement:   app.placement,
		Lifecycle:   app.nodeLifecycle,
		Replication: app.replicateExecutor,
		Config:      config,
	})
	if err != nil {
		return nil, fmt.Errorf("client facade service init: %w", err)
	}

	if err := app.initAPI(config); err != nil {
		return nil, err
	}

	return app, nil
}

func (app *App) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go app.masterMode.MasterState().WatchState(ctx, func(ctx context.Context) {
		go app.replicatePlanner.RunLoop(ctx)
		go app.replicateExecutor.RunLoop(ctx)
		go app.nodeCleanup.RunLoop(ctx)
		go app.catalogCleanup.RunLoop(ctx)
	})

	go app.metricsService.Serve(ctx)

	reg := func(s *grpc.Server) {
		mpb.RegisterAdminServiceServer(s, app.adminAPI)
		mpb.RegisterMasterClientServiceServer(s, app.clientAPI)
		mpb.RegisterMasterStorageServiceServer(s, app.storageAPI)
		mpb.RegisterMasterDiscoveryServiceServer(s, app.discoveryAPI)
	}

	guard := api.NewMasterGuard(app.masterMode.MasterState())
	err := listener.RunGRPCServer(ctx, &app.config.Listen, reg, guard.AsOption())
	return err
}

func (app *App) initCatalog(config *Config) (err error) {

	catalogMetrics := catalog.NewCatalogMetrics(app.metricsService.Provider())
	app.catalogService, err = catalog.NewCatalogService(catalog.CatalogDeps{
		ObjectAuthority: app.masterMode.ObjectAuthority(),
		ChunkRepository: app.chunkRepository,
		Metrics:         catalogMetrics,
	})
	if err != nil {
		return fmt.Errorf("catalog service init: %w", err)
	}

	app.catalogCleanup, err = catalog.NewCleanupService(catalog.CleanupDeps{
		ObjectAuthority: app.masterMode.ObjectAuthority(),
		ChunkRepository: app.chunkRepository,
		Config:          config,
		Metrics:         catalogMetrics,
	})
	if err != nil {
		return fmt.Errorf("catalog cleanup init: %w", err)
	}
	return nil
}

func (app *App) initStorageNodeServices(config *Config) (err error) {

	lifecycleMetrics := storagenode.NewLifecycleMetrics(app.metricsService.Provider())
	app.nodeLifecycle, err = storagenode.NewLifecycleService(storagenode.LifecycleDeps{
		ChunkRepository: app.chunkRepository,
		ChunkNodeIndex:  app.chunkNodeIndex,
		NodeRegistry:    app.nodeRegistry,
		Metrics:         lifecycleMetrics,
	})
	if err != nil {
		return fmt.Errorf("storage lifecycle service init: %w", err)
	}

	reportMetrics := storagenode.NewReportMetrics(app.metricsService.Provider())
	app.nodeReport, err = storagenode.NewReportService(storagenode.ReportDeps{
		ChunkRepo:      app.chunkRepository,
		NodeRegistry:   app.nodeRegistry,
		ChunkNodeIndex: app.chunkNodeIndex,
		Replication:    app.replicateExecutor,
		Metrics:        reportMetrics,
	})
	if err != nil {
		return fmt.Errorf("storage report serivce init: %w", err)
	}

	app.nodeCleanup, err = storagenode.NewCleanupWorker(storagenode.CleanupDeps{
		Lifecycle:   app.nodeLifecycle,
		Replication: app.replicateExecutor,
		Config:      config,
	})
	if err != nil {
		return fmt.Errorf("storage node cleanulp service init: %w", err)
	}
	return nil
}

func (app *App) initReplication(config *Config) (err error) {
	executorMetrics := replicate.NewExecutorMetrics(app.metricsService.Provider())
	app.replicateExecutor, err = replicate.NewExecutorService(replicate.ExecutorDeps{
		ObjectReader:    app.masterMode.ObjectAuthority(),
		ChunkRepository: app.chunkRepository,
		Placement:       app.placement,
		ChunkTransport:  app.chunkTransport,
		Config:          config,
		Metrics:         executorMetrics,
	})
	if err != nil {
		return fmt.Errorf("replicate executor service init: %w", err)
	}

	plannerMetrics := replicate.NewPlannerMetrics(app.metricsService.Provider())
	app.replicatePlanner, err = replicate.NewPlannerService(replicate.PlannerDeps{
		ObjectReader:    app.masterMode.ObjectAuthority(),
		ChunkRepository: app.chunkRepository,
		Replication:     app.replicateExecutor,
		Config:          config,
		Metrics:         plannerMetrics,
	})
	if err != nil {
		return fmt.Errorf("replicate planner init: %w", err)
	}
	return nil
}

func (app *App) initAPI(config *Config) (err error) {
	app.adminAPI, err = api.NewAdminServer(app.clientFacade)
	if err != nil {
		return fmt.Errorf("admin api init: %w", err)
	}
	app.clientAPI, err = api.NewClientServer(app.clientFacade)
	if err != nil {
		return fmt.Errorf("client api init: %w", err)
	}
	app.storageAPI, err = api.NewStorageServer(app.nodeLifecycle, app.nodeReport)
	if err != nil {
		return fmt.Errorf("storage api init: %w", err)
	}
	app.discoveryAPI, err = api.NewMasterDiscoveryServer(app.masterMode.MasterState())
	if err != nil {
		return fmt.Errorf("discovery api init: %w", err)
	}
	return nil
}

func (app *App) Close() error {
	return app.conn.Close()
}
