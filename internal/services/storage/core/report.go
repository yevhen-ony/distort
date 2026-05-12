package core

import (
	"context"
	"dos/internal/common/dosctx"
	t "dos/internal/common/types"
	"dos/internal/services/storage/transport"
	"log/slog"
	"math/rand/v2"
	"time"
)


type ReportQueue struct {
	ch chan t.StorageNodeReport
}

func NewReportQueue(size int) *ReportQueue {
	return &ReportQueue{
		ch: make(chan t.StorageNodeReport, size),
	}
}

func (rq *ReportQueue) Enqueue(ctx context.Context, rec t.StorageNodeReport) {
	select {
	case <-ctx.Done():
	case rq.ch <- rec:
	}
}

func (rq *ReportQueue) Drain() []t.StorageNodeReport {
	n := len(rq.ch)
	reports := make([]t.StorageNodeReport, 0, n)
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

func (rs *ReportService) Enqueue(ctx context.Context, report t.StorageNodeReport) {
	rs.queue.Enqueue(ctx, report)
}

func (rs *ReportService) Flush(ctx context.Context) {
	select {
	case rs.wake <- struct{}{}:
	case <-ctx.Done():
	}
}

func (rs *ReportService) Report(ctx context.Context) {
	reports := rs.queue.Drain()
	if len(reports) == 0 {
		return
	}

	result, err := rs.SendReport(ctx, reports)
	if err != nil {
		slog.ErrorContext(ctx, "report chunks failed", "error", err)
		return
	}
	if len(result.Rejected) != 0 {
		slog.WarnContext(ctx, "rejected chunks", "chunk_ids", result.Rejected)
	}
}

func (rs *ReportService) SendReport(ctx context.Context, reports []t.StorageNodeReport ) (t.ReportResult, error) {
	nodeID, err := rs.identity.GetID()
	if err != nil {
		return t.ReportResult{}, err
	}

	result, err := rs.master.ReportChunks(ctx, nodeID, reports)
	if err != nil {
		return t.ReportResult{}, err
	}
	return result, nil
}

func (rs *ReportService) RunReportLoop(ctx context.Context) {
	timer := time.NewTimer(0)
	defer timer.Stop()
	
	ctx = dosctx.WithService(ctx, "report")

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		case <-rs.wake:
		}

		slog.DebugContext(ctx, "exec report chunks")
		rs.Report(ctx)

		timer.Reset(jitter(rs.config.ReportInterval(), 0.2))
	}
}

func jitter(base time.Duration, frac float64) time.Duration {
	delta := float64(base) * frac
	j := (rand.Float64()*2 - 1) * delta
	return time.Duration(float64(base) + j)
}

