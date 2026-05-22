package storagenode

import (
	"context"
	"dos/internal/common/dosctx"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
	m "dos/internal/services/master"
	"errors"
	"fmt"
	"time"
)

type StorageNodeConfig interface {
	ChunkAllocationMarginBytes() int64
}

type LifecycleDeps struct {
	NodeRegistry   m.NodeRegistry
	ChunkRepo      m.ChunkRepo
	ChunkNodeIndex m.ChunkNodeIndex
	Metrics        *LifecycleMetrics
}

type LifecycleService struct {
	nodeReg        m.NodeRegistry
	chunkRepo      m.ChunkRepo
	chunkNodeIndex m.ChunkNodeIndex
	metrics        *LifecycleMetrics
}

func NewLifecycleService(deps LifecycleDeps) (*LifecycleService, error) {
	if deps.NodeRegistry == nil {
		return nil, errors.New("missing node registry")
	}
	if deps.ChunkRepo == nil {
		return nil, errors.New("missing chunk repository")
	}
	if deps.ChunkNodeIndex == nil {
		return nil, errors.New("missing chunk-node index")
	}
	if deps.Metrics == nil {
		return nil, errors.New("missing metrics")
	}
	service := &LifecycleService{
		nodeReg:        deps.NodeRegistry,
		chunkRepo:      deps.ChunkRepo,
		chunkNodeIndex: deps.ChunkNodeIndex,
		metrics:        deps.Metrics,
	}
	return service, nil
}

func (s *LifecycleService) Register(ctx context.Context, addr string) (t.NodeRef, error) {

	nref, err := s.nodeReg.Register(ctx, addr)
	if err != nil {
		return t.NodeRef{}, fmt.Errorf("register node: %w", err)
	}
	s.metrics.RegisteredNodesCount.Add(1)
	return nref, nil
}

func (s *LifecycleService) UpdateStats(ctx context.Context, nodeID t.NodeID, stats t.NodeStats) error {
	if _, err := s.nodeReg.Get(ctx, nodeID); err != nil {
		return fmt.Errorf("get node %s: %w", nodeID, err)
	}
	if err := s.nodeReg.UpdateStats(ctx, nodeID, stats); err != nil {
		return fmt.Errorf("update stats for node %s: %w", nodeID, err)
	}
	return nil
}

func (s *LifecycleService) Remove(ctx context.Context, nodeID t.NodeID) ([]t.ChunkID, error) {

	ctx = dosctx.WithService(ctx, "lifecycle")
	ctx = dosctx.WithOperation(ctx, "remove node")
	ctx = dosctx.WithNodeID(ctx, nodeID)

	if _, err := s.nodeReg.Get(ctx, nodeID); err != nil {
		return nil, fmt.Errorf("get node %s: %w", nodeID, err)
	}

	chunks := s.chunkNodeIndex.GetNodeChunks(ctx, nodeID)
	s.chunkNodeIndex.DetachNode(ctx, nodeID)
	s.nodeReg.Unregister(ctx, nodeID)
	for _, chunkID := range chunks {
		_ = s.chunkRepo.DecReplicaCount(ctx, chunkID)
	}
	s.metrics.RegisteredNodesCount.Add(-1)
	return chunks, nil
}

func (s *LifecycleService) GetInactive(ctx context.Context, cutoff time.Time) []t.NodeID {
	return s.nodeReg.GetInactive(ctx, cutoff)
}

func (s *LifecycleService) ListNodes(ctx context.Context) []t.NodeInfo {
	nodes := s.nodeReg.Find(ctx, m.NodeQuery{})
	infos := utils.Map(nodes, func(n m.Node) t.NodeInfo {
		return t.NodeInfo{
			ID:         n.ID,
			Addr:       n.Addr,
			ChunkCount: n.Stats.ChunkCount,
			UsedBytes:  n.Stats.UsedBytes,
		}
	})
	return infos
}

func (s *LifecycleService) GetNodeCount(ctx context.Context) int {
	return s.nodeReg.Count(ctx)
}
