package reconcile 

import (
	"context"

	t "dos/internal/common/types"
)

type Queue struct {
	queued map[t.ChunkID]struct{}
	ch     chan t.ChunkID
}

func NewQueue(size int) *Queue {
	return &Queue{
		queued: make(map[t.ChunkID]struct{}),
		ch:     make(chan t.ChunkID, size),
	}
}

func (q *Queue) Enqueue(ctx context.Context, id t.ChunkID) {
	if _, ok := q.queued[id]; ok {
		return
	}

	select {
	case <-ctx.Done():
	case q.ch <- id:
		q.queued[id] = struct{}{}
	}
}

func (q *Queue) Pop(ctx context.Context) (t.ChunkID, error) {
	select {
	case id := <- q.ch:
		return id, nil 
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
