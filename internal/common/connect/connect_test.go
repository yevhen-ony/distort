package connect

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/connectivity"
)

func TestNewConnCache_GetCaches(tt *testing.T) {
	cache := NewConnCache()
	tt.Cleanup(func() {
		require.NoError(tt, cache.Close())
	})

	conn1, err := cache.Get("localhost:1")
	require.NoError(tt, err)

	tt.Run("same", func(tt *testing.T) {
		conn2, err := cache.Get("localhost:1")
		require.NoError(tt, err)

		require.Same(tt, conn1, conn2)
	})
	tt.Run("different", func(tt *testing.T) {
		conn2, err := cache.Get("localhost:2")
		require.NoError(tt, err)

		require.NotSame(tt, conn1, conn2)
	})
}

func TestConnCache_Close(tt *testing.T) {
	cache := NewConnCache()

	conn, err := cache.Get("localhost:1")
	require.NoError(tt, err)

	require.NoError(tt, cache.Close())
	require.Equal(tt, connectivity.Shutdown, conn.GetState())
}
