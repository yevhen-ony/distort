package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	t "dos/internal/common/types"
	s "dos/internal/services/storage"
)

type ChunkCatalogConfig interface {
	MaxStorageBytes() int64
}

type ChunkCatalogDeps struct {
	Config  ChunkCatalogConfig
	Metrics *ChunkCatalogMetrics
}

type ChunkCatalogService struct {
	catalog    ChunkCatalog
	mu         sync.RWMutex
	totalBytes int64

	config  ChunkCatalogConfig
	metrics *ChunkCatalogMetrics
}

func NewChunkCatalogService(deps ChunkCatalogDeps) (*ChunkCatalogService, error) {
	if deps.Config == nil {
		return nil, errors.New("missing config")
	}
	if deps.Metrics == nil {
		return nil, errors.New("missing metrics")
	}

	service := &ChunkCatalogService{
		config:  deps.Config,
		metrics: deps.Metrics,
		catalog: make(ChunkCatalog),
	}
	return service, nil
}

func (cs *ChunkCatalogService) Add(meta *t.ChunkMeta) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if _, ok := cs.catalog[meta.ID]; ok {
		return s.ErrChunkConflict
	}

	size := meta.Digest.Size

	cs.metrics.ChunksCount.Add(1)
	cs.metrics.ChunksTotalBytes.Add(float64(size))

	cs.catalog[meta.ID] = NewChunkRecord(*meta)
	cs.totalBytes += meta.Digest.Size

	return nil
}

func (cs *ChunkCatalogService) Has(chunkID t.ChunkID) bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	_, ok := cs.catalog[chunkID]
	return ok
}

func (cs *ChunkCatalogService) Get(chunkID t.ChunkID) (t.ChunkMeta, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	rec, ok := cs.catalog[chunkID]
	if !ok {
		return t.ChunkMeta{}, s.ErrChunkNotFound
	}
	return *rec.Meta.Clone(), nil
}

func (cs *ChunkCatalogService) Remove(chunkID t.ChunkID) bool {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	rec, ok := cs.catalog[chunkID]
	if !ok {
		return false
	}
	size := rec.Meta.Digest.Size

	cs.metrics.ChunksCount.Add(-1)
	cs.metrics.ChunksTotalBytes.Add(float64(-size))

	cs.totalBytes -= size
	delete(cs.catalog, chunkID)
	return true
}

func (cs *ChunkCatalogService) GetStats() t.NodeStats {

	cs.mu.RLock()
	defer cs.mu.RUnlock()

	return t.NodeStats{
		FreeBytes:  cs.config.MaxStorageBytes() - cs.totalBytes,
		UsedBytes:  cs.totalBytes,
		ChunkCount: len(cs.catalog),
	}
}

type CatalogSource interface {
	List() ([]t.ChunkID, error)
	GetMeta(t.ChunkID) (t.ChunkMeta, error)
}

func (cs *ChunkCatalogService) BuildCatalog(
	ctx context.Context, source CatalogSource,
) ([]t.ChunkMeta, error) {

	ids, err := source.List()
	if err != nil {
		return nil, fmt.Errorf("list chunks: %w", err)
	}

	catalog := make(ChunkCatalog, len(ids))
	var totalBytes int64

	metas := make([]t.ChunkMeta, 0, len(ids))
	for _, id := range ids {
		meta, err := source.GetMeta(id)
		if err != nil {
			slog.Error("read chunk", "id", id, "error", err)
			continue
		}
		catalog[id] = NewChunkRecord(meta)
		totalBytes += meta.Digest.Size

		metas = append(metas, *meta.Clone())
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.catalog = catalog
	cs.totalBytes = totalBytes

	cs.metrics.ChunksCount.Set(float64(len(catalog)))
	cs.metrics.ChunksTotalBytes.Set(float64(totalBytes))

	return metas, nil
}
