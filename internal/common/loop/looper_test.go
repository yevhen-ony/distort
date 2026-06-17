package loop

import (
	"context"
	"testing"
	"time"
)

func TestLooper_SkipFirstWait(tt *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	called := make(chan struct{}, 1)
	looper := NewLooper(time.Hour).SkipFirstWait()

	go looper.Run(ctx, func(context.Context) {
		called <- struct{}{}
		cancel()
	})

	select {
	case <-called:
	case <-time.After(100 * time.Millisecond):
		tt.Fatal("wait not skipped")
	}
}

func TestLooper_Flush(tt *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	called := make(chan struct{}, 1)
	looper := NewLooper(time.Hour)

	go looper.Run(ctx, func(context.Context) {
		called <- struct{}{}
		cancel()
	})

	looper.Flush()

	select {
	case <-called:
	case <-time.After(100 * time.Millisecond):
		tt.Fatal("flush failed")
	}
}

func TestLooper_OnCancel(tt *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	exec := make(chan struct{})

	looper := NewLooper(time.Hour)

	go func() {
		defer close(done)
		looper.Run(ctx, func(context.Context) {
			exec <- struct{}{}
		})
	}()

	cancel()

	select {
	case <-exec:
		tt.Fatal("fn should not run")
	case <-done:
	case <-time.After(100 * time.Millisecond):
		tt.Fatal("looper did not stop")
	}
}
