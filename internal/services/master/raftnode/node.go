package raftnode

import (
	"dos/internal/common/master/resolve"
	"dos/internal/common/utils"
	"dos/internal/services/master/domain/object"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/raft"
)

type ObjectNode struct {
	Submitter object.CommandSubmitter
	Raft      *raft.Raft
}

type ObjectNodeDeps struct {
	Config   *Config
	Codec    object.CommandCodec
	Applier  object.CommandApplier
	Resolver *resolve.ResolverWithRaft
}

func NewObjectNode(deps ObjectNodeDeps) (*ObjectNode, error) {
	if err := deps.Config.Validate(); err != nil {
		return nil, err
	}
	if deps.Resolver == nil {
		return nil, errors.New("missing resolver")
	}

	self, err := deps.Resolver.RaftSelfRef()
	if err != nil {
		return nil, fmt.Errorf("resolve self ref: %w", err)
	}

	fsm, err := NewObjectFSM(deps.Codec, deps.Applier)
	if err != nil {
		return nil, fmt.Errorf("object fsm init: %w", err)
	}

	logStore := raft.NewInmemStore()
	stableStore := raft.NewInmemStore()
	snapshotStore := raft.NewInmemSnapshotStore()

	addr := raft.ServerAddress(self.Addr)
	_, transport := raft.NewInmemTransport(addr)

	raftConfig := deps.Config.RaftConfig(self.ID)
	r, err := raft.NewRaft(raftConfig, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return nil, fmt.Errorf("create raft node: %w", err)
	}

	if deps.Config.Bootstrap {
		if err := bootstrapRaft(r, deps.Resolver); err != nil {
			return nil, fmt.Errorf("raft bootstrap: %w", err)
		}
	}

	timeout := deps.Config.ApplyTimeout
	if timeout <= 0 {
		timeout = time.Second
	}

	submitter, err := NewCommandSubmitter(CommandSubmitterDeps{
		Codec:   deps.Codec,
		Raft:    r,
		Timeout: timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("create raft command submitter: %w", err)
	}

	node := &ObjectNode{
		Raft:      r,
		Submitter: submitter,
	}
	return node, nil
}

func bootstrapRaft(r *raft.Raft, resolver *resolve.ResolverWithRaft) error {
	servers := utils.Map(resolver.RaftRefs(), func(ref resolve.RaftRef) raft.Server {
		return raft.Server{
			ID:      raft.ServerID(ref.ID),
			Address: raft.ServerAddress(ref.Addr),
		}
	})
	return r.BootstrapCluster(raft.Configuration{
		Servers: servers,
	}).Error()
}
