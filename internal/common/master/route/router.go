package route

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"dos/internal/common/connect"
	"dos/internal/common/dosctx"
	t "dos/internal/common/types"
	"dos/internal/common/utils"

	"google.golang.org/grpc"
)

var (
	ErrRediscoveryExhausted = errors.New("all seeds exhausted")
)

type Resolver interface {
	Refs() []t.MasterRef
}

type ConnCache interface {
	Close() error
	Get(string) (*grpc.ClientConn, error)
}

type MasterDiscovery interface {
	DiscoverActive(context.Context, string) (t.MasterRef, error)
	Close() error
}

type MasterRouter struct {
	apiConn ConnCache

	discovery MasterDiscovery
	resolver  Resolver

	mu     sync.RWMutex
	active t.MasterRef

	discoveryMu sync.Mutex

	onMasterChange func(context.Context)
}

func New(resolver Resolver) (*MasterRouter, error) {
	if resolver == nil {
		return nil, errors.New("missing resolver")
	}

	interceptor := NewOnUnavailableInterceptor()
	router := &MasterRouter{
		resolver:  resolver,
		discovery: NewMasterDiscoveryService(),
		apiConn:   connect.NewConnCache(grpc.WithUnaryInterceptor(interceptor.UnaryIntercept)),
	}
	interceptor.SetOnUnavailable(func(ctx context.Context) error {
		_, err := router.Rediscover(ctx)
		return err
	})
	return router, nil
}

func (r *MasterRouter) SetOnMasterChange(fn func(ctx context.Context)) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.onMasterChange = fn
}

func (r *MasterRouter) Close() error {
	var errs []error
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.discovery.Close(); err != nil {
		errs = append(errs, err)
	}

	if err := r.apiConn.Close(); err != nil {
		errs = append(errs, err)
	} else {
		r.apiConn = nil
	}
	return errors.Join(errs...)
}

func (r *MasterRouter) Conn(ctx context.Context) (*grpc.ClientConn, error) {
	r.mu.RLock()
	active := r.active
	r.mu.RUnlock()

	if err := active.Validate(); err != nil {
		active, err = r.Rediscover(ctx)
		if err != nil {
			return nil, err
		}
	}

	return r.apiConn.Get(active.Addr)
}

func (r *MasterRouter) Rediscover(ctx context.Context) (t.MasterRef, error) {

	if !r.discoveryMu.TryLock() {
		return r.waitRediscovery()
	}
	defer r.discoveryMu.Unlock()

	ctx = dosctx.WithOperation(ctx, "rediscover")
	ctx = dosctx.WithService(ctx, "master_router")

	refs := r.resolver.Refs()
	for _, seed := range utils.RandomSelect(refs, len(refs)) {
		active, err := r.discovery.DiscoverActive(ctx, seed.Addr)
		if err != nil {
			slog.ErrorContext(ctx, "rediscovery", "error", err)
			continue
		}
		if err := active.Validate(); err != nil {
			slog.ErrorContext(ctx, "got invalid master ref")
			continue
		}

		if r.setActive(active) {
			if r.onMasterChange != nil {
				go r.onMasterChange(context.WithoutCancel(ctx))
			}
		}
		return active, nil
	}
	return t.MasterRef{}, ErrRediscoveryExhausted
}

func (r *MasterRouter) setActive(ref t.MasterRef) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.active.ID == ref.ID {
		return false
	}
	r.active = ref
	return true
}

func (r *MasterRouter) waitRediscovery() (t.MasterRef, error) {
	r.discoveryMu.Lock()
	r.discoveryMu.Unlock()

	r.mu.Lock()
	active := r.active
	r.mu.Unlock()

	if err := active.Validate(); err != nil {
		return t.MasterRef{}, err
	}
	return active, active.Validate()
}
