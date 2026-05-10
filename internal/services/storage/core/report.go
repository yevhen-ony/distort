package core

import (
	"context"
	t "dos/internal/common/types"
	"dos/internal/services/storage/transport"
	"log/slog"
	"math/rand/v2"
	"time"
)


type ReportQueue struct {
	ch chan t.ReplicaReport
}

func NewReportQueue(size int) *ReportQueue {
	return &ReportQueue{
		ch: make(chan t.ReplicaReport, size),
	}
}

func (rq *ReportQueue) Enqueue(ctx context.Context, rec t.ReplicaReport) {
	select {
	case <-ctx.Done():
	case rq.ch <- rec:
	}
}

func (rq *ReportQueue) Drain() []t.ReplicaReport {
	n := len(rq.ch)
	reports := make([]t.ReplicaReport, 0, n)
	for range n {
		reports = append(reports, <-rq.ch)
	}
	return reports
}

type Config interface {
	ReportInterval() time.Duration
	QueueCapacity() int
}


type ReportService struct {
	identity *IdentityService
	master *transport.Master

	config Config

	queue *ReportQueue
	wake chan struct{}
}

func NewReportService(
	identity *IdentityService, master *transport.Master, cfg Config, 
) *ReportService {

	return &ReportService {
		identity: identity,
		master: master,
		config: cfg,
		queue: NewReportQueue(cfg.QueueCapacity()),
	}
}

func (rs *ReportService) Enqueue(ctx context.Context, report t.ReplicaReport) {
	rs.queue.Enqueue(ctx, report)
}

func (rs *ReportService) Flush(ctx context.Context) {
	select {
	case rs.wake <- struct{}{}:
	case <-ctx.Done():
	}
}

func (rs *ReportService) SendReport(ctx context.Context) (t.ReportResult, error) {
	nodeID, err := rs.identity.GetID()
	if err != nil {
		return t.ReportResult{}, err
	}

	reports := rs.queue.Drain()
	result, err := rs.master.ReportChunks(ctx, nodeID, reports)
	if err != nil {
		return t.ReportResult{}, err
	}
	return result, nil
}

func (rs *ReportService) RunReportLoop(ctx context.Context) {
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		case <-rs.wake:
		}

		slog.DebugContext(ctx, "exec report chunks")
		res, err := rs.SendReport(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "report chunks failed", "error", err)
		}
		slog.WarnContext(ctx, "rejected chunks", "chunk_ids", res.Rejected)

		timer.Reset(jitter(rs.config.ReportInterval(), 0.2))
	}
}

func jitter(base time.Duration, frac float64) time.Duration {
	delta := float64(base) * frac
	j := (rand.Float64()*2 - 1) * delta
	return time.Duration(float64(base) + j)
}

