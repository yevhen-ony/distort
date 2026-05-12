package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"dos/internal/common/dosctx"
	"dos/internal/common/transport/chunkrpc"
	t "dos/internal/common/types"
	s "dos/internal/services/storage"
)

type StorageServiceConfig interface {
	AdvertiseAddr() string
	MaxStorage() int64
	HeartbeatInterval() time.Duration
	ReportInterval() time.Duration
	ReplicationTimeout() time.Duration
}

type ReportSink interface {
	Enqueue(context.Context, t.StorageNodeReport)
	Flush(context.Context)
}

type StorageService struct {
	catalog    s.ChunkCatalog
	totalBytes int64
	mu         sync.RWMutex

	diskStore       s.ChunkStorage
	masterTransport s.MasterTransport
	chunkTransport  *chunkrpc.Transport

	identity   *IdentityService
	reportSink ReportSink

	config StorageServiceConfig
}

func NewStorageService(
	diskStore s.ChunkStorage,
	masterTransport s.MasterTransport,
	chunkTransport *chunkrpc.Transport,
	identity *IdentityService,
	reportSink ReportSink,
	config StorageServiceConfig,
) (*StorageService, error) {

	if diskStore == nil {
		return nil, errors.New("missing store")
	}
	if masterTransport == nil {
		return nil, errors.New("missing master transport")
	}
	if chunkTransport == nil {
		return nil, errors.New("missing storage transport")
	}
	if identity == nil {
		return nil, errors.New("missing identity service")
	}

	service := &StorageService{
		diskStore:       diskStore,
		masterTransport: masterTransport,
		chunkTransport:  chunkTransport,
		identity:        identity,
		reportSink:      reportSink,
		config:          config,
	}
	if err := service.buildCatalog(); err != nil {
		return nil, fmt.Errorf("build catalog: %w", err)
	}
	return service, nil
}

func (svc *StorageService) StartUploadSession(desc *t.ChunkMeta) (s.ChunkWriter, error) {
	svc.mu.RLock()
	_, ok := svc.catalog[desc.ID]
	svc.mu.RUnlock()

	if ok {
		return nil, s.ErrChunkConflict
	}
	w, err := svc.diskStore.NewWriter()
	if err != nil {
		return nil, fmt.Errorf("create chunk writer: %w", err)
	}
	return w, nil
}

func (svc *StorageService) CommitUploadSession(
	ctx context.Context, w s.ChunkWriter, meta *t.ChunkMeta,
) error {

	if err := meta.Digest.Match(w.Digest()); err != nil {
		return err
	}

	svc.mu.Lock()
	defer svc.mu.Unlock()

	if _, ok := svc.catalog[meta.ID]; ok {
		return s.ErrChunkConflict
	}

	if err := w.Commit(meta.ID); err != nil {
		return fmt.Errorf("session commit: %w", err)
	}
	svc.catalog[meta.ID] = s.NewChunkRecord(*meta)
	svc.reportSink.Enqueue(ctx, t.NewReplicaStaged(*meta).ToRecord())
	return nil
}

func (svc *StorageService) GetChunk(chunkID t.ChunkID) (t.Chunk, error) {
	svc.mu.RLock()
	state, ok := svc.catalog[chunkID]
	svc.mu.RUnlock()

	if !ok {
		return t.Chunk{}, s.ErrChunkNotFound
	}
	reader, err := svc.diskStore.Get(chunkID)
	if err != nil {
		return t.Chunk{}, fmt.Errorf("get from store: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return t.Chunk{}, fmt.Errorf("read chunk: %w", err)
	}
	chunk := t.Chunk{
		Meta: state.Meta,
		Data: data,
	}
	return chunk, nil
}

func (svc *StorageService) ReplicateChunk(
	ctx context.Context, chunkID t.ChunkID, targets []t.NodeRef,
) error {

	ctx = dosctx.WithChunkID(ctx, chunkID)
	ctx, cancel := context.WithTimeout(ctx, svc.config.ReplicationTimeout())
	defer cancel()

	chunk, err := svc.GetChunk(chunkID)
	if err != nil {
		return fmt.Errorf("get chunk: %w", err)
	}

	session := svc.chunkTransport.NewTransferSession(targets)
	if _, err = session.Upload(ctx, &chunk); err != nil {
		slog.ErrorContext(ctx, "chank replication failed", "targets", targets, "error", err)

		report := t.NewReplicaChainFailed(chunkID, targets).ToRecord()
		svc.reportSink.Enqueue(ctx, report)
		svc.reportSink.Flush(ctx)

		return fmt.Errorf("upload replica: %w", err)
	}

	return nil
}

func (svc *StorageService) DeleteChunk(ctx context.Context, chunkID t.ChunkID) error {

	svc.mu.RLock()
	_, ok := svc.catalog[chunkID]
	svc.mu.RUnlock()

	if !ok {
		slog.WarnContext(ctx, "delete non-existing chunk", "chunk_id", chunkID)
		return nil
	}

	svc.mu.Lock()
	delete(svc.catalog, chunkID)
	svc.mu.Unlock()

	if err := svc.diskStore.Delete(chunkID); err != nil {
		return fmt.Errorf("delete data from disk: %w", err)
	}
	return nil
}

func (svc *StorageService) Heartbeat(ctx context.Context) error {
	ctx = dosctx.WithOperation(ctx, "heartbeat")

	svc.mu.RLock()
	stats := t.NodeStats{
		FreeBytes:  svc.config.MaxStorage() - svc.totalBytes,
		UsedBytes:  svc.totalBytes,
		ChunkCount: len(svc.catalog),
	}
	svc.mu.RUnlock()

	nodeID, err := svc.identity.GetID()
	if err != nil {
		return fmt.Errorf("read node id: %w", err)
	}

	res, err := svc.masterTransport.Heartbeat(ctx, nodeID, stats)
	if err != nil {
		return err
	}

	if res.NodeUnknown {
		slog.Warn("request new node id")
		if err := svc.identity.RequestNewID(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (svc *StorageService) RunHearbeatLoop(ctx context.Context) {
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {

		slog.DebugContext(ctx, "exec heartbeat")
		if err := svc.Heartbeat(ctx); err != nil {
			slog.ErrorContext(ctx, "heartbeat failed", "error", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}

		timer.Reset(jitter(svc.config.HeartbeatInterval(), 0.2))
	}
}

func (svc *StorageService) buildCatalog() error {

	ids, err := svc.diskStore.List()
	if err != nil {
		return fmt.Errorf("list chunks: %w", err)
	}

	catalog := make(map[t.ChunkID]*s.ChunkRecord, len(ids))
	var totalBytes int64
	for _, id := range ids {
		meta, err := svc.diskStore.GetMeta(id)
		if err != nil {
			slog.Error("read chunk", "id", id, "error", err)
			continue
		}
		catalog[id] = &s.ChunkRecord{Meta: meta, State: s.ChunkStateStaged}
		totalBytes += meta.Digest.Size
	}

	svc.mu.Lock()
	defer svc.mu.Unlock()

	svc.catalog = catalog
	svc.totalBytes = totalBytes
	return nil
}
