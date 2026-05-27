package domain

import (
	"context"
	"dos/internal/common/master/resolve"
	t "dos/internal/common/types"
	"errors"
)

type StaticDiscoveryConfig interface {
	AdvertiseAddr() string
}

type StaticDiscoveryService struct {
	resolver *resolve.Resolver
}

func NewStaticDiscoveryService(resolver *resolve.Resolver) (*StaticDiscoveryService, error) {
	if resolver == nil {
		return nil, errors.New("missing resolver")
	}

	s := &StaticDiscoveryService{
		resolver: resolver,
	}	
	return s, nil
}

func (s *StaticDiscoveryService) GetActiveMaster(_ context.Context) (t.MasterRef, error) {
	return s.resolver.SelfRef()
}

