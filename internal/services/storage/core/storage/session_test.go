package storage

import (
	"context"
	"testing"

	t "dos/internal/common/types"

	"github.com/stretchr/testify/require"
)

func TestUploadSession_Commit(tt *testing.T) {
	ctx := context.Background()

	session := NewUploadSession("chunk-1", 10)
	expected := t.NewChunk("chunk-1", []byte("hello"))

	var got t.Chunk
	commitCalls := 0
	abortCalls := 0
	session.onCommit = func(_ context.Context, chunk t.Chunk) error {
		commitCalls++
		got = chunk
		return nil
	}
	session.onAbort = func() error {
		abortCalls++
		return nil
	}
	// write
	n, err := session.Write([]byte("he"))
	require.NoError(tt, err)
	require.Equal(tt, 2, n)

	n, err = session.Write([]byte("llo"))
	require.NoError(tt, err)
	require.Equal(tt, 3, n)

	// exec commit
	require.NoError(tt, session.Commit(ctx))
	require.Equal(tt, 1, commitCalls)
	require.Equal(tt, expected, got)

	// close after commit is nop
	require.NoError(tt, session.Close())
	require.Zero(tt, abortCalls)
}

func TestUploadSession_Close(tt *testing.T) {
	session := NewUploadSession("chunk-1", 10)

	commitCalls := 0
	abortCalls := 0
	session.onCommit = func(context.Context, t.Chunk) error {
		commitCalls++
		return nil
	}
	session.onAbort = func() error {
		abortCalls++
		return nil
	}

	require.NoError(tt, session.Close())

	// close after close is nop
	require.NoError(tt, session.Close())
	require.Equal(tt, 1, abortCalls)

	// commit after close is nop
	require.NoError(tt, session.Commit(context.Background()))
	require.Equal(tt, 0, commitCalls)
}
