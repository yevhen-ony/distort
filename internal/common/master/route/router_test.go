package route

import (
	"context"
	"errors"
	"testing"
	"time"

	t "dos/internal/common/types"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestMasterRouter_Rediscover_Success(tt *testing.T) {
	active := t.MasterRef{ID: "master-1", Addr: "master-1:10000"}
	changed := make(chan struct{}, 1)

	router := &MasterRouter{
		resolver: testResolver{
			refs: []t.MasterRef{{ID: "seed-1", Addr: "seed-1:10000"}},
		},
		discovery: &testDiscovery{active: active},
	}
	router.SetOnMasterChange(func(context.Context) {
		changed <- struct{}{}
	})

	got, err := router.Rediscover(context.Background())

	require.NoError(tt, err)
	require.Equal(tt, active, got)
	require.Equal(tt, active, router.active)

	select {
	case <-changed:
	case <-time.After(100 * time.Millisecond):
		tt.Fatal("on master change not called")
	}
}

func TestMasterRouter_Rediscover_OnError(tt *testing.T) {
	router := &MasterRouter{
		resolver: testResolver{
			refs: []t.MasterRef{{ID: "seed-1", Addr: "seed-1:10000"}},
		},
		discovery: &testDiscovery{err: errors.New("discovery failed")},
	}

	_, err := router.Rediscover(context.Background())

	require.ErrorIs(tt, err, ErrRediscoveryExhausted)
}

func TestMasterRouter_Conn_OnActive(tt *testing.T) {
	apiConn := &testConnCache{}
	active := t.MasterRef{ID: "master-1", Addr: "master-1:10000"}

	router := &MasterRouter{
		active:  active,
		apiConn: apiConn,
	}

	_, err := router.Conn(context.Background())

	require.NoError(tt, err)
	require.Equal(tt, active.Addr, apiConn.addr)
}

func TestMasterRouter_Conn_Discovered(tt *testing.T) {
	apiConn := &testConnCache{}
	active := t.MasterRef{ID: "master-1", Addr: "master-1:10000"}

	router := &MasterRouter{
		resolver: testResolver{
			refs: []t.MasterRef{{ID: "seed-1", Addr: "seed-1:10000"}},
		},
		discovery: &testDiscovery{active: active},
		apiConn:   apiConn,
	}

	_, err := router.Conn(context.Background())

	require.NoError(tt, err)
	require.Equal(tt, active.Addr, apiConn.addr)
}

type testResolver struct {
	refs []t.MasterRef
}

func (r testResolver) Refs() []t.MasterRef {
	return r.refs
}

type testDiscovery struct {
	active t.MasterRef
	err    error
	closed bool
}

func (d *testDiscovery) DiscoverActive(context.Context, string) (t.MasterRef, error) {
	if d.err != nil {
		return t.MasterRef{}, d.err
	}
	return d.active, nil
}

func (d *testDiscovery) Close() error {
	d.closed = true
	return nil
}

type testConnCache struct {
	addr   string
	closed bool
}

func (c *testConnCache) Get(addr string) (*grpc.ClientConn, error) {
	c.addr = addr
	return nil, nil
}

func (c *testConnCache) Close() error {
	c.closed = true
	return nil
}
