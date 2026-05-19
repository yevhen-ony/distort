package loop

import (
	"context"
	"math/rand/v2"
	"time"
)

type Looper struct {
	nextWait time.Duration
	firstWait time.Duration

	flush chan struct{}
}

func NewLooper(interval time.Duration) *Looper {
	return &Looper {
		nextWait: interval,
		firstWait: interval,
		flush: make(chan struct{}, 1),
	}
}

func (l *Looper) SkipFirstWait() *Looper {
	l.firstWait	= 0
	return l
}

func (l *Looper) Run(ctx context.Context, fn func(context.Context) ) {
	timer := time.NewTimer(l.firstWait)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		case <-l.flush:
		}

		fn(ctx)	

		interval := jitter(l.nextWait, 0.2)
		timer.Reset(interval)
	}
}

func (l *Looper) Flush() {
	select{
	case l.flush <- struct{}{}:
	default:
	}
}

func jitter(base time.Duration, frac float64) time.Duration {
	delta := float64(base) * frac
	j := (rand.Float64()*2 - 1) * delta
	return time.Duration(float64(base) + j)
}
