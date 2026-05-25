package identity

import (
	"context"
	"dos/internal/common/dosctx"
	"dos/internal/common/retry"
	t "dos/internal/common/types"
	"dos/internal/services/storage/transport"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type IdentityConfig interface {
	AdvertiseAddr() string
	RegistrationTimeout() time.Duration
}

type IdentityDeps struct {
	MasterT *transport.Master
	Config  IdentityConfig
}

type IdentityService struct {
	masterT *transport.Master
	config  IdentityConfig
	mu      sync.RWMutex

	nodeID     t.NodeID
	obtainedAt time.Time
}

func NewIdentityService(deps IdentityDeps) (*IdentityService, error) {
	if deps.MasterT == nil {
		return nil, errors.New("missing master transport")
	}
	if deps.Config == nil {
		return nil, errors.New("missing config")
	}
	service := &IdentityService{
		masterT: deps.MasterT,
		config:  deps.Config,
	}
	return service, nil
}

var (
	ErrInvalidNodeID     = errors.New("invalid node id")
	ErrNodeNotRegistered = errors.New("node not registered")
)

func (is *IdentityService) getNewID(ctx context.Context) (t.NodeID, error) {

	var nodeID t.NodeID

	retry := retry.Retry{
		Delay: time.Second,
		Timeout: 2 * time.Second,
	}
	err := retry.Run(ctx, func(ctx context.Context) error {
		var innerErr error
		nodeID, innerErr = is.masterT.RegisterNode(ctx, is.config.AdvertiseAddr())
		return innerErr
	})

	if err != nil {
		return "", fmt.Errorf("register storage node: %w", err)
	}
	return nodeID, nil
}

func (is *IdentityService) RequestNewID(ctx context.Context) error {

	ctx = dosctx.WithService(ctx, "identity")

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

	slog.DebugContext(ctx, "received new id", "node_id", nodeID)

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
