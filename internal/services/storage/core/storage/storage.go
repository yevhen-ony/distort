package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	t "dos/internal/common/types"
)

type NopReporter struct{}

func (*NopReporter) Report(context.Context, t.StorageNodeReport) {}
func (*NopReporter) Flush(context.Context)                       {}

type StorageDeps struct {
	Inventory Inventory
	Identity  Identity
	StorageBE ChunkStorage
	ChunkT    ChunkTransport
	Config    StorageConfig
	Metrics   *StorageMetrics
}

type StorageService struct {
	inventory Inventory
	identity  Identity

	storageBE ChunkStorage
	chunkT    ChunkTransport
	config    StorageConfig

	reporter Reporter

	sem     chan struct{}
	metrics *StorageMetrics
}

func NewStorageService(deps StorageDeps) (*StorageService, error) {
	if deps.Inventory == nil {
		return nil, errors.New("missing catalog service")
	}
	if deps.Identity == nil {
		return nil, errors.New("missing identity service")
	}
	if deps.StorageBE == nil {
		return nil, errors.New("missing store")
	}
	if deps.ChunkT == nil {
		return nil, errors.New("missing storage transport")
	}
	if deps.Config == nil {
		return nil, errors.New("missing config")
	}

	service := &StorageService{
		inventory: deps.Inventory,
		identity:  deps.Identity,
		reporter:  &NopReporter{},

		storageBE: deps.StorageBE,
		chunkT:    deps.ChunkT,
		config:    deps.Config,
		metrics:   deps.Metrics,

		sem: make(chan struct{}, deps.Config.MaxParallelHeavyOps()),
	}
	return service, nil
}

func (cs *StorageService) SetReporter(reporter Reporter) {
	cs.reporter = reporter
}

func (cs *StorageService) Start(ctx context.Context) error {

	if err := cs.inventory.BuildCatalog(ctx, cs.storageBE); err != nil {
		return fmt.Errorf("build catalog: %w", err)
	}

	res := cs.StageAndReportAll(ctx)
	if len(res.Failed) != 0 {
		slog.WarnContext(ctx,
			"stage and report on start partially failed",
			"failed_chunks", res.Failed,
		)
	}
	return nil
}
