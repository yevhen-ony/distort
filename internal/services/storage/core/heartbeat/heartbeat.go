package heartbeat 

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"dos/internal/common/dosctx"
	"dos/internal/common/loop"
	t "dos/internal/common/types"
	s "dos/internal/services/storage"
)

type HeartbeatConfig interface {
	HeartbeatInterval() time.Duration
}

type StatsProvider interface {
	GetStats() t.NodeStats
}

type IdentityProvider interface {
	GetID() (t.NodeID, error)
	RequestNewID(context.Context) error
}

type CatalogRestager interface {
	StageAndReportAll(context.Context) *s.TriggerReportResult
}

type MasterTransport interface {
	Heartbeat(context.Context, t.NodeID, t.NodeStats) (s.HeartbeatResult, error)
}

type HeartbeatDeps struct {
	Inventory StatsProvider
	Identity  IdentityProvider
	Storage   CatalogRestager

	MasterT MasterTransport
	Config  HeartbeatConfig
	Metrics *HeartbeatMetrics
}

type HeartbeatService struct {
	inventory StatsProvider
	identity  IdentityProvider
	storage   CatalogRestager

	masterT MasterTransport

	config  HeartbeatConfig
	metrics *HeartbeatMetrics
	looper  *loop.Looper

	mu     sync.RWMutex
	paused bool
}

func NewHeartbeatService(deps HeartbeatDeps) (*HeartbeatService, error) {
	if deps.Inventory == nil {
		return nil, errors.New("missing inventory service")
	}
	if deps.Identity == nil {
		return nil, errors.New("missing identity service")
	}
	if deps.Storage == nil {
		return nil, errors.New("missing restager")
	}
	if deps.MasterT == nil {
		return nil, errors.New("missing master transport")
	}
	if deps.Config == nil {
		return nil, errors.New("missing heartbeat config")
	}
	if deps.Metrics == nil {
		return nil, errors.New("missing heartbeat metrics")
	}
	service := &HeartbeatService{
		inventory: deps.Inventory,
		identity:  deps.Identity,
		storage:   deps.Storage,
		masterT:   deps.MasterT,
		config:    deps.Config,
		metrics:   deps.Metrics,
		looper:    loop.NewLooper(deps.Config.HeartbeatInterval()),
	}
	return service, nil
}

func (s *HeartbeatService) doIteration(ctx context.Context) {
	s.mu.RLock()
	paused := s.paused
	s.mu.RUnlock()

	if paused {
		slog.DebugContext(ctx, "heartbeat paused")
		return
	}

	nodeID, err := s.identity.GetID()
	if err != nil {
		slog.ErrorContext(ctx, "read node id failed", "error", err)
		return
	}

	stats := s.inventory.GetStats()

	res, err := s.masterT.Heartbeat(ctx, nodeID, stats)
	if err != nil {
		s.metrics.HeartbeatFailedTotal.Inc()
		slog.ErrorContext(ctx, "heartbeat transport failed", "node_id", nodeID, "error", err)
	}

	if res.NodeUnknown {
		slog.WarnContext(ctx, "node id is unknown", "node_id", nodeID)
		if err := s.identity.RequestNewID(ctx); err != nil {
			slog.WarnContext(ctx, "request new node id failed", "error", err)
			return
		}
		res := s.storage.StageAndReportAll(ctx)
		if len(res.Failed) != 0 {
			slog.WarnContext(ctx,
				"stage and report on re-registration partially failed",
				"failed_chunks", res.Failed,
			)
		}
	}
}

func (s *HeartbeatService) RunLoop(ctx context.Context) {
	ctx = dosctx.WithService(ctx, "heartbeat")
	s.looper.Run(ctx, s.doIteration)
}

func (s *HeartbeatService) Flush() {
	s.looper.Flush()
}

func (s *HeartbeatService) Pause() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.paused {
		return false
	}

	s.paused = true
	return true 
}

func (s *HeartbeatService) Resume() bool {
	defer s.Flush()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.paused {
		s.paused = false
		return true
	}
	return false	
}

func (s *HeartbeatService) IsPaused() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.paused
}
