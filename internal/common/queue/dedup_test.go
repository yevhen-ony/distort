package queue

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDedupQueue_TryEnq_Deduplicates(tt *testing.T) {
	q := NewDedupQueue[string](2)

	require.True(tt, q.TryEnq("a"))
	require.False(tt, q.TryEnq("a"))

	got, ok := q.TryDeq()
	require.True(tt, ok)
	require.Equal(tt, "a", got)

	require.True(tt, q.TryEnq("a"))
}

func TestDedupQueue_Drain(tt *testing.T) {
	q := NewDedupQueue[string](2)

	require.True(tt, q.TryEnq("a"))
	require.True(tt, q.TryEnq("b"))

	require.Equal(tt, []string{"a", "b"}, q.Drain())

	require.True(tt, q.TryEnq("a"))
	require.True(tt, q.TryEnq("b"))
}

func TestDedupQueue_TryDeq(tt *testing.T) {
	q := NewDedupQueue[string](1)

	require.True(tt, q.TryEnq("a"))
	require.False(tt, q.TryEnq("b"))

	got, ok := q.TryDeq()
	require.True(tt, ok)
	require.Equal(tt, "a", got)

	require.True(tt, q.TryEnq("b"))
}

func TestDedupQueue_Enq_Cancel(tt *testing.T) {
	q := NewDedupQueue[string](1)

	require.True(tt, q.TryEnq("a"))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ok, err := q.Enq(ctx, "b")
	require.False(tt, ok)
	require.ErrorIs(tt, err, context.Canceled)

	got, ok := q.TryDeq()
	require.True(tt, ok)
	require.Equal(tt, "a", got)

	require.True(tt, q.TryEnq("b"))
}

func TestDedupQueue_Full(tt *testing.T) {
	q := NewDedupQueue[string](1)

	require.False(tt, q.Full())

	require.True(tt, q.TryEnq("a"))
	require.True(tt, q.Full())

	_, ok := q.TryDeq()
	require.True(tt, ok)
	require.False(tt, q.Full())
}
