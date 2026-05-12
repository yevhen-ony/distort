package domain

import (
	"context"
	t "dos/internal/common/types"
	m "dos/internal/services/master"
	"dos/internal/services/master/domain/storagenode"
	"fmt"
)

type ClientFacadeConfig interface {
	ReplicationCount() int
}

type ClientFacadeService struct {
	catalog *CatalogService
	placement *storagenode.PlacementService

	config ClientFacadeConfig
}

func NewClientFacadeService(
	objectCatalog *CatalogService,
	placement *storagenode.PlacementService,
	config ClientFacadeConfig,
) *ClientFacadeService {
	return &ClientFacadeService{
		catalog: objectCatalog,
		placement: placement,
		config: config,
	}
}

func (s *ClientFacadeService) CreateObject(ctx context.Context, oid t.ObjectID) error {
	return s.catalog.Create(ctx, oid, s.config.ReplicationCount())
}

func (s *ClientFacadeService) AllocateChunk(
	ctx context.Context,
	cmd m.AllocateChunkCommand,
) (t.ChunkPlacement, error) {

	replicaCount, err := s.catalog.GetReplicaCount(ctx, cmd.ObjectID)
	if err != nil {
		return t.ChunkPlacement{}, err
	}

	candidates, err := s.placement.GetCandidates(ctx, m.CandidateNodesQuery{
		MinFreeBytes: cmd.ChunkSize,
		MaxCount:          replicaCount,
	})
	if err != nil {
		return t.ChunkPlacement{}, fmt.Errorf("get candidate nodes: %w", err)
	}

	chunkDesc, err := s.catalog.AllocateChunk(ctx, cmd.ObjectID, cmd.ChunkKey, cmd.ChunkSize)
	if err != nil {
		return t.ChunkPlacement{}, fmt.Errorf("allocate chunk: %w", err)
	}

	res := t.ChunkPlacement{
		ChunkDesc: chunkDesc,
		Nodes:     candidates,
	}
	return res, nil
}

func (s *ClientFacadeService) GetObjectAccess(
	ctx context.Context, objectID t.ObjectID,
) (t.ObjectAccess, error) {

	var totalSize int64
	chunks, err := s.catalog.GetChunks(ctx, objectID)
	if err != nil {
		return t.ObjectAccess{}, err
	}

	placements := []t.ChunkPlacement{}
	for _, chunk := range chunks {

		nodes, err := s.placement.GetChunkNodes(ctx, chunk.ChunkID)
		if err != nil {
			return t.ObjectAccess{}, fmt.Errorf("get chunk %s nodes: %w", chunk.ChunkID, err)
		}

		totalSize += chunk.ChunkSize
		placements = append(placements, t.ChunkPlacement{
			ChunkDesc: chunk,
			Nodes:     nodes,
		})
	}
	objectAccess := t.ObjectAccess{
		ObjectDesc: t.ObjectDesc{
			ID:        objectID,
			TotalSize: totalSize,
		},
		Chunks: placements,
	}

	return objectAccess, nil
}

func (s *ClientFacadeService) ListObjects(ctx context.Context) []t.ObjectInfo {
	return s.catalog.ListObjects(ctx)
}

func (s *ClientFacadeService) ListChunks(ctx context.Context) []t.ChunkInfo {
	return s.catalog.ListChunks(ctx)
}
