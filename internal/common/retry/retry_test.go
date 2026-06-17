package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRetry_FirstAttemptSucceeded(tt *testing.T) {
	calls := 0

	r := Retry{Delay: time.Millisecond, MaxAttempts: 3}
	err := r.Run(context.Background(), func(context.Context) error {
		calls++
		return nil
	})

	require.NoError(tt, err)
	require.Equal(tt, 1, calls)
}

func TestRetry_RetriesUntilSuccess(tt *testing.T) {
	calls := 0
	r := Retry{Delay: time.Millisecond, MaxAttempts: 3}

	err := r.Run(context.Background(), func(context.Context) error {
		calls++
		if calls < 3 {
			return errors.New("fail")
		}
		return nil
	})

	require.NoError(tt, err)
	require.Equal(tt, 3, calls)
}

func TestRetry_ReachMaxAttempts(tt *testing.T) {
	expectedErr := errors.New("fail")
	calls := 0

	r := Retry{Delay: time.Millisecond, MaxAttempts: 2}
	err := r.Run(context.Background(), func(context.Context) error {
		calls++
		return expectedErr
	})

	require.ErrorIs(tt, err, expectedErr)
	require.ErrorContains(tt, err, "retry failed after 2 attempts")
	require.Equal(tt, 2, calls)
}

func TestRetry_OnCancel(tt *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0

	r := Retry{Delay: time.Hour}
	err := r.Run(ctx, func(context.Context) error {
		calls++
		cancel()
		return errors.New("fail")
	})

	require.ErrorIs(tt, err, context.Canceled)
	require.Equal(tt, 1, calls)
}

func TestRetry_OnTimeout(tt *testing.T) {
	var attemptErr error

	r := Retry{
		Delay:       time.Millisecond,
		Timeout:     time.Millisecond,
		MaxAttempts: 1,
	}
	err := r.Run(context.Background(), func(ctx context.Context) error {
		<-ctx.Done()
		attemptErr = ctx.Err()
		return attemptErr
	})

	require.ErrorIs(tt, attemptErr, context.DeadlineExceeded)
	require.ErrorIs(tt, err, context.DeadlineExceeded)
}
