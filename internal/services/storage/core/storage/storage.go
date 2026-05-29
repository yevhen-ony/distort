package storage

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
	"dos/internal/services/storage/core/identity"
)

type StorageConfig interface {
	AdvertiseAddr() string
	ReportInterval() time.Duration
	ReplicationTimeout() time.Duration
	MaxParallelHeavyOps() int
}

type Reporter interface {
	Report(context.Context, t.StorageNodeReport)
	Flush(context.Context)
}

type StorageDeps struct {
	Catalog   *ChunkInventory
	Identity  *identity.IdentityService
	Reporter  Reporter
	StorageBE s.ChunkStorage
	MasterT   s.MasterTransport
	ChunkT    *chunkrpc.Transport
	Config    StorageConfig
	Metrics   *StorageMetrics
}

type StorageService struct {
	inventory  *ChunkInventory
	identity *identity.IdentityService
	reporter Reporter

	storageBE s.ChunkStorage
	masterT   s.MasterTransport
	chunkT    *chunkrpc.Transport
	config    StorageConfig

	sem     chan struct{}
	metrics *StorageMetrics
}

func NewStorageService(deps StorageDeps) (*StorageService, error) {
	if deps.Catalog == nil {
		return nil, errors.New("missing catalog service")
	}
	if deps.Identity == nil {
		return nil, errors.New("missing identity service")
	}
	if deps.Reporter == nil {
		return nil, errors.New("missing reporter")
	}

	if deps.StorageBE == nil {
		return nil, errors.New("missing store")
	}
	if deps.MasterT == nil {
		return nil, errors.New("missing master transport")
	}
	if deps.ChunkT == nil {
		return nil, errors.New("missing storage transport")
	}
	if deps.Config == nil {
		return nil, errors.New("missing config")
	}

	service := &StorageService{
		inventory:  deps.Catalog,
		reporter: deps.Reporter,
		identity: deps.Identity,

		storageBE: deps.StorageBE,
		masterT:   deps.MasterT,
		chunkT:    deps.ChunkT,
		config:    deps.Config,
		metrics:   deps.Metrics,

		sem: make(chan struct{}, deps.Config.MaxParallelHeavyOps()),
	}
	return service, nil
}

func (cs *StorageService) Start(ctx context.Context) error {

	if err := cs.inventory.BuildCatalog(ctx, cs.storageBE); err != nil {
		return fmt.Errorf("build catalog: %w", err)
	}

	metas := cs.inventory.ListStaged()
	for _, meta := range metas {
		cs.reporter.Report(ctx, t.NewReplicaStaged(meta).ToRecord())
	}

	cs.reporter.Flush(ctx)
	slog.DebugContext(ctx, "catalog built", "chunks", len(metas))
	return nil
}

func (cs *StorageService) AcquireOpSlot(ctx context.Context) (func(), error) {
	acqCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	start := time.Now()

	select {
	case <-acqCtx.Done():
		cs.metrics.OpSlotsAcquireDuration.Observe(time.Since(start).Seconds())
		return nil, s.ErrServiceBusy

	case cs.sem <- struct{}{}:
		cs.metrics.OpSlotsAcquireDuration.Observe(time.Since(start).Seconds())
		cs.metrics.OpSlotsInUse.Add(1)
		start = time.Now()
		release := func() {
			<-cs.sem
			cs.metrics.OpSlotsInUse.Add(-1)
			cs.metrics.OpSlotsHoldDuration.Observe(time.Since(start).Seconds())
		}
		return release, nil
	}
}

func (cs *StorageService) StartUpload(ctx context.Context, meta *t.ChunkMeta) (*UploadSession, error) {
	if cs.inventory.Has(meta.ID) {
		return nil, s.ErrChunkConflict
	}

	release, err := cs.AcquireOpSlot(ctx)
	if err != nil {
		return nil, err
	}

	start := time.Now()

	session := &UploadSession{
		id:   meta.ID,
		data: make([]byte, meta.Digest.Size),
		onCommit: func(ctx context.Context, chunk t.Chunk) error {
			defer release()
			err := cs.commitUpload(ctx, chunk, meta)
			if err != nil {
				cs.metrics.UploadsFailedDuration.Observe(time.Since(start).Seconds())
			} else {
				cs.metrics.UploadsSuccessDuration.Observe(time.Since(start).Seconds())
			}
			return err
		},
		onAbort: func() error {
			defer release()
			cs.metrics.UploadsFailedDuration.Observe(time.Since(start).Seconds())
			return nil
		},
	}
	return session, nil
}

func (cs *StorageService) commitUpload(
	ctx context.Context, chunk t.Chunk, meta *t.ChunkMeta,
) error {

	ctx = dosctx.WithOperation(ctx, "commit_upload")

	if err := meta.Digest.Match(&chunk.Meta.Digest); err != nil {
		return err
	}

	if err := cs.storageBE.Store(chunk); err != nil {
		return fmt.Errorf("store chunk: %w", err)
	}

	if err := cs.inventory.Add(meta); err != nil {
		if err := cs.storageBE.Delete(meta.ID); err != nil {
			slog.ErrorContext(ctx, "rollback failed", "error", err)
		}
		return err
	}

	cs.reporter.Report(ctx, t.NewReplicaStaged(*meta).ToRecord())
	return nil
}

