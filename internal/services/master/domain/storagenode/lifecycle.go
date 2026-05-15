package storagenode

import (
	"context"
	"dos/internal/common/dosctx"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
	m "dos/internal/services/master"
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
}


func NewLifecycleService(
	chunkNodeIndex  m.ChunkNodeIndex,
	chunkRepository m.ChunkRepo,
	nodeRegistry m.NodeRegistry,
) *LifecycleService {
	return &LifecycleService{
		nodeRegistry: nodeRegistry,
		chunkRepository: chunkRepository,
		chunkNodeIndex: chunkNodeIndex,
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


func (s *LifecycleService) Remove(ctx context.Context, nodeID t.NodeID) ([]t.ChunkID, error) {

	ctx = dosctx.WithService(ctx, "lifecycle")
	ctx = dosctx.WithOperation(ctx, "remove node")
	ctx = dosctx.WithNodeID(ctx, nodeID)

	if _, err := s.nodeRegistry.Get(ctx, nodeID); err != nil {
		return nil, fmt.Errorf("get node %s: %w", nodeID, err)
	}
	
	chunks := s.chunkNodeIndex.GetNodeChunks(ctx, nodeID)
	s.chunkNodeIndex.DetachNode(ctx, nodeID)
	s.nodeRegistry.Unregister(ctx, nodeID)
	for _, chunkID := range chunks {
		s.chunkRepository.DecReplication(ctx, chunkID)
	}

	return chunks, nil 
}

func (s *LifecycleService) GetInactive(ctx context.Context, cutoff time.Time) []t.NodeID {
	return s.nodeRegistry.GetInactive(ctx, cutoff)
}

func (s *LifecycleService) ListNodes(ctx context.Context) []t.NodeInfo {
	nodes := s.nodeRegistry.Find(ctx, m.NodeQuery{})
	infos := utils.Map(nodes, func(n m.Node) t.NodeInfo {
		return t.NodeInfo{
			ID: n.ID,
			Addr: n.Addr,
			ChunkCount: n.Stats.ChunkCount,
			UsedBytes: n.Stats.UsedBytes,
		}
	})
	return infos
}

func (s *LifecycleService) GetNodeCount(ctx context.Context) int {
	return s.nodeRegistry.Count(ctx)
}

