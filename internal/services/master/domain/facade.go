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

type ReplicationScheduler interface {
	Schedule(context.Context, t.ChunkID)
}

type ClientFacadeService struct {
	catalog   *CatalogService
	placement *storagenode.PlacementService
	lifecycle *storagenode.LifecycleService
	replicate ReplicationScheduler

	config ClientFacadeConfig
}

func NewClientFacadeService(
	objectCatalog *CatalogService,
	placement *storagenode.PlacementService,
	lifecycle *storagenode.LifecycleService,
	replicate ReplicationScheduler,
	config ClientFacadeConfig,
) *ClientFacadeService {
	return &ClientFacadeService{
		catalog:   objectCatalog,
		placement: placement,
		lifecycle: lifecycle,
		replicate: replicate,
		config:    config,
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
		MaxCount:     replicaCount,
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
	chunkIDs, err := s.catalog.GetObjectChunks(ctx, objectID)
	if err != nil {
		return t.ObjectAccess{}, err
	}

	placements := []t.ChunkPlacement{}
	for _, chunkID := range chunkIDs {

		chunk, err := s.catalog.DescribeChunk(ctx, chunkID)
		if err != nil {
			return t.ObjectAccess{}, fmt.Errorf("describe chunk %s: %w", chunkID, err)
		}
		nodes, err := s.placement.GetChunkNodes(ctx, chunkID)
		if err != nil {
			return t.ObjectAccess{}, fmt.Errorf("get chunk %s nodes: %w", chunkID, err)
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

func (s *ClientFacadeService) ListNodes(ctx context.Context) []t.NodeInfo {
	return s.lifecycle.ListNodes(ctx)
}

func (s *ClientFacadeService) SetReplication(ctx context.Context, objectID t.ObjectID, count int) error {
	nodesCount := s.lifecycle.GetNodeCount(ctx)
	if count > nodesCount{
		return fmt.Errorf("requested replica count %d exceeds number of nodes %d", count, nodesCount)
	}
	if err := s.catalog.SetReplicaCount(ctx, objectID, count); err != nil {
		return fmt.Errorf("set replication count: %w", err)
	}

	chunkIDs, err := s.catalog.GetObjectChunks(ctx, objectID)
	if err != nil {
		return fmt.Errorf("get object chunk ids: %w", err)
	}
	for _, chunkID := range chunkIDs {
		s.replicate.Schedule(ctx, chunkID)
	}
	return nil
}
