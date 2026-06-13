package identity

import (
	"context"
	"testing"
	"time"

	t "dos/internal/common/types"

	"github.com/stretchr/testify/require"
)

func TestIdentityService_RequestNewID_SingleCall(tt *testing.T) {
	ctx := context.Background()

	f := newIdentityFixture()
	service, err := NewIdentityService(f.deps())
	require.NoError(tt, err)

	err = service.RequestNewID(ctx)

	require.NoError(tt, err)
	require.Equal(tt, 1, f.transport.calls)
	require.Equal(tt, "storage-1:10001", f.transport.addr)

	got, err := service.GetID()
	require.NoError(tt, err)
	require.Equal(tt, t.NodeID("node-1"), got)
}

func TestIdentityService_RequestNewID_SkipsWithinTimeout(t *testing.T) {
	ctx := context.Background()

	f := newIdentityFixture()
	service, err := NewIdentityService(f.deps())
	require.NoError(t, err)

	require.NoError(t, service.RequestNewID(ctx))
	require.NoError(t, service.RequestNewID(ctx))

	require.Equal(t, 1, f.transport.calls)
}

func TestIdentityService_RequestNewID_RefreshesAfterTimeout(tt *testing.T) {
	ctx := context.Background()
	f := newIdentityFixture()
	f.config.registrationTimeout = -time.Second

	service, err := NewIdentityService(f.deps())
	require.NoError(tt, err)

	require.NoError(tt, service.RequestNewID(ctx))

	f.transport.nodeID = "node-2"
	require.NoError(tt, service.RequestNewID(ctx))

	require.Equal(tt, 2, f.transport.calls)

	got, err := service.GetID()
	require.NoError(tt, err)
	require.Equal(tt, t.NodeID("node-2"), got)
}

func TestIdentityService_Validate(t *testing.T) {
	ctx := context.Background()
	f := newIdentityFixture()
	service, err := NewIdentityService(f.deps())
	require.NoError(t, err)

	_, err = service.GetID()
	require.ErrorIs(t, err, ErrNodeNotRegistered)

	require.NoError(t, service.RequestNewID(ctx))

	require.NoError(t, service.Validate("node-1"))
	require.ErrorIs(t, service.Validate("node-2"), ErrInvalidNodeID)
}

// fixture

type identityFixture struct {
	config    *fakeIdentityConfig
	transport *fakeMasterTransport
}

func newIdentityFixture() *identityFixture {
	return &identityFixture{
		transport: &fakeMasterTransport{nodeID: "node-1"},
		config: &fakeIdentityConfig{
			advertiseAddr:       "storage-1:10001",
			registrationTimeout: time.Minute,
		},
	}
}

func (f *identityFixture) deps() IdentityDeps{
	return IdentityDeps{
		MasterT: f.transport,
		Config: f.config,
	}
}

// fake config

type fakeIdentityConfig struct {
	advertiseAddr       string
	registrationTimeout time.Duration
}

func (c fakeIdentityConfig) AdvertiseAddr() string {
	return c.advertiseAddr
}

func (c fakeIdentityConfig) RegistrationTimeout() time.Duration {
	return c.registrationTimeout
}

// fake master transport

type fakeMasterTransport struct {
	nodeID t.NodeID
	err    error
	calls  int
	addr   string
}

func (t *fakeMasterTransport) RegisterNode(_ context.Context, addr string) (t.NodeID, error) {
	t.calls++
	t.addr = addr
	if t.err != nil {
		return "", t.err
	}
	return t.nodeID, nil
}
