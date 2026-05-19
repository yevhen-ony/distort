package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"dos/internal/common/dosctx"
	"dos/internal/common/loop"
	"dos/internal/common/transport/chunkrpc"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
	s "dos/internal/services/storage"
	"dos/internal/services/storage/core/identity"
)

type StorageServiceConfig interface {
	AdvertiseAddr() string
	MaxStorage() int64
	HeartbeatInterval() time.Duration
	ReportInterval() time.Duration
	ReplicationTimeout() time.Duration
	MaxParallelHeavyOps() int
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
	identity        *identity.IdentityService
	reporter        Reporter
	config          StorageServiceConfig

	looper *loop.Looper
	sem    chan struct{}

	metrics *StorageServiceMetrics
}

func NewStorageService(
	diskStore s.ChunkStorage,
	masterTransport s.MasterTransport,
	chunkTransport *chunkrpc.Transport,
	identity *identity.IdentityService,
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

		looper: loop.NewLooper(config.HeartbeatInterval()),
		sem:    make(chan struct{}, config.MaxParallelHeavyOps()),
		
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

func (svc *StorageService) AcquireOpSlot(ctx context.Context) (func(), error) {
	acqCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	start := time.Now()

	select {
	case <-acqCtx.Done():
		svc.metrics.OpSlotsAcquireDuration.Observe(time.Since(start).Seconds())
		return nil, s.ErrServiceBusy

	case svc.sem <- struct{}{}:
		svc.metrics.OpSlotsAcquireDuration.Observe(time.Since(start).Seconds())
		svc.metrics.OpSlotsInUse.Add(1)
		start = time.Now()
		release := func() { 
			<-svc.sem
			svc.metrics.OpSlotsInUse.Add(-1)
			svc.metrics.OpSlotsHoldDuration.Observe(time.Since(start).Seconds())
		}
		return release, nil
	}
}

func (svc *StorageService) StartUpload(ctx context.Context, meta *t.ChunkMeta) (*UploadSession, error) {
	svc.state.Mu.RLock()
	_, ok := svc.state.Catalog[meta.ID]
	svc.state.Mu.RUnlock()

	if ok {
		return nil, s.ErrChunkConflict
	}

	release, err := svc.AcquireOpSlot(ctx) 
	if err != nil {
		return nil, err
	}

	start := time.Now()

	session := &UploadSession{
		id: meta.ID, 
		data: make([]byte, meta.Digest.Size),
		onCommit: func(ctx context.Context, chunk t.Chunk) error {
			defer release()
			err := svc.commitUpload(ctx, chunk, meta)
			if err != nil {
				svc.metrics.UploadsFailedDuration.Observe(time.Since(start).Seconds())
			} else {
				svc.metrics.UploadsSuccessDuration.Observe(time.Since(start).Seconds())
			}
			return err
		},
		onAbort: func() error {
			defer release() 
			svc.metrics.UploadsFailedDuration.Observe(time.Since(start).Seconds())
			return nil
		},

	}
	return session, nil
}

func (svc *StorageService) commitUpload(
	ctx context.Context, chunk t.Chunk, meta *t.ChunkMeta,
) error {
	ctx = dosctx.WithOperation(ctx, "commit_upload")

	if err := meta.Digest.Match(chunk.Meta.Digest); err != nil {
		return err
	}

	if err := svc.diskStore.Store(chunk); err != nil {
		return fmt.Errorf("store chunk: %w", err)
	}

	if err := svc.AddToCatalog(meta); err != nil {
		if err := svc.diskStore.Delete(meta.ID); err != nil {
			slog.ErrorContext(ctx, "rollback failed", "error", err)
		}
		return err
	}

	svc.reporter.Report(ctx, t.NewReplicaStaged(*meta).ToRecord())
	return nil
}

func (svc *StorageService) AddToCatalog(meta *t.ChunkMeta) error {
	svc.state.Mu.Lock()
	defer svc.state.Mu.Unlock()

	if _, ok := svc.state.Catalog[meta.ID]; ok {
		return s.ErrChunkConflict
	}

	size := meta.Digest.Size

	svc.metrics.ChunksCount.Add(1)
	svc.metrics.ChunksTotalBytes.Add(float64(size))

	svc.state.Catalog[meta.ID] = NewChunkRecord(*meta)
	svc.state.TotalBytes += meta.Digest.Size

	return nil
}

func (svc *StorageService) RemoveFromCatalog(chunkID t.ChunkID) bool {
	svc.state.Mu.Lock()
	defer svc.state.Mu.Unlock()

	rec, ok := svc.state.Catalog[chunkID]
	if !ok {
		return false
	}
	size := rec.Meta.Digest.Size
	
	svc.metrics.ChunksCount.Add(-1)
	svc.metrics.ChunksTotalBytes.Add(float64(size))

	svc.state.TotalBytes -= size 
	delete(svc.state.Catalog, chunkID)
	return true
}

func (svc *StorageService) LoadChunk(chunkID t.ChunkID) (t.Chunk, error) {
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


func (svc *StorageService) ForwardChunk(
	ctx context.Context, chunkID t.ChunkID, targets []t.NodeRef,
) error {

	slog.DebugContext(ctx, "forward chunk")
	nodeID, err := svc.identity.GetID()
	if err != nil {
		return fmt.Errorf("access node id: %w", err)
	}

	targets = utils.Select(targets, func(r t.NodeRef) bool {
		return r.ID != nodeID
	})
	if len(targets) == 0 {
		return s.ErrNoValidTargets
	}

	slog.DebugContext(ctx, "load chunk")
	chunk, err := svc.LoadChunk(chunkID)
	if err != nil {
		return fmt.Errorf("get chunk: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, svc.config.ReplicationTimeout())
	defer cancel()

	slog.DebugContext(ctx, "send chunk")
	chosen, err := svc.SendChunk(ctx, chunk, targets)
	if err != nil {
		return fmt.Errorf("send chunk: %w", err)
	}

	targets = utils.Select(targets, func(r t.NodeRef) bool {
		return r.ID != chosen.ID
	})

	if len(targets) == 0 {
		return nil
	}

	slog.DebugContext(ctx, "handoff replicate chunk", "source", chosen.ID)
	err = svc.ReplicateChunk(ctx, chunkID, chosen, targets)
	if err != nil {
		return fmt.Errorf("replicate chunk: %w", err)
	}

	return nil
}

func (svc *StorageService) ScheduleForwardChunk(
	ctx context.Context, chunkID t.ChunkID, targets []t.NodeRef,
) error {

	svc.state.Mu.RLock()
	_, ok := svc.state.Catalog[chunkID]
	svc.state.Mu.RUnlock()
	if !ok {
		return s.ErrChunkNotFound
	}

	release, err := svc.AcquireOpSlot(ctx)
	if err != nil {
		return err
	}

	fwdCtx := context.WithoutCancel(ctx)
	go func() {
		defer release()
		_ = svc.ForwardChunk(fwdCtx, chunkID, targets)
	}()

	return nil
}

func (svc *StorageService) SendChunk(
	ctx context.Context, chunk t.Chunk, targets []t.NodeRef,
) (t.NodeRef, error) {

	start := time.Now()

	session := svc.chunkTransport.NewTransferSession(targets)
	chosen, err := session.Upload(ctx, &chunk)
	if err != nil {
		svc.metrics.SendsFailedDuration.Observe(time.Since(start).Seconds())
		slog.ErrorContext(ctx, "push chunk failed", "targets", targets, "error", err)

		svc.reporter.Report(ctx, t.NewReplicaChainFailed(chunk.Meta.ID, targets).ToRecord())
		svc.reporter.Flush(ctx)

		return t.NodeRef{}, fmt.Errorf("upload replica: %w", err)
	}

	svc.metrics.SendsSuccessDuration.Observe(time.Since(start).Seconds())
	return chosen, nil
}

func (svc *StorageService) ReplicateChunk(
	ctx context.Context, chunkID t.ChunkID, source t.NodeRef, targets []t.NodeRef,
) error {
	
	start := time.Now()

	err := svc.chunkTransport.ReplicateChunk(ctx, chunkID, source, targets)
	if err != nil {
		svc.metrics.ReplicateFailedDuration.Observe(time.Since(start).Seconds())

		slog.ErrorContext(ctx, "replicate chunk failed",
			"source", source, "targets", targets, "error", err)

		svc.reporter.Report(ctx, t.NewReplicaChainFailed(chunkID, targets).ToRecord())
		svc.reporter.Flush(ctx)

		return fmt.Errorf("replicate chunk: %w", err)
	}

	svc.metrics.ReplicateSuccessDuration.Observe(time.Since(start).Seconds())
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

	if svc.RemoveFromCatalog(chunkID) {
		svc.reporter.Report(ctx, t.NewReplicaDeleted(chunkID).ToRecord())
	}
	return nil
}

func (svc *StorageService) Heartbeat(ctx context.Context) {

	svc.state.Mu.RLock()
	stats := t.NodeStats{
		FreeBytes:  svc.config.MaxStorage() - svc.state.TotalBytes,
		UsedBytes:  svc.state.TotalBytes,
		ChunkCount: len(svc.state.Catalog),
	}
	svc.state.Mu.RUnlock()

	nodeID, err := svc.identity.GetID()
	if err != nil {
		slog.ErrorContext(ctx, "read node id failed", "error", err)
		return
	}

	res, err := svc.masterTransport.Heartbeat(ctx, nodeID, stats)
	if err != nil {
		svc.metrics.HeartbeatFailedTotal.Inc()
		slog.ErrorContext(ctx, "heartbeat transport failed", "node_id", nodeID, "error", err)
	}

	if res.NodeUnknown {
		slog.WarnContext(ctx, "node id is unknown", "node_id", nodeID)
		if err := svc.identity.RequestNewID(ctx); err != nil {
			slog.WarnContext(ctx, "request new node id failed", "error", err)
		}
	}
}

func (svc *StorageService) RunHearbeatLoop(ctx context.Context) {
	ctx = dosctx.WithOperation(ctx, "heartbeat")
	svc.looper.SkipFirstWait().Run(ctx, svc.Heartbeat)
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
	
	svc.metrics.ChunksCount.Set(float64(len(catalog)))
	svc.metrics.ChunksTotalBytes.Set(float64(totalBytes))
	
	return nil
}

