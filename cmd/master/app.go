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
	"dos/internal/services/master/domain/object"
	"dos/internal/services/master/domain/replicate"
	"dos/internal/services/master/domain/storagenode"
	"dos/internal/services/master/repo"

	"google.golang.org/grpc"
)

type App struct {
	conn *connect.ConnCache

	objectHodler   *ObjectAuthorityHolder
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
	adminAPI   *api.AdminServer
	storageAPI *api.StorageServer

	config *Config
}

func NewApp(config *Config) (*App, error) {
	conn := connect.NewConnCache()

	object, err := InitObjectAuthority()
	if err != nil {
		return nil, fmt.Errorf("object authority init: %w", err)
	}
	chunkRepository := repo.NewInMemChunkRepo()
	nodeRegistry := repo.NewInMemNodeRegistry()
	chunkNodeIndex := domain.NewInMemChunkNodeIndex()

	metricsService := prom.NewService(config.Metrics)

	chunkT, err := chunkrpc.NewTransport(conn, config)
	if err != nil {
		return nil, fmt.Errorf("chunk transport init: %w", err)
	}

	catalogMetrics := catalog.NewCatalogMetrics(metricsService.Provider())

	catalogService, err := catalog.NewCatalogService(catalog.CatalogDeps{
		ObjectAuthority: object.Authority,
		ChunkRepository: chunkRepository,
		Metrics:         catalogMetrics,
	})
	if err != nil {
		return nil, fmt.Errorf("catalog service init: %w", err)
	}

	catalogCleanup, err := catalog.NewCleanupService(catalog.CleanupDeps{
		ObjectAuthority: object.Authority,
		ChunkRepository: chunkRepository,
		Config:          config,
		Metrics:         catalogMetrics,
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
		ObjectReader:    object.Authority,
		ChunkRepository: chunkRepository,
		Placement:       nodePlacement,
		ChunkT:          chunkT,
		Config:          config,
		Metrics:         replicate.NewExecutorMetrics(metricsService.Provider()),
	})
	if err != nil {
		return nil, fmt.Errorf("replicate executor service init: %w", err)
	}

	replicatePlanner, err := replicate.NewPlannerService(replicate.PlannerDeps{
		ObjectReader:    object.Authority,
		ChunkRepository: chunkRepository,
		Replication:     replicateExecutor,
		Config:          config,
	})
	if err != nil {
		return nil, fmt.Errorf("replicate planner init: %w", err)
	}

	nodeLifecycle, err := storagenode.NewLifecycleService(storagenode.LifecycleDeps{
		ChunkRepository: chunkRepository,
		ChunkNodeIndex:  chunkNodeIndex,
		NodeRegistry:    nodeRegistry,
		Metrics:         storagenode.NewLifecycleMetrics(metricsService.Provider()),
	})
	if err != nil {
		return nil, fmt.Errorf("storage lifecycle service init: %w", err)
	}

	nodeReport, err := storagenode.NewReportService(storagenode.ReportDeps{
		ChunkRepo:      chunkRepository,
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

	adminAPI := api.NewAdminServer(clientFacade)
	clientAPI := api.NewClientServer(clientFacade)
	storageAPI := api.NewStorageServer(nodeLifecycle, nodeReport)

	app := &App{
		conn: conn,

		objectHodler:   object,
		chunkRepo:      chunkRepository,
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

		adminAPI:   adminAPI,
		clientAPI:  clientAPI,
		storageAPI: storageAPI,

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
		mpb.RegisterAdminServiceServer(s, app.adminAPI)
		mpb.RegisterMasterClientServiceServer(s, app.clientAPI)
		mpb.RegisterMasterStorageServiceServer(s, app.storageAPI)
	})
	return err
}

func (app *App) Close() error {
	return app.conn.Close()
}

type ObjectAuthorityHolder struct {
	repository *repo.InMemObjectRepo
	applier    *object.LocalObjectCommandApplier
	writer     *object.ObjectWriterImpl

	Authority *object.Authority
}

func InitObjectAuthority() (*ObjectAuthorityHolder, error) {
	repo := repo.NewInMemObjectRepo()

	applier, err := object.NewLocalObjectCommandApplier(repo)
	if err != nil {
		return nil, fmt.Errorf("command applier init: %w", err)
	}

	writer, err := object.NewObjectWriterImpl(applier)
	if err != nil {
		return nil, fmt.Errorf("object writer init: %w", err)
	}

	authority, err := object.NewObjectAuthority(repo, writer)
	if err != nil {
		return nil, fmt.Errorf("object authority init: %w", err)
	}
	holder := &ObjectAuthorityHolder{
		repository: repo,
		applier:    applier,
		writer:     writer,
		Authority:  authority,
	}

	return holder, nil
}
