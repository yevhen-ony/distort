
package queue

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueue_HappyPath(tt *testing.T) {
	q := NewQueue[string](2)

	require.False(tt, q.Full())

	require.True(tt, q.TryEnq("a"))
	require.True(tt, q.TryEnq("b"))
	require.True(tt, q.Full())
	require.False(tt, q.TryEnq("c"))

	got, ok := q.TryDeq()
	require.True(tt, ok)
	require.Equal(tt, "a", got)

	got, ok = q.TryDeq()
	require.True(tt, ok)
	require.Equal(tt, "b", got)

	_, ok = q.TryDeq()
	require.False(tt, ok)
}

func TestQueue_Drain(tt *testing.T) {
	q := NewQueue[string](3)

	require.True(tt, q.TryEnq("a"))
	require.True(tt, q.TryEnq("b"))

	require.Equal(tt, []string{"a", "b"}, q.Drain())

	_, ok := q.TryDeq()
	require.False(tt, ok)
}

func TestQueue_Enq_WhenFull(tt *testing.T) {
	q := NewQueue[string](1)

	require.True(tt, q.TryEnq("a"))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := q.Enq(ctx, "b")
	require.ErrorIs(tt, err, context.Canceled)
}
