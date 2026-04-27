package domain

import (
	"context"
	t "dos/internal/common/types"
	m "dos/internal/services/master"
	"fmt"
)

type MasterServiceConfig struct{
	ReplicationCount int
	ChunkAllocationMarginBytes int64
}

type MasterService struct {
	chunkRepo m.ChunkRepo
	objectRepo m.ObjectRepo
	nodeReg m.NodeRegistry

	index m.ChunkNodeIndex
	placementPolicy m.PlacementPolicy
	config *MasterServiceConfig
}

func NewMasterService(
	chunkRepo m.ChunkRepo,
	objectRepo m.ObjectRepo,
	nodeReg m.NodeRegistry,
	config *MasterServiceConfig,
) *MasterService {
	return &MasterService{
		chunkRepo: chunkRepo,
		objectRepo: objectRepo,
		nodeReg: nodeReg,
		placementPolicy: &RandomPlacementPolicy{},
		index: NewInMemChunkNodeIndex(),
		config: config,
	}
}

func (s *MasterService) CreateObject(ctx context.Context, oid t.ObjectID) error {
	return s.objectRepo.Create(ctx, oid)
}

func (s *MasterService) AllocateChunk(
	ctx context.Context,
	cmd *m.AllocateChunkCommand,
) (t.ChunkPlacement, error) {

	_, err := s.objectRepo.Get(ctx, cmd.ObjectID)
	if err != nil {
		return t.ChunkPlacement{}, fmt.Errorf("ensure object exists: %w", err)
	}
	
	candidateNodes, err := s.GetCandidateNodes(ctx, m.CandidateNodesQuery{
		MinFreeBytes: cmd.ChunkSize + s.config.ChunkAllocationMarginBytes,
	})
	if err != nil {
		return t.ChunkPlacement{}, fmt.Errorf("get candidate nodes: %w", err)
	}
	
	nodes := s.placementPolicy.Select(candidateNodes, s.config.ReplicationCount)
	if len(nodes) == 0 {
		return t.ChunkPlacement{}, m.ErrNoCandidateNodes
	}

	chunkID := s.chunkRepo.NewChunkID()
	err = s.objectRepo.AddChunk(ctx, cmd.ObjectID, cmd.ChunkKey, chunkID)
	if err != nil {
		return t.ChunkPlacement{}, fmt.Errorf("add chunk to object: %w", err)
	}

	if err := s.chunkRepo.Create(ctx, chunkID); err  != nil {
		return t.ChunkPlacement{}, fmt.Errorf("create chunk: %w", err)
	}
	
	res := t.ChunkPlacement{
		ID: chunkID,
		Key: cmd.ChunkKey, 
		Nodes: toNodeRef(nodes...),
	}
	return res, nil
}

func (s *MasterService) GetObjectAccess(ctx context.Context, oid t.ObjectID) (m.ObjectAccess, error) {

	obj, err := s.objectRepo.Get(ctx, oid)
	if err != nil {
		return m.ObjectAccess{}, fmt.Errorf("access object: %w", err) 
	}
	
	objectAccess := m.ObjectAccess{ObjectID: obj.ID}
	for key, chunkID := range obj.Chunks {
		chunk, err := s.chunkRepo.Get(ctx, chunkID)
		if err != nil {
			return m.ObjectAccess{}, fmt.Errorf("access chunk %s: %w", chunkID, err)
		}
		
		chunkPlacement := t.ChunkPlacement{ID: chunkID, Key: key}
		nodeIDs := s.index.GetChunkNodes(ctx, chunkID)
		nodes := s.nodeReg.GetMany(ctx, nodeIDs...)
		chunkPlacement.Nodes = toNodeRef(nodes...)
		
		objectAccess.TotalSize += chunk.Digest.Size
		objectAccess.Chunks = append(objectAccess.Chunks, chunkPlacement)
	}
	return objectAccess, nil
}

func (s *MasterService) RegisterStorageNode(ctx context.Context, addr string) (t.NodeRef, error) {
	nref, err := s.nodeReg.Register(ctx, addr)
	if err != nil {
		return t.NodeRef{}, fmt.Errorf("register node: %w", err)
	}
	return nref, err
}

func (s *MasterService) ReportChunkStorage(
	ctx context.Context, nodeID t.NodeID, chunks []t.ChunkDesc,
) (map[t.ChunkID]string, error) {

	if _, err := s.nodeReg.Get(ctx, nodeID); err != nil {
		return nil, fmt.Errorf("get node %s: %w", nodeID, err)
	}

	rejected := map[t.ChunkID]string{}
	for _, chunk := range chunks {
		if err := s.chunkRepo.SetDigest(ctx, chunk.ID, chunk.Digest); err != nil {
			rejected[chunk.ID] = m.ErrChunkDigestConflict.Error()
			continue
		}

		s.index.AttachChunk(ctx, nodeID, chunk.ID)
	}
	return rejected, nil
}

func (s *MasterService) GetCandidateNodes(
	ctx context.Context, query m.CandidateNodesQuery,
) ([]m.Node, error) {

	nodesToExclude := s.index.GetChunkNodes(ctx, query.ExcludeChunk)
	nodes, err := s.nodeReg.Find(ctx, m.NodeQuery{
		MinFreeBytes: query.MinFreeBytes,
		ExcludeIDs: nodesToExclude,
	})
	if err != nil {
		return []m.Node{}, err
	}
	return nodes, nil	
}

func toNodeRef(nodes ...m.Node) []t.NodeRef {
	refs := make([]t.NodeRef, 0, len(nodes))
	for _, node := range nodes {
		refs = append(refs, node.NodeRef)
	}
	return refs
}





