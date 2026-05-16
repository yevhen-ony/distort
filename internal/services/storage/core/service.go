package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"dos/internal/common/dosctx"
	"dos/internal/common/transport/chunkrpc"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
	s "dos/internal/services/storage"
)

type StorageServiceConfig interface {
	AdvertiseAddr() string
	MaxStorage() int64
	HeartbeatInterval() time.Duration
	ReportInterval() time.Duration
	ReplicationTimeout() time.Duration
}

type Reporter interface {
	Report(context.Context, t.StorageNodeReport)
	Flush(context.Context)
}

type NOOPReporter struct{}

func (*NOOPReporter) Report(context.Context, t.StorageNodeReport) {}
func (*NOOPReporter) Flush(context.Context)                       {}

type StorageService struct {
	state ChunkCatalogState

	diskStore       s.ChunkStorage
	masterTransport s.MasterTransport
	chunkTransport  *chunkrpc.Transport

	identity *IdentityService
	reporter Reporter

	config StorageServiceConfig
}

func NewStorageService(
	diskStore s.ChunkStorage,
	masterTransport s.MasterTransport,
	chunkTransport *chunkrpc.Transport,
	identity *IdentityService,
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
		reporter:        &NOOPReporter{},
		config:          config,
	}
	return service, nil
}

func (svc *StorageService) SetReporter(r Reporter) {
	svc.reporter = r
}

func (svc *StorageService) Start(ctx context.Context) error {
	if err := svc.buildCatalog(context.Background()); err != nil {
		return fmt.Errorf("build catalog: %w", err)
	}
	slog.Debug("catalog built", "chunks", len(svc.state.Catalog))
	return nil
}

func (svc *StorageService) StartUpload(_ context.Context, meta *t.ChunkMeta) (*ChunkBuilder, error) {
	svc.state.Mu.RLock()
	_, ok := svc.state.Catalog[meta.ID]
	svc.state.Mu.RUnlock()

	if ok {
		return nil, s.ErrChunkConflict
	}
	builder := NewChunkBuilder(meta.ID, meta.Digest.Size)
	return builder, nil
}

func (svc *StorageService) CommitUpload(
	ctx context.Context, chunk t.Chunk, meta *t.ChunkMeta,
) error {
	ctx = dosctx.WithOperation(ctx, "commit upload")

	if err := meta.Digest.Match(chunk.Meta.Digest); err != nil {
		return err
	}

	if err := svc.diskStore.Store(chunk); err != nil {
		return fmt.Errorf("store chunk: %w", err)
	}

	svc.state.Mu.Lock()
	defer svc.state.Mu.Unlock()

	if _, ok := svc.state.Catalog[meta.ID]; ok {
		if err := svc.diskStore.Delete(meta.ID); err != nil {
			slog.ErrorContext(ctx, "rollback failed after catalog conflict", "error", err)
		}
		return s.ErrChunkConflict
	}

	svc.state.Catalog[meta.ID] = NewChunkRecord(*meta)
	svc.state.TotalBytes += meta.Digest.Size
	svc.reporter.Report(ctx, t.NewReplicaStaged(*meta).ToRecord())
	return nil
}

func (svc *StorageService) GetChunk(chunkID t.ChunkID) (t.Chunk, error) {
	svc.state.Mu.RLock()
	state, ok := svc.state.Catalog[chunkID]
	svc.state.Mu.RUnlock()

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

	chunk, err := svc.GetChunk(chunkID)
	if err != nil {
		return fmt.Errorf("get chunk: %w", err)
	}
	return svc.SendChunk(ctx, chunk, targets)
}

