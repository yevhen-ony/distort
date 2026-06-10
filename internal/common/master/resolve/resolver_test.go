package resolve

import (
	"testing"

	t "dos/internal/common/types"

	"github.com/stretchr/testify/require"
)

func testConfig() *Config {
	return &Config{
		APIPort:  10000,
		RaftPort: 12000,
		Self:     "master-1",
		Peers:    []string{"master-1", "master-2"},
	}
}

func TestNew_Refs(tt *testing.T) {
	resolver, err := New(testConfig())
	require.NoError(tt, err)

	require.Equal(tt, []t.MasterRef{
		{ID: "master-1", Addr: "master-1:10000"},
		{ID: "master-2", Addr: "master-2:10000"},
	}, resolver.Refs())

	ref, err := resolver.Ref("master-2")
	require.NoError(tt, err)
	require.Equal(tt, t.MasterRef{ID: "master-2", Addr: "master-2:10000"}, ref)

	_, err = resolver.Ref("missing")
	require.ErrorIs(tt, err, ErrPeerNotFound)
}

func TestResolver_Self(tt *testing.T) {
	cfg := testConfig()
	cfg.Self = ""

	resolver, err := New(cfg)
	require.NoError(tt, err)

	_, err = resolver.Self()
	require.ErrorIs(tt, err, ErrMissingSelf)
}

func TestNewWithRaft(tt *testing.T) {
	resolver, err := NewWithRaft(testConfig())
	require.NoError(tt, err)

	require.Equal(tt, []RaftRef{
		{ID: "master-1", Addr: "master-1:12000"},
		{ID: "master-2", Addr: "master-2:12000"},
	}, resolver.RaftRefs())

	self, err := resolver.RaftSelfRef()
	require.NoError(tt, err)
	require.Equal(tt, RaftRef{ID: "master-1", Addr: "master-1:12000"}, self)
}
