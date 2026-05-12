package storagenode

import (
	"context"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
	m "dos/internal/services/master"
)
type PlacementConfig interface {
	ChunkAllocationMarginBytes() int64
}

type PlacementService struct {
	chunkNodeIndex  m.ChunkNodeIndex
	nodeRegistry m.NodeRegistry

	config PlacementConfig
}

func NewPlacementService(
	index m.ChunkNodeIndex, nodeRegistry m.NodeRegistry, config PlacementConfig,
) *PlacementService {
	return &PlacementService{
		chunkNodeIndex: index,
		nodeRegistry: nodeRegistry,
		config: config,
	}
}

func (s *PlacementService) GetCandidates(
	ctx context.Context, query m.CandidateNodesQuery,
) ([]t.NodeRef, error) {

	nodesToExclude := s.chunkNodeIndex.GetChunkNodes(ctx, query.ExcludeChunk)
	nodes := s.nodeRegistry.Find(ctx, m.NodeQuery{
		MinFreeBytes: query.MinFreeBytes + s.config.ChunkAllocationMarginBytes(),
		ExcludeIDs: nodesToExclude,
	})

	count := query.MaxCount
	if count == 0 {
		count = len(nodes)
	}
	nodeRefs := utils.Map(nodes, func(n m.Node) t.NodeRef { return n.NodeRef })
	nodeRefs = utils.RandomSelect(nodeRefs, count)
	
	return nodeRefs, nil
}

func (s *PlacementService) GetChunkNodes(ctx context.Context, chunkID t.ChunkID) ([]t.NodeRef, error) {
	nodeIDs := s.chunkNodeIndex.GetChunkNodes(ctx, chunkID)
	if len(nodeIDs) == 0 {
		return nil, m.ErrNodeNotFound 
	}
	nodes := s.nodeRegistry.GetMany(ctx, nodeIDs...)
	nodeRefs := utils.Map(nodes, func(n m.Node) t.NodeRef { return n.NodeRef })
	return nodeRefs, nil
}
