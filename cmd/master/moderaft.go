package main

import (
	"errors"
	"fmt"

	"dos/internal/common/master/resolve"
	m "dos/internal/services/master"
	"dos/internal/services/master/domain/object"
	"dos/internal/services/master/raftnode"
	"dos/internal/services/master/repo"
)

type MasterRaftMode struct {
	authority *object.Authority
	state     *raftnode.RaftMasterStateService

	node       *raftnode.ObjectNode
	repository *repo.InMemObjectRepo
	applier    *object.LocalCommandApplier
	writer     *object.ObjectWriterImpl
	codec      *object.JSONCommandCodec

	resolver *resolve.ResolverWithRaft
}

func (mode *MasterRaftMode) ObjectAuthority() *object.Authority {
	return mode.authority
}

func (mode *MasterRaftMode) MasterState() m.MasterState {
	return mode.state
}

func NewMasterRaftMode(config *Config) (*MasterRaftMode, error) {
	if config == nil {
		return nil, errors.New("missing config")
	}

	var err error
	mode := &MasterRaftMode{}

	mode.resolver, err = resolve.NewWithRaft(&config.Master)
	if err != nil {
		return nil, fmt.Errorf("raft resolver init: %w", err)
	}

	if err = mode.initObjectAuthority(mode.resolver, config); err != nil {
		return nil, fmt.Errorf("object authority init: %w", err)
	}

	mode.state, err = raftnode.NewRaftMasterStateService(raftnode.RaftMasterStateDeps{
		Raft: mode.node.Raft,
		Resolver: mode.resolver,
		Config: &config.Raft, 
	})
	if err != nil {
		return nil, fmt.Errorf("raft discovery service: %w", err)
	}

	return mode, nil
}

func (mode *MasterRaftMode) initObjectAuthority(
	resolver *resolve.ResolverWithRaft,
	config *Config,
) (err error) {

	mode.repository = repo.NewInMemObjectRepo()

	mode.applier, err = object.NewLocalCommandApplier(mode.repository)
	if err != nil {
		return fmt.Errorf("local command applier init: %w", err)
	}

	mode.codec = object.NewJSONCommandCodec()

	mode.node, err = raftnode.NewObjectNode(raftnode.ObjectNodeDeps{
		Config:   &config.Raft,
		Applier:  mode.applier,
		Codec:    mode.codec,
		Resolver: resolver,
	})
	if err != nil {
		return fmt.Errorf("object node init: %w", err)
	}

	mode.writer, err = object.NewObjectWriterImpl(mode.node.Submitter)
	if err != nil {
		return fmt.Errorf("object writer init: %w", err)
	}

	mode.authority, err = object.NewAuthority(object.AuthorityDeps{
		Reader: mode.repository,
		Writer: mode.writer,
	})
	if err != nil {
		return fmt.Errorf("object authority init: %w", err)
	}

	return nil
}
