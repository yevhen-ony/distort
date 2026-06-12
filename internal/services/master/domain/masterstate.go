package domain

import (
	"context"
	"errors"

	t "dos/internal/common/types"
)

type LocalMasterStateService struct {
	resolver MasterSelfResolver 
}

type MasterSelfResolver interface {
	SelfRef() (t.MasterRef, error)
}

func NewLocalMasterStateService(resolver MasterSelfResolver) (*LocalMasterStateService, error) {
	if resolver == nil {
		return nil, errors.New("missing resolver")
	}

	s := &LocalMasterStateService{
		resolver: resolver,
	}	
	return s, nil
}

func (s *LocalMasterStateService) GetActiveMaster(_ context.Context) (t.MasterRef, error) {
	return s.resolver.SelfRef()
}

func (s *LocalMasterStateService) IsActiveMaster() bool {
	return true
}

func (s *LocalMasterStateService) WatchState(
  	ctx context.Context,
  	onActive func(context.Context),
) {
  	onActive(ctx)
  	<-ctx.Done()
}

func (s *LocalMasterStateService) TransferLeadership(_ context.Context) error {
 	return errors.New("leadership transfer is not supported")
}
