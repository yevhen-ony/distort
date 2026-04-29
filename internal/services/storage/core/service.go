package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	t "dos/internal/common/types"
	s "dos/internal/services/storage"
)

type StorageServiceConfig struct{
	AdvertiseAddr string `yaml:"advertise_addr"`
	MaxStorageBytes int64 `yaml:"max_storage_bytes"`
}

type Service struct {
	catalog s.ChunkCatalog
	totalBytes int64
	mu      sync.RWMutex

	store   s.ChunkStorage
	master s.MasterTransport
	
	config StorageServiceConfig
	nodeID t.NodeID

}

func New(store s.ChunkStorage, master s.MasterTransport, config StorageServiceConfig) (*Service, error) {
	catalog := map[t.ChunkID]t.ChunkMeta{}

	if store == nil {
		return nil, errors.New("missing store") 
	}
	if master == nil {
		return nil, errors.New("missing master transport")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()

	nodeID, err := master.RegisterStorageNode(ctx, config.AdvertiseAddr)
	if err != nil {
		return nil, fmt.Errorf("register node: %w", err)
	}

	ids, err := store.GetAllIDs()
	if err != nil {
		return nil, fmt.Errorf("get all ids: %w", err)
	}

	var totalBytes int64
	for _, id := range ids {
		meta, err := store.GetMeta(id)
		if err != nil {
			slog.Error("read chunk", "id", id, "error", err)
			continue
		}
		catalog[id] = *meta
		totalBytes += meta.Digest.Size
	}

	service := &Service{
		catalog: catalog,
		totalBytes: totalBytes,

		store: store,
		master: master,

		config: config,
		nodeID: nodeID,
	}
	return service, nil
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

func (svc *Service) Heartbeat(ctx context.Context) error {
	svc.mu.RLock()
	stats := t.NodeStats{
		FreeBytes: svc.config.MaxStorageBytes - svc.totalBytes,	
		UsedBytes: svc.totalBytes,
		ChunkCount: len(svc.catalog),
	}
	svc.mu.RUnlock()

	res, err := svc.master.Heartbeat(ctx, svc.nodeID, stats)
	if err != nil {
		return err
	}

	if res.NodeUnknown {
		slog.Warn("request new node id")
		if err := svc.Register(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (svc *Service) Register(ctx context.Context) error {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	nodeID, err := svc.master.RegisterStorageNode(ctx, svc.config.AdvertiseAddr)
	if err != nil {
		return fmt.Errorf("register storage node: %w", err) 
	}
	svc.nodeID = nodeID
	return nil
}

func (svc *Service) ValidateNodeID(nodeID t.NodeID) error {
	if nodeID != svc.nodeID {
		return s.ErrInvalidNodeID
	}
	return nil
}


