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
		ChunkDesc: t.ChunkDesc{
			ChunkID: chunkID,
			ChunkKey: cmd.ChunkKey, 
			ChunkSize: cmd.ChunkSize, 
		},
		Nodes: toNodeRef(nodes...),
	}
	return res, nil
}

func (s *MasterService) GetObjectAccess(ctx context.Context, oid t.ObjectID) (t.ObjectAccess, error) {

	obj, err := s.objectRepo.Get(ctx, oid)
	if err != nil {
		return t.ObjectAccess{}, fmt.Errorf("access object: %w", err) 
	}

	var totalSize int64
	placements := []t.ChunkPlacement{}
	for key, chunkID := range obj.Chunks {
		chunk, err := s.chunkRepo.Get(ctx, chunkID)
		if err != nil {
			return t.ObjectAccess{}, fmt.Errorf("access chunk %s: %w", chunkID, err)
		}
		if chunk.ReplicaCount == 0 {
			return t.ObjectAccess{}, fmt.Errorf("replica count = 0: %w", m.ErrChunkNotAvailable) 
		}
		
		nodeIDs := s.index.GetChunkNodes(ctx, chunkID)
		nodes := s.nodeReg.GetMany(ctx, nodeIDs...)
		placement := t.ChunkPlacement{
			ChunkDesc: t.ChunkDesc{
				ChunkID: chunkID, 
				ChunkKey: key,
				ChunkSize: chunk.Digest.Size,
			},
			Nodes: toNodeRef(nodes...),
		}
		
		totalSize += chunk.Digest.Size
		placements = append(placements, placement)
	}
	objectAccess := t.ObjectAccess{
		ObjectDesc: t.ObjectDesc{
			ID: obj.ID,
			TotalSize: totalSize,
		},
		Chunks: placements,
	}
		
	return objectAccess, nil
}

func (s *MasterService) ListObjects(ctx context.Context) ([]t.ObjectItem, error) {
	return s.objectRepo.List(ctx), nil
}

func toNodeRef(nodes ...m.Node) []t.NodeRef {
	refs := make([]t.NodeRef, 0, len(nodes))
	for _, node := range nodes {
		refs = append(refs, node.NodeRef)
	}
	return refs
}
