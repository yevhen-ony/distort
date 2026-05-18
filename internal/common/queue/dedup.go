package queue

import (
	"context"
	"sync"
)

type DedupQueue[T comparable] struct {
	ch     chan T
	queued map[T]struct{}
	cap    int
	mu     sync.Mutex
}

func NewDedupQueue[T comparable](cap int) *DedupQueue[T] {
	return &DedupQueue[T]{
		ch:     make(chan T, cap),
		queued: make(map[T]struct{}),
		cap:    cap,
	}
}

func (q *DedupQueue[T]) Enq(ctx context.Context, t T) (bool, error) {
	if !q.tryMark(t) {
		return false, nil
	}

	select {
	case q.ch <- t:
		return true, nil
	case <-ctx.Done():
		q.unmark(t)
		return false, ctx.Err()
	}
}

func (q *DedupQueue[T]) TryEnq(t T) bool {
	if !q.tryMark(t) {
		return false
	}

	select {
	case q.ch <- t:
		return true
	default:
		q.unmark(t)
		return false
	}
}

func (q *DedupQueue[T]) TryDeq() (T, bool) {
	select {
	case v := <-q.ch:
		q.unmark(v)
		return v, true
	default:
		var zero T
		return zero, false
	}
}

func (q *DedupQueue[T]) tryMark(t T) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	_, ok := q.queued[t]
	if ok {
		return false
	}

	q.queued[t] = struct{}{}
	return true
}

func (q *DedupQueue[T]) unmark(t T) {
	q.mu.Lock()
	defer q.mu.Unlock()

	delete(q.queued, t)
}

func (q *DedupQueue[T]) Drain() []T {
	l := len(q.ch)
	out := make([]T, 0, l)
	for range l {
		select {
		case v := <-q.ch:
			q.unmark(v)
			out = append(out, v)
		default:
			return out
		}
	}
	return out
}

func (q *DedupQueue[T]) Full() bool {
	return len(q.ch) == q.cap
}
