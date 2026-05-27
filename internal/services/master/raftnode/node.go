package raftnode

import (
	"dos/internal/services/master/domain/object"
	"fmt"
	"time"

	"github.com/hashicorp/raft"
)


type ObjectNode struct {
	Submitter object.CommandSubmitter
	Raft      *raft.Raft
}

type ObjectNodeDeps struct {
	Config  *Config
	Codec   object.CommandCodec
	Applier object.CommandApplier
}

func NewObjectNode(deps ObjectNodeDeps) (*ObjectNode, error) {
	if err := deps.Config.Validate(); err != nil {
		return nil, err
	}

	fsm, err := NewObjectFSM(deps.Codec, deps.Applier)
	if err != nil {
		return nil, fmt.Errorf("object fsm init: %w", err)
	}

  	logStore := raft.NewInmemStore()
  	stableStore := raft.NewInmemStore()
  	snapshotStore := raft.NewInmemSnapshotStore()

  	addr := raft.ServerAddress(deps.Config.BindAddr)
  	_, transport := raft.NewInmemTransport(addr)

	raftConfig := deps.Config.RaftConfig()
  	r, err := raft.NewRaft(raftConfig, fsm, logStore, stableStore, snapshotStore, transport)
  	if err != nil {
  		return nil, fmt.Errorf("create raft node: %w", err)
  	}

  	if deps.Config.Bootstrap {
		servers := []raft.Server{
			{ ID: raftConfig.LocalID, Address: addr},
		}
  		err := r.BootstrapCluster(raft.Configuration{
  			Servers: servers,
  		}).Error()
  		if err != nil {
  			return nil, fmt.Errorf("bootstrap raft cluster: %w", err)
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
