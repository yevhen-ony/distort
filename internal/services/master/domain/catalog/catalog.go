package catalog 

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
	return s.objectRepo.GetReplication(ctx, objectID)
}

func (s *CatalogService) SetReplicaCount(ctx context.Context, objectID t.ObjectID, count int) error {
	return s.objectRepo.SetReplication(ctx, objectID, count)
}

func (s *CatalogService) AllocateChunk(
	ctx context.Context, objectID t.ObjectID, chunkKey t.ChunkKey, chunkSize int64,
) (t.ChunkDesc, error) {

	chunkID := s.chunkRepo.NewChunkID()
	if err := s.chunkRepo.Create(ctx, chunkID, objectID, chunkKey); err != nil {
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

func (s *CatalogService) GetObjectChunks(ctx context.Context, objectID t.ObjectID) ([]t.ChunkID, error) {

	object, err := s.objectRepo.Get(ctx, objectID)
	if err != nil {
		return nil, err
	}

	result := make([]t.ChunkID, 0, len(object.Chunks))
	for _, chunkID := range object.Chunks {
		result = append(result, chunkID)
	}
	return result, nil
}

func (s *CatalogService) DescribeChunk(ctx context.Context, chunkID t.ChunkID) (t.ChunkDesc, error) {
		chunk, err := s.chunkRepo.Get(ctx, chunkID)
		if err != nil {
			return t.ChunkDesc{}, fmt.Errorf("access chunk %s: %w", chunkID, err)
		}
		
		size := int64(0)
		if chunk.ReplicaCount > 0 {
			size = chunk.Digest.Size
		}
		desc := t.ChunkDesc{
			ChunkID:   chunk.ID,
			ChunkKey:  chunk.ChunkKey,
			ChunkSize: size,
		}
		return desc, nil 
}

func (s *CatalogService) ListObjects(ctx context.Context) []t.ObjectInfo {
	return utils.Map(s.objectRepo.List(ctx), func(o m.Object) t.ObjectInfo {
		return t.ObjectInfo {
			ID: o.ID,
			ChunkCount: len(o.Chunks),
			Replication: o.Replication,
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
