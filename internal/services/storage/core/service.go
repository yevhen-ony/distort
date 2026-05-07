package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"sync"
	"time"

	"dos/internal/common/retry"
	"dos/internal/common/transport/chunkrpc"
	t "dos/internal/common/types"
	s "dos/internal/services/storage"
)

type StorageServiceConfig struct {
	AdvertiseAddr   string        `yaml:"advertise_addr"`
	MaxStorageBytes int64         `yaml:"max_storage_bytes"`
	HeartbeatDelay  time.Duration `yaml:"heartbeat_delay"`
	ReportDelay     time.Duration `yaml:"report_delay"`
}

type Service struct {
	catalog    s.ChunkCatalog
	totalBytes int64
	mu         sync.RWMutex

	diskStore  s.ChunkStorage
	masterTransport s.MasterTransport
	chunkTransport *chunkrpc.Transport

	config StorageServiceConfig
	nodeID t.NodeID

	started    bool

	reportWake chan struct{}
}

func New(
	diskStore s.ChunkStorage,
	masterTransport s.MasterTransport,
	chunkTransport *chunkrpc.Transport,
	config StorageServiceConfig,
) (*Service, error) {

	if diskStore == nil {
		return nil, errors.New("missing store")
	}
	if masterTransport == nil {
		return nil, errors.New("missing master transport")
	}
	if chunkTransport == nil {
		return nil, errors.New("missing storage transport")
	}

	service := &Service{
		diskStore:  diskStore,
		masterTransport: masterTransport,
		chunkTransport: chunkTransport,
		config: config,
		reportWake: make(chan struct{}, 1),
	}
	return service, nil
}

func (svc *Service) StartUploadSession(desc *t.ChunkMeta) (s.ChunkWriter, error) {
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

func (svc *Service) CommitUploadSession(w s.ChunkWriter, meta *t.ChunkMeta) error {
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
	svc.catalog[meta.ID] = &s.ChunkState{ChunkMeta: *meta}
	return nil
}

func (svc *Service) GetChunk(chunkID t.ChunkID) (t.Chunk, error) {
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
	data, err := io.ReadAll(reader)
	if err != nil {
		return t.Chunk{}, fmt.Errorf("read chunk: %w", err)
	}
	chunk := t.Chunk{
		Meta: state.ChunkMeta,
		Data: data,
	}
	return chunk, nil
}

// func (svc *Service) ReplicateChunk(chunkID t.ChunkID, nodeRefs []t.NodeRef) error {
// 	chunk, err := svc.GetChunk(chunkID)
// 	if err != nil {
// 		return err
// 	}
//
// }

func (svc *Service) Heartbeat(ctx context.Context) error {
	svc.mu.RLock()
	stats := t.NodeStats{
		FreeBytes:  svc.config.MaxStorageBytes - svc.totalBytes,
		UsedBytes:  svc.totalBytes,
		ChunkCount: len(svc.catalog),
	}
	svc.mu.RUnlock()

	res, err := svc.masterTransport.Heartbeat(ctx, svc.nodeID, stats)
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

	var nodeID t.NodeID
	err := retry.Retry{Delay: time.Second}.Run(ctx, func(ctx context.Context) error {
		var innerErr error
		nodeID, innerErr = svc.masterTransport.RegisterNode(ctx, svc.config.AdvertiseAddr)
		return innerErr
	})

	if err != nil {
		return fmt.Errorf("register storage node: %w", err)
	}
	svc.mu.Lock()
	svc.nodeID = nodeID
	svc.mu.Unlock()
	return nil
}

func (svc *Service) ValidateNodeID(nodeID t.NodeID) error {
	if svc.nodeID == "" {
		return s.ErrNodeNotRegistered
	}
	if nodeID != svc.nodeID {
		return s.ErrInvalidNodeID
	}
	return nil
}

func (svc *Service) RunHearbeatLoop(ctx context.Context) {
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

		timer.Reset(jitter(svc.config.HeartbeatDelay, 0.2))
	}
}

func (svc *Service) Start(ctx context.Context) error {

	svc.mu.Lock()
	if svc.started {
		svc.mu.Unlock()
		return nil
	}
	svc.started = true
	svc.mu.Unlock()

	if err := svc.Register(ctx); err != nil {
		return fmt.Errorf("register node: %w", err)
	}

	if err := svc.BuildCatalog(); err != nil {
		return fmt.Errorf("build catalog: %w", err)
	}

	go svc.RunReportLoop(ctx)
	go svc.RunHearbeatLoop(ctx)

	return nil
}

func (svc *Service) BuildCatalog() error {

	ids, err := svc.diskStore.GetAllIDs()
	if err != nil {
		return fmt.Errorf("get all ids: %w", err)
	}

	catalog := make(map[t.ChunkID]*s.ChunkState, len(ids))
	var totalBytes int64
	for _, id := range ids {
		meta, err := svc.diskStore.GetMeta(id)
		if err != nil {
			slog.Error("read chunk", "id", id, "error", err)
			continue
		}
		catalog[id] = &s.ChunkState{ChunkMeta: meta, Reported: false}
		totalBytes += meta.Digest.Size
	}

	svc.mu.Lock()
	defer svc.mu.Unlock()

	svc.catalog = catalog
	svc.totalBytes = totalBytes
	return nil
}

func (svc *Service) ReportChunks(ctx context.Context) error {
	toReport := []t.ChunkMeta{}
	svc.mu.RLock()
	for _, s := range svc.catalog {
		if !s.Reported {
			toReport = append(toReport, *s.ChunkMeta.Clone())
		}
	}
	nodeID := svc.nodeID
	svc.mu.RUnlock()
	if nodeID == "" {
		return s.ErrNodeNotRegistered
	}
	if len(toReport) == 0 {
		return nil
	}

	_, err := svc.masterTransport.ReportChunks(ctx, nodeID, toReport)
	if err != nil {
		return fmt.Errorf("request chunk report: %w", err)
	}

	svc.mu.Lock()
	defer svc.mu.Unlock()
	if svc.nodeID != nodeID {
		return fmt.Errorf("registration changed: invalid report")
	}

	for _, chunk := range toReport {
		if state, ok := svc.catalog[chunk.ID]; ok {
			state.Reported = true
		}
	}
	return nil
}

func (svc *Service) RunReportLoop(ctx context.Context) {
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		case <- svc.reportWake:
		}

		slog.DebugContext(ctx, "exec report chunks")
		if err := svc.ReportChunks(ctx); err != nil {
			slog.ErrorContext(ctx, "report chunks failed", "error", err)
		}

		timer.Reset(jitter(svc.config.ReportDelay, 0.2))
	}
}

func jitter(base time.Duration, frac float64) time.Duration {
	delta := float64(base) * frac
	j := (rand.Float64() * 2 - 1) * delta
	return time.Duration(float64(base) + j)
}
