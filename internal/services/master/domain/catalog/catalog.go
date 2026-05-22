package catalog

import (
	"context"
	"errors"
	"fmt"

	t "dos/internal/common/types"
	"dos/internal/common/utils"
	m "dos/internal/services/master"
)

type CatalogDeps struct {
	ObjectRepo m.ObjectRepo
	ChunkRepo  m.ChunkRepo
	Metrics    *CatalogMetrics
}

type CatalogService struct {
	objectRepo m.ObjectRepo
	chunkRepo  m.ChunkRepo

	metrics *CatalogMetrics
}

func NewCatalogService(deps CatalogDeps) (*CatalogService, error) {
	if deps.ObjectRepo == nil {
		return nil, errors.New("missing object repository")
	}
	if deps.ChunkRepo == nil {
		return nil, errors.New("missing chunk repository")
	}
	if deps.Metrics == nil {
		return nil, errors.New("missing metrics")
	}

	service := &CatalogService{
		objectRepo: deps.ObjectRepo,
		chunkRepo:  deps.ChunkRepo,
		metrics:    deps.Metrics,
	}
	return service, nil
}

func (s *CatalogService) Create(ctx context.Context, objectID t.ObjectID, replicaCount int) error {

	err := s.objectRepo.Create(ctx, objectID, replicaCount)
	if err != nil {
		return fmt.Errorf("create object: %w", err)
	}
	s.metrics.ObjectCount.Add(1)
	return nil
}

func (s *CatalogService) GetReplication(ctx context.Context, objectID t.ObjectID) (int, error) {
	return s.objectRepo.GetReplication(ctx, objectID)
}

func (s *CatalogService) SetReplication(ctx context.Context, objectID t.ObjectID, count int) error {
	return s.objectRepo.SetReplication(ctx, objectID, count)
}

func (s *CatalogService) ExistsChunk(
	ctx context.Context, objectID t.ObjectID, chunkKey t.ChunkKey,
) (bool, error) {
	obj, err := s.objectRepo.Get(ctx, objectID)
	if err != nil {
		return false, fmt.Errorf("get object: %w", err)
	}
	_, ok := obj.Chunks[chunkKey]
	return ok, nil
}

func (s *CatalogService) GetChunk(
	ctx context.Context, objectID t.ObjectID, chunkKey t.ChunkKey,
) (t.ChunkID, error) {

	obj, err := s.objectRepo.Get(ctx, objectID)
	if err != nil {
		return "", fmt.Errorf("get object: %w", err)
	}
	chunkID, ok := obj.Chunks[chunkKey]
	if !ok {
		return "", m.ErrChunkNotFound
	}
	return chunkID, nil
}

func (s *CatalogService) AddChunk(
	ctx context.Context, objectID t.ObjectID, chunkKey t.ChunkKey, chunkSize int64,
) (t.ChunkDesc, error) {

	chunkID := s.chunkRepo.NewChunkID()
	err := s.chunkRepo.Create(ctx, chunkID, objectID, chunkKey)
	if err != nil {
		return t.ChunkDesc{}, fmt.Errorf("create chunk: %w", err)
	}

	err = s.objectRepo.AddChunk(ctx, objectID, chunkKey, chunkID)
	if err != nil {
		s.chunkRepo.Delete(ctx, chunkID)
		return t.ChunkDesc{}, fmt.Errorf("add chunk to object %s: %w", objectID, err)
	}

	desc := t.ChunkDesc{
		ChunkID:   chunkID,
		ChunkKey:  chunkKey,
		ChunkSize: chunkSize,
	}
	s.metrics.ChunkCount.Add(1)
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
		return t.ChunkDesc{}, fmt.Errorf("get chunk: %w", err)
	}

	size := int64(0)
	if chunk.ReplicaCount > 0 {
		size = chunk.Meta.Digest.Size
	}
	desc := t.ChunkDesc{
		ChunkID:   chunk.Meta.ID,
		ChunkKey:  chunk.ChunkKey,
		ChunkSize: size,
	}
	return desc, nil
}

func (s *CatalogService) ListObjects(ctx context.Context) []t.ObjectInfo {
	return utils.Map(s.objectRepo.List(ctx), func(o m.Object) t.ObjectInfo {
		return t.ObjectInfo{
			ID:          o.ID,
			ChunkCount:  len(o.Chunks),
			Replication: o.Replication,
		}
	})
}

func (s *CatalogService) ListChunks(ctx context.Context) []t.ChunkInfo {
	chunks, _ := s.chunkRepo.List(ctx)
	return utils.Map(chunks, func(c m.Chunk) t.ChunkInfo {
		size := int64(0)
		if c.ReplicaCount > 0 {
			size = c.Meta.Digest.Size
		}
		return t.ChunkInfo{
			ID:           c.Meta.ID,
			Size:         size,
			ReplicaCount: c.ReplicaCount,
			ObjectID:     c.ObjectID,
		}
	})
}
