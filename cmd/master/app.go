package main

import (
	"context"
	"fmt"

	mpb "dos/gen/proto/master/v1"
	"dos/internal/common/connect"
	"dos/internal/common/listener"
	"dos/internal/common/metrics/prom"
	"dos/internal/common/transport/chunkrpc"
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

	objectRepo     *repo.InMemObjectRepo
	chunkRepo      *repo.InMemChunkRepo
	nodeRegistry   *repo.InMemNodeRegistry
	chunkNodeIndex *domain.InMemChunkNodeIndex

	metricsService    *prom.Service
	replicateExecutor *replicate.ExecutorService
	replicatePlanner  *replicate.PlannerService
	catalogService    *catalog.CatalogService
	catalogCleanup    *catalog.CleanupService
	storageLifecycle  *storagenode.LifecycleService
	storagePlacement  *storagenode.PlacementService
	storageReport     *storagenode.ReportService
	nodeCleanup       *storagenode.CleanupWorker
	clientFacade      *domain.ClientFacadeService

	clientAPI  *api.ClientServer
	storageAPI *api.StorageServer

	config *Config
}

func NewApp(config *Config) (*App, error) {
	conn := connect.NewConnCache()

	objectRepo := repo.NewInMemObjectRepo()
	chunkRepo := repo.NewInMemChunkRepo()
	nodeRegistry := repo.NewInMemNodeRegistry()
	chunkNodeIndex := domain.NewInMemChunkNodeIndex()

	metricsService := prom.NewService(config.Metrics)

	chunkT, err := chunkrpc.NewTransport(conn, config)
	if err != nil {
		return nil, fmt.Errorf("chunk transport init: %w", err)
	}

	catalogMetrics := catalog.NewCatalogMetrics(metricsService.Provider())

	catalogService, err := catalog.NewCatalogService(catalog.CatalogDeps{
		ObjectRepo: objectRepo,
		ChunkRepo:  chunkRepo,
		Metrics:    catalogMetrics,
	})
	if err != nil {
		return nil, fmt.Errorf("catalog service init: %w", err)
	}

	catalogCleanup, err := catalog.NewCleanupService(catalog.CleanupDeps{
		ObjectRepo: objectRepo,
		ChunkRepo:  chunkRepo,
		Config:     config,
		Metrics:    catalogMetrics,
	})
	if err != nil {
		return nil, fmt.Errorf("catalog cleanup init: %w", err)
	}

	nodePlacement, err := storagenode.NewPlacementService(storagenode.PlacementDeps{
		ChunkNodeIndex: chunkNodeIndex,
		NodeRegistry:   nodeRegistry,
		Config:         config,
	})
	if err != nil {
		return nil, fmt.Errorf("storage placement service init: %w", err)
	}

	replicateExecutor, err := replicate.NewExecutorService(replicate.ExecutorDeps{
		ObjectRepo: objectRepo,
		ChunkRepo:  chunkRepo,
		Placement:  nodePlacement,
		ChunkT:     chunkT,
		Config:     config,
		Metrics:    replicate.NewExecutorMetrics(metricsService.Provider()),
	})
	if err != nil {
		return nil, fmt.Errorf("replicate executor service init: %w", err)
	}

	replicatePlanner, err := replicate.NewPlannerService(replicate.PlannerDeps{
		ObjectRepo:  objectRepo,
		ChunkRepo:   chunkRepo,
		Replication: replicateExecutor,
		Config:      config,
	})
	if err != nil {
		return nil, fmt.Errorf("replicate planner init: %w", err)
	}

	nodeLifecycle, err := storagenode.NewLifecycleService(storagenode.LifecycleDeps{
		ChunkRepo:      chunkRepo,
		ChunkNodeIndex: chunkNodeIndex,
		NodeRegistry:   nodeRegistry,
		Metrics:        storagenode.NewLifecycleMetrics(metricsService.Provider()),
	})
	if err != nil {
		return nil, fmt.Errorf("storage lifecycle service init: %w", err)
	}

	nodeReport, err := storagenode.NewReportService(storagenode.ReportDeps{
		ChunkRepo:      chunkRepo,
		NodeRegistry:   nodeRegistry,
		ChunkNodeIndex: chunkNodeIndex,
		Replication:    replicateExecutor,
		Metrics:        storagenode.NewReportMetrics(metricsService.Provider()),
	})
	if err != nil {
		return nil, fmt.Errorf("storage report serivce init: %w", err)
	}

	nodeCleanup, err := storagenode.NewCleanupWorker(storagenode.CleanupDeps{
		Lifecycle:   nodeLifecycle,
		Replication: replicateExecutor,
		Config:      config,
	})

	clientFacade, err := domain.NewClientFacadeService(domain.ClientFacadeDeps{
		Catalog:     catalogService,
		Placement:   nodePlacement,
		Lifecycle:   nodeLifecycle,
		Replication: replicateExecutor,
		Config:      config,
	})
	if err != nil {
		return nil, fmt.Errorf("client facade service init: %w", err)
	}

	clientAPI := api.NewClientServer(clientFacade)
	storageAPI := api.NewStorageServer(nodeLifecycle, nodeReport)

	app := &App{
		conn: conn,

		objectRepo:     objectRepo,
		chunkRepo:      chunkRepo,
		nodeRegistry:   nodeRegistry,
		chunkNodeIndex: chunkNodeIndex,

		metricsService:    metricsService,
		catalogService:    catalogService,
		catalogCleanup:    catalogCleanup,
		storageLifecycle:  nodeLifecycle,
		storagePlacement:  nodePlacement,
		storageReport:     nodeReport,
		nodeCleanup:       nodeCleanup,
		replicateExecutor: replicateExecutor,
		replicatePlanner:  replicatePlanner,

		storageAPI: storageAPI,
		clientAPI:  clientAPI,

		config: config,
	}
	return app, nil
}

func (app *App) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go app.replicatePlanner.RunLoop(ctx)
	go app.replicateExecutor.RunLoop(ctx)
	go app.nodeCleanup.RunLoop(ctx)
	go app.catalogCleanup.RunLoop(ctx)
	go app.metricsService.Serve(ctx)

	err := listener.RunGRPCServer(ctx, &app.config.Listen, func(s *grpc.Server) {
		mpb.RegisterMasterStorageServiceServer(s, app.storageAPI)
		mpb.RegisterMasterClientServiceServer(s, app.clientAPI)
	})
	return err
}

func (app *App) Close() error {
	return app.conn.Close()
}
