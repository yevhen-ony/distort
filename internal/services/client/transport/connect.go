package transport

import (
	"errors"
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ConnectionPool struct {
	mu    sync.Mutex
	conns map[string]*grpc.ClientConn
}

func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		conns: map[string]*grpc.ClientConn{},
	}
}

func (cp *ConnectionPool) Get(addr string) (*grpc.ClientConn, error) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	
	if conn, ok := cp.conns[addr]; ok {
		return conn, nil
	}

	conn, err := newConn(addr)
	if err != nil {
		return nil, fmt.Errorf("new conn: %w", err)
	}
	cp.conns[addr] = conn
	return conn, nil
}

func (cp *ConnectionPool) Close() error {
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

func newConn(addr string) (*grpc.ClientConn, error){
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	return grpc.NewClient(addr, opts...)
}
