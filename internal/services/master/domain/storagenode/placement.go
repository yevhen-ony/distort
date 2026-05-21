package storagenode

import (
	"context"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
	m "dos/internal/services/master"
	"errors"
)

type PlacementConfig interface {
	ChunkAllocationMarginBytes() int64
}

type PlacementDeps struct {
	ChunkNodeIndex m.ChunkNodeIndex
	NodeRegistry        m.NodeRegistry
	Config         PlacementConfig
}

type PlacementService struct {
	chunkNodeIndex m.ChunkNodeIndex
	nodeReg        m.NodeRegistry
	config PlacementConfig
}

func NewPlacementService(deps PlacementDeps) (*PlacementService, error) {
	if deps.ChunkNodeIndex == nil {
		return nil, errors.New("missing chunk-node index")
	}
	if deps.NodeRegistry == nil {
		return nil, errors.New("missing node registry")
	}
	if deps.Config == nil {
		return nil, errors.New("missing config")
	}
	service := &PlacementService{
		chunkNodeIndex: deps.ChunkNodeIndex,
		nodeReg:        deps.NodeRegistry,
		config:         deps.Config,
	}
	return service, nil
}

func (s *PlacementService) GetCandidates(
	ctx context.Context, query m.CandidateNodesQuery,
) ([]t.NodeRef, error) {

	nodesToExclude := s.chunkNodeIndex.GetChunkNodes(ctx, query.ExcludeChunk)
	nodes := s.nodeReg.Find(ctx, m.NodeQuery{
		MinFreeBytes: query.MinFreeBytes + s.config.ChunkAllocationMarginBytes(),
		ExcludeIDs:   nodesToExclude,
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
	nodes := s.nodeReg.GetMany(ctx, nodeIDs...)
	nodeRefs := utils.Map(nodes, func(n m.Node) t.NodeRef { return n.NodeRef })
	return nodeRefs, nil
}
