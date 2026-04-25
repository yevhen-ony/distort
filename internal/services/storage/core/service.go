package core 

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	s "dos/internal/services/storage"
)

type Service struct {
	store   s.ChunkStorage
	catalog s.ChunkCatalog
	mu      sync.RWMutex
}

func New(store s.ChunkStorage) (*Service, error) {
	catalog := map[s.ChunkID]s.ChunkMeta{}

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

func (svc *Service) StartUploadSession(info *s.ChunkInfo) (s.ChunkWriter, error) {
	svc.mu.RLock()
	_, ok := svc.catalog[info.ID]
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

func (svc *Service) CommitUploadSession(w s.ChunkWriter, info *s.ChunkInfo) error {
	meta := s.ChunkMeta{Digest: w.Digest()} 
	if !info.Digest.Equal(&meta.Digest) {
		return fmt.Errorf("session validation: digest missmatch") 
	}

	svc.mu.Lock()	
	defer svc.mu.Unlock()

	if _, ok := svc.catalog[info.ID]; ok {
		return s.ErrChunkConflict
	}
	if ts, err := w.Commit(info.ID); err != nil {
		return fmt.Errorf("session commit: %w", err)
	} else {
		meta.ModifiedAt = ts
	}
	svc.catalog[info.ID] = meta
	return nil
}

func (svc *Service) GetChunk(chunkID s.ChunkID) (*s.Chunk, error) {
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
		ID: chunkID,
		Meta: &meta,
		Reader: reader,
	}
	return chunk, nil
}

func (svc *Service) validate(want, got s.ChunkDigest) error {
	// skip if not set
	if want.Checksum != "" && want.Checksum != got.Checksum {
		return fmt.Errorf("checksum: %w", s.ErrDigestInvalid)
	}
	if want.Size != got.Size {
		return fmt.Errorf("size: %w", s.ErrDigestInvalid)
	}
	return nil
}

