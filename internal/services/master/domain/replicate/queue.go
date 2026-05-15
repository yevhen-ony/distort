package replicate

import (
	"context"
	"log/slog"
	"sync"

	t "dos/internal/common/types"
)

type Queue struct {
	queued map[t.ChunkID]struct{}
	mu     sync.Mutex

	ch     chan t.ChunkID
}

func NewQueue(size int) *Queue {
	return &Queue{
		queued: make(map[t.ChunkID]struct{}),
		ch:     make(chan t.ChunkID, size),
	}
}


func (q *Queue) tryMarkQueued(id t.ChunkID) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	
	_, ok := q.queued[id]
	if ok {
		return false
	}

	q.queued[id] = struct{}{}
	return true
}

func (q *Queue) unmarkQueued(id t.ChunkID) {
	q.mu.Lock()
	defer q.mu.Unlock()

	delete(q.queued, id)
}



func (q *Queue) Enqueue(ctx context.Context, id t.ChunkID) {
	if !q.tryMarkQueued(id) {
		slog.DebugContext(ctx, "chunk is already enqueued; skip")
		return
	}

	select {
	case <-ctx.Done():
		q.unmarkQueued(id)
	case q.ch <- id:
	}
}

func (q *Queue) Pop(ctx context.Context) (t.ChunkID, error) {
	select {
	case id := <-q.ch:
		q.unmarkQueued(id)
		return id, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
