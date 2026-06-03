package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
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
	catalog    s.ChunkCatalog
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
		catalog: make(s.ChunkCatalog),
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

	ci.catalog[meta.ID] = s.NewChunkRecord(*meta)
	ci.totalBytes += meta.Digest.Size

	return nil
}

func (ci *ChunkInventory) Has(chunkID t.ChunkID) bool {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	_, ok := ci.catalog[chunkID]
	return ok
}

func (ci *ChunkInventory) GetRecord(chunkID t.ChunkID) (*s.ChunkRecord, error) {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	rec, ok := ci.catalog[chunkID]
	if !ok {
		return nil, s.ErrChunkNotFound
	}
	return rec.Clone(), nil
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

func (ci *ChunkInventory) ListStaged() []t.ChunkMeta {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	metas := make([]t.ChunkMeta, 0, len(ci.catalog))
	for _, record := range ci.catalog {
		if record.State == s.ChunkStateStaged {
			metas = append(metas, *record.Meta.Clone())
		}
	}
	return metas
}

func (ci *ChunkInventory) Activate(chunkID t.ChunkID) (t.ChunkMeta, error) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	
	rec, ok := ci.catalog[chunkID]
	if !ok {
		return t.ChunkMeta{}, s.ErrChunkNotFound
	}

	rec.State = s.ChunkStateActive
	return *rec.Meta.Clone(), nil 
}

func (ci *ChunkInventory) Stage(chunkID t.ChunkID) (t.ChunkMeta, error) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	
	rec, ok := ci.catalog[chunkID]
	if !ok {
		return t.ChunkMeta{}, s.ErrChunkNotFound
	}

	rec.State = s.ChunkStateStaged
	return *rec.Meta.Clone(), nil
}

func (ci *ChunkInventory) GetState(chunkID t.ChunkID) (s.ChunkState, error)  {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	rec, ok := ci.catalog[chunkID]
	if !ok {
		return 0, s.ErrChunkNotFound
	}
	return rec.State, nil
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

	catalog := make(s.ChunkCatalog, len(ids))
	var totalBytes int64

	for _, id := range ids {
		meta, err := source.GetMeta(id)
		if err != nil {
			slog.Error("read chunk", "id", id, "error", err)
			continue
		}
		catalog[id] = s.NewChunkRecord(meta)
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

func (ci *ChunkInventory) ListRecords() []s.ChunkRecord {
  	ci.mu.RLock()
  	defer ci.mu.RUnlock()

  	records := make([]s.ChunkRecord, 0, len(ci.catalog))
  	for _, record := range ci.catalog {
  		records = append(records, *record.Clone())
  	}
  	return records
}

func (ci *ChunkInventory) ListIDs() []t.ChunkID {
	return slices.Collect(maps.Keys(ci.catalog))
}
