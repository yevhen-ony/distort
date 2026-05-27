package raftnode

import (
	"context"
	"errors"

	"github.com/hashicorp/raft"

	t "dos/internal/common/types"
)

var (
	ErrUnknownActiveMaster = errors.New("unknown active master")
)

type MasterResolver interface {
	Ref(t.MasterID) (t.MasterRef, error)
}

type RaftDiscoveryService struct {
	raft     *raft.Raft
	resolver MasterResolver
}

func NewRaftDiscoveryService(raft *raft.Raft, resolver MasterResolver) (*RaftDiscoveryService, error) {
	if raft == nil {
		return nil, errors.New("missing raft")
	}
	if resolver == nil {
		return nil, errors.New("missing resolver")
	}

	s := &RaftDiscoveryService{
		raft:     raft,
		resolver: resolver,
	}
	return s, nil
}

func (s *RaftDiscoveryService) GetActiveMaster(ctx context.Context) (t.MasterRef, error) {
	_, id := s.raft.LeaderWithID()
	if id == "" {
		return t.MasterRef{}, ErrUnknownActiveMaster
	}
	return s.resolver.Ref(t.MasterID(id))
}
