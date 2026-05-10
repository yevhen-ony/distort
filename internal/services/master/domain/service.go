package domain

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	t "dos/internal/common/types"
	m "dos/internal/services/master"
)

var (
	ErrInvalidConfig = errors.New("invalid config")
)

type MasterServiceConfig struct {
	ReplicationCount           int           `yaml:"replication_count"`
	ChunkAllocationMarginBytes int64         `yaml:"chunk_allocation_mergin_bytes"`
	NodeInactivityTimeout      time.Duration `yaml:"node_inactivity_timeout"`
	NodeCleanupInterval        time.Duration `yaml:"node_cleanup_interval"`
}

type ReconcileSink interface {
	Enqueue(context.Context, t.ChunkID)
}

type MasterService struct {
	chunkRepo  m.ChunkRepo
	objectRepo m.ObjectRepo
	nodeReg    m.NodeRegistry

	index           m.ChunkNodeIndex
	placementPolicy m.PlacementPolicy
	config          *MasterServiceConfig

	reconcileSink ReconcileSink
}

func NewMasterService(
	chunkRepo m.ChunkRepo,
	objectRepo m.ObjectRepo,
	nodeReg m.NodeRegistry,
	config *MasterServiceConfig,
) (*MasterService, error) {
	if err := validateConfig(config); err != nil {
		return nil, err 
	}

	service := &MasterService{
		chunkRepo:       chunkRepo,
		objectRepo:      objectRepo,
		nodeReg:         nodeReg,
		placementPolicy: &RandomPlacementPolicy{},
		index:           NewInMemChunkNodeIndex(),
		config:          config,
	}
	return service, nil
}

func (s *MasterService) EvictInactiveNodes(ctx context.Context) (int, error) {

	var errs []error
	cutoff := time.Now().UTC().Add(-s.config.NodeInactivityTimeout)
	nodes := s.nodeReg.GetInactive(ctx, cutoff)

	var count int
	for _, node := range nodes {
		if err := s.EvictStorageNode(ctx, node); err != nil {
			errs = append(errs, err)
		} else {
			count++
		}
	}
	return count, errors.Join(errs...)
}

func (s *MasterService) RunNodeCleanupLoop(ctx context.Context) {
	
	timer := time.NewTimer(s.config.NodeCleanupInterval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}

		count, err := s.EvictInactiveNodes(ctx)
		slog.DebugContext(ctx, "evicted inactive nodes", "count", count)
		if err != nil {
			slog.ErrorContext(ctx, "evict inactive nodes failed", "error", err)
		}

		timer.Reset(s.config.NodeCleanupInterval)
	}
}


func validateConfig(config *MasterServiceConfig) error {
	if config == nil {
		return fmt.Errorf("config nil: %w", ErrInvalidConfig)
	}

	if config.NodeInactivityTimeout <= 0 {
		return fmt.Errorf("node inactivity interval: %w", ErrInvalidConfig)
	}
	if config.NodeCleanupInterval <= 0 {
		return fmt.Errorf("node cleanup interval: %w", ErrInvalidConfig)
	}
	if config.ReplicationCount <= 0 {
		return fmt.Errorf("replication count: %w", ErrInvalidConfig)
	}

	if config.ChunkAllocationMarginBytes < 0 {
		return fmt.Errorf("chunk allocation margin: %w", ErrInvalidConfig)
	}
	return nil
}


