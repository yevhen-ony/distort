package main

import (
	"dos/internal/common/master/resolve"
	m "dos/internal/services/master"
	"dos/internal/services/master/domain"
	"dos/internal/services/master/domain/object"
	"dos/internal/services/master/repo"
	"errors"
	"fmt"
)

type MasterLocalMode struct {
	authority *object.Authority
	discovery *domain.LocalMasterStateService

	repository *repo.InMemObjectRepo
	applier    *object.LocalCommandApplier
	submitter  *object.LocalCommandSubmitter
	writer     *object.CommandBackedObjectWriter

	resolver *resolve.Resolver
}

func (mode *MasterLocalMode) ObjectAuthority() *object.Authority {
	return mode.authority
}

func (mode *MasterLocalMode) MasterState() m.MasterState {
	return mode.discovery
}

func NewMasterLocalMode(config *Config) (*MasterLocalMode, error) {
	if config == nil {
		return nil, errors.New("missing config")
	}

	mode := &MasterLocalMode{}
	if err := mode.initLocalObjectAuthority(); err != nil {
		return nil, fmt.Errorf("object authority init: %w", err)
	}

	if err := mode.initLocalMasterDiscovery(config); err != nil {
		return nil, fmt.Errorf("local master discovery init: %w", err)
	}
	return mode, nil
}

func (mode *MasterLocalMode) initLocalObjectAuthority() (err error) {

	mode.repository = repo.NewInMemObjectRepo()

	mode.applier, err = object.NewLocalCommandApplier(mode.repository)
	if err != nil {
		return fmt.Errorf("local command applier init: %w", err)
	}

	mode.submitter, err = object.NewLocalCommandSubmitter(mode.applier)
	if err != nil {
		return fmt.Errorf("local command submitter init: %w", err)
	}

	mode.writer, err = object.NewCommandBackedObjectWriter(mode.submitter)
	if err != nil {
		return fmt.Errorf("object writer init: %w", err)
	}

	mode.authority, err = object.NewAuthority(object.AuthorityDeps{
		Reader: mode.repository, // direct from repo
		Writer: mode.writer,     // via submitter
	})
	if err != nil {
		return fmt.Errorf("object authority init: %w", err)
	}

	return nil
}

func (mode *MasterLocalMode) initLocalMasterDiscovery(config *Config) (err error) {
	mode.resolver, err = resolve.New(&config.Master)
	if err != nil {
		return fmt.Errorf("resolver init: %w", err) 
	}

	mode.discovery, err = domain.NewLocalMasterStateService(mode.resolver)
	if err != nil {
		return fmt.Errorf("discovery init: %w", err)
	}

	return nil	
}
