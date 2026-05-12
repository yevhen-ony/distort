package main

import (
	"context"
	mpb "dos/gen/proto/master/v1"
	"dos/internal/common/connect"
	"dos/internal/common/listener"
	"dos/internal/services/master/api"
	"dos/internal/services/master/domain"
	"dos/internal/services/master/domain/reconcile"
	"dos/internal/services/master/domain/storagenode"
	"dos/internal/services/master/repo"
	"dos/internal/services/master/transport"

	"google.golang.org/grpc"
)

type App struct {
	conn *connect.ConnCache

	objectRepository *repo.InMemObjectRepo
	chunkRepository  *repo.InMemChunkRepo
	nodeRegistry     *repo.InMemNodeRegistry
	chunkNodeIndex   *domain.InMemChunkNodeIndex

	reconcile        *reconcile.ReconcileWorker
	catalog          *domain.CatalogService
	storageLifecycle *storagenode.LifecycleService
	storagePlacement *storagenode.PlacementService
	storageReport    *storagenode.ReportService
	storageCleanup   *storagenode.CleanupWorker

	clientFacade *domain.ClientFacadeService

	clientAPI  *api.ClientServer
	storageAPI *api.StorageServer

	config *Config
}

func NewApp(config *Config) (*App, error) {
	conn := connect.NewConnCache()

	storageTransport := transport.NewStorage(conn)

	objectRepo := repo.NewInMemObjectRepo()
	chunkRepo := repo.NewInMemChunkRepo()
	nodeRegistry := repo.NewInMemNodeRegistry()
	chunkNodeIndex := domain.NewInMemChunkNodeIndex()

	catalog := domain.NewCatalogService(
		objectRepo,
		chunkRepo,
	)

	storagePlacement := storagenode.NewPlacementService(
		chunkNodeIndex,
		nodeRegistry,
		config,
	)

	reconcile := reconcile.NewReconcileWorker(
		chunkRepo,
		objectRepo,
		storagePlacement,
		storageTransport,
		config,
	)

	storageLifecycle := storagenode.NewLifecycleService(
		chunkNodeIndex,
		chunkRepo,
		nodeRegistry,
		reconcile,
	)

	storageReport := storagenode.NewReportService(
		chunkNodeIndex,
		chunkRepo,
		nodeRegistry,
		reconcile,
	)

	storageCleanup := storagenode.NewCleanupWorker(
		storageLifecycle,
		config,
	)

	clientFacade := domain.NewClientFacadeService(
		catalog,
		storagePlacement,
		storageLifecycle,
		config,
	)

	clientAPI := api.NewClientServer(clientFacade)
	storageAPI := api.NewStorageServer(storageLifecycle, storageReport)

	app := &App{
		conn: conn,

		objectRepository: objectRepo,
		chunkRepository:  chunkRepo,
		nodeRegistry:     nodeRegistry,
		chunkNodeIndex:   chunkNodeIndex,

		catalog:          catalog,
		storageLifecycle: storageLifecycle,
		storagePlacement: storagePlacement,
		storageReport:    storageReport,
		storageCleanup:   storageCleanup,
		reconcile:        reconcile,

		storageAPI: storageAPI,
		clientAPI:  clientAPI,

		config: config,
	}
	return app, nil
}

func (app *App) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go app.reconcile.RunLoop(ctx)
	go app.storageCleanup.RunLoop(ctx)

	err := listener.RunGRPCServer(ctx, &app.config.Listen, func(s *grpc.Server) {
		mpb.RegisterMasterStorageServiceServer(s, app.storageAPI)
		mpb.RegisterMasterClientServiceServer(s, app.clientAPI)
	})
	return err
}

func (app *App) Close() error {
	return app.conn.Close()
}