func (svc *StorageService) SendChunk( 
	ctx context.Context, chunk t.Chunk, targets []t.NodeRef,
) error {
	ctx, cancel := context.WithTimeout(ctx, svc.config.ReplicationTimeout())
	defer cancel()

	targets = utils.Select(targets, func(r t.NodeRef) bool {
		return r.ID != svc.identity.nodeID
	})
	if len(targets) == 0 {
		return s.ErrNoValidTargets 
	}

	session := svc.chunkTransport.NewTransferSession(targets)
	if _, err := session.Upload(ctx, &chunk); err != nil {
		slog.ErrorContext(ctx, "chunk replication failed", "targets", targets, "error", err)

		svc.reporter.Report(ctx, t.NewReplicaChainFailed(chunk.Meta.ID, targets).ToRecord())
		svc.reporter.Flush(ctx)

		return fmt.Errorf("upload replica: %w", err)
	}

	return nil
}

func (svc *StorageService) DeleteChunk(ctx context.Context, chunkID t.ChunkID) error {

	ctx = dosctx.WithChunkID(ctx, chunkID)
	ctx = dosctx.WithOperation(ctx, "delete")

	svc.state.Mu.Lock()
	_, ok := svc.state.Catalog[chunkID]
	svc.state.Mu.Unlock()

	if !ok {
		slog.WarnContext(ctx, "delete non-existing chunk")
		return nil
	}
	
	if err := svc.diskStore.Delete(chunkID); err != nil {
		return fmt.Errorf("delete data from disk: %w", err)
	}

	svc.state.Mu.Lock()
	if chunk, ok := svc.state.Catalog[chunkID]; ok {
		delete(svc.state.Catalog, chunkID)
		svc.state.TotalBytes -= chunk.Meta.Digest.Size
	}
	svc.state.Mu.Unlock()

	svc.reporter.Report(ctx, t.NewReplicaDeleted(chunkID).ToRecord())

	return nil
}

func (svc *StorageService) Heartbeat(ctx context.Context) error {
	ctx = dosctx.WithOperation(ctx, "heartbeat")

	svc.state.Mu.RLock()
	stats := t.NodeStats{
		FreeBytes:  svc.config.MaxStorage() - svc.state.TotalBytes,
		UsedBytes:  svc.state.TotalBytes,
		ChunkCount: len(svc.state.Catalog),
	}
	svc.state.Mu.RUnlock()

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
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}

		slog.DebugContext(ctx, "exec heartbeat")
		if err := svc.Heartbeat(ctx); err != nil {
			slog.ErrorContext(ctx, "heartbeat failed", "error", err)
		}

		interval := utils.Jitter(svc.config.HeartbeatInterval(), 0.2)
		timer.Reset(interval)
	}
}

func (svc *StorageService) buildCatalog(ctx context.Context) error {

	ids, err := svc.diskStore.List()
	if err != nil {
		return fmt.Errorf("list chunks: %w", err)
	}

	catalog := make(ChunkCatalog, len(ids))
	var totalBytes int64
	for _, id := range ids {
		meta, err := svc.diskStore.GetMeta(id)
		if err != nil {
			slog.Error("read chunk", "id", id, "error", err)
			continue
		}
		catalog[id] = NewChunkRecord(meta) 
		totalBytes += meta.Digest.Size
		svc.reporter.Report(ctx, t.NewReplicaStaged(meta).ToRecord())
	}
	svc.reporter.Flush(ctx)

	svc.state.Mu.Lock()
	defer svc.state.Mu.Unlock()

	svc.state.Catalog = catalog
	svc.state.TotalBytes = totalBytes
	return nil
}

type ChunkBuilder struct {
	id t.ChunkID
	data []byte	
	n int
}

func NewChunkBuilder(chunkID t.ChunkID, size int64) *ChunkBuilder {
	return &ChunkBuilder{
		id: chunkID,
		data: make([]byte, size),
	}
}

func (b *ChunkBuilder) Write(p []byte) (int, error) {
  	if b.n+len(p) > len(b.data) {
  		return 0, io.ErrShortBuffer
  	}

  	n := copy(b.data[b.n:], p)
	if n != len(p) {
		return 0, io.ErrShortWrite
	}

  	b.n += n
  	return n, nil
}

func (b *ChunkBuilder) Chunk() t.Chunk {
	return t.NewChunk(b.id, b.data[:b.n])
}

