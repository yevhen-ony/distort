package route

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	mpb "dos/gen/proto/master/v1"
	"dos/internal/common/connect"
	"dos/internal/common/convert"
	"dos/internal/common/dosctx"
	t "dos/internal/common/types"
	"dos/internal/common/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrRediscoveryExhausted = errors.New("all seeds exhausted")
)

type Resolver interface {
	Refs() []t.MasterRef
}

type MasterRouter struct {
	discoveryConn *connect.ConnCache
	apiConn *connect.ConnCache

	resolver Resolver

	mu     sync.RWMutex
	active t.MasterRef

	discoveryMu sync.Mutex
}

func New(resolver Resolver) (*MasterRouter, error) {
	if resolver == nil {
		return nil, errors.New("missing resolver")
	}
	
	router := &MasterRouter{
		resolver: resolver,
	}
	router.setupConnCaches()
	return router, nil
}

func (r *MasterRouter) setupConnCaches() {
	r.discoveryConn = connect.NewConnCache()
	r.apiConn = connect.NewConnCache(grpc.WithUnaryInterceptor(r.UnaryIntercept))
}

func (r *MasterRouter) Close() error {
	var errs []error
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.discoveryConn.Close(); err != nil {
		errs = append(errs, err)
	} else {
		r.discoveryConn = nil
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
		if err := r.Rediscover(ctx); err != nil {
			return nil, err
		}

		r.mu.RLock()
		active = r.active
		r.mu.RUnlock()
	}

	return r.apiConn.Get(active.Addr)
}

func (r *MasterRouter) Rediscover(ctx context.Context) error {

	if !r.discoveryMu.TryLock() {
		return r.waitRediscovery()
	}
	defer r.discoveryMu.Unlock()

	ctx = dosctx.WithOperation(ctx, "rediscover")
	ctx = dosctx.WithService(ctx, "master_router")

	refs := r.resolver.Refs()
	for _, seed := range utils.RandomSelect(refs, len(refs)) {

		conn, err := r.discoveryConn.Get(seed.Addr)
		if err != nil {
			slog.ErrorContext(ctx, "create connection", "addr", seed.Addr, "error", err)
			continue
		}
		client := mpb.NewMasterDiscoveryServiceClient(conn)
		rsp, err := client.GetActiveMaster(ctx, &mpb.GetActiveMasterRequest{})
		if err != nil {
			slog.ErrorContext(ctx, "get active master", "addr", seed.Addr, "error", err)
			continue
		}

		active := convert.MasterRefFromPB(rsp.GetActive())
		if err := active.Validate(); err != nil {
			slog.ErrorContext(ctx, "got invalid master ref")
			continue
		}

		r.setActive(active)
		return nil
	}
	return ErrRediscoveryExhausted
}

func (r *MasterRouter) UnaryIntercept(
	ctx context.Context,
	method string,
	req any,
	reply any,
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {

	err := invoker(ctx, method, req, reply, cc, opts...)
	if err == nil {
		return nil
	}

	if status.Code(err) != codes.Unavailable {
		return err
	}

	if rediscoverErr := r.Rediscover(ctx); rediscoverErr != nil {
		slog.ErrorContext(ctx,
			"rediscover active master failed",
			"method", method,
			"error", rediscoverErr,
		)
	}

	return err
}

func (r *MasterRouter) setActive(id t.MasterRef) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.active = id
}

func (r *MasterRouter) waitRediscovery() error {
	r.discoveryMu.Lock()
	r.discoveryMu.Unlock()

	r.mu.Lock()
	active := r.active
	r.mu.Unlock()

	return active.Validate()
}
