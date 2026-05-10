package core

import (
	"context"
	"dos/internal/common/retry"
	t "dos/internal/common/types"
	"dos/internal/services/storage/transport"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type IdentityServiceConfig interface {
	AdvertiseAddr() string
	RegistrationTimeout() time.Duration
}

type IdentityService struct {
	master *transport.Master
	config IdentityServiceConfig
	mu sync.RWMutex

	nodeID t.NodeID
	obtainedAt time.Time 
}

func NewIdentityService(master *transport.Master, config IdentityServiceConfig) *IdentityService {
	return &IdentityService{master: master, config: config}
}

var (
	ErrInvalidNodeID = errors.New("invalid node id")
	ErrNodeNotRegistered = errors.New("node not registered")
)


func (is *IdentityService) getNewID(ctx context.Context) (t.NodeID, error) {

	var nodeID t.NodeID

	retry := retry.Retry{Delay: time.Second}
	err := retry.Run(ctx, func(ctx context.Context) error {
		var innerErr error
		nodeID, innerErr = is.master.RegisterNode(ctx, is.config.AdvertiseAddr())
		return innerErr
	})

	if err != nil {
		return "", fmt.Errorf("register storage node: %w", err)
	}
	return nodeID, nil
}

func (is *IdentityService) RequestNewID(ctx context.Context) error {
	requestedAt := time.Now()	

	is.mu.Lock() // intentionally blocking while long getNewID
	defer is.mu.Unlock()

	if requestedAt.Before(is.obtainedAt.Add(is.config.RegistrationTimeout())) {
		return nil
	}
	
	nodeID, err := is.getNewID(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "fail to get new node ID: %w", err)
		return err
	}
	is.nodeID = nodeID
	is.obtainedAt = time.Now()
	return nil
}

func (is *IdentityService) Validate(nodeID t.NodeID) error {
	is.mu.RLock()
	defer is.mu.RUnlock()
	if nodeID != is.nodeID {
		return ErrInvalidNodeID
	}
	return nil
}

func (is *IdentityService) GetID() (t.NodeID, error) {
	is.mu.RLock()
	defer is.mu.RUnlock()
	
	if is.nodeID == "" {
		return "", ErrNodeNotRegistered
	}
	return is.nodeID, nil
}

