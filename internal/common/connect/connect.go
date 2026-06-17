package connect

import (
	"errors"
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ConnCache struct {
	mu    sync.Mutex
	conns map[string]*grpc.ClientConn
	opts  []grpc.DialOption
}

func NewConnCache(opts ...grpc.DialOption) *ConnCache {
	base := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	return &ConnCache{
		conns: map[string]*grpc.ClientConn{},
		opts:  append(base, opts...),
	}
}

func (cp *ConnCache) Get(addr string) (*grpc.ClientConn, error) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if conn, ok := cp.conns[addr]; ok {
		return conn, nil
	}

	conn, err := newConn(addr, cp.opts...)
	if err != nil {
		return nil, fmt.Errorf("new conn: %w", err)
	}
	cp.conns[addr] = conn
	return conn, nil
}

func (cp *ConnCache) Close() error {
	if cp == nil {
		return nil
	}

	cp.mu.Lock()
	defer cp.mu.Unlock()

	var errs []error
	for addr, conn := range cp.conns {
		if err := conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close %s: %w", addr, err))
		}
	}
	return errors.Join(errs...)
}

func newConn(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr, opts...)
}
