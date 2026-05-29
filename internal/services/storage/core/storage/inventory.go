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

type ChunkInventoryDeps struct {
	Config  ChunkCatalogConfig
	Metrics *ChunkCatalogMetrics
}

type ChunkInventory struct {
	catalog    ChunkCatalog
	mu         sync.RWMutex
	totalBytes int64

	config  ChunkCatalogConfig
	metrics *ChunkCatalogMetrics
}

func NewChunkInventory(deps ChunkInventoryDeps) (*ChunkInventory, error) {
	if deps.Config == nil {
		return nil, errors.New("missing config")
	}
	if deps.Metrics == nil {
		return nil, errors.New("missing metrics")
	}

	service := &ChunkInventory{
		config:  deps.Config,
		metrics: deps.Metrics,
		catalog: make(ChunkCatalog),
	}
	return service, nil
}

func (ci *ChunkInventory) Add(meta *t.ChunkMeta) error {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	if _, ok := ci.catalog[meta.ID]; ok {
		return s.ErrChunkConflict
	}

	size := meta.Digest.Size

	ci.metrics.ChunksCount.Add(1)
	ci.metrics.ChunksTotalBytes.Add(float64(size))

	ci.catalog[meta.ID] = NewChunkRecord(*meta)
	ci.totalBytes += meta.Digest.Size

	return nil
}

func (ci *ChunkInventory) Has(chunkID t.ChunkID) bool {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	_, ok := ci.catalog[chunkID]
	return ok
}

func (ci *ChunkInventory) Get(chunkID t.ChunkID) (t.ChunkMeta, error) {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	rec, ok := ci.catalog[chunkID]
	if !ok {
		return t.ChunkMeta{}, s.ErrChunkNotFound
	}
	return *rec.Meta.Clone(), nil
}

func (ci *ChunkInventory) Remove(chunkID t.ChunkID) bool {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	rec, ok := ci.catalog[chunkID]
	if !ok {
		return false
	}
	size := rec.Meta.Digest.Size

	ci.metrics.ChunksCount.Add(-1)
	ci.metrics.ChunksTotalBytes.Add(float64(-size))

	ci.totalBytes -= size
	delete(ci.catalog, chunkID)
	return true
}

func (ci *ChunkInventory) GetStats() t.NodeStats {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	return t.NodeStats{
		FreeBytes:  ci.config.MaxStorageBytes() - ci.totalBytes,
		UsedBytes:  ci.totalBytes,
		ChunkCount: len(ci.catalog),
	}
}

func (ci *ChunkInventory) RestageActive() {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	for _, record := range ci.catalog {
		if record.State == ChunkStateActive {
			record.State = ChunkStateStaged
		}
	}
}

func (ci *ChunkInventory) ListStaged() []t.ChunkMeta {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	metas := make([]t.ChunkMeta, 0, len(ci.catalog))
	for _, record := range ci.catalog {
		if record.State == ChunkStateStaged {
			metas = append(metas, *record.Meta.Clone())
		}
	}
	return metas
}

type CatalogSource interface {
	List() ([]t.ChunkID, error)
	GetMeta(t.ChunkID) (t.ChunkMeta, error)
}

func (cs *ChunkInventory) BuildCatalog(
	ctx context.Context, source CatalogSource,
) error {

	ids, err := source.List()
	if err != nil {
		return fmt.Errorf("list chunks: %w", err)
	}

	catalog := make(ChunkCatalog, len(ids))
	var totalBytes int64

	for _, id := range ids {
		meta, err := source.GetMeta(id)
		if err != nil {
			slog.Error("read chunk", "id", id, "error", err)
			continue
		}
		catalog[id] = NewChunkRecord(meta)
		totalBytes += meta.Digest.Size
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.catalog = catalog
	cs.totalBytes = totalBytes

	cs.metrics.ChunksCount.Set(float64(len(catalog)))
	cs.metrics.ChunksTotalBytes.Set(float64(totalBytes))

	return nil
}
