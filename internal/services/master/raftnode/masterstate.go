package raftnode

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/hashicorp/raft"

	t "dos/internal/common/types"
)

var (
	ErrUnknownActiveMaster = errors.New("unknown active master")
)

type RaftMasterStateConfig interface {
	MasterStatePollingInterval() time.Duration
}

type MasterResolver interface {
	Ref(t.MasterID) (t.MasterRef, error)
}

type RaftMasterStateService struct {
	raft     *raft.Raft
	resolver MasterResolver
	config   RaftMasterStateConfig
}

type RaftMasterStateDeps struct {
	Raft     *raft.Raft
	Resolver MasterResolver
	Config   RaftMasterStateConfig
}

func NewRaftMasterStateService(deps RaftMasterStateDeps) (*RaftMasterStateService, error) {
	if deps.Raft == nil {
		return nil, errors.New("missing raft")
	}
	if deps.Resolver == nil {
		return nil, errors.New("missing resolver")
	}
	if deps.Config == nil {
		return nil, errors.New("missing config")
	}

	s := &RaftMasterStateService{
		raft:     deps.Raft,
		resolver: deps.Resolver,
		config:   deps.Config,
	}
	return s, nil
}

func (s *RaftMasterStateService) GetActiveMaster(ctx context.Context) (t.MasterRef, error) {
	_, id := s.raft.LeaderWithID()
	if id == "" {
		return t.MasterRef{}, ErrUnknownActiveMaster
	}
	return s.resolver.Ref(t.MasterID(id))
}

func (s *RaftMasterStateService) IsActiveMaster() bool {
	return raft.Leader == s.raft.State()
}

func (s *RaftMasterStateService) WatchState(
	ctx context.Context,
	onActive func(ctx context.Context),
) {

	var activeCancel context.CancelFunc
	active := false

	start := func() {
		if active {
			return
		}
		active = true

		slog.DebugContext(ctx, "*** onActive is triggerd ***")

		activeCtx, cancel := context.WithCancel(ctx)
		activeCancel = cancel
		onActive(activeCtx)
	}

	stop := func() {
		if !active {
			return
		}
		active = false

		if activeCancel != nil {
			activeCancel()
			activeCancel = nil
		}
	}

	defer stop()

	if s.IsActiveMaster() {
		start()
	}

	ticker := time.NewTicker(s.config.MasterStatePollingInterval())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			if s.IsActiveMaster() {
				start()
			} else {
				stop()
			}
		}
	}
}
