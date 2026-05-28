package domain

import (
	"context"
	t "dos/internal/common/types"
	"errors"
)

type StaticDiscoveryConfig interface {
	AdvertiseAddr() string
}

type LocalDiscoveryService struct {
	resolver MasterSelfResolver 
}

type MasterSelfResolver interface {
	SelfRef() (t.MasterRef, error)
}

func NewLocalDiscoveryService(resolver MasterSelfResolver) (*LocalDiscoveryService, error) {
	if resolver == nil {
		return nil, errors.New("missing resolver")
	}

	s := &LocalDiscoveryService{
		resolver: resolver,
	}	
	return s, nil
}

func (s *LocalDiscoveryService) GetActiveMaster(_ context.Context) (t.MasterRef, error) {
	return s.resolver.SelfRef()
}

