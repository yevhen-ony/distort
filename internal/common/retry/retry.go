package retry

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Retry struct {
	Delay time.Duration
	Timeout time.Duration
	MaxAttempts int
	ExpBackoff bool 
	MaxDelay time.Duration
}

func (r Retry) Run(ctx context.Context, fn func(context.Context) error) (err error) {
	delay := r.Delay
	if delay  == 0 {
		delay = time.Second
	}

	maxDelay := r.MaxDelay
	if maxDelay == 0 {
		maxDelay = 30 * time.Second
	}

	
	attempt := 0
	for {
		var attemptCtx context.Context = ctx
		var cancel context.CancelFunc = func() {}

		if r.Timeout > 0 {
			attemptCtx, cancel = context.WithTimeout(ctx, r.Timeout)
		}

		err = fn(attemptCtx)
		cancel()

		if err == nil {
			return nil
		}
		attempt++
		slog.DebugContext(ctx, "retry: attempt failed", "attempt", attempt, "error", err)

		if r.MaxAttempts > 0 && attempt >= r.MaxAttempts {
			return fmt.Errorf("retry failed after %d attempts: %w", attempt, err)
		}
		
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
		if r.ExpBackoff  {
			delay *= 2
			if delay > maxDelay {
				delay = maxDelay
			}
		}
	}
}

