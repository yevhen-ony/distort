package domain

import (
	"context"
	"fmt"

	t "dos/internal/common/types"
	"dos/internal/common/utils"
	m "dos/internal/services/master"
)

type CatalogService struct {
	objectRepo m.ObjectRepo
	chunkRepo  m.ChunkRepo
}

func NewCatalogService(objectRepo m.ObjectRepo, chunkRepo m.ChunkRepo) *CatalogService {
	return &CatalogService{
		objectRepo: objectRepo,
		chunkRepo:  chunkRepo,
	}
}

func (s *CatalogService) Create(
	ctx context.Context, objectID t.ObjectID, replicaCount int,
) error {

	return s.objectRepo.Create(ctx, objectID, replicaCount)
}

func (s *CatalogService) GetReplicaCount(ctx context.Context, objectID t.ObjectID) (int, error) {
	obj, err := s.objectRepo.Get(ctx, objectID)
	if err != nil {
		return 0, err
	}
	return obj.DesiredReplication, nil
}

func (s *CatalogService) AllocateChunk(
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
		ChunkID:   chunkID,
		ChunkKey:  chunkKey,
		ChunkSize: chunkSize,
	}
	return desc, nil
}

func (s *CatalogService) GetChunks(ctx context.Context, objectID t.ObjectID) ([]t.ChunkDesc, error) {

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
			ChunkID:   chunkID,
			ChunkKey:  chunkKey,
			ChunkSize: chunk.Digest.Size,
		})
	}
	return result, nil
}

func (s *CatalogService) ListObjects(ctx context.Context) []t.ObjectInfo {
	return utils.Map(s.objectRepo.List(ctx), func(o m.Object) t.ObjectInfo {
		return t.ObjectInfo {
			ID: o.ID,
			ChunkCount: len(o.Chunks),
		}
	})
}

func (s *CatalogService) ListChunks(ctx context.Context) []t.ChunkInfo {
	return utils.Map(s.chunkRepo.List(ctx), func(c m.Chunk) t.ChunkInfo {
		size := int64(0)
		if c.ReplicaCount > 0 {
			size = c.Digest.Size
		}
		return t.ChunkInfo{
			ID:           c.ID,
			Size:         size,
			ReplicaCount: c.ReplicaCount,
			ObjectID:     c.ObjectID,
		}
	})
}
