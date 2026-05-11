package storagenode 

import (
	"context"
	t "dos/internal/common/types"
	m "dos/internal/services/master"
	"errors"
	"fmt"
	"time"
)

type StorageNodeConfig interface {
	ChunkAllocationMarginBytes() int64
}

type LifecycleService struct {
	nodeRegistry m.NodeRegistry
	chunkRepository m.ChunkRepo
	chunkNodeIndex  m.ChunkNodeIndex

	reconcileSink m.ReconcileSink
}

func NewLifecycleService(
	chunkNodeIndex  m.ChunkNodeIndex,
	chunkRepository m.ChunkRepo,
	nodeRegistry m.NodeRegistry,
	reconcileSink m.ReconcileSink,
) *LifecycleService {
	return &LifecycleService{
		nodeRegistry: nodeRegistry,
		chunkRepository: chunkRepository,
		chunkNodeIndex: chunkNodeIndex,
		reconcileSink: reconcileSink,
	}
}

func (s *LifecycleService) Register(ctx context.Context, addr string) (t.NodeRef, error) {

	nref, err := s.nodeRegistry.Register(ctx, addr)
	if err != nil {
		return t.NodeRef{}, fmt.Errorf("register node: %w", err)
	}
	return nref, err
}


func (s *LifecycleService) UpdateStats(ctx context.Context, nodeID t.NodeID, stats t.NodeStats) error {
	if _, err := s.nodeRegistry.Get(ctx, nodeID); err != nil {
		return fmt.Errorf("get node %s: %w", nodeID, err)
	}
	if err := s.nodeRegistry.UpdateStats(ctx, nodeID, stats); err != nil {
		return fmt.Errorf("update stats for node %s: %w", nodeID, err)
	}
	return nil 
}


func (s *LifecycleService) Remove(ctx context.Context, nodeID t.NodeID) error {
	if _, err := s.nodeRegistry.Get(ctx, nodeID); err != nil {
		return fmt.Errorf("get node %s: %w", nodeID, err)
	}
	
	var errs []error
	chunks := s.chunkNodeIndex.GetNodeChunks(ctx, nodeID)
	for _, chunk := range chunks {
		if err := s.chunkRepository.DecReplication(ctx, chunk); err != nil {
			errs = append(errs, fmt.Errorf("dec replica for chunk %s: %w", chunk, err))
		}
	}

	s.chunkNodeIndex.DetachNode(ctx, nodeID)

	if err := s.nodeRegistry.Unregister(ctx, nodeID); err != nil {
		errs = append(errs, fmt.Errorf("unregister node %s: %w", nodeID, err))
	}
	return errors.Join(errs...)
}

func (s *LifecycleService) RemoveInactive(ctx context.Context, cutoff time.Time) (int, error) {
	var errs []error
	nodes := s.nodeRegistry.GetInactive(ctx, cutoff)

	var count int
	for _, node := range nodes {
		if err := s.Remove(ctx, node); err != nil {
			errs = append(errs, err)
		} else {
			count++
		}
	}
	return count, errors.Join(errs...)
}

