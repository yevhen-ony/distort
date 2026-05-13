package report

import (
	"context"
	"dos/internal/common/dosctx"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
	"dos/internal/services/storage/transport"
	"log/slog"
	"time"
)


type Config interface {
	ReportInterval() time.Duration
	QueueCapacity() int
}

type IdentityProvider interface {
	GetID() (t.NodeID, error)
}

type ReportProcessor interface{
	Process(context.Context, t.ReportResult)
}

type NOPReportProcessor struct {}
func (*NOPReportProcessor) Process(context.Context, t.ReportResult) {}


type ReportService struct {
	identity IdentityProvider 
	master *transport.Master

	config Config

	queue *Queue
	pending []t.StorageNodeReport
	processor ReportProcessor

	wake chan struct{}
}

func NewReportService(
	identity IdentityProvider, master *transport.Master, cfg Config, 
) *ReportService {

	return &ReportService {
		identity: identity,
		master: master,
		config: cfg,
		queue: NewReportQueue(cfg.QueueCapacity()),
		wake: make(chan struct{}, 1),
		processor: &NOPReportProcessor{},
	}
}

func (rs *ReportService) SetReportProcessor(p ReportProcessor) {
	rs.processor = p
}

func (rs *ReportService) Report(ctx context.Context, report t.StorageNodeReport) {
	rs.queue.Enqueue(ctx, report)
	if rs.queue.IsFull() {
		rs.Flush(ctx)
	}
}

func (rs *ReportService) Flush(ctx context.Context) {
	select {
	case rs.wake <- struct{}{}:
	case <-ctx.Done():
	}
}

func (rs *ReportService) RunReportIteration(ctx context.Context) {

	if len(rs.pending) > 0 {
		result, err := rs.SendReports(ctx, rs.pending)
		if err != nil {
			slog.ErrorContext(ctx, "report chunks failed", "error", err)
			return
		}
		rs.ProcessReportResult(ctx, result)
		rs.pending = nil
	}
	
	rs.pending = rs.queue.Drain()
	slog.DebugContext(ctx, "drain report queue", "length", len(rs.pending))
	if len(rs.pending) == 0 {
		return
	}

	result, err := rs.SendReports(ctx, rs.pending)
	if err != nil {
		slog.ErrorContext(ctx, "report chunks failed", "error", err)
		return
	}
	rs.ProcessReportResult(ctx, result)
	rs.pending = nil
}

func (rs *ReportService) ProcessReportResult(ctx context.Context, result t.ReportResult) {
	if len(result.Rejected) != 0 {
		slog.WarnContext(ctx, "rejected chunks", "chunk_ids", result.Rejected)
	}
	rs.processor.Process(ctx, result)
}

func (rs *ReportService) SendReports(
	ctx context.Context, reports []t.StorageNodeReport,
) (t.ReportResult, error) {
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
		rs.RunReportIteration(ctx)

		interval := utils.Jitter(rs.config.ReportInterval(), 0.2) 
		timer.Reset(interval)
	}
}


