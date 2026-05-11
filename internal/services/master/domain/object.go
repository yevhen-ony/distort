package domain

import (
	"context"
	"fmt"
	
	t "dos/internal/common/types"
	m "dos/internal/services/master"
)

type ObjectCatalogService struct {
	objectRepo m.ObjectRepo
	chunkRepo m.ChunkRepo
}

func NewObjectCatalogService(objectRepo m.ObjectRepo, chunkRepo m.ChunkRepo) *ObjectCatalogService {
	return &ObjectCatalogService{
		objectRepo: objectRepo,
		chunkRepo: chunkRepo,
	}
}

func (s *ObjectCatalogService) Create(
	ctx context.Context, objectID t.ObjectID, replicaCount int,
) error {
	
	return s.objectRepo.Create(ctx, objectID, replicaCount)
}

func (s *ObjectCatalogService) GetReplicaCount(ctx context.Context, objectID t.ObjectID) (int, error) {
	obj, err := s.objectRepo.Get(ctx, objectID)
	if err != nil {
		return 0, err
	}
	return obj.DesiredReplication, nil
}

func (s *ObjectCatalogService) AllocateChunk(
	ctx context.Context, objectID t.ObjectID, chunkKey t.ChunkKey, chunkSize int64,
) (t.ChunkDesc, error) {

	chunkID := s.chunkRepo.NewChunkID()
	if err := s.chunkRepo.Create(ctx, chunkID, objectID); err != nil {
		return t.ChunkDesc{}, fmt.Errorf("create chunk: %w", err)
	}

	if err := s.objectRepo.AddChunk(ctx, objectID, chunkKey, chunkID); err != nil {
		return t.ChunkDesc{}, fmt.Errorf("add chunk to object %s: %w", objectID, err)
	}

	desc := t.ChunkDesc{
		ChunkID: chunkID,
		ChunkKey: chunkKey,
		ChunkSize: chunkSize, 
	}
	return desc, nil
}

func (s *ObjectCatalogService) GetChunks(ctx context.Context, objectID t.ObjectID) ([]t.ChunkDesc, error) {

	object, err := s.objectRepo.Get(ctx, objectID)
	if err != nil {
		return nil, err
	}

	result := make([]t.ChunkDesc, 0, len(object.Chunks))
	for chunkKey, chunkID := range object.Chunks {
		chunk, err := s.chunkRepo.Get(ctx, chunkID)	
		if err != nil {
			return nil, fmt.Errorf("access chunk %s: %w", chunkID, err)
		}
		if chunk.ReplicaCount == 0 {
			return nil, fmt.Errorf("access chunk %s: %w", chunkID, m.ErrChunkNotAvailable)
		}
		result = append(result, t.ChunkDesc{
			ChunkID: chunkID,
			ChunkKey: chunkKey,
			ChunkSize: chunk.Digest.Size,
		})
	}
	return result, nil
}

func (s *ObjectCatalogService) List(ctx context.Context) []t.ObjectItem {
	return s.objectRepo.List(ctx)
}

