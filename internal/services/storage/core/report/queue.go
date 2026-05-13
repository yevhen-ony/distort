package report

import (
	"context"

	t "dos/internal/common/types"
)

type Queue struct {
	ch chan t.StorageNodeReport
	size int
}

func NewReportQueue(size int) *Queue {
	return &Queue{
		ch: make(chan t.StorageNodeReport, size),
		size: size,
	}
}

func (q *Queue) Enqueue(ctx context.Context, rec t.StorageNodeReport) {
	select {
	case <-ctx.Done():
	case q.ch <- rec:
	}
}

func (q *Queue) Drain() []t.StorageNodeReport {
	n := len(q.ch)
	reports := make([]t.StorageNodeReport, 0, n)
	for range n {
		reports = append(reports, <-q.ch)
	}
	return reports
}

func (q *Queue) IsFull() bool {
	return len(q.ch) == q.size 
}

