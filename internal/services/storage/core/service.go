package core

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	t "dos/internal/common/types"
	s "dos/internal/services/storage"
)

type Service struct {
	store   s.ChunkStorage
	catalog s.ChunkCatalog
	mu      sync.RWMutex
}

func New(store s.ChunkStorage) (*Service, error) {
	catalog := map[t.ChunkID]t.ChunkMeta{}

	if store == nil {
		return nil, errors.New("store must not be nil") 
	}

	ids, err := store.GetAllIDs()
	if err != nil {
		return nil, fmt.Errorf("get all ids: %w", err)
	}

	for _, id := range ids {
		meta, err := store.GetMeta(id)
		if err != nil {
			slog.Error("read chunk", "id", id, "error", err)
			continue
		}
		catalog[id] = *meta
	}

	service := &Service{
		store: store,
		catalog: catalog,
	}
	return service, nil
}


func (svc *Service) GetServerID() string {
	return "service-id-123"
}

func (svc *Service) StartUploadSession(desc *t.ChunkDesc) (s.ChunkWriter, error) {
	svc.mu.RLock()
	_, ok := svc.catalog[desc.ID]
	svc.mu.RUnlock()

	if ok {
		return nil, s.ErrChunkConflict
	}
	w, err := svc.store.NewWriter()
	if err != nil {
		return nil, fmt.Errorf("create chunk writer: %w", err)
	}
	return w, nil
}

func (svc *Service) CommitUploadSession(w s.ChunkWriter, desc *t.ChunkDesc) error {
	if !desc.Digest.Equal(w.Digest()) {
		return fmt.Errorf("session validation: digest missmatch") 
	}

	svc.mu.Lock()	
	defer svc.mu.Unlock()

	if _, ok := svc.catalog[desc.ID]; ok {
		return s.ErrChunkConflict
	}

	ts, err := w.Commit(desc.ID)
	if err != nil {
		return fmt.Errorf("session commit: %w", err)
	} 
	meta := &t.ChunkMeta{
		ChunkDesc: *desc,
		ModifiedAt: ts, 
	}
	svc.catalog[meta.ID] = *meta
	return nil
}

func (svc *Service) GetChunk(chunkID t.ChunkID) (*s.Chunk, error) {
	svc.mu.RLock()
	meta, ok := svc.catalog[chunkID]
	svc.mu.RUnlock()

	if !ok {
		return nil, s.ErrChunkNotFound
	}
	reader, err := svc.store.Get(chunkID)
	if err != nil {
		return nil, fmt.Errorf("get from store: %w", err)
	}
	chunk := &s.Chunk{
		ChunkMeta: meta,
		Data: reader,
	}
	return chunk, nil
}

