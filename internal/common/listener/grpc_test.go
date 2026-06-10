package listener

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type testListenerConfig struct {
	addr string
}

func (c testListenerConfig) ListeningAddr() string {
	return c.addr
}

func TestRunGRPCServer_OnBadAddress(tt *testing.T) {
	err := RunGRPCServer(
		context.Background(),
		testListenerConfig{addr: "bad-address"},
		func(*grpc.Server) {},
	)

	require.Error(tt, err)
	require.True(tt, strings.Contains(err.Error(), "bad-address"))
}

func TestRunGRPCServer_OnContextCancel(tt *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	called := false
	err := RunGRPCServer(ctx,
		testListenerConfig{addr: "localhost:0"},
		func(*grpc.Server) { called = true },
	)

	require.NoError(tt, err)
	require.True(tt, called)
}