func (cs *StorageService) LoadChunk(chunkID t.ChunkID) (t.Chunk, error) {

	meta, err := cs.inventory.Get(chunkID)
	if err != nil {
		return t.Chunk{}, err
	}
	reader, err := cs.storageBE.Get(chunkID)
	if err != nil {
		return t.Chunk{}, fmt.Errorf("get from store: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return t.Chunk{}, fmt.Errorf("read chunk: %w", err)
	}
	chunk := t.Chunk{
		Meta: meta,
		Data: data,
	}
	return chunk, nil
}

func (cs *StorageService) ForwardChunk(
	ctx context.Context, chunkID t.ChunkID, targets []t.NodeRef,
) error {

	slog.DebugContext(ctx, "forward chunk")
	nodeID, err := cs.identity.GetID()
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
	chunk, err := cs.LoadChunk(chunkID)
	if err != nil {
		return fmt.Errorf("get chunk: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, cs.config.ReplicationTimeout())
	defer cancel()

	slog.DebugContext(ctx, "send chunk")
	chosen, err := cs.SendChunk(ctx, chunk, targets)
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
	err = cs.ReplicateChunk(ctx, chunkID, chosen, targets)
	if err != nil {
		return fmt.Errorf("replicate chunk: %w", err)
	}

	return nil
}

func (cs *StorageService) ScheduleForwardChunk(
	ctx context.Context, chunkID t.ChunkID, targets []t.NodeRef,
) error {

	if !cs.inventory.Has(chunkID) {
		return s.ErrChunkNotFound
	}

	release, err := cs.AcquireOpSlot(ctx)
	if err != nil {
		return err
	}

	fwdCtx := context.WithoutCancel(ctx)
	go func() {
		defer release()
		_ = cs.ForwardChunk(fwdCtx, chunkID, targets)
	}()

	return nil
}

func (cs *StorageService) SendChunk(
	ctx context.Context, chunk t.Chunk, targets []t.NodeRef,
) (t.NodeRef, error) {

	start := time.Now()

	session := cs.chunkT.NewTransferSession(targets)
	chosen, err := session.Upload(ctx, &chunk)
	if err != nil {
		cs.metrics.SendsFailedDuration.Observe(time.Since(start).Seconds())
		slog.ErrorContext(ctx, "push chunk failed", "targets", targets, "error", err)

		cs.reporter.Report(ctx, t.NewReplicaChainFailed(chunk.Meta.ID, targets).ToRecord())
		cs.reporter.Flush(ctx)

		return t.NodeRef{}, fmt.Errorf("upload replica: %w", err)
	}

	cs.metrics.SendsSuccessDuration.Observe(time.Since(start).Seconds())
	return chosen, nil
}

func (cs *StorageService) ReplicateChunk(
	ctx context.Context, chunkID t.ChunkID, source t.NodeRef, targets []t.NodeRef,
) error {

	start := time.Now()

	err := cs.chunkT.ReplicateChunk(ctx, chunkID, source, targets)
	if err != nil {
		cs.metrics.ReplicateFailedDuration.Observe(time.Since(start).Seconds())

		slog.ErrorContext(ctx, "replicate chunk failed",
			"source", source, "targets", targets, "error", err)

		cs.reporter.Report(ctx, t.NewReplicaChainFailed(chunkID, targets).ToRecord())
		cs.reporter.Flush(ctx)

		return fmt.Errorf("replicate chunk: %w", err)
	}

	cs.metrics.ReplicateSuccessDuration.Observe(time.Since(start).Seconds())
	return nil
}

func (cs *StorageService) DeleteChunk(ctx context.Context, chunkID t.ChunkID) error {

	ctx = dosctx.WithChunkID(ctx, chunkID)
	ctx = dosctx.WithOperation(ctx, "delete")

	if !cs.inventory.Has(chunkID) {
		slog.WarnContext(ctx, "delete non-existing chunk")
		return nil
	}

	if err := cs.storageBE.Delete(chunkID); err != nil {
		return fmt.Errorf("delete data from disk: %w", err)
	}

	if cs.inventory.Remove(chunkID) {
		cs.reporter.Report(ctx, t.NewReplicaDeleted(chunkID).ToRecord())
	}
	return nil
}

func (cs *StorageService) RestageCatalog(ctx context.Context) {
	slog.InfoContext(ctx, "restage catalog")
	cs.inventory.RestageActive()
	for _, meta := range cs.inventory.ListStaged() {
		cs.reporter.Report(ctx, t.NewReplicaStaged(meta).ToRecord())
	}
	cs.reporter.Flush(ctx)
}
