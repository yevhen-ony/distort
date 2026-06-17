package raftnode

import (
	"context"
	"errors"
	"log/slog"

	"github.com/hashicorp/raft"

	t "dos/internal/common/types"
)

var (
	ErrUnknownActiveMaster = errors.New("unknown active master")
)

type MasterResolver interface {
	Ref(t.MasterID) (t.MasterRef, error)
}

type RaftMasterStateService struct {
	raft     *raft.Raft
	resolver MasterResolver
}

type RaftMasterStateDeps struct {
	Raft     *raft.Raft
	Resolver MasterResolver
}

func NewRaftMasterStateService(deps RaftMasterStateDeps) (*RaftMasterStateService, error) {
	if deps.Raft == nil {
		return nil, errors.New("missing raft")
	}
	if deps.Resolver == nil {
		return nil, errors.New("missing resolver")
	}

	s := &RaftMasterStateService{
		raft:     deps.Raft,
		resolver: deps.Resolver,
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

		slog.DebugContext(ctx, "master activated")

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

	for {
		select {
		case <-ctx.Done():
			return

		case isLeader := <-s.raft.LeaderCh():
			if isLeader {
				start()
			} else {
				stop()
			}

		}
	}
}

func (s *RaftMasterStateService) TransferLeadership(_ context.Context) error {
	if !s.IsActiveMaster() {
		return errors.New("not a leader")
	}

	f := s.raft.LeadershipTransfer()
	return f.Error()
}
