package catalog

import (
	"context"
	"errors"
	"fmt"

	t "dos/internal/common/types"
	"dos/internal/common/utils"
	m "dos/internal/services/master"
	"dos/internal/services/master/domain/object"
)

type CatalogDeps struct {
	ObjectAuthority object.ObjectAuthority
	ChunkRepository m.ChunkRepo
	Metrics         *CatalogMetrics
}

type CatalogService struct {
	objects object.ObjectAuthority
	chunks  m.ChunkRepo

	metrics *CatalogMetrics
}

func NewCatalogService(deps CatalogDeps) (*CatalogService, error) {
	if deps.ObjectAuthority == nil {
		return nil, errors.New("missing object repository")
	}
	if deps.ChunkRepository == nil {
		return nil, errors.New("missing chunk repository")
	}
	if deps.Metrics == nil {
		return nil, errors.New("missing metrics")
	}

	service := &CatalogService{
		objects: deps.ObjectAuthority,
		chunks:  deps.ChunkRepository,
		metrics: deps.Metrics,
	}
	return service, nil
}

func (s *CatalogService) CreateObject(ctx context.Context, objectID t.ObjectID, replicaCount int) error {

	err := s.objects.Create(ctx, objectID, replicaCount)
	if err != nil {
		return fmt.Errorf("create object: %w", err)
	}
	s.metrics.ObjectCount.Add(1)
	return nil
}

func (s *CatalogService) GetObject(ctx context.Context, objectID t.ObjectID) (m.Object, error) {
	return s.objects.Get(ctx, objectID)
}

func (s *CatalogService) GetReplication(ctx context.Context, objectID t.ObjectID) (int, error) {
	return s.objects.GetReplication(ctx, objectID)
}

func (s *CatalogService) SetReplication(ctx context.Context, objectID t.ObjectID, count int) error {
	return s.objects.SetReplication(ctx, objectID, count)
}

func (s *CatalogService) ExistsChunk(
	ctx context.Context, slot t.ObjectSlot,
) (bool, error) {
	obj, err := s.objects.Get(ctx, slot.ObjectID)
	if err != nil {
		return false, fmt.Errorf("get object: %w", err)
	}
	_, ok := obj.Chunks[slot.ChunkKey]
	return ok, nil
}

func (s *CatalogService) GetChunkID(ctx context.Context, slot t.ObjectSlot) (t.ChunkID, error) {

	obj, err := s.objects.Get(ctx, slot.ObjectID)
	if err != nil {
		return "", fmt.Errorf("get object: %w", err)
	}
	chunkID, ok := obj.Chunks[slot.ChunkKey]
	if !ok {
		return "", m.ErrChunkNotFound
	}
	return chunkID, nil
}

func (s *CatalogService) GetChunk(ctx context.Context, chunkID t.ChunkID) (m.Chunk, error) {
	return s.chunks.Get(ctx, chunkID)
}

func (s *CatalogService) AddChunk(
	ctx context.Context, slot t.ObjectSlot, chunkSize int64,
) (t.ChunkID, error) {

	chunkID := s.chunks.NewChunkID()
	err := s.chunks.Create(ctx, chunkID, slot)
	if err != nil {
		return "", fmt.Errorf("create chunk: %w", err)
	}

	err = s.objects.AddChunk(ctx, slot, chunkID)
	if err != nil {
		s.chunks.Delete(ctx, chunkID)
		return "", fmt.Errorf("add chunk to object %s: %w", slot.ObjectID, err)
	}

	s.metrics.ChunkCount.Add(1)
	return chunkID, nil
}

func (s *CatalogService) GetObjectChunks(ctx context.Context, objectID t.ObjectID) ([]t.ChunkID, error) {

	object, err := s.objects.Get(ctx, objectID)
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

	chunk, err := s.chunks.Get(ctx, chunkID)
	if err != nil {
		return t.ChunkDesc{}, fmt.Errorf("get chunk: %w", err)
	}

	size := int64(0)
	if chunk.ReplicaCount > 0 {
		size = chunk.Meta.Digest.Size
	}
	desc := t.ChunkDesc{
		ChunkID:   chunk.Meta.ID,
		ChunkKey:  chunk.Slot.ChunkKey,
		ChunkSize: size,
	}
	return desc, nil
}

func (s *CatalogService) ListObjects(ctx context.Context) []t.ObjectInfo {
	return utils.Map(s.objects.List(ctx), func(o m.Object) t.ObjectInfo {
		return t.ObjectInfo{
			ID:          o.ID,
			ChunkCount:  len(o.Chunks),
			Replication: o.Replication,
		}
	})
}

func (s *CatalogService) ListChunks(ctx context.Context) []t.ChunkInfo {
	chunks, _ := s.chunks.List(ctx)
	return utils.Map(chunks, func(c m.Chunk) t.ChunkInfo {
		size := int64(0)
		if c.ReplicaCount > 0 {
			size = c.Meta.Digest.Size
		}
		return t.ChunkInfo{
			ID:           c.Meta.ID,
			Size:         size,
			ReplicaCount: c.ReplicaCount,
			ObjectID:     c.Slot.ObjectID,
		}
	})
}
