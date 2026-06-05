package report

import (
	"context"
	"dos/internal/common/dosctx"
	"dos/internal/common/loop"
	"dos/internal/common/queue"
	t "dos/internal/common/types"
	"dos/internal/services/storage/transport"
	"errors"
	"log/slog"
	"time"
)

type ReportConfig interface {
	ReportInterval() time.Duration
	QueueCapacity() int
}

type IdentityProvider interface {
	GetID() (t.NodeID, error)
}

type ReportProcessor interface {
	ProcessReport(context.Context, t.ReportResult)
}

type NOPReportProcessor struct{}

func (*NOPReportProcessor) ProcessReport(context.Context, t.ReportResult) {}

type ReportDeps struct {
	Identity IdentityProvider
	MasterT  *transport.Master
	Config   ReportConfig
	Metrics  *ReportMetrics
}

type ReportService struct {
	identity IdentityProvider
	masterT  *transport.Master
	metrics  *ReportMetrics

	processor ReportProcessor

	config ReportConfig

	pending []t.StorageNodeReport

	queue  *queue.Queue[t.StorageNodeReport]
	looper *loop.Looper
}

func NewReportService(deps ReportDeps) (*ReportService, error) {
	if deps.Identity == nil {
		return nil, errors.New("missing identity")
	}
	if deps.MasterT == nil {
		return nil, errors.New("missing master transport")
	}
	if deps.Config == nil {
		return nil, errors.New("missing config")
	}
	if deps.Metrics == nil {
		return nil, errors.New("missing metrics")
	}

	config := deps.Config 
	queue := queue.NewQueue[t.StorageNodeReport](config.QueueCapacity())
	looper := loop.NewLooper(config.ReportInterval())
	service := &ReportService{
		identity: deps.Identity,
		masterT:  deps.MasterT,
		config:   deps.Config,
		metrics:  deps.Metrics,

		queue:     queue,
		looper:    looper,
		processor: &NOPReportProcessor{},
	}
	return service, nil
}

func (rs *ReportService) SetReportProcessor(p ReportProcessor) {
	rs.processor = p
}

func (rs *ReportService) Report(ctx context.Context, report t.StorageNodeReport) {
	if err := rs.queue.Enq(ctx, report); err != nil {
		return
	}
	if rs.queue.Full() {
		rs.Flush(ctx)
	}
}

func (rs *ReportService) RunReportIteration(ctx context.Context) {

	if len(rs.pending) > 0 {
		slog.DebugContext(ctx, "send pending reports")
		result, err := rs.SendReports(ctx, rs.pending)
		if err != nil {
			slog.ErrorContext(ctx, "report chunks failed", "error", err)
			return
		}
		rs.ProcessReportResult(ctx, result)
		rs.pending = nil
	}

	rs.pending = rs.queue.Drain()
	if len(rs.pending) == 0 {
		return
	}

	slog.DebugContext(ctx, "send drained reports", "length", len(rs.pending))
	rs.metrics.ReportsQueueBatchSize.Observe(float64(len(rs.pending)))

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
	rs.processor.ProcessReport(ctx, result)
	rs.metrics.ReportsRejectedTotal.Add(float64(len(result.Rejected)))
}

func (rs *ReportService) SendReports(
	ctx context.Context, reports []t.StorageNodeReport,
) (t.ReportResult, error) {

	rs.metrics.ReportsSentTotal.Inc()

	nodeID, err := rs.identity.GetID()
	if err != nil {
		return t.ReportResult{}, err
	}

	result, err := rs.masterT.ReportChunks(ctx, nodeID, reports)
	if err != nil {
		rs.metrics.ReportsFailedTotal.Inc()
		return t.ReportResult{}, err
	}
	rs.metrics.ReportsRecordsTotal.Add(float64(len(reports)))
	return result, nil
}

func (rs *ReportService) RunLoop(ctx context.Context) {
	ctx = dosctx.WithService(ctx, "reporter")
	rs.looper.SkipFirstWait().Run(ctx, rs.RunReportIteration)
}

func (rs *ReportService) Flush(_ context.Context) {
	rs.looper.Flush()
}
