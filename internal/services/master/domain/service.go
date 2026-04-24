package domain

import (
	"context"
	m "dos/internal/services/master"
	"fmt"
)

type ServiceConfig struct{
	ReplicationCount int
	ChunkAllocationMarginBytes int64
}

type Service struct {
	chunkRepo m.ChunkRepo
	objectRepo m.ObjectRepo
	nodeReg m.NodeRegistry

	placementPolicy m.PlacementPolicy
	config *ServiceConfig
}

func (s *Service) CreateObject(ctx context.Context, oid m.ObjectID) error {
	return s.objectRepo.Create(ctx, oid)
}

func (s *Service) AllocateChunk(
	ctx context.Context,
	cmd *m.AllocateChunkCommand,
) (placement m.ChunkPlacement, err error) {
	_, err = s.objectRepo.Get(ctx, cmd.ObjectID)
	if err != nil {
		return placement, fmt.Errorf("ensure object exists: %w", err)
	}
	
	candidateNodes, err := s.nodeReg.GetCandidateNodes(
		ctx, &m.CandidateNodesQuery{
			MinFreeBytes: cmd.ChunkSize + s.config.ChunkAllocationMarginBytes,
		},
	)
	if err != nil {
		return placement, fmt.Errorf("get candidate nodes: %w", err)
	}
	
	nodes := s.placementPolicy.Select(candidateNodes, s.config.ReplicationCount)
	if len(nodes) == 0 {
		return placement, m.ErrNoCandidateNodes
	}

	chunkID, err := s.chunkRepo.Create(ctx) 
	if err != nil {
		return placement, fmt.Errorf("create chunk: %w", err)
	}
	
	err = s.objectRepo.AddChunk(ctx, cmd.ObjectID, cmd.ChunkKey, chunkID)
	if err != nil {
		return placement, fmt.Errorf("add chunk to object: %w", err)
	}
		
	placement.ChunkID = chunkID	
	placement.Nodes = toNodeAccess(nodes...) 
	return placement, nil
}

func (s *Service) GetObjectAccess(ctx context.Context, oid m.ObjectID) (m.ObjectAccess, error) {
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
		
		chunkPlacement := m.ChunkPlacement{ChunkID: chunkID, ChunkKey: key}
		nodes, err := s.nodeReg.GetChunkNodes(ctx, chunkID)
		if err != nil {
			return m.ObjectAccess{}, fmt.Errorf("access %s chunk's nodes: %w", chunkID, err)
		}
		chunkPlacement.Nodes = toNodeAccess(nodes...)
		
		objectAccess.TotalSize += chunk.Digest.Size
		objectAccess.Chunks = append(objectAccess.Chunks, chunkPlacement)
	}
	return objectAccess, nil
}



