package core 

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	cs "dos/internal/services/chunkserver"
)

type Service struct {
	store   cs.ChunkStorage
	catalog cs.ChunkCatalog
	mu      sync.RWMutex
}

func New(store cs.ChunkStorage) (*Service, error) {
	catalog := map[cs.ChunkID]cs.ChunkMeta{}

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


func (s *Service) GetServerID() string {
	return "service-id-123"
}

func (s *Service) StartUploadSession(info *cs.ChunkInfo) (cs.ChunkWriter, error) {
	s.mu.RLock()
	_, ok := s.catalog[info.ID]
	s.mu.RUnlock()

	if ok {
		return nil, ErrConflict
	}
	w, err := s.store.NewWriter()
	if err != nil {
		return nil, fmt.Errorf("create chunk writer: %w", err)
	}
	return w, nil
}

func (s *Service) CommitUploadSession(w cs.ChunkWriter, info *cs.ChunkInfo) error {
	meta := cs.ChunkMeta{Digest: w.Digest()} 
	if !info.Digest.Equal(meta.Digest) {
		return fmt.Errorf("session validation: digest missmatch") 
	}

	s.mu.Lock()	
	defer s.mu.Unlock()

	if _, ok := s.catalog[info.ID]; ok {
		return ErrConflict
	}
	if ts, err := w.Commit(info.ID); err != nil {
		return fmt.Errorf("session commit: %w", err)
	} else {
		meta.ModifiedAt = ts
	}
	s.catalog[info.ID] = meta
	return nil
}

func (s *Service) GetChunk(chunkID cs.ChunkID) (*cs.Chunk, error) {
	s.mu.RLock()
	meta, ok := s.catalog[chunkID]
	s.mu.RUnlock()

	if !ok {
		return nil, ErrNotFound
	}
	reader, err := s.store.Get(chunkID)
	if err != nil {
		return nil, fmt.Errorf("get from store: %w", err)
	}
	chunk := &cs.Chunk{
		ID: chunkID,
		Meta: &meta,
		Reader: reader,
	}
	return chunk, nil
}

func (s *Service) validate(want, got cs.ChunkDigest) error {
	// skip if not set
	if want.Checksum != "" && want.Checksum != got.Checksum {
		return fmt.Errorf("checksum: %w", ErrInvalid)
	}
	if want.Size != got.Size {
		return fmt.Errorf("size: %w", ErrInvalid)
	}
	return nil
}

