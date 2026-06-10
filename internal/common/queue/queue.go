package queue

import "context"

type Queue[T any] struct {
	ch  chan T
	cap int
}

func NewQueue[T any](cap int) *Queue[T] {
	return &Queue[T]{
		ch:  make(chan T, cap),
		cap: cap,
	}
}

func (q *Queue[T]) Enq(ctx context.Context, t T) error {
	select {
	case q.ch <- t:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (q *Queue[T]) TryEnq(t T) bool {
	select {
	case q.ch <- t:
		return true
	default:
		return false
	}
}

func (q *Queue[T]) TryDeq() (T, bool) {
	select {
	case v := <-q.ch:
		return v, true
	default:
		var zero T
		return zero, false
	}
}

func (q *Queue[T]) Drain() []T {
	l := len(q.ch)
	out := make([]T, 0, l)
	for range l {
		select {
		case v := <-q.ch:
			out = append(out, v)
		default:
			return out
		}
	}
	return out
}

func (q *Queue[T]) Full() bool {
	return len(q.ch) == q.cap
}
