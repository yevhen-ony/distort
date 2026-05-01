package domain

import (
	"context"
	t "dos/internal/common/types"
	m "dos/internal/services/master"
	"fmt"
)

func (s *MasterService) CreateObject(ctx context.Context, oid t.ObjectID) error {
	return s.objectRepo.Create(ctx, oid)
}

func (s *MasterService) AllocateChunk(
	ctx context.Context,
	cmd *m.AllocateChunkCommand,
) (t.ChunkLocation, error) {

	_, err := s.objectRepo.Get(ctx, cmd.ObjectID)
	if err != nil {
		return t.ChunkLocation{}, fmt.Errorf("ensure object exists: %w", err)
	}
	
	candidateNodes, err := s.GetCandidateNodes(ctx, m.CandidateNodesQuery{
		MinFreeBytes: cmd.ChunkSize + s.config.ChunkAllocationMarginBytes,
	})
	if err != nil {
		return t.ChunkLocation{}, fmt.Errorf("get candidate nodes: %w", err)
	}
	
	nodes := s.placementPolicy.Select(candidateNodes, s.config.ReplicationCount)
	if len(nodes) == 0 {
		return t.ChunkLocation{}, m.ErrNoCandidateNodes
	}

	chunkID := s.chunkRepo.NewChunkID()
	err = s.objectRepo.AddChunk(ctx, cmd.ObjectID, cmd.ChunkKey, chunkID)
	if err != nil {
		return t.ChunkLocation{}, fmt.Errorf("add chunk to object: %w", err)
	}

	if err := s.chunkRepo.Create(ctx, chunkID); err  != nil {
		return t.ChunkLocation{}, fmt.Errorf("create chunk: %w", err)
	}
	
	res := t.ChunkLocation{
		ChunkID: chunkID,
		ChunkKey: cmd.ChunkKey, 
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
		
		chunkPlacement := t.ChunkLocation{ChunkID: chunkID, ChunkKey: key}
		nodeIDs := s.index.GetChunkNodes(ctx, chunkID)
		nodes := s.nodeReg.GetMany(ctx, nodeIDs...)
		chunkPlacement.Nodes = toNodeRef(nodes...)
		
		objectAccess.TotalSize += chunk.Digest.Size
		objectAccess.Chunks = append(objectAccess.Chunks, chunkPlacement)
	}
	return objectAccess, nil
}

func toNodeRef(nodes ...m.Node) []t.NodeRef {
	refs := make([]t.NodeRef, 0, len(nodes))
	for _, node := range nodes {
		refs = append(refs, node.NodeRef)
	}
	return refs
}
